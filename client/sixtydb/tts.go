package sixtydb

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DefaultVoiceID is the 60db docs example voice; override per-request as needed.
const DefaultVoiceID = "fbb75ed2-975a-40c7-9e06-38e30524a9a1"

// SynthesisOptions mirrors the ElevenLabs SynthesisOptions surface so
// callers swapping providers don't have to relearn the parameter shape.
// Internally values are clamped/converted to 60db's 0-100 integer scale.
type SynthesisOptions struct {
	Stability       float64 // 0.0-1.0
	SimilarityBoost float64 // 0.0-1.0
	Speed           float64 // 0.5-2.0 (default 1.0)
	Enhance         bool    // default true
}

func (o SynthesisOptions) toPayload() (int, int, float64, bool) {
	stab := int(o.Stability * 100)
	if stab < 0 {
		stab = 0
	} else if stab > 100 {
		stab = 100
	}
	sim := int(o.SimilarityBoost * 100)
	if sim < 0 {
		sim = 0
	} else if sim > 100 {
		sim = 100
	}
	speed := o.Speed
	if speed == 0 {
		speed = 1.0
	}
	enhance := o.Enhance
	if !enhance && o.SimilarityBoost == 0 && o.Stability == 0 && o.Speed == 0 {
		// Zero-value SynthesisOptions{} → safe defaults instead of muting.
		enhance = true
	}
	return stab, sim, speed, enhance
}

type ttsRequest struct {
	Text         string  `json:"text"`
	VoiceID      string  `json:"voice_id"`
	Enhance      bool    `json:"enhance"`
	Speed        float64 `json:"speed"`
	Stability    int     `json:"stability"`
	Similarity   int     `json:"similarity"`
	OutputFormat string  `json:"output_format"`
}

type ttsResponse struct {
	Success         bool    `json:"success"`
	AudioBase64     string  `json:"audio_base64"`
	SampleRate      int     `json:"sample_rate"`
	DurationSeconds float64 `json:"duration_seconds"`
	Encoding        string  `json:"encoding"`
	OutputFormat    string  `json:"output_format"`
	Message         string  `json:"message"`
}

// TTS performs a one-shot synthesis via POST /tts-synthesize and returns
// the decoded audio bytes (mp3 by default).
//
// Endpoint reference: https://docs.60db.ai/api-reference/tts/text-to-speech
func (c Client) TTS(ctx context.Context, text, voiceID string, options SynthesisOptions) ([]byte, error) {
	if voiceID == "" {
		voiceID = DefaultVoiceID
	}
	stab, sim, speed, enhance := options.toPayload()
	body, _ := json.Marshal(ttsRequest{
		Text:         text,
		VoiceID:      voiceID,
		Enhance:      enhance,
		Speed:        speed,
		Stability:    stab,
		Similarity:   sim,
		OutputFormat: "mp3",
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/tts-synthesize", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs/client/sixtydb")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case http.StatusUnauthorized:
		return nil, ErrUnauthorized
	case http.StatusOK:
		// Success path falls through.
	default:
		preview, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return nil, fmt.Errorf("60db /tts-synthesize %d: %s", res.StatusCode, string(preview))
	}
	var data ttsResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}
	if !data.Success || data.AudioBase64 == "" {
		return nil, fmt.Errorf("60db /tts-synthesize empty: %s", data.Message)
	}
	return base64.StdEncoding.DecodeString(data.AudioBase64)
}

// TTSWriter is a convenience wrapper that writes the synthesized bytes
// directly to w — matches the parent package's TTSWriter signature
// (minus modelID, which 60db's /tts-synthesize does not expose).
func (c Client) TTSWriter(ctx context.Context, w io.Writer, text, voiceID string, options SynthesisOptions) error {
	audio, err := c.TTS(ctx, text, voiceID, options)
	if err != nil {
		return err
	}
	_, err = w.Write(audio)
	return err
}

// TTSStream pipes mp3 chunks from POST /tts-stream into w as they arrive,
// matching the ElevenLabs TTSStream contract — both let callers feed an
// mp3 decoder (e.g. gopxl/beep) with low time-to-first-audio.
//
// The upstream emits newline-delimited JSON like
//
//	{"type":"chunk","audioContent":"<base64-mp3>"}
//	...
//	{"type":"complete"}
//
// Reference: https://docs.60db.ai/api-reference/tts/text-to-speech-stream
func (c Client) TTSStream(ctx context.Context, w io.Writer, text, voiceID string, options SynthesisOptions) error {
	if voiceID == "" {
		voiceID = DefaultVoiceID
	}
	stab, sim, speed, enhance := options.toPayload()
	body, _ := json.Marshal(ttsRequest{
		Text:         text,
		VoiceID:      voiceID,
		Enhance:      enhance,
		Speed:        speed,
		Stability:    stab,
		Similarity:   sim,
		OutputFormat: "mp3",
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/tts-stream", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("User-Agent", "github.com/taigrr/elevenlabs/client/sixtydb")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	switch res.StatusCode {
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusOK:
		// fall through
	default:
		preview, _ := io.ReadAll(io.LimitReader(res.Body, 1024))
		return fmt.Errorf("60db /tts-stream %d: %s", res.StatusCode, string(preview))
	}

	// Read NDJSON line-by-line, decode each `audioContent` base64 chunk
	// to w. We use Scanner with a large buffer to handle big chunks
	// (audio_base64 lines can be tens of KB).
	scanner := bufio.NewScanner(res.Body)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var msg struct {
			Type         string `json:"type"`
			AudioContent string `json:"audioContent"`
		}
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}
		switch msg.Type {
		case "error":
			return fmt.Errorf("60db /tts-stream emitted error: %s", line)
		case "complete":
			return nil
		}
		if msg.AudioContent == "" {
			continue
		}
		chunk, err := base64.StdEncoding.DecodeString(msg.AudioContent)
		if err != nil {
			return err
		}
		if _, err := w.Write(chunk); err != nil {
			return err
		}
	}
	return scanner.Err()
}
