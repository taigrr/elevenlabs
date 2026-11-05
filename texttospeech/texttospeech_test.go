package texttospeech

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/coder/websocket"
	"github.com/taigrr/elevenlabs/v2/apierror"
	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/realtime"
	"github.com/taigrr/elevenlabs/v2/types"
)

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, []testutil.Endpoint[*Service]{
		{
			Name: "textToSpeech.convert",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.ReadAll(c.Convert(ctx, "v1", fullTTSRequest()))
			},
			Method:   http.MethodPost,
			Path:     "/v1/text-to-speech/v1",
			Query:    "enable_logging=false&optimize_streaming_latency=2&output_format=pcm_16000",
			Body:     fullTTSBody,
			Accept:   transport.AcceptAny,
			RespType: "audio/mpeg",
			RespBody: "audio",
			Check:    testutil.CheckString("audio"),
		},
		{
			Name: "textToSpeech.stream",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.ReadAll(c.Stream(ctx, "v1", &ConvertRequest{Text: "hi", ModelID: "eleven_flash_v2_5"}))
			},
			Method:   http.MethodPost,
			Path:     "/v1/text-to-speech/v1/stream",
			Body:     `{"text":"hi","model_id":"eleven_flash_v2_5"}`,
			Accept:   transport.AcceptAny,
			RespType: "audio/mpeg",
			RespBody: "chunked-audio",
			Check:    testutil.CheckString("chunked-audio"),
		},
		{
			Name: "textToSpeech.convertWithTimestamps",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.ConvertWithTimestamps(ctx, "v1", &ConvertRequest{Text: "hi", OutputFormat: types.OutputFormatMP3_44100_128})
			},
			Method:   http.MethodPost,
			Path:     "/v1/text-to-speech/v1/with-timestamps",
			Query:    "output_format=mp3_44100_128",
			Body:     `{"text":"hi"}`,
			RespBody: `{"audio_base64":"AAA=","alignment":{"characters":["h","i"],"character_start_times_seconds":[0,0.1],"character_end_times_seconds":[0.1,0.2]}}`,
			Check: func(t *testing.T, r any) {
				a := r.(*AudioWithTimestamps)
				if a.AudioBase64 != "AAA=" || len(a.Alignment.Characters) != 2 || a.NormalizedAlignment != nil {
					t.Errorf("timestamps = %+v", a)
				}
			},
		},
		{
			Name: "textToSpeech.streamWithTimestamps",
			Call: func(ctx context.Context, c *Service) (any, error) {
				stream, err := c.StreamWithTimestamps(ctx, "v1", &ConvertRequest{Text: "hi"})
				if err != nil {
					return nil, err
				}
				var chunks []string
				for chunk, err := range stream.All() {
					if err != nil {
						return nil, err
					}
					chunks = append(chunks, chunk.AudioBase64)
				}
				return strings.Join(chunks, "|"), nil
			},
			Method:   http.MethodPost,
			Path:     "/v1/text-to-speech/v1/stream/with-timestamps",
			Body:     `{"text":"hi"}`,
			RespBody: "{\"audio_base64\":\"AA==\"}\n{\"audio_base64\":\"AQ==\"}\n",
			Check:    testutil.CheckString("AA==|AQ=="),
		},
	})
}

