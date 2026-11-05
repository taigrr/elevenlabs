//go:build integration

package elevenlabs

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/taigrr/elevenlabs/v2/convai"
	"github.com/taigrr/elevenlabs/v2/convai/agents"
	"github.com/taigrr/elevenlabs/v2/dubbing"
	"github.com/taigrr/elevenlabs/v2/realtime"
	"github.com/taigrr/elevenlabs/v2/speechtotext"
	"github.com/taigrr/elevenlabs/v2/texttodialogue"
	"github.com/taigrr/elevenlabs/v2/texttospeech"
	"github.com/taigrr/elevenlabs/v2/translation"
)

// Live tests against api.elevenlabs.io. Run with:
//
//	ELEVENLABS_API_KEY=... go test -tags integration -run Live -v .
//
// They consume a small amount of credits.

const (
	liveVoiceA = "JBFqnCBsd6RMkjVDRZzb"
	liveVoiceB = "21m00Tcm4TlvDq8ikWAM"
)

func liveClient(t *testing.T) *Client {
	t.Helper()
	key := os.Getenv(APIKeyEnvVar)
	if key == "" {
		key = os.Getenv("XI_API_KEY")
	}
	if key == "" {
		t.Skip("ELEVENLABS_API_KEY or XI_API_KEY not set")
	}
	return NewClient(WithAPIKey(key), WithTimeout(60*time.Second))
}

// skipIfForbidden skips when the key lacks a permission for the feature
// under test, so restricted keys can still run the rest of the suite.
func skipIfForbidden(t *testing.T, err error) {
	t.Helper()
	var apiErr *APIError
	if errors.As(err, &apiErr) && (apiErr.StatusCode == 403 || apiErr.Status == "missing_permissions") {
		t.Skipf("api key not permitted: %v", err)
	}
}

func liveCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// livePCM synthesizes a short 16 kHz PCM utterance for the audio-input sockets.
func livePCM(t *testing.T, c *Client, text string) []byte {
	t.Helper()
	rc, err := c.TextToSpeech.Convert(liveCtx(t), liveVoiceA, &texttospeech.ConvertRequest{
		Text: text, ModelID: "eleven_flash_v2_5", OutputFormat: "pcm_16000",
	})
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	defer rc.Close()
	pcm, err := io.ReadAll(rc)
	if err != nil || len(pcm) == 0 {
		t.Fatalf("read pcm: %d bytes, %v", len(pcm), err)
	}
	return pcm
}

func TestLiveREST(t *testing.T) {
	c := liveClient(t)
	ctx := liveCtx(t)
	models, err := c.Models.List(ctx)
	if err != nil || len(models) == 0 {
		t.Fatalf("models: %d, %v", len(models), err)
	}
	user, err := c.User.Get(ctx)
	if err != nil || user.UserID == "" {
		t.Fatalf("user: %+v, %v", user, err)
	}
	catalog, err := c.AudioEffects.GetCatalog(ctx)
	if err != nil {
		t.Fatalf("catalog: %v", err)
	}
	t.Logf("catalog: %d presets, %d environments, %d noises", len(catalog.FilterPresets), len(catalog.Environments), len(catalog.BackgroundNoises))
	voices, err := c.Voices.Search(ctx, nil)
	if err != nil || len(voices.Voices) == 0 {
		t.Fatalf("voices: %v", err)
	}
	if _, err := c.Voices.Get(ctx, "does-not-exist-voice", nil); !errors.Is(err, ErrNotFound) && !errors.Is(err, ErrBadRequest) {
		t.Errorf("missing voice err = %v", err)
	}
	if len(livePCM(t, c, "Hello.")) == 0 {
		t.Fatal("no audio")
	}
}

