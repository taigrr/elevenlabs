package texttodialogue

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
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
			Name: "textToDialogue.convert",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.ReadAll(c.Convert(ctx, &ConvertRequest{
					OutputFormat: types.OutputFormatMP3_44100_128, EnableLogging: types.Ptr(true),
					Inputs: []Input{{Text: "hi", VoiceID: "a"}}, ModelID: "eleven_v3", LanguageCode: "en",
					Settings: &Settings{Stability: types.Ptr(0.5)}, PreviousText: "p", FutureText: "f", Seed: types.Ptr(int64(1)),
					ApplyTextNormalization: types.TextNormalizationOn, PreviousRequestIDs: []string{"r"},
				}))
			},
			Method: http.MethodPost, Path: "/v1/text-to-dialogue", Accept: transport.AcceptAny, RespType: "audio/mpeg", RespBody: "dlg",
			Query: "enable_logging=true&output_format=mp3_44100_128",
			Body:  `{"inputs":[{"text":"hi","voice_id":"a"}],"model_id":"eleven_v3","language_code":"en","settings":{"stability":0.5},"previous_text":"p","future_text":"f","seed":1,"apply_text_normalization":"on","previous_request_ids":["r"]}`,
			Check: testutil.CheckString("dlg"),
		},
		{
			Name: "textToDialogue.stream",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.ReadAll(c.Stream(ctx, &ConvertRequest{Inputs: []Input{{Text: "hi", VoiceID: "a"}}}))
			},
			Method: http.MethodPost, Path: "/v1/text-to-dialogue/stream", Accept: transport.AcceptAny, RespType: "audio/mpeg", RespBody: "dlg",
			Body:  `{"inputs":[{"text":"hi","voice_id":"a"}]}`,
			Check: testutil.CheckString("dlg"),
		},
		{
			Name: "textToDialogue.convertWithTimestamps",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.ConvertWithTimestamps(ctx, &ConvertRequest{Inputs: []Input{{Text: "hi", VoiceID: "a"}}})
			},
			Method: http.MethodPost, Path: "/v1/text-to-dialogue/with-timestamps",
			Body:     `{"inputs":[{"text":"hi","voice_id":"a"}]}`,
			RespBody: `{"audio_base64":"AA==","voice_segments":[{"voice_id":"a","start_time_seconds":0,"end_time_seconds":1,"character_start_index":0,"character_end_index":2,"dialogue_input_index":0}]}`,
			Check: func(t *testing.T, r any) {
				a := r.(*AudioWithTimestampsAndVoiceSegments)
				if a.AudioBase64 != "AA==" || len(a.VoiceSegments) != 1 || a.VoiceSegments[0].CharacterEndIndex != 2 {
					t.Errorf("dialogue = %+v", a)
				}
			},
		},
		{
			Name: "textToDialogue.streamWithTimestamps",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.DrainStream(c.StreamWithTimestamps(ctx, &ConvertRequest{Inputs: []Input{{Text: "hi", VoiceID: "a"}}}))
			},
			Method: http.MethodPost, Path: "/v1/text-to-dialogue/stream/with-timestamps",
			Body:     `{"inputs":[{"text":"hi","voice_id":"a"}]}`,
			RespBody: "{\"audio_base64\":\"AA==\",\"voice_segments\":[]}\n{\"audio_base64\":\"AQ==\",\"voice_segments\":[]}\n",
			Check: func(t *testing.T, r any) {
				if c := r.([]AudioWithTimestampsAndVoiceSegments); len(c) != 2 || c[1].AudioBase64 != "AQ==" {
					t.Errorf("chunks = %+v", c)
				}
			},
		},
	})
}

func TestStreamInput(t *testing.T) {
	done := make(chan struct{})
	client, ts := newTestServer(t, testutil.WSHandler(t, func(ctx context.Context, c *websocket.Conn) {
		defer close(done)
		testutil.ExpectMsg(t, ctx, c, `{"voices":["a","b"],"xi_api_key":"k","voice_settings":{"stability":0.4}}`)
		testutil.ExpectMsg(t, ctx, c, `{"inputs":[{"text":"Hi","voice_id":"a"},{"text":"Yo","voice_id":"b","new_turn":true}]}`)
		testutil.ExpectMsg(t, ctx, c, `{"flush":true}`)
		testutil.ExpectMsg(t, ctx, c, `{"keep_alive":true}`)
		testutil.ExpectMsg(t, ctx, c, `{"close_socket":true}`)
		testutil.WriteMsg(t, ctx, c, `{"audio":"`+testutil.B64("A")+`","alignment":{"chars":["H"],"char_start_times_ms":[0],"char_durations_ms":[1]}}`)
		testutil.WriteMsg(t, ctx, c, `{"is_final_audio_for_turn":true}`)
		testutil.WriteMsg(t, ctx, c, `{"message":"bad voice","error":"invalid_voice","code":400,"param":"voice_id"}`)
		_ = c.Close(websocket.StatusNormalClosure, "")
	}))
	ctx := testutil.Ctx(t)
	stream, err := client.StreamInput(ctx, &StreamInputRequest{ModelID: "eleven_v3", EnableLogging: types.Ptr(false)})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	req := ts.Last(t)
	if req.Path != "/v1/text-to-dialogue/stream-input" {
		t.Errorf("path = %s", req.Path)
	}
	testutil.AssertQuery(t, req.RawQuery, url.Values{"model_id": {"eleven_v3"}, "enable_logging": {"false"}})
	err = errors.Join(
		stream.Send(ctx, &StreamMessage{Voices: []string{"a", "b"}, XIAPIKey: "k", VoiceSettings: &StreamVoiceSettings{Stability: types.Ptr(0.4)}}),
		stream.SendInputs(ctx, StreamInput{Text: "Hi", VoiceID: "a"}, StreamInput{Text: "Yo", VoiceID: "b", NewTurn: types.Ptr(true)}),
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
	var rtErr *realtime.Error
	if !errors.As(err, &rtErr) || !errors.Is(err, realtime.ErrRealtime) || !errors.Is(err, apierror.ErrBadRequest) {
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

func TestMultiStreamInput(t *testing.T) {
	done := make(chan struct{})
	client, ts := newTestServer(t, testutil.WSHandler(t, func(ctx context.Context, c *websocket.Conn) {
		defer close(done)
		testutil.ExpectMsg(t, ctx, c, `{"context_id":"x","inputs":[{"text":"Hi","voice_id":"a"}]}`)
		testutil.ExpectMsg(t, ctx, c, `{"context_id":"x","flush":true}`)
		testutil.ExpectMsg(t, ctx, c, `{"context_id":"x","keep_alive":true}`)
		testutil.ExpectMsg(t, ctx, c, `{"context_id":"x","close_context":true}`)
		testutil.ExpectMsg(t, ctx, c, `{"close_socket":true}`)
		testutil.WriteMsg(t, ctx, c, `{"is_final":true,"context_id":"x"}`)
		_ = c.Close(websocket.StatusNormalClosure, "")
	}))
	ctx := testutil.Ctx(t)
	stream, err := client.MultiStreamInput(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	if p := ts.Last(t).Path; p != "/v1/text-to-dialogue/multi-stream-input" {
		t.Errorf("path = %s", p)
	}
	err = errors.Join(
		stream.SendInputs(ctx, "x", StreamInput{Text: "Hi", VoiceID: "a"}),
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

func newTestServer(t *testing.T, h http.HandlerFunc, opts ...option.RequestOption) (*Service, *testutil.Server) {
	t.Helper()
	ts, o := testutil.NewServer(t, h, opts...)
	return NewService(o...), ts
}
