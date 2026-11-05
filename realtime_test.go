package elevenlabs

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

// wsHandler accepts a WebSocket and runs script on it. Failures inside the
// script are reported with t.Errorf.
func wsHandler(t *testing.T, script func(ctx context.Context, c *websocket.Conn)) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			t.Errorf("accept: %v", err)
			return
		}
		c.SetReadLimit(realtimeReadLimit)
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		script(ctx, c)
	}
}

func readMsg(t *testing.T, ctx context.Context, c *websocket.Conn) map[string]any {
	t.Helper()
	_, b, err := c.Read(ctx)
	if err != nil {
		t.Errorf("server read: %v", err)
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Errorf("server decode %s: %v", b, err)
	}
	return m
}

func expectMsg(t *testing.T, ctx context.Context, c *websocket.Conn, want string) {
	t.Helper()
	got := readMsg(t, ctx, c)
	var w map[string]any
	if err := json.Unmarshal([]byte(want), &w); err != nil {
		t.Fatalf("bad want %s: %v", want, err)
	}
	if !reflect.DeepEqual(got, w) {
		gb, _ := json.Marshal(got)
		t.Errorf("server got %s, want %s", gb, want)
	}
}

func writeMsg(t *testing.T, ctx context.Context, c *websocket.Conn, msg string) {
	t.Helper()
	if err := c.Write(ctx, websocket.MessageText, []byte(msg)); err != nil {
		t.Errorf("server write: %v", err)
	}
}

func b64(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }

func testCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func assertQuery(t *testing.T, raw string, want url.Values) {
	t.Helper()
	got, err := url.ParseQuery(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("query = %v, want %v", got, want)
	}
}

