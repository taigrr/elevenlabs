//go:build integration

// Package e2e is a live end-to-end harness that exercises the public
// packages of the SDK against api.elevenlabs.io and writes every produced
// artifact to an output directory.
package e2e

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/taigrr/elevenlabs/v2"
	"github.com/taigrr/elevenlabs/v2/history"
	"github.com/taigrr/elevenlabs/v2/realtime"
	"github.com/taigrr/elevenlabs/v2/soundeffects"
	"github.com/taigrr/elevenlabs/v2/speechtotext"
	"github.com/taigrr/elevenlabs/v2/texttospeech"
	"github.com/taigrr/elevenlabs/v2/texttovoice"
	"github.com/taigrr/elevenlabs/v2/types"
	"github.com/taigrr/elevenlabs/v2/voices"
)

const (
	// VoiceID is a premade voice available on every account.
	VoiceID = "JBFqnCBsd6RMkjVDRZzb"
	// TTSModel is the cheapest low-latency TTS model.
	TTSModel = "eleven_flash_v2_5"
	// DefaultOutDir is where artifacts are written by default.
	DefaultOutDir = "/Volumes/Store1/tmp/el-e2e"
	// ResourcePrefix prefixes every resource created by the harness.
	ResourcePrefix = "go-e2e-"

	sampleRate       = 16000
	bytesPerSecond   = sampleRate * 2
	chunkBytes       = bytesPerSecond / 10
	chunkInterval    = 100 * time.Millisecond
	gapSilence       = 300 * time.Millisecond
	voiceDescription = "A warm, calm middle-aged narrator with a gentle British accent, speaking clearly at a relaxed pace."
	ttsText          = "The quick brown fox jumps over the lazy dog near the river bank."
)

// Status is the outcome of an area.
type Status string

// Area outcomes.
const (
	Pass    Status = "PASS"
	Fail    Status = "FAIL"
	Skipped Status = "SKIPPED"
)

// Result is the outcome of one tested area.
type Result struct {
	Area   string
	Status Status
	Detail string
}

// ErrSkip marks an area as skipped; wrap it with the reason.
var ErrSkip = errors.New("skipped")

// Config configures a harness run.
type Config struct {
	Client *elevenlabs.Client
	OutDir string
	FFmpeg string
	Logf   func(format string, args ...any)
}

// Area is a named step of the harness.
type Area struct {
	Name string
	Run  func(ctx context.Context, h *Harness) (string, error)
}

// Harness holds state shared between areas.
type Harness struct {
	Config
	Stamp string

	mu       sync.Mutex
	cleanups []func()

	convertMP3 []byte
	// voicesWrite is set once a voices_write operation succeeded, proving
	// created voices can be deleted again.
	voicesWrite      bool
	generatedVoiceID string
}

// New returns a harness for cfg.
func New(cfg Config) (*Harness, error) {
	if cfg.OutDir == "" {
		cfg.OutDir = DefaultOutDir
	}
	if cfg.FFmpeg == "" {
		cfg.FFmpeg = "/opt/homebrew/bin/ffmpeg"
		if _, err := os.Stat(cfg.FFmpeg); err != nil {
			path, lookErr := exec.LookPath("ffmpeg")
			if lookErr != nil {
				return nil, fmt.Errorf("ffmpeg not found: %w", lookErr)
			}
			cfg.FFmpeg = path
		}
	}
	if cfg.Logf == nil {
		cfg.Logf = func(string, ...any) {}
	}
	if err := os.MkdirAll(cfg.OutDir, 0o755); err != nil {
		return nil, err
	}
	return &Harness{Config: cfg, Stamp: time.Now().Format("20060102-150405")}, nil
}

// Areas lists every area in execution order.
func Areas() []Area {
	return []Area{
		{"models.List", areaModels},
		{"user.Get+Subscription", areaUser},
		{"voices.List+Get", areaVoices},
		{"tts.Convert", areaTTSConvert},
		{"tts.Stream", areaTTSStream},
		{"stt.Convert", areaSTT},
		{"history.List+Get+Audio", areaHistory},
		{"soundeffects.Convert", areaSoundEffects},
		{"voices.IVC.Create", areaIVC},
		{"texttovoice.Design", areaVoiceDesign},
		{"texttovoice.Create+Delete", areaVoiceCreate},
		{"ws.tts.StreamInput", areaWSStreamInput},
		{"convai.Conversation", areaConversation},
	}
}