func TestLiveTextToSpeechStreamInput(t *testing.T) {
	c := liveClient(t)
	ctx := liveCtx(t)
	stream, err := c.TextToSpeech.StreamInput(ctx, liveVoiceA, &texttospeech.StreamInputRequest{
		ModelID: "eleven_flash_v2_5", OutputFormat: "pcm_16000", SyncAlignment: Ptr(true),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	if err := errors.Join(
		stream.Send(ctx, &texttospeech.InitializeConnection{}),
		stream.SendText(ctx, "Hello from the Go "),
		stream.SendText(ctx, "streaming client. "),
		stream.CloseSend(ctx),
	); err != nil {
		t.Fatal(err)
	}
	audio, err := io.ReadAll(stream.AudioReader(ctx))
	if err != nil || len(audio) == 0 {
		t.Fatalf("audio: %d bytes, %v", len(audio), err)
	}
	t.Logf("stream-input: %d bytes", len(audio))
}

func TestLiveTextToSpeechMultiStreamInput(t *testing.T) {
	c := liveClient(t)
	ctx := liveCtx(t)
	stream, err := c.TextToSpeech.MultiStreamInput(ctx, liveVoiceA, &texttospeech.StreamInputRequest{
		ModelID: "eleven_flash_v2_5", OutputFormat: "pcm_16000",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	if err := errors.Join(
		stream.Send(ctx, &texttospeech.InitializeConnectionMulti{ContextID: "ctx1"}),
		stream.SendText(ctx, "ctx1", "Multi context streaming works. "),
		stream.FlushContext(ctx, "ctx1"),
		stream.CloseContext(ctx, "ctx1"),
	); err != nil {
		t.Fatal(err)
	}
	var total int
	for {
		out, err := stream.Recv(ctx)
		if err != nil {
			t.Fatalf("recv after %d bytes: %v", total, err)
		}
		total += len(out.Audio)
		if out.IsFinal {
			if out.ContextID != "ctx1" {
				t.Errorf("final context = %q", out.ContextID)
			}
			break
		}
	}
	if total == 0 {
		t.Fatal("no audio")
	}
	_ = stream.CloseSend(ctx)
	t.Logf("multi-stream-input: %d bytes", total)
}

func TestLiveTextToDialogueStreamInput(t *testing.T) {
	c := liveClient(t)
	ctx := liveCtx(t)
	stream, err := c.TextToDialogue.StreamInput(ctx, &texttodialogue.StreamInputRequest{ModelID: "eleven_v3", OutputFormat: "pcm_16000"})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	if err := errors.Join(
		stream.Send(ctx, &texttodialogue.StreamMessage{Voices: []string{liveVoiceA, liveVoiceB}}),
		stream.SendInputs(ctx,
			texttodialogue.StreamInput{Text: "Did the Go client work?", VoiceID: liveVoiceA},
			texttodialogue.StreamInput{Text: "It did!", VoiceID: liveVoiceB, NewTurn: Ptr(true)}),
		stream.Flush(ctx),
		stream.CloseSend(ctx),
	); err != nil {
		t.Fatal(err)
	}
	audio, err := io.ReadAll(stream.AudioReader(ctx))
	if err != nil || len(audio) == 0 {
		t.Fatalf("audio: %d bytes, %v", len(audio), err)
	}
	t.Logf("dialogue: %d bytes", len(audio))
}

func TestLiveTextToDialogueMultiStreamInput(t *testing.T) {
	c := liveClient(t)
	ctx := liveCtx(t)
	stream, err := c.TextToDialogue.MultiStreamInput(ctx, &texttodialogue.StreamInputRequest{ModelID: "eleven_v3", OutputFormat: "pcm_16000"})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	if err := errors.Join(
		stream.Send(ctx, &texttodialogue.MultiStreamMessage{ContextID: "d1", Voices: []string{liveVoiceA, liveVoiceB}}),
		stream.SendInputs(ctx, "d1", texttodialogue.StreamInput{Text: "Context one speaking.", VoiceID: liveVoiceA}),
		stream.FlushContext(ctx, "d1"),
		stream.CloseContext(ctx, "d1"),
	); err != nil {
		t.Fatal(err)
	}
	var total int
	for {
		out, err := stream.Recv(ctx)
		if err != nil {
			t.Fatalf("recv after %d bytes: %v", total, err)
		}
		total += len(out.Audio)
		if out.IsFinal {
			break
		}
	}
	if total == 0 {
		t.Fatal("no audio")
	}
	_ = stream.CloseSend(ctx)
	t.Logf("dialogue multi: %d bytes", total)
}

func TestLiveSpeechToTextRealtime(t *testing.T) {
	c := liveClient(t)
	pcm := livePCM(t, c, "The quick brown fox jumps over the lazy dog.")
	ctx := liveCtx(t)
	conn, err := c.SpeechToText.Realtime(ctx, &speechtotext.RealtimeRequest{
		ModelID: "scribe_v2_realtime", AudioFormat: speechtotext.RealtimeAudioPCM16000, SampleRate: 16000,
		IncludeTimestamps: Ptr(true),
	})
	if err != nil {
		t.Fatal(err)
	}
	sendErr := make(chan error, 1)
	go func() { sendErr <- conn.SendFrom(ctx, bytes.NewReader(pcm)) }()
	defer func() {
		_ = conn.Close()
		if err := <-sendErr; err != nil {
			t.Logf("send: %v", err)
		}
	}()
	for {
		msg, err := conn.Recv(ctx)
		if err != nil {
			t.Fatalf("recv: %v", err)
		}
		t.Logf("scribe %s: %q", msg.MessageType, msg.Text)
		if msg.MessageType == speechtotext.RealtimeCommittedTranscript || msg.MessageType == speechtotext.RealtimeCommittedTranscriptWithTimestamp {
			if msg.Text == "" {
				t.Error("empty committed transcript")
			}
			return
		}
	}
}

type liveTranslateConn interface {
	SendAudio(context.Context, []byte) error
	EndOfStream(context.Context) error
	Recv(context.Context) (*dubbing.RealtimeMessage, error)
	Close() error
}

func runLiveTranslate(t *testing.T, ctx context.Context, conn liveTranslateConn, pcm []byte) {
	t.Helper()
	sendDone := make(chan struct{})
	defer func() {
		_ = conn.Close()
		<-sendDone
	}()
	go func() {
		defer close(sendDone)
		const chunk = 3200
		for i := 0; i < len(pcm); i += chunk {
			if err := conn.SendAudio(ctx, pcm[i:min(i+chunk, len(pcm))]); err != nil {
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
		silence := make([]byte, chunk)
		for range 20 {
			if err := conn.SendAudio(ctx, silence); err != nil {
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
		_ = conn.EndOfStream(ctx)
	}()
	var gotSession, gotOutput bool
	for {
		msg, err := conn.Recv(ctx)
		if errors.Is(err, io.EOF) {
			break
		}
		var rtErr *realtime.Error
		if errors.As(err, &rtErr) && rtErr.MessageType == "auth_error" && strings.Contains(rtErr.Message, "permission") {
			t.Skipf("api key lacks permission: %v", err)
		}
		if err != nil {
			t.Fatalf("recv: %v", err)
		}
		if msg.MessageType == dubbing.RealtimeAudio {
			t.Logf("%s: %d bytes", msg.MessageType, len(msg.Data))
		} else {
			t.Logf("%s: %q %s", msg.MessageType, msg.Text, msg.Status)
		}
		switch msg.MessageType {
		case dubbing.RealtimeSessionStarted:
			gotSession = true
		case dubbing.RealtimeTranslation, dubbing.RealtimeAudio:
			gotOutput = true
		}
		if gotOutput && msg.MessageType == dubbing.RealtimeAudio {
			break
		}
	}
	if !gotSession || !gotOutput {
		t.Errorf("session=%v output=%v", gotSession, gotOutput)
	}
}

func TestLiveTranslationRealtime(t *testing.T) {
	c := liveClient(t)
	pcm := livePCM(t, c, "Good morning, how are you today?")
	ctx := liveCtx(t)
	conn, err := c.Translation.Realtime(ctx, &translation.RealtimeRequest{
		SourceLanguage: "en", TargetLanguage: "es", VoiceID: liveVoiceA,
		InputFormat: speechtotext.RealtimeAudioPCM16000, OutputFormat: "pcm_16000",
	})
	if err != nil {
		t.Fatal(err)
	}
	runLiveTranslate(t, ctx, conn, pcm)
}

func TestLiveDubbingRealtime(t *testing.T) {
	c := liveClient(t)
	pcm := livePCM(t, c, "Good morning, how are you today?")
	ctx := liveCtx(t)
	conn, err := c.Dubbing.Realtime(ctx, &dubbing.RealtimeRequest{
		SourceLanguage: "en", TargetLanguage: "es",
		InputFormat: speechtotext.RealtimeAudioPCM16000, OutputFormat: speechtotext.RealtimeAudioPCM16000,
	})
	skipIfForbidden(t, err)
	if err != nil {
		t.Fatal(err)
	}
	runLiveTranslate(t, ctx, conn, pcm)
}

func TestLiveConversation(t *testing.T) {
	c := liveClient(t)
	ctx := liveCtx(t)
	agentID := os.Getenv("ELEVENLABS_AGENT_ID")
	if agentID == "" {
		created, err := c.ConvAI.Agents.Create(ctx, &agents.CreateRequest{
			Name: "go-sdk-integration-test",
			ConversationConfig: agents.ConversationalConfig{
				Agent: &agents.AgentConfig{
					FirstMessage: "Hello from the integration test.",
					Prompt:       &agents.AgentPrompt{Prompt: "Reply with one short sentence."},
				},
			},
		})
		skipIfForbidden(t, err)
		if err != nil {
			t.Fatalf("create agent: %v", err)
		}
		agentID = created.AgentID
		t.Cleanup(func() {
			if err := c.ConvAI.Agents.Delete(context.Background(), agentID, nil); err != nil {
				t.Errorf("delete agent: %v", err)
			}
		})
	}
	for _, auth := range []bool{false, true} {
		conv, err := c.ConvAI.Connect(ctx, &convai.ConnectRequest{AgentID: agentID, RequiresAuth: auth})
		if err != nil {
			t.Fatalf("connect (auth=%v): %v", auth, err)
		}
		var gotMeta, gotResponse bool
		for !gotResponse {
			ev, err := conv.Recv(ctx)
			if err != nil {
				t.Fatalf("recv (auth=%v): %v", auth, err)
			}
			switch ev.Type {
			case convai.ConversationEventInitiationMetadata:
				gotMeta = true
				if err := conv.SendUserMessage(ctx, "Say hi."); err != nil {
					t.Fatal(err)
				}
			case convai.ConversationEventAgentResponse:
				t.Logf("agent (auth=%v): %s", auth, ev.AgentResponse.AgentResponse)
				gotResponse = true
			}
		}
		if !gotMeta || conv.ConversationID() == "" {
			t.Errorf("metadata=%v id=%q", gotMeta, conv.ConversationID())
		}
		_ = conv.Close()
	}
}