func TestTextToSpeechStreamInput(t *testing.T) {
	done := make(chan struct{})
	client, ts := newTestServer(t, wsHandler(t, func(ctx context.Context, c *websocket.Conn) {
		defer close(done)
		expectMsg(t, ctx, c, `{"text":" ","voice_settings":{"stability":0.5},"generation_config":{"chunk_length_schedule":[120,160]}}`)
		expectMsg(t, ctx, c, `{"text":"Hello world "}`)
		expectMsg(t, ctx, c, `{"text":" ","flush":true}`)
		expectMsg(t, ctx, c, `{"text":"Bye ","try_trigger_generation":true}`)
		expectMsg(t, ctx, c, `{"text":""}`)
		writeMsg(t, ctx, c, `{"audio":"`+b64("abc")+`","normalizedAlignment":{"chars":["a"],"charStartTimesMs":[0],"charDurationsMs":[5]}}`)
		writeMsg(t, ctx, c, `{"audio":"`+b64("def")+`"}`)
		writeMsg(t, ctx, c, `{"isFinal":true}`)
		_ = c.Close(websocket.StatusNormalClosure, "")
	}))
	ctx := testCtx(t)
	stream, err := client.TextToSpeech.StreamInput(ctx, "voice 1", &TextToSpeechStreamInputRequest{
		ModelID: "eleven_flash_v2_5", OutputFormat: "pcm_16000", SyncAlignment: Ptr(true),
		InactivityTimeout: Ptr(30), Seed: Ptr(uint32(7)), ApplyTextNormalization: TextNormalizationOn,
	}, WithHeader("X-Extra", "1"), WithQueryParam("extra", "x"))
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	req := ts.last(t)
	if req.Path != "/v1/text-to-speech/voice%201/stream-input" {
		t.Errorf("path = %s", req.Path)
	}
	assertQuery(t, req.RawQuery, url.Values{
		"model_id": {"eleven_flash_v2_5"}, "output_format": {"pcm_16000"}, "sync_alignment": {"true"},
		"inactivity_timeout": {"30"}, "seed": {"7"}, "apply_text_normalization": {"on"}, "extra": {"x"},
	})
	if req.Header.Get(headerAPIKey) != "test-key" || req.Header.Get("X-Extra") != "1" {
		t.Errorf("headers = %v", req.Header)
	}
	if err := stream.Send(ctx, &InitializeConnection{
		VoiceSettings:    &RealtimeVoiceSettings{Stability: Ptr(0.5)},
		GenerationConfig: &GenerationConfig{ChunkLengthSchedule: []float64{120, 160}},
	}); err != nil {
		t.Fatal(err)
	}
	for _, step := range []func() error{
		func() error { return stream.SendText(ctx, "Hello world ") },
		func() error { return stream.Flush(ctx) },
		func() error { return stream.Send(ctx, &SendText{Text: "Bye ", TryTriggerGeneration: Ptr(true)}) },
		func() error { return stream.CloseSend(ctx) },
	} {
		if err := step(); err != nil {
			t.Fatal(err)
		}
	}
	first, err := stream.Recv(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if string(first.Audio) != "abc" || first.NormalizedAlignment == nil || first.NormalizedAlignment.Chars[0] != "a" {
		t.Errorf("first = %+v", first)
	}
	rest, err := io.ReadAll(stream.AudioReader(ctx))
	if err != nil {
		t.Fatal(err)
	}
	if string(rest) != "def" {
		t.Errorf("audio = %q", rest)
	}
	if _, err := stream.Recv(ctx); !errors.Is(err, io.EOF) {
		t.Errorf("recv after close = %v, want io.EOF", err)
	}
	<-done
}

func TestTextToSpeechStreamInputTokenSkipsAPIKey(t *testing.T) {
	client, ts := newTestServer(t, wsHandler(t, func(ctx context.Context, c *websocket.Conn) {
		_ = c.Close(websocket.StatusNormalClosure, "")
	}))
	stream, err := client.TextToSpeech.StreamInput(testCtx(t), "v", &TextToSpeechStreamInputRequest{SingleUseToken: "tok"})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	req := ts.last(t)
	if req.Header.Get(headerAPIKey) != "" {
		t.Error("api key header sent with single use token")
	}
	assertQuery(t, req.RawQuery, url.Values{"single_use_token": {"tok"}})
}

func TestTextToSpeechMultiStreamMessages(t *testing.T) {
	tests := []struct {
		name string
		msg  TextToSpeechMultiStreamInputMessage
		want string
	}{
		{"init connection", &InitializeConnectionMulti{ContextID: "a", XIAPIKey: "k"}, `{"text":" ","xi_api_key":"k","context_id":"a"}`},
		{"init context", &InitializeContext{Text: "Hi ", ContextID: "b"}, `{"text":"Hi ","context_id":"b"}`},
		{"send text", &SendTextMulti{Text: "x ", ContextID: "a"}, `{"text":"x ","context_id":"a"}`},
		{"flush", &FlushContext{ContextID: "a"}, `{"context_id":"a","flush":true}`},
		{"close context", &CloseContext{ContextID: "a"}, `{"context_id":"a","close_context":true}`},
		{"close socket", &CloseSocket{}, `{"close_socket":true}`},
		{"keep alive", &KeepContextAlive{ContextID: "a"}, `{"text":"","context_id":"a"}`},
	}
	done := make(chan struct{})
	client, ts := newTestServer(t, wsHandler(t, func(ctx context.Context, c *websocket.Conn) {
		defer close(done)
		for _, tt := range tests {
			expectMsg(t, ctx, c, tt.want)
		}
		expectMsg(t, ctx, c, `{"text":"y ","context_id":"c"}`)
		expectMsg(t, ctx, c, `{"context_id":"c","flush":true}`)
		expectMsg(t, ctx, c, `{"context_id":"c","close_context":true}`)
		expectMsg(t, ctx, c, `{"text":"","context_id":"c"}`)
		expectMsg(t, ctx, c, `{"close_socket":true}`)
		writeMsg(t, ctx, c, `{"audio":"`+b64("zz")+`","contextId":"c"}`)
		writeMsg(t, ctx, c, `{"isFinal":true,"contextId":"c"}`)
		_ = c.Close(websocket.StatusNormalClosure, "")
	}))
	ctx := testCtx(t)
	stream, err := client.TextToSpeech.MultiStreamInput(ctx, "v1", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	if p := ts.last(t).Path; p != "/v1/text-to-speech/v1/multi-stream-input" {
		t.Errorf("path = %s", p)
	}
	for _, tt := range tests {
		if err := stream.Send(ctx, tt.msg); err != nil {
			t.Fatalf("%s: %v", tt.name, err)
		}
	}
	steps := []error{
		stream.SendText(ctx, "c", "y "),
		stream.FlushContext(ctx, "c"),
		stream.CloseContext(ctx, "c"),
		stream.KeepContextAlive(ctx, "c"),
		stream.CloseSend(ctx),
	}
	if err := errors.Join(steps...); err != nil {
		t.Fatal(err)
	}
	out, err := stream.Recv(ctx)
	if err != nil || string(out.Audio) != "zz" || out.ContextID != "c" {
		t.Fatalf("recv = %+v, %v", out, err)
	}
	out, err = stream.Recv(ctx)
	if err != nil || !out.IsFinal || out.ContextID != "c" {
		t.Fatalf("final = %+v, %v", out, err)
	}
	if _, err := stream.Recv(ctx); !errors.Is(err, io.EOF) {
		t.Errorf("recv = %v, want io.EOF", err)
	}
	<-done
}

func TestTextToDialogueStreamInput(t *testing.T) {
	done := make(chan struct{})
	client, ts := newTestServer(t, wsHandler(t, func(ctx context.Context, c *websocket.Conn) {
		defer close(done)
		expectMsg(t, ctx, c, `{"voices":["a","b"],"xi_api_key":"k","voice_settings":{"stability":0.4}}`)
		expectMsg(t, ctx, c, `{"inputs":[{"text":"Hi","voice_id":"a"},{"text":"Yo","voice_id":"b","new_turn":true}]}`)
		expectMsg(t, ctx, c, `{"flush":true}`)
		expectMsg(t, ctx, c, `{"keep_alive":true}`)
		expectMsg(t, ctx, c, `{"close_socket":true}`)
		writeMsg(t, ctx, c, `{"audio":"`+b64("A")+`","alignment":{"chars":["H"],"char_start_times_ms":[0],"char_durations_ms":[1]}}`)
		writeMsg(t, ctx, c, `{"is_final_audio_for_turn":true}`)
		writeMsg(t, ctx, c, `{"message":"bad voice","error":"invalid_voice","code":400,"param":"voice_id"}`)
		_ = c.Close(websocket.StatusNormalClosure, "")
	}))
	ctx := testCtx(t)
	stream, err := client.TextToDialogue.StreamInput(ctx, &TextToDialogueStreamInputRequest{ModelID: "eleven_v3", EnableLogging: Ptr(false)})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	req := ts.last(t)
	if req.Path != "/v1/text-to-dialogue/stream-input" {
		t.Errorf("path = %s", req.Path)
	}
	assertQuery(t, req.RawQuery, url.Values{"model_id": {"eleven_v3"}, "enable_logging": {"false"}})
	err = errors.Join(
		stream.Send(ctx, &DialogueStreamMessage{Voices: []string{"a", "b"}, XIAPIKey: "k", VoiceSettings: &DialogueStreamVoiceSettings{Stability: Ptr(0.4)}}),
		stream.SendInputs(ctx, DialogueStreamInput{Text: "Hi", VoiceID: "a"}, DialogueStreamInput{Text: "Yo", VoiceID: "b", NewTurn: Ptr(true)}),
		stream.Flush(ctx),
		stream.KeepAlive(ctx),
		stream.CloseSend(ctx),
	)
	if err != nil {
		t.Fatal(err)
	}
	out, err := stream.Recv(ctx)
	if err != nil || string(out.Audio) != "A" || out.Alignment.Chars[0] != "H" {
		t.Fatalf("recv = %+v, %v", out, err)
	}
	out, err = stream.Recv(ctx)
	if err != nil || !out.IsFinalAudioForTurn {
		t.Fatalf("turn = %+v, %v", out, err)
	}
	_, err = stream.Recv(ctx)
	var rtErr *RealtimeError
	if !errors.As(err, &rtErr) || !errors.Is(err, ErrRealtime) || !errors.Is(err, ErrBadRequest) {
		t.Fatalf("err = %v", err)
	}
	if rtErr.Param != "voice_id" || rtErr.Code != 400 || rtErr.Message != "bad voice" {
		t.Errorf("realtime error = %+v", rtErr)
	}
	if _, err := stream.Recv(ctx); !errors.Is(err, io.EOF) {
		t.Errorf("recv = %v, want io.EOF", err)
	}
	<-done
}

func TestTextToDialogueMultiStreamInput(t *testing.T) {
	done := make(chan struct{})
	client, ts := newTestServer(t, wsHandler(t, func(ctx context.Context, c *websocket.Conn) {
		defer close(done)
		expectMsg(t, ctx, c, `{"context_id":"x","inputs":[{"text":"Hi","voice_id":"a"}]}`)
		expectMsg(t, ctx, c, `{"context_id":"x","flush":true}`)
		expectMsg(t, ctx, c, `{"context_id":"x","keep_alive":true}`)
		expectMsg(t, ctx, c, `{"context_id":"x","close_context":true}`)
		expectMsg(t, ctx, c, `{"close_socket":true}`)
		writeMsg(t, ctx, c, `{"is_final":true,"context_id":"x"}`)
		_ = c.Close(websocket.StatusNormalClosure, "")
	}))
	ctx := testCtx(t)
	stream, err := client.TextToDialogue.MultiStreamInput(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	if p := ts.last(t).Path; p != "/v1/text-to-dialogue/multi-stream-input" {
		t.Errorf("path = %s", p)
	}
	err = errors.Join(
		stream.SendInputs(ctx, "x", DialogueStreamInput{Text: "Hi", VoiceID: "a"}),
		stream.FlushContext(ctx, "x"),
		stream.KeepContextAlive(ctx, "x"),
		stream.CloseContext(ctx, "x"),
		stream.CloseSend(ctx),
	)
	if err != nil {
		t.Fatal(err)
	}
	out, err := stream.Recv(ctx)
	if err != nil || !out.IsFinal || out.ContextID != "x" {
		t.Fatalf("recv = %+v, %v", out, err)
	}
	if _, err := stream.Recv(ctx); !errors.Is(err, io.EOF) {
		t.Errorf("recv = %v, want io.EOF", err)
	}
	<-done
}

func TestSpeechToTextRealtimeValidation(t *testing.T) {
	client := NewClient(WithAPIKey("k"), WithBaseURL("http://127.0.0.1:1"))
	tests := []struct {
		name string
		req  *SpeechToTextRealtimeRequest
	}{
		{"nil", nil},
		{"missing model", &SpeechToTextRealtimeRequest{}},
		{"vad silence", &SpeechToTextRealtimeRequest{ModelID: "m", VADSilenceThresholdSecs: Ptr(0.1)}},
		{"vad threshold", &SpeechToTextRealtimeRequest{ModelID: "m", VADThreshold: Ptr(0.95)}},
		{"min speech", &SpeechToTextRealtimeRequest{ModelID: "m", MinSpeechDurationMs: Ptr(10)}},
		{"min silence", &SpeechToTextRealtimeRequest{ModelID: "m", MinSilenceDurationMs: Ptr(3000)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := client.SpeechToText.Realtime(testCtx(t), tt.req); !errors.Is(err, ErrInvalidParameter) {
				t.Errorf("err = %v, want ErrInvalidParameter", err)
			}
		})
	}
}

func TestSpeechToTextRealtime(t *testing.T) {
	done := make(chan struct{})
	client, ts := newTestServer(t, wsHandler(t, func(ctx context.Context, c *websocket.Conn) {
		defer close(done)
		expectMsg(t, ctx, c, `{"message_type":"input_audio_chunk","audio_base_64":"`+b64("pcm")+`","commit":false,"sample_rate":24000}`)
		expectMsg(t, ctx, c, `{"message_type":"input_audio_chunk","audio_base_64":"`+b64("x")+`","commit":false,"sample_rate":8000,"previous_text":"ctx"}`)
		expectMsg(t, ctx, c, `{"message_type":"input_audio_chunk","audio_base_64":"`+b64("w")+`","commit":false,"sample_rate":24000}`)
		expectMsg(t, ctx, c, `{"message_type":"input_audio_chunk","audio_base_64":"","commit":true,"sample_rate":24000}`)
		writeMsg(t, ctx, c, `{"message_type":"session_started","session_id":"s1","config":{"sample_rate":24000,"model_id":"scribe_v2_realtime","commit_strategy":"vad"}}`)
		writeMsg(t, ctx, c, `{"message_type":"committed_transcript_with_timestamps","text":"hi","language_code":"en","words":[{"text":"hi","start":0,"end":0.4,"type":"word","characters":["h","i"]}]}`)
		writeMsg(t, ctx, c, `{"message_type":"final_transcript_with_timestamps","text":"hi","words":[{"text":"hi","characters":[{"text":"h","start":0,"end":0.2}],"channel_index":null}]}`)
		writeMsg(t, ctx, c, `{"message_type":"committed_transcript_entities","text":"hi bob","entities":[{"text":"bob","entity_type":"person_name","start_char":3,"end_char":6}]}`)
		writeMsg(t, ctx, c, `{"message_type":"auth_error","error":"invalid key"}`)
		_ = c.Close(websocket.StatusPolicyViolation, "bye")
	}))
	ctx := testCtx(t)
	conn, err := client.SpeechToText.Realtime(ctx, &SpeechToTextRealtimeRequest{
		ModelID: "scribe_v2_realtime", CommitStrategy: CommitStrategyVAD, AudioFormat: RealtimeAudioPCM24000,
		VADThreshold: Ptr(0.5), SecondaryLanguages: []string{"de", "fr"}, Keyterms: []string{"Go"},
		EntityDetection: []string{"pii"}, EntityRedactionMode: EntityRedactionEntityType, NoVerbatim: Ptr(true),
		SampleRate: 24000,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	req := ts.last(t)
	if req.Path != "/v1/speech-to-text/realtime" || req.Header.Get(headerAPIKey) != "test-key" {
		t.Errorf("path = %s headers = %v", req.Path, req.Header)
	}
	assertQuery(t, req.RawQuery, url.Values{
		"model_id": {"scribe_v2_realtime"}, "commit_strategy": {"vad"}, "audio_format": {"pcm_24000"},
		"vad_threshold": {"0.5"}, "secondary_languages": {"de", "fr"}, "keyterms": {"Go"},
		"entity_detection": {"pii"}, "entity_redaction_mode": {"entity_type"}, "no_verbatim": {"true"},
	})
	if err := conn.SendAudio(ctx, []byte("pcm")); err != nil {
		t.Fatal(err)
	}
	if err := conn.Send(ctx, &RealtimeAudioChunk{Audio: []byte("x"), SampleRate: 8000, PreviousText: "ctx"}); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.AudioWriter(ctx).Write([]byte("w")); err != nil {
		t.Fatal(err)
	}
	if err := conn.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	msg, err := conn.Recv(ctx)
	if err != nil || msg.MessageType != RealtimeSessionStarted || msg.SessionID != "s1" || msg.Config.SampleRate != 24000 {
		t.Fatalf("session = %+v, %v", msg, err)
	}
	msg, err = conn.Recv(ctx)
	if err != nil || len(msg.Words) != 1 || msg.Words[0].Characters[1].Text != "i" || msg.LanguageCode != "en" {
		t.Fatalf("committed = %+v, %v", msg, err)
	}
	msg, err = conn.Recv(ctx)
	if err != nil || *msg.Words[0].Characters[0].End != 0.2 || msg.Words[0].ChannelIndex != nil {
		t.Fatalf("final = %+v, %v", msg, err)
	}
	msg, err = conn.Recv(ctx)
	if err != nil || msg.Entities[0].EntityType != "person_name" {
		t.Fatalf("entities = %+v, %v", msg, err)
	}
	msg, err = conn.Recv(ctx)
	if msg == nil || msg.MessageType != RealtimeErrorAuth || !errors.Is(err, ErrUnauthorized) || !errors.Is(err, ErrRealtime) {
		t.Fatalf("auth error = %+v, %v", msg, err)
	}
	_, err = conn.Recv(ctx)
	var closeErr *CloseError
	if !errors.As(err, &closeErr) || !errors.Is(err, ErrSocketClosed) || closeErr.Code != int(websocket.StatusPolicyViolation) || closeErr.Reason != "bye" {
		t.Fatalf("close err = %v", err)
	}
	<-done
}

func TestSpeechToTextRealtimeSendFrom(t *testing.T) {
	audio := bytes.Repeat([]byte{1}, realtimeChunkSize+10)
	done := make(chan struct{})
	client, ts := newTestServer(t, wsHandler(t, func(ctx context.Context, c *websocket.Conn) {
		defer close(done)
		var got []byte
		for {
			m := readMsg(t, ctx, c)
			if m == nil {
				return
			}
			chunk, _ := base64.StdEncoding.DecodeString(m["audio_base_64"].(string))
			got = append(got, chunk...)
			if m["commit"] == true {
				break
			}
		}
		if !bytes.Equal(got, audio) {
			t.Errorf("server got %d bytes, want %d", len(got), len(audio))
		}
		_ = c.Close(websocket.StatusNormalClosure, "")
	}))
	ctx := testCtx(t)
	conn, err := client.SpeechToText.Realtime(ctx, &SpeechToTextRealtimeRequest{ModelID: "m", Token: "tok"})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if ts.last(t).Header.Get(headerAPIKey) != "" {
		t.Error("api key sent with token")
	}
	if err := conn.SendFrom(ctx, bytes.NewReader(audio)); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Recv(ctx); !errors.Is(err, io.EOF) {
		t.Errorf("recv = %v, want io.EOF", err)
	}
	<-done
}

func TestRealtimeTranslateSockets(t *testing.T) {
	script := func(t *testing.T, done chan struct{}) http.HandlerFunc {
		return wsHandler(t, func(ctx context.Context, c *websocket.Conn) {
			defer close(done)
			expectMsg(t, ctx, c, `{"message_type":"input_audio_chunk","audio_base_64":"`+b64("w")+`"}`)
			expectMsg(t, ctx, c, `{"message_type":"end_of_stream"}`)
			_ = c.Close(websocket.StatusNormalClosure, "")
		})
	}
	type conn interface {
		EndOfStream(context.Context) error
		AudioWriter(context.Context) io.Writer
		Recv(context.Context) (*RealtimeTranslateMessage, error)
		Close() error
	}
	tests := []struct {
		name   string
		open   func(ctx context.Context, c *Client) (conn, error)
		path   string
		query  url.Values
		apiKey string
	}{
		{
			name: "dubbing",
			open: func(ctx context.Context, c *Client) (conn, error) {
				return c.Dubbing.Realtime(ctx, &DubbingRealtimeRequest{SourceLanguage: "en", TargetLanguage: "es", InputFormat: RealtimeAudioPCM16000})
			},
			path:   "/v1/dubbing/realtime",
			query:  url.Values{"source_language": {"en"}, "target_language": {"es"}, "input_format": {"pcm_16000"}},
			apiKey: "test-key",
		},
		{
			name: "translation",
			open: func(ctx context.Context, c *Client) (conn, error) {
				return c.Translation.Realtime(ctx, &TranslationRealtimeRequest{
					TargetLanguage: "es", TTSModelID: TranslationTTSModelV4Turbo, Eagerness: TranslationEagernessHigh,
					EnableZRM: Ptr(true), Authorization: "bearer",
				})
			},
			path: "/v1/translation/realtime",
			query: url.Values{"target_language": {"es"}, "tts_model_id": {"eleven_v4_turbo"}, "eagerness": {"high"},
				"enable_zrm": {"true"}, "authorization": {"bearer"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			done := make(chan struct{})
			client, ts := newTestServer(t, script(t, done))
			ctx := testCtx(t)
			c, err := tt.open(ctx, client)
			if err != nil {
				t.Fatal(err)
			}
			defer c.Close()
			req := ts.last(t)
			if req.Path != tt.path || req.Header.Get(headerAPIKey) != tt.apiKey {
				t.Errorf("path = %s api key = %q", req.Path, req.Header.Get(headerAPIKey))
			}
			assertQuery(t, req.RawQuery, tt.query)
			if _, err := c.AudioWriter(ctx).Write([]byte("w")); err != nil {
				t.Fatal(err)
			}
			if err := c.EndOfStream(ctx); err != nil {
				t.Fatal(err)
			}
			if _, err := c.Recv(ctx); !errors.Is(err, io.EOF) {
				t.Errorf("recv = %v, want io.EOF", err)
			}
			<-done
		})
	}
}

func TestRealtimeTranslateMessages(t *testing.T) {
	client, _ := newTestServer(t, wsHandler(t, func(ctx context.Context, c *websocket.Conn) {
		expectMsg(t, ctx, c, `{"message_type":"input_audio_chunk","audio_base_64":"`+b64("in")+`"}`)
		expectMsg(t, ctx, c, `{"message_type":"input_audio_chunk","audio_base_64":"`+b64("w")+`"}`)
		expectMsg(t, ctx, c, `{"message_type":"end_of_stream"}`)
		writeMsg(t, ctx, c, `{"message_type":"session_started","session_id":"s","client_session_id":"c"}`)
		writeMsg(t, ctx, c, `{"message_type":"translation","text":"hola","segment_id":"g","start_s":1.5,"end_s":2,"speaker_id":"sp"}`)
		writeMsg(t, ctx, c, `{"message_type":"audio","data":"`+b64("out1")+`","sample_rate":16000}`)
		writeMsg(t, ctx, c, `{"message_type":"status","status":"stopped"}`)
		writeMsg(t, ctx, c, `{"message_type":"audio","data":"`+b64("out2")+`","sample_rate":16000}`)
		writeMsg(t, ctx, c, `{"message_type":"rate_limited","error":"slow down"}`)
		_ = c.Close(websocket.StatusNormalClosure, "")
	}))
	ctx := testCtx(t)
	c, err := client.Dubbing.Realtime(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if err := errors.Join(c.SendAudio(ctx, []byte("in")), func() error { _, err := c.AudioWriter(ctx).Write([]byte("w")); return err }(), c.EndOfStream(ctx)); err != nil {
		t.Fatal(err)
	}
	msg, err := c.Recv(ctx)
	if err != nil || msg.SessionID != "s" || msg.ClientSessionID != "c" {
		t.Fatalf("started = %+v, %v", msg, err)
	}
	msg, err = c.Recv(ctx)
	if err != nil || msg.Text != "hola" || msg.SegmentID != "g" || msg.StartS != 1.5 || msg.SpeakerID != "sp" {
		t.Fatalf("translation = %+v, %v", msg, err)
	}
	audio, err := io.ReadAll(c.AudioReader(ctx))
	if string(audio) != "out1out2" || !errors.Is(err, ErrTooManyRequests) {
		t.Fatalf("audio = %q, err = %v", audio, err)
	}
}

func TestConversationSession(t *testing.T) {
	done := make(chan struct{})
	client, ts := newTestServer(t, wsHandler(t, func(ctx context.Context, c *websocket.Conn) {
		defer close(done)
		expectMsg(t, ctx, c, `{"type":"conversation_initiation_client_data","dynamic_variables":{"name":"Tai"},"user_id":"u1"}`)
		writeMsg(t, ctx, c, `{"type":"conversation_initiation_metadata","conversation_initiation_metadata_event":{"conversation_id":"conv1","agent_output_audio_format":"pcm_16000"}}`)
		writeMsg(t, ctx, c, `{"type":"ping","ping_event":{"event_id":9,"ping_ms":40}}`)
		expectMsg(t, ctx, c, `{"type":"pong","event_id":9}`)
		writeMsg(t, ctx, c, `{"type":"interruption","interruption_event":{"event_id":5}}`)
		writeMsg(t, ctx, c, `{"type":"audio","audio_event":{"audio_base_64":"`+b64("old")+`","event_id":3}}`)
		writeMsg(t, ctx, c, `{"type":"audio","audio_event":{"audio_base_64":"`+b64("new")+`","event_id":6}}`)
		writeMsg(t, ctx, c, `{"type":"client_tool_call","client_tool_call":{"tool_name":"lookup","tool_call_id":"t1","parameters":{"q":"x"},"event_id":7,"expects_response":true}}`)
		expectMsg(t, ctx, c, `{"type":"client_tool_result","tool_call_id":"t1","result":"found x via t1","is_error":false}`)
		writeMsg(t, ctx, c, `{"type":"client_tool_call","client_tool_call":{"tool_name":"missing","tool_call_id":"t2","parameters":{},"event_id":8,"expects_response":true}}`)
		got := readMsg(t, ctx, c)
		if got["is_error"] != true || got["tool_call_id"] != "t2" || !strings.Contains(got["result"].(string), "not registered") {
			t.Errorf("missing tool result = %v", got)
		}
		expectMsg(t, ctx, c, `{"type":"user_message","text":"hello"}`)
		expectMsg(t, ctx, c, `{"type":"user_activity"}`)
		expectMsg(t, ctx, c, `{"type":"contextual_update","text":"ctx"}`)
		expectMsg(t, ctx, c, `{"user_audio_chunk":"`+b64("mic")+`"}`)
		expectMsg(t, ctx, c, `{"user_audio_chunk":"`+b64("w")+`"}`)
		expectMsg(t, ctx, c, `{"type":"feedback","event_id":6,"score":"like"}`)
		expectMsg(t, ctx, c, `{"type":"feedback","event_id":6,"score":null}`)
		expectMsg(t, ctx, c, `{"type":"multimodal_message","text":{"type":"user_message","text":"see"},"files":[{"type":"file_input","file_id":"f1"}]}`)
		writeMsg(t, ctx, c, `{"type":"agent_response","agent_response_event":{"agent_response":"hi","event_id":10}}`)
		writeMsg(t, ctx, c, `{"type":"client_error","error_event":{"code":1008,"error_name":"policy","message":"nope"}}`)
		writeMsg(t, ctx, c, `{"type":"queue_status","queue_status_event":{"status":"queued"}}`)
		_ = c.Close(websocket.StatusNormalClosure, "")
	}))
	tools := &ClientTools{}
	if err := tools.Register("lookup", func(_ context.Context, p map[string]any) (any, error) {
		return "found " + p["q"].(string) + " via " + p["tool_call_id"].(string), nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := tools.Register("lookup", func(context.Context, map[string]any) (any, error) { return nil, nil }); !errors.Is(err, ErrToolAlreadyRegistered) {
		t.Errorf("duplicate register err = %v", err)
	}
	ctx := testCtx(t)
	conv, err := client.ConversationalAI.Connect(ctx, &ConversationRequest{
		AgentID:        "agent1",
		BranchID:       "b1",
		InitiationData: &ConversationInitiationData{DynamicVariables: map[string]any{"name": "Tai"}, UserID: "u1"},
		ClientTools:    tools,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conv.Close()
	req := ts.last(t)
	if req.Path != "/v1/convai/conversation" || req.Header.Get(headerAPIKey) != "test-key" {
		t.Errorf("path = %s", req.Path)
	}
	assertQuery(t, req.RawQuery, url.Values{"agent_id": {"agent1"}, "branch_id": {"b1"}})

	ev, err := conv.Recv(ctx)
	if err != nil || ev.Type != ConversationEventInitiationMetadata || conv.ConversationID() != "conv1" {
		t.Fatalf("metadata = %+v, %v", ev, err)
	}
	ev, err = conv.Recv(ctx)
	if err != nil || ev.Ping.PingMs != 40 {
		t.Fatalf("ping = %+v, %v", ev, err)
	}
	ev, err = conv.Recv(ctx)
	if err != nil || ev.Type != ConversationEventInterruption {
		t.Fatalf("interruption = %+v, %v", ev, err)
	}
	ev, err = conv.Recv(ctx)
	if err != nil || string(ev.Audio.Audio) != "new" {
		t.Fatalf("audio = %+v, %v", ev, err)
	}
	for range 2 {
		if ev, err = conv.Recv(ctx); err != nil || ev.Type != ConversationEventClientToolCall {
			t.Fatalf("tool call = %+v, %v", ev, err)
		}
	}
	time.Sleep(50 * time.Millisecond)
	err = errors.Join(
		conv.SendUserMessage(ctx, "hello"),
		conv.SendUserActivity(ctx),
		conv.SendContextualUpdate(ctx, "ctx"),
		conv.SendUserAudio(ctx, []byte("mic")),
		func() error { _, err := conv.AudioWriter(ctx).Write([]byte("w")); return err }(),
		conv.SendFeedback(ctx, 6, FeedbackLike),
		conv.SendFeedback(ctx, 6, ""),
		conv.SendMultimodalMessage(ctx, "see", MultimodalFile{FileID: "f1"}),
	)
	if err != nil {
		t.Fatal(err)
	}
	ev, err = conv.Recv(ctx)
	if err != nil || ev.AgentResponse.AgentResponse != "hi" {
		t.Fatalf("agent response = %+v, %v", ev, err)
	}
	ev, err = conv.Recv(ctx)
	var rtErr *RealtimeError
	if ev == nil || !errors.As(err, &rtErr) || rtErr.Code != 1008 || rtErr.MessageType != "policy" {
		t.Fatalf("client error = %+v, %v", ev, err)
	}
	ev, err = conv.Recv(ctx)
	if err != nil || ev.Type != ConversationEventQueueStatus || !bytes.Contains(ev.Raw, []byte("queued")) {
		t.Fatalf("queue status = %+v, %v", ev, err)
	}
	if _, err := conv.Recv(ctx); !errors.Is(err, io.EOF) {
		t.Errorf("end = %v", err)
	}
	<-done
}

func TestConversationSignedURL(t *testing.T) {
	var serverURL string
	client, ts := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/convai/conversation/get-signed-url":
			respond(http.StatusOK, contentTypeJSON, `{"signed_url":"`+strings.Replace(serverURL, "http", "ws", 1)+`/signed?conversation_signature=sig"}`)(w, r)
		case "/signed":
			wsHandler(t, func(ctx context.Context, c *websocket.Conn) {
				expectMsg(t, ctx, c, `{"type":"conversation_initiation_client_data"}`)
				_ = c.Close(websocket.StatusNormalClosure, "")
			})(w, r)
		default:
			http.NotFound(w, r)
		}
	})
	serverURL = ts.URL
	ctx := testCtx(t)
	conv, err := client.ConversationalAI.Connect(ctx, &ConversationRequest{AgentID: "a", RequiresAuth: true})
	if err != nil {
		t.Fatal(err)
	}
	defer conv.Close()
	reqs := ts.all()
	if len(reqs) != 2 || reqs[0].RawQuery != "agent_id=a" || reqs[1].Path != "/signed" ||
		reqs[1].RawQuery != "conversation_signature=sig" || reqs[1].Header.Get(headerAPIKey) != "" {
		t.Errorf("requests = %+v", reqs)
	}
	if _, err := conv.Recv(ctx); !errors.Is(err, io.EOF) {
		t.Errorf("recv = %v", err)
	}
	if _, err := client.ConversationalAI.Connect(ctx, &ConversationRequest{}); !errors.Is(err, ErrInvalidParameter) {
		t.Errorf("missing agent id err = %v", err)
	}
}

func TestRealtimeHandshakeErrors(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   error
	}{
		{"unauthorized", http.StatusUnauthorized, ErrUnauthorized},
		{"forbidden", http.StatusForbidden, ErrForbidden},
		{"not found", http.StatusNotFound, ErrNotFound},
		{"rate limited", http.StatusTooManyRequests, ErrTooManyRequests},
		{"server", http.StatusInternalServerError, ErrServer},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, ts := newTestServer(t, respond(tt.status, contentTypeJSON, `{"detail":{"status":"bad","message":"denied"}}`))
			_, err := client.TextToSpeech.StreamInput(testCtx(t), "v", nil)
			var apiErr *APIError
			if !errors.Is(err, tt.want) || !errors.As(err, &apiErr) || apiErr.StatusCode != tt.status || apiErr.Message != "denied" {
				t.Fatalf("err = %v", err)
			}
			if n := len(ts.all()); n != 1 {
				t.Errorf("handshake attempts = %d, want 1", n)
			}
		})
	}
}

func TestRealtimeContextCancelAndClose(t *testing.T) {
	release := make(chan struct{})
	client, _ := newTestServer(t, wsHandler(t, func(ctx context.Context, c *websocket.Conn) {
		select {
		case <-release:
		case <-ctx.Done():
		}
	}))
	stream, err := client.TextToSpeech.StreamInput(testCtx(t), "v", nil)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := stream.Recv(ctx); err == nil {
		t.Fatal("recv with cancelled context succeeded")
	}
	close(release)
	_ = stream.Close()
	if err := stream.SendText(testCtx(t), "x"); !errors.Is(err, ErrSocketClosed) {
		t.Errorf("send after close = %v, want ErrSocketClosed", err)
	}
}

func TestRealtimeErrorIs(t *testing.T) {
	tests := []struct {
		err  *RealtimeError
		want error
	}{
		{&RealtimeError{MessageType: "auth_error"}, ErrUnauthorized},
		{&RealtimeError{MessageType: "quota_exceeded"}, ErrTooManyRequests},
		{&RealtimeError{MessageType: "commit_throttled"}, ErrTooManyRequests},
		{&RealtimeError{MessageType: "input_error"}, ErrBadRequest},
		{&RealtimeError{MessageType: "unaccepted_terms"}, ErrForbidden},
		{&RealtimeError{MessageType: "x", Code: 404}, ErrNotFound},
		{&RealtimeError{MessageType: "transcriber_error"}, ErrRealtime},
	}
	for _, tt := range tests {
		if !errors.Is(tt.err, tt.want) || !errors.Is(tt.err, ErrRealtime) {
			t.Errorf("%v: Is(%v) = false", tt.err, tt.want)
		}
	}
	if errors.Is(&RealtimeError{MessageType: "transcriber_error"}, ErrUnauthorized) {
		t.Error("transcriber_error matched ErrUnauthorized")
	}
	if got := (&RealtimeError{MessageType: "e", Code: 3, Message: "m", Param: "p"}).Error(); got != "elevenlabs: realtime e (3): m [param p]" {
		t.Errorf("Error() = %q", got)
	}
	if got := (&CloseError{Code: 1011, Reason: "boom"}).Error(); got != "elevenlabs: websocket closed with status 1011: boom" {
		t.Errorf("CloseError.Error() = %q", got)
	}
}

func TestClientToolsRegistry(t *testing.T) {
	var tools ClientTools
	ctx := context.Background()
	if _, err := tools.Handle(ctx, "x", nil); !errors.Is(err, ErrToolNotRegistered) {
		t.Errorf("handle unknown = %v", err)
	}
	if err := tools.Register("x", nil); !errors.Is(err, ErrInvalidParameter) {
		t.Errorf("nil handler = %v", err)
	}
	_ = tools.Register("b", func(context.Context, map[string]any) (any, error) { return 1, nil })
	_ = tools.Register("a", func(context.Context, map[string]any) (any, error) { return nil, errors.New("boom") })
	if got := tools.Names(); !reflect.DeepEqual(got, []string{"a", "b"}) {
		t.Errorf("names = %v", got)
	}
	if v, err := tools.Handle(ctx, "b", nil); v != 1 || err != nil {
		t.Errorf("handle = %v, %v", v, err)
	}
	if !tools.Unregister("a") || tools.Unregister("a") || tools.IsRegistered("a") {
		t.Error("unregister misbehaved")
	}
	tools.Clear()
	if len(tools.Names()) != 0 {
		t.Error("clear left tools")
	}
}

func TestSpeechToTextRealtimeStreamURLWithoutFFmpeg(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	conn := &SpeechToTextRealtime{}
	if err := conn.StreamURL(testCtx(t), "https://example.com/a.mp3"); !errors.Is(err, ErrFFmpegNotFound) {
		t.Errorf("err = %v, want ErrFFmpegNotFound", err)
	}
}