// Defer registers fn to run in Cleanup.
func (h *Harness) Defer(fn func()) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cleanups = append(h.cleanups, fn)
}

// Cleanup runs every registered cleanup in reverse order.
func (h *Harness) Cleanup() {
	h.mu.Lock()
	fns := h.cleanups
	h.cleanups = nil
	h.mu.Unlock()
	for i := len(fns) - 1; i >= 0; i-- {
		fns[i]()
	}
}

// RunArea runs one area, classifying plan and permission errors as skipped.
func (h *Harness) RunArea(ctx context.Context, area Area) Result {
	detail, err := area.Run(ctx, h)
	switch {
	case err == nil:
		return Result{area.Name, Pass, detail}
	case errors.Is(err, ErrSkip):
		return Result{area.Name, Skipped, err.Error()}
	case IsPlanLimit(err):
		return Result{area.Name, Skipped, "plan/permission limit: " + err.Error()}
	default:
		return Result{area.Name, Fail, err.Error()}
	}
}

// IsPlanLimit reports whether err comes from a plan, permission or quota limit.
func IsPlanLimit(err error) bool {
	var apiErr *elevenlabs.APIError
	if errors.As(err, &apiErr) {
		if apiErr.StatusCode == 401 || apiErr.StatusCode == 402 || apiErr.StatusCode == 403 {
			return true
		}
		switch apiErr.Status {
		case "missing_permissions", "quota_exceeded", "payment_required", "subscription_required", "voice_limit_reached", "can_not_use_instant_voice_cloning":
			return true
		}
		msg := strings.ToLower(apiErr.Message)
		return strings.Contains(msg, "subscription") || strings.Contains(msg, "upgrade") || strings.Contains(msg, "quota")
	}
	var rtErr *realtime.Error
	if errors.As(err, &rtErr) {
		msg := strings.ToLower(rtErr.Message)
		return rtErr.MessageType == "auth_error" && strings.Contains(msg, "permission") || strings.Contains(msg, "quota")
	}
	return false
}

func (h *Harness) path(name string) string { return filepath.Join(h.OutDir, name) }

func (h *Harness) write(name string, data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("%s: no data", name)
	}
	return os.WriteFile(h.path(name), data, 0o644)
}

// pcmToMP3 encodes 16 kHz mono s16le PCM to an MP3 file.
func (h *Harness) pcmToMP3(ctx context.Context, pcm []byte, name string) error {
	if len(pcm) == 0 {
		return fmt.Errorf("%s: no audio", name)
	}
	cmd := exec.CommandContext(ctx, h.FFmpeg, "-hide_banner", "-loglevel", "error", "-y",
		"-f", "s16le", "-ar", fmt.Sprint(sampleRate), "-ac", "1", "-i", "pipe:0",
		"-codec:a", "libmp3lame", "-b:a", "64k", h.path(name))
	cmd.Stdin = bytes.NewReader(pcm)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg %s: %w: %s", name, err, out)
	}
	return nil
}

func readAll(rc io.ReadCloser, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

func (h *Harness) tts(ctx context.Context, text string, format types.OutputFormat) ([]byte, error) {
	return readAll(h.Client.TextToSpeech.Convert(ctx, VoiceID, &texttospeech.ConvertRequest{
		Text: text, ModelID: TTSModel, OutputFormat: format,
	}))
}

func areaModels(ctx context.Context, h *Harness) (string, error) {
	list, err := h.Client.Models.List(ctx)
	if err != nil {
		return "", err
	}
	for _, m := range list {
		if m.ModelID == TTSModel {
			return fmt.Sprintf("%d models, %s present", len(list), TTSModel), nil
		}
	}
	return "", fmt.Errorf("%d models, %s missing", len(list), TTSModel)
}

func areaUser(ctx context.Context, h *Harness) (string, error) {
	u, err := h.Client.User.Get(ctx)
	if err != nil {
		return "", err
	}
	if u.UserID == "" {
		return "", errors.New("empty user id")
	}
	sub, err := h.Client.User.Subscription.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("subscription: %w", err)
	}
	return fmt.Sprintf("tier=%s chars=%d/%d", sub.Tier, sub.CharacterCount, sub.CharacterLimit), nil
}