func TestStreamInput(t *testing.T) {
	done := make(chan struct{})
	client, ts := newTestServer(t, testutil.WSHandler(t, func(ctx context.Context, c *websocket.Conn) {
		defer close(done)
		testutil.ExpectMsg(t, ctx, c, `{"text":" ","voice_settings":{"stability":0.5},"generation_config":{"chunk_length_schedule":[120,160]}}`)
		testutil.ExpectMsg(t, ctx, c, `{"text":"Hello world "}`)
		testutil.ExpectMsg(t, ctx, c, `{"text":" ","flush":true}`)
		testutil.ExpectMsg(t, ctx, c, `{"text":"Bye ","try_trigger_generation":true}`)
		testutil.ExpectMsg(t, ctx, c, `{"text":""}`)
		testutil.WriteMsg(t, ctx, c, `{"audio":"`+testutil.B64("abc")+`","normalizedAlignment":{"chars":["a"],"charStartTimesMs":[0],"charDurationsMs":[5]}}`)
		testutil.WriteMsg(t, ctx, c, `{"audio":"`+testutil.B64("def")+`"}`)
		testutil.WriteMsg(t, ctx, c, `{"isFinal":true}`)
		_ = c.Close(websocket.StatusNormalClosure, "")
	}))
	ctx := testutil.Ctx(t)
	stream, err := client.StreamInput(ctx, "voice 1", &StreamInputRequest{
		ModelID: "eleven_flash_v2_5", OutputFormat: "pcm_16000", SyncAlignment: types.Ptr(true),
		InactivityTimeout: types.Ptr(30), Seed: types.Ptr(uint32(7)), ApplyTextNormalization: types.TextNormalizationOn,
	}, option.WithHeader("X-Extra", "1"), option.WithQueryParam("extra", "x"))
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	req := ts.Last(t)
	if req.Path != "/v1/text-to-speech/voice%201/stream-input" {
		t.Errorf("path = %s", req.Path)
	}
	testutil.AssertQuery(t, req.RawQuery, url.Values{
		"model_id": {"eleven_flash_v2_5"}, "output_format": {"pcm_16000"}, "sync_alignment": {"true"},
		"inactivity_timeout": {"30"}, "seed": {"7"}, "apply_text_normalization": {"on"}, "extra": {"x"},
	})
	if req.Header.Get(transport.HeaderAPIKey) != "test-key" || req.Header.Get("X-Extra") != "1" {
		t.Errorf("headers = %v", req.Header)
	}
	if err := stream.Send(ctx, &InitializeConnection{
		VoiceSettings:    &RealtimeVoiceSettings{Stability: types.Ptr(0.5)},
		GenerationConfig: &GenerationConfig{ChunkLengthSchedule: []float64{120, 160}},
	}); err != nil {
		t.Fatal(err)
	}
	for _, step := range []func() error{
		func() error { return stream.SendText(ctx, "Hello world ") },
		func() error { return stream.Flush(ctx) },
		func() error { return stream.Send(ctx, &SendText{Text: "Bye ", TryTriggerGeneration: types.Ptr(true)}) },
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

func TestStreamInputTokenSkipsAPIKey(t *testing.T) {
	client, ts := newTestServer(t, testutil.WSHandler(t, func(ctx context.Context, c *websocket.Conn) {
		_ = c.Close(websocket.StatusNormalClosure, "")
	}))
	stream, err := client.StreamInput(testutil.Ctx(t), "v", &StreamInputRequest{SingleUseToken: "tok"})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	req := ts.Last(t)
	if req.Header.Get(transport.HeaderAPIKey) != "" {
		t.Error("api key header sent with single use token")
	}
	testutil.AssertQuery(t, req.RawQuery, url.Values{"single_use_token": {"tok"}})
}

func TestMultiStreamMessages(t *testing.T) {
	tests := []struct {
		name string
		msg  MultiStreamInputMessage
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
	client, ts := newTestServer(t, testutil.WSHandler(t, func(ctx context.Context, c *websocket.Conn) {
		defer close(done)
		for _, tt := range tests {
			testutil.ExpectMsg(t, ctx, c, tt.want)
		}
		testutil.ExpectMsg(t, ctx, c, `{"text":"y ","context_id":"c"}`)
		testutil.ExpectMsg(t, ctx, c, `{"context_id":"c","flush":true}`)
		testutil.ExpectMsg(t, ctx, c, `{"context_id":"c","close_context":true}`)
		testutil.ExpectMsg(t, ctx, c, `{"text":"","context_id":"c"}`)
		testutil.ExpectMsg(t, ctx, c, `{"close_socket":true}`)
		testutil.WriteMsg(t, ctx, c, `{"audio":"`+testutil.B64("zz")+`","contextId":"c"}`)
		testutil.WriteMsg(t, ctx, c, `{"isFinal":true,"contextId":"c"}`)
		_ = c.Close(websocket.StatusNormalClosure, "")
	}))
	ctx := testutil.Ctx(t)
	stream, err := client.MultiStreamInput(ctx, "v1", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	if p := ts.Last(t).Path; p != "/v1/text-to-speech/v1/multi-stream-input" {
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

func TestRealtimeHandshakeErrors(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   error
	}{
		{"unauthorized", http.StatusUnauthorized, apierror.ErrUnauthorized},
		{"forbidden", http.StatusForbidden, apierror.ErrForbidden},
		{"not found", http.StatusNotFound, apierror.ErrNotFound},
		{"rate limited", http.StatusTooManyRequests, apierror.ErrTooManyRequests},
		{"server", http.StatusInternalServerError, apierror.ErrServer},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, ts := newTestServer(t, testutil.Respond(tt.status, transport.ContentTypeJSON, `{"detail":{"status":"bad","message":"denied"}}`))
			_, err := client.StreamInput(testutil.Ctx(t), "v", nil)
			var apiErr *apierror.APIError
			if !errors.Is(err, tt.want) || !errors.As(err, &apiErr) || apiErr.StatusCode != tt.status || apiErr.Message != "denied" {
				t.Fatalf("err = %v", err)
			}
			if n := len(ts.All()); n != 1 {
				t.Errorf("handshake attempts = %d, want 1", n)
			}
		})
	}
}

func TestRealtimeContextCancelAndClose(t *testing.T) {
	release := make(chan struct{})
	client, _ := newTestServer(t, testutil.WSHandler(t, func(ctx context.Context, c *websocket.Conn) {
		select {
		case <-release:
		case <-ctx.Done():
		}
	}))
	stream, err := client.StreamInput(testutil.Ctx(t), "v", nil)
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
	if err := stream.SendText(testutil.Ctx(t), "x"); !errors.Is(err, realtime.ErrSocketClosed) {
		t.Errorf("send after close = %v, want ErrSocketClosed", err)
	}
}

const fullTTSBody = `{
	"text":"hello","model_id":"eleven_multilingual_v2","language_code":"en",
	"voice_settings":{"stability":0.5,"use_speaker_boost":false},
	"pronunciation_dictionary_locators":[{"pronunciation_dictionary_id":"pd","version_id":"v"}],
	"seed":7,"previous_text":"a","next_text":"b","previous_request_ids":["r1"],"next_request_ids":["r2"],
	"use_pvc_as_ivc":false,"apply_text_normalization":"off","apply_language_text_normalization":true,
	"audio_effects":{"filter_preset_id":"radio","pan":-0.5}
}`

func fullTTSRequest() *ConvertRequest {
	return &ConvertRequest{
		EnableLogging:                   types.Ptr(false),
		OptimizeStreamingLatency:        types.Ptr(2),
		OutputFormat:                    types.OutputFormatPCM16000,
		Text:                            "hello",
		ModelID:                         "eleven_multilingual_v2",
		LanguageCode:                    "en",
		VoiceSettings:                   &types.VoiceSettings{Stability: types.Ptr(0.5), UseSpeakerBoost: types.Ptr(false)},
		PronunciationDictionaryLocators: []types.PronunciationDictionaryLocator{{PronunciationDictionaryID: "pd", VersionID: "v"}},
		Seed:                            types.Ptr(int64(7)),
		PreviousText:                    "a",
		NextText:                        "b",
		PreviousRequestIDs:              []string{"r1"},
		NextRequestIDs:                  []string{"r2"},
		UsePVCAsIVC:                     types.Ptr(false),
		ApplyTextNormalization:          types.TextNormalizationOff,
		ApplyLanguageTextNormalization:  types.Ptr(true),
		AudioEffects:                    &types.AudioEffects{FilterPresetID: "radio", Pan: types.Ptr(-0.5)},
	}
}

func TestNilRequestsSendEmptyBodies(t *testing.T) {
	client, ts := newTestServer(t, testutil.Respond(200, "audio/mpeg", "a"))
	rc, err := client.Convert(context.Background(), "v", nil)
	if err != nil {
		t.Fatal(err)
	}
	rc.Close()
	testutil.AssertJSON(t, ts.Last(t).Body, `{"text":""}`)
}

func newTestServer(t *testing.T, h http.HandlerFunc, opts ...option.RequestOption) (*Service, *testutil.Server) {
	t.Helper()
	ts, o := testutil.NewServer(t, h, opts...)
	return NewService(o...), ts
}