func areaVoices(ctx context.Context, h *Harness) (string, error) {
	all, err := h.Client.Voices.GetAll(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("get all: %w", err)
	}
	search, err := h.Client.Voices.Search(ctx, &voices.SearchRequest{PageSize: elevenlabs.Ptr(5)})
	if err != nil {
		return "", fmt.Errorf("search: %w", err)
	}
	v, err := h.Client.Voices.Get(ctx, VoiceID, nil)
	if err != nil {
		return "", fmt.Errorf("get: %w", err)
	}
	if v.VoiceID != VoiceID || v.Name == nil {
		return "", fmt.Errorf("get returned %q %q", v.VoiceID, nameOf(v))
	}
	return fmt.Sprintf("getall=%d search=%d get=%s", len(all.Voices), len(search.Voices), *v.Name), nil
}

func areaTTSConvert(ctx context.Context, h *Harness) (string, error) {
	mp3, err := h.tts(ctx, ttsText, types.OutputFormatMP3_44100_128)
	if err != nil {
		return "", err
	}
	if !isMP3(mp3) {
		return "", fmt.Errorf("not mp3: % x", mp3[:min(8, len(mp3))])
	}
	h.convertMP3 = mp3
	return fmt.Sprintf("%d bytes", len(mp3)), h.write("tts_convert.mp3", mp3)
}

func areaTTSStream(ctx context.Context, h *Harness) (string, error) {
	rc, err := h.Client.TextToSpeech.Stream(ctx, VoiceID, &texttospeech.ConvertRequest{
		Text: "Streaming speech arrives in small chunks.", ModelID: TTSModel, OutputFormat: types.OutputFormatMP3_44100_128,
	})
	if err != nil {
		return "", err
	}
	defer rc.Close()
	var buf bytes.Buffer
	chunks := 0
	p := make([]byte, 4096)
	for {
		n, err := rc.Read(p)
		if n > 0 {
			chunks++
			buf.Write(p[:n])
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", err
		}
	}
	if !isMP3(buf.Bytes()) {
		return "", errors.New("stream is not mp3")
	}
	return fmt.Sprintf("%d bytes in %d reads", buf.Len(), chunks), h.write("tts_stream.mp3", buf.Bytes())
}

func isMP3(b []byte) bool {
	return len(b) > 3 && (string(b[:3]) == "ID3" || b[0] == 0xFF && b[1]&0xE0 == 0xE0)
}

var nonWord = regexp.MustCompile(`[^a-z0-9 ]+`)

func words(s string) []string {
	return strings.Fields(nonWord.ReplaceAllString(strings.ToLower(s), " "))
}

// overlap returns the fraction of want's words present in got.
func overlap(want, got string) float64 {
	have := map[string]bool{}
	for _, w := range words(got) {
		have[w] = true
	}
	ws := words(want)
	if len(ws) == 0 {
		return 0
	}
	hit := 0
	for _, w := range ws {
		if have[w] {
			hit++
		}
	}
	return float64(hit) / float64(len(ws))
}

func areaSTT(ctx context.Context, h *Harness) (string, error) {
	if h.convertMP3 == nil {
		return "", fmt.Errorf("%w: tts.Convert produced no audio", ErrSkip)
	}
	resp, err := h.Client.SpeechToText.Convert(ctx, &speechtotext.ConvertRequest{
		ModelID:      speechtotext.ModelScribeV1,
		File:         &types.File{Reader: bytes.NewReader(h.convertMP3), Name: "tts_convert.mp3", ContentType: "audio/mpeg"},
		LanguageCode: "en",
	})
	if err != nil {
		return "", err
	}
	if resp.Chunk == nil {
		return "", errors.New("no single-channel transcript")
	}
	score := overlap(ttsText, resp.Chunk.Text)
	if score < 0.7 {
		return "", fmt.Errorf("transcript %q matches %.0f%%", resp.Chunk.Text, score*100)
	}
	return fmt.Sprintf("%q (%.0f%% match, %d words)", resp.Chunk.Text, score*100, len(resp.Chunk.Words)), nil
}

func areaHistory(ctx context.Context, h *Harness) (string, error) {
	text := fmt.Sprintf("History check %s.", strings.ReplaceAll(h.Stamp, "-", " "))
	if _, err := h.tts(ctx, text, types.OutputFormatMP3_44100_128); err != nil {
		return "", fmt.Errorf("tts: %w", err)
	}
	var item *history.Item
	for attempt := 0; attempt < 10 && item == nil; attempt++ {
		page, err := h.Client.History.List(ctx, &history.ListRequest{PageSize: elevenlabs.Ptr(10), VoiceID: VoiceID})
		if err != nil {
			return "", fmt.Errorf("list: %w", err)
		}
		for i := range page.History {
			if it := page.History[i]; it.Text != nil && *it.Text == text {
				item = &it
				break
			}
		}
		if item == nil {
			time.Sleep(time.Second)
		}
	}
	if item == nil {
		return "", errors.New("created item not found in history")
	}
	id := item.HistoryItemID
	cleanupNote := "deleted in cleanup"
	if _, err := h.Client.History.Delete(ctx, ResourcePrefix+"probe-missing"); IsPlanLimit(err) {
		cleanupNote = "not deletable, key lacks speech_history_write"
	} else {
		h.Defer(func() {
			if _, err := h.Client.History.Delete(context.Background(), id); err != nil {
				h.Logf("cleanup history %s: %v", id, err)
			}
		})
	}
	got, err := h.Client.History.Get(ctx, id)
	if err != nil {
		return "", fmt.Errorf("get: %w", err)
	}
	if got.HistoryItemID != id {
		return "", fmt.Errorf("get returned %q", got.HistoryItemID)
	}
	audio, err := readAll(h.Client.History.GetAudio(ctx, id))
	if err != nil {
		return "", fmt.Errorf("audio: %w", err)
	}
	if !isMP3(audio) {
		return "", errors.New("history audio is not mp3")
	}
	return fmt.Sprintf("item %s, %d audio bytes (%s)", id, len(audio), cleanupNote), nil
}

func areaSoundEffects(ctx context.Context, h *Harness) (string, error) {
	mp3, err := readAll(h.Client.TextToSoundEffects.Convert(ctx, &soundeffects.ConvertRequest{
		Text: "a single short doorbell ding", DurationSeconds: elevenlabs.Ptr(1.0),
		OutputFormat: types.OutputFormatMP3_44100_128,
	}))
	if err != nil {
		return "", err
	}
	if !isMP3(mp3) {
		return "", errors.New("not mp3")
	}
	return fmt.Sprintf("%d bytes", len(mp3)), h.write("sound_effect.mp3", mp3)
}

func (h *Harness) deleteVoice(id string) {
	h.Defer(func() {
		if _, err := h.Client.Voices.Delete(context.Background(), id); err != nil {
			h.Logf("cleanup voice %s: %v", id, err)
		}
	})
}

func areaVoiceDesign(ctx context.Context, h *Harness) (string, error) {
	previews, err := h.Client.TextToVoice.Design(ctx, &texttovoice.DesignRequest{
		VoiceDescription: voiceDescription,
		ModelID:          texttovoice.ModelMultilingualTTVV2,
		AutoGenerateText: elevenlabs.Ptr(true),
		OutputFormat:     types.OutputFormatMP3_44100_128,
	})
	if err != nil {
		return "", fmt.Errorf("design: %w", err)
	}
	if len(previews.Previews) == 0 {
		return "", errors.New("no previews")
	}
	preview := previews.Previews[0]
	audio, err := base64.StdEncoding.DecodeString(preview.AudioBase64)
	if err != nil {
		return "", fmt.Errorf("preview audio: %w", err)
	}
	if err := h.write("voice_design_preview.mp3", audio); err != nil {
		return "", err
	}
	h.generatedVoiceID = preview.GeneratedVoiceID
	return fmt.Sprintf("%d previews, %.1fs preview saved", len(previews.Previews), preview.DurationSecs), nil
}

func areaVoiceCreate(ctx context.Context, h *Harness) (string, error) {
	if h.generatedVoiceID == "" {
		return "", fmt.Errorf("%w: no voice design preview", ErrSkip)
	}
	if !h.voicesWrite {
		return "", fmt.Errorf("%w: key lacks voices_write, a created voice could not be deleted again", ErrSkip)
	}
	voice, err := h.Client.TextToVoice.Create(ctx, &texttovoice.CreateRequest{
		VoiceName:        ResourcePrefix + "design-" + h.Stamp,
		VoiceDescription: voiceDescription,
		GeneratedVoiceID: h.generatedVoiceID,
	})
	if err != nil {
		return "", fmt.Errorf("create: %w", err)
	}
	if voice.VoiceID == "" {
		return "", errors.New("create returned empty voice id")
	}
	h.deleteVoice(voice.VoiceID)
	return fmt.Sprintf("created %s (deleted in cleanup)", voice.VoiceID), nil
}

func areaIVC(ctx context.Context, h *Harness) (string, error) {
	if h.convertMP3 == nil {
		return "", fmt.Errorf("%w: tts.Convert produced no audio", ErrSkip)
	}
	resp, err := h.Client.Voices.IVC.Create(ctx, &voices.IVCCreateRequest{
		Name:        ResourcePrefix + "ivc-" + h.Stamp,
		Description: "temporary e2e clone",
		Files:       []types.File{{Reader: bytes.NewReader(h.convertMP3), Name: "sample.mp3", ContentType: "audio/mpeg"}},
		Labels:      map[string]string{"purpose": "e2e"},
	})
	if err != nil {
		return "", err
	}
	if resp.VoiceID == "" {
		return "", errors.New("empty voice id")
	}
	h.voicesWrite = true
	h.deleteVoice(resp.VoiceID)
	v, err := h.Client.Voices.Get(ctx, resp.VoiceID, nil)
	if err != nil {
		return "", fmt.Errorf("get clone: %w", err)
	}
	return fmt.Sprintf("created %s %q (deleted in cleanup)", resp.VoiceID, nameOf(v)), nil
}

func areaWSStreamInput(ctx context.Context, h *Harness) (string, error) {
	stream, err := h.Client.TextToSpeech.StreamInput(ctx, VoiceID, &texttospeech.StreamInputRequest{
		ModelID: TTSModel, OutputFormat: types.OutputFormatMP3_44100_128,
	})
	if err != nil {
		return "", err
	}
	defer stream.Close()
	chunks := []string{"Hello there, ", "this audio was ", "streamed over a ", "WebSocket from Go. "}
	if err := stream.Send(ctx, &texttospeech.InitializeConnection{}); err != nil {
		return "", err
	}
	for _, c := range chunks {
		if err := stream.SendText(ctx, c); err != nil {
			return "", err
		}
	}
	if err := stream.CloseSend(ctx); err != nil {
		return "", err
	}
	audio, err := io.ReadAll(stream.AudioReader(ctx))
	if err != nil {
		return "", err
	}
	if !isMP3(audio) {
		return "", fmt.Errorf("not mp3 (%d bytes)", len(audio))
	}
	return fmt.Sprintf("%d text chunks -> %d bytes", len(chunks), len(audio)), h.write("ws_tts_stream_input.mp3", audio)
}

func silence(d time.Duration) []byte {
	return make([]byte, int(d.Seconds()*bytesPerSecond)&^1)
}

// WriteResults writes results.md to the output directory.
func (h *Harness) WriteResults(results []Result) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# ElevenLabs Go SDK live E2E (%s)\n\n| Area | Status | Detail |\n| --- | --- | --- |\n", h.Stamp)
	for _, r := range results {
		fmt.Fprintf(&b, "| %s | %s | %s |\n", r.Area, r.Status, strings.ReplaceAll(r.Detail, "|", "\\|"))
	}
	return os.WriteFile(h.path("results.md"), []byte(b.String()), 0o644)
}

// RunAll runs every area, cleans up created resources and writes results.md.
func (h *Harness) RunAll(ctx context.Context) []Result {
	defer h.Cleanup()
	var results []Result
	for _, area := range Areas() {
		actx, cancel := context.WithTimeout(ctx, 3*time.Minute)
		r := h.RunArea(actx, area)
		cancel()
		h.Logf("%-28s %-7s %s", r.Area, r.Status, r.Detail)
		results = append(results, r)
	}
	if err := h.WriteResults(results); err != nil {
		h.Logf("write results: %v", err)
	}
	return results
}

func nameOf(v *voices.Voice) string {
	if v.Name == nil {
		return ""
	}
	return *v.Name
}
