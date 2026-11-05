package speechtotext

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
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
			Name: "speechToText.convert",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Convert(ctx, fullSTTRequest())
			},
			Method:   http.MethodPost,
			Path:     "/v1/speech-to-text",
			Query:    "enable_logging=false&token=tok",
			Form:     fullSTTForm,
			RespBody: `{"language_code":"en","language_probability":0.98,"text":"hello","words":[{"text":"hello","start":0,"end":0.5,"type":"word","speaker_id":"speaker_0","logprob":-0.1}]}`,
			Check: func(t *testing.T, r any) {
				resp := r.(*ConvertResponse)
				if resp.Chunk == nil || resp.Chunk.Text != "hello" || resp.Chunk.Words[0].Type != WordTypeWord || *resp.Chunk.Words[0].SpeakerID != "speaker_0" {
					t.Errorf("stt = %+v", resp)
				}
			},
		},
		{
			Name: "speechToText.transcripts.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Transcripts.Get(ctx, "t1")
			},
			Method: http.MethodGet, Path: "/v1/speech-to-text/transcripts/t1",
			RespBody: `{"language_code":"en","language_probability":1,"text":"hi","words":[]}`,
			Check: func(t *testing.T, r any) {
				if tr := r.(*Transcript); tr.Chunk == nil || tr.Chunk.Text != "hi" || tr.Multichannel != nil {
					t.Errorf("transcript = %+v", tr)
				}
			},
		},
		{
			Name: "speechToText.transcripts.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Transcripts.Delete(ctx, "t1")
			},
			Method: http.MethodDelete, Path: "/v1/speech-to-text/transcripts/t1", RespBody: `{"ok":true}`,
			Check: func(t *testing.T, r any) {
				if string(r.(json.RawMessage)) != `{"ok":true}` {
					t.Errorf("raw = %s", r)
				}
			},
		},
	})
}

func TestRealtimeValidation(t *testing.T) {
	client := NewService(option.WithAPIKey("k"), option.WithBaseURL("http://127.0.0.1:1"))
	tests := []struct {
		name string
		req  *RealtimeRequest
	}{
		{"nil", nil},
		{"missing model", &RealtimeRequest{}},
		{"vad silence", &RealtimeRequest{ModelID: "m", VADSilenceThresholdSecs: types.Ptr(0.1)}},
		{"vad threshold", &RealtimeRequest{ModelID: "m", VADThreshold: types.Ptr(0.95)}},
		{"min speech", &RealtimeRequest{ModelID: "m", MinSpeechDurationMs: types.Ptr(10)}},
		{"min silence", &RealtimeRequest{ModelID: "m", MinSilenceDurationMs: types.Ptr(3000)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := client.Realtime(testutil.Ctx(t), tt.req); !errors.Is(err, realtime.ErrInvalidParameter) {
				t.Errorf("err = %v, want ErrInvalidParameter", err)
			}
		})
	}
}

func TestRealtime(t *testing.T) {
	done := make(chan struct{})
	client, ts := newTestServer(t, testutil.WSHandler(t, func(ctx context.Context, c *websocket.Conn) {
		defer close(done)
		testutil.ExpectMsg(t, ctx, c, `{"message_type":"input_audio_chunk","audio_base_64":"`+testutil.B64("pcm")+`","commit":false,"sample_rate":24000}`)
		testutil.ExpectMsg(t, ctx, c, `{"message_type":"input_audio_chunk","audio_base_64":"`+testutil.B64("x")+`","commit":false,"sample_rate":8000,"previous_text":"ctx"}`)
		testutil.ExpectMsg(t, ctx, c, `{"message_type":"input_audio_chunk","audio_base_64":"`+testutil.B64("w")+`","commit":false,"sample_rate":24000}`)
		testutil.ExpectMsg(t, ctx, c, `{"message_type":"input_audio_chunk","audio_base_64":"","commit":true,"sample_rate":24000}`)
		testutil.WriteMsg(t, ctx, c, `{"message_type":"session_started","session_id":"s1","config":{"sample_rate":24000,"model_id":"scribe_v2_realtime","commit_strategy":"vad"}}`)
		testutil.WriteMsg(t, ctx, c, `{"message_type":"committed_transcript_with_timestamps","text":"hi","language_code":"en","words":[{"text":"hi","start":0,"end":0.4,"type":"word","characters":["h","i"]}]}`)
		testutil.WriteMsg(t, ctx, c, `{"message_type":"final_transcript_with_timestamps","text":"hi","words":[{"text":"hi","characters":[{"text":"h","start":0,"end":0.2}],"channel_index":null}]}`)
		testutil.WriteMsg(t, ctx, c, `{"message_type":"committed_transcript_entities","text":"hi bob","entities":[{"text":"bob","entity_type":"person_name","start_char":3,"end_char":6}]}`)
		testutil.WriteMsg(t, ctx, c, `{"message_type":"auth_error","error":"invalid key"}`)
		_ = c.Close(websocket.StatusPolicyViolation, "bye")
	}))
	ctx := testutil.Ctx(t)
	conn, err := client.Realtime(ctx, &RealtimeRequest{
		ModelID: "scribe_v2_realtime", CommitStrategy: CommitStrategyVAD, AudioFormat: RealtimeAudioPCM24000,
		VADThreshold: types.Ptr(0.5), SecondaryLanguages: []string{"de", "fr"}, Keyterms: []string{"Go"},
		EntityDetection: []string{"pii"}, EntityRedactionMode: EntityRedactionEntityType, NoVerbatim: types.Ptr(true),
		SampleRate: 24000,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	req := ts.Last(t)
	if req.Path != "/v1/speech-to-text/realtime" || req.Header.Get(transport.HeaderAPIKey) != "test-key" {
		t.Errorf("path = %s headers = %v", req.Path, req.Header)
	}
	testutil.AssertQuery(t, req.RawQuery, url.Values{
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
	if msg == nil || msg.MessageType != RealtimeErrorAuth || !errors.Is(err, apierror.ErrUnauthorized) || !errors.Is(err, realtime.ErrRealtime) {
		t.Fatalf("auth error = %+v, %v", msg, err)
	}
	_, err = conn.Recv(ctx)
	var closeErr *realtime.CloseError
	if !errors.As(err, &closeErr) || !errors.Is(err, realtime.ErrSocketClosed) || closeErr.Code != int(websocket.StatusPolicyViolation) || closeErr.Reason != "bye" {
		t.Fatalf("close err = %v", err)
	}
	<-done
}

func TestRealtimeSendFrom(t *testing.T) {
	audio := bytes.Repeat([]byte{1}, realtimeChunkSize+10)
	done := make(chan struct{})
	client, ts := newTestServer(t, testutil.WSHandler(t, func(ctx context.Context, c *websocket.Conn) {
		defer close(done)
		var got []byte
		for {
			m := testutil.ReadMsg(t, ctx, c)
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
	ctx := testutil.Ctx(t)
	conn, err := client.Realtime(ctx, &RealtimeRequest{ModelID: "m", Token: "tok"})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if ts.Last(t).Header.Get(transport.HeaderAPIKey) != "" {
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

func TestRealtimeStreamURLWithoutFFmpeg(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	conn := &RealtimeSession{}
	if err := conn.StreamURL(testutil.Ctx(t), "https://example.com/a.mp3"); !errors.Is(err, ErrFFmpegNotFound) {
		t.Errorf("err = %v, want ErrFFmpegNotFound", err)
	}
}

func TestTranscriptMultichannel(t *testing.T) {
	var tr Transcript
	if err := json.Unmarshal([]byte(`{"transcripts":[],"transcription_id":"t"}`), &tr); err != nil {
		t.Fatal(err)
	}
	if tr.Multichannel == nil || tr.Chunk != nil {
		t.Errorf("transcript = %+v", tr)
	}
}

func fullSTTRequest() *ConvertRequest {
	return &ConvertRequest{
		Token:                   "tok",
		EnableLogging:           types.Ptr(false),
		ModelID:                 ModelScribeV1,
		File:                    &types.File{Reader: bytes.NewReader([]byte("wav")), Name: "a.wav"},
		LanguageCode:            "en",
		TranscriptEdit:          `{"x":1}`,
		TagAudioEvents:          types.Ptr(true),
		NumSpeakers:             types.Ptr(2),
		TimestampsGranularity:   TimestampsGranularityCharacter,
		Diarize:                 types.Ptr(true),
		DiarizationThreshold:    types.Ptr(0.25),
		AdditionalFormats:       []ExportOptions{{Format: ExportFormatSRT, MaxCharactersPerLine: types.Ptr(42)}},
		FileFormat:              FileFormatOther,
		CloudStorageURL:         "https://example.com/a.wav",
		SourceURL:               "https://example.com/v",
		Webhook:                 types.Ptr(false),
		WebhookID:               "wh",
		Temperature:             types.Ptr(0.5),
		Seed:                    types.Ptr(int64(9)),
		UseMultiChannel:         types.Ptr(false),
		MultichannelOutputStyle: MultichannelOutputStyleCombined,
		WebhookMetadata:         map[string]any{"job": "1"},
		EntityDetection:         []string{"pii", "phi"},
		NoVerbatim:              types.Ptr(true),
		UseSpeakerLibrary:       types.Ptr(false),
		DetectSpeakerRoles:      types.Ptr(true),
		EntityRedaction:         []string{"all"},
		EntityRedactionMode:     "mask",
		Keyterms:                []string{"ElevenLabs", "Scribe"},
	}
}

var fullSTTForm = map[string][]string{
	"model_id":                  {"scribe_v1"},
	"file":                      {"file:a.wav:wav"},
	"language_code":             {"en"},
	"transcript_edit":           {`{"x":1}`},
	"tag_audio_events":          {"true"},
	"num_speakers":              {"2"},
	"timestamps_granularity":    {"character"},
	"diarize":                   {"true"},
	"diarization_threshold":     {"0.25"},
	"additional_formats":        {`[{"format":"srt","max_characters_per_line":42}]`},
	"file_format":               {"other"},
	"cloud_storage_url":         {"https://example.com/a.wav"},
	"source_url":                {"https://example.com/v"},
	"webhook":                   {"false"},
	"webhook_id":                {"wh"},
	"temperature":               {"0.5"},
	"seed":                      {"9"},
	"use_multi_channel":         {"false"},
	"multichannel_output_style": {"combined"},
	"webhook_metadata":          {`{"job":"1"}`},
	"entity_detection":          {"pii", "phi"},
	"no_verbatim":               {"true"},
	"use_speaker_library":       {"false"},
	"detect_speaker_roles":      {"true"},
	"entity_redaction":          {"all"},
	"entity_redaction_mode":     {"mask"},
	"keyterms":                  {"ElevenLabs", "Scribe"},
}

func TestResponseVariants(t *testing.T) {
	tests := []struct {
		name string
		body string
		want func(r ConvertResponse) bool
	}{
		{"chunk", `{"language_code":"en","language_probability":1,"text":"hi","words":[]}`, func(r ConvertResponse) bool {
			return r.Chunk != nil && r.Chunk.Text == "hi" && r.Multichannel == nil && r.Webhook == nil
		}},
		{"multichannel", `{"transcripts":[{"language_code":"en","language_probability":1,"text":"a","words":[],"channel_index":0}],"transcription_id":"t"}`, func(r ConvertResponse) bool {
			return r.Multichannel != nil && len(r.Multichannel.Transcripts) == 1 && *r.Multichannel.TranscriptionID == "t"
		}},
		{"webhook", `{"message":"queued","request_id":"r","transcription_id":"t"}`, func(r ConvertResponse) bool {
			return r.Webhook != nil && r.Webhook.RequestID == "r"
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r ConvertResponse
			if err := json.Unmarshal([]byte(tt.body), &r); err != nil {
				t.Fatal(err)
			}
			if !tt.want(r) {
				t.Errorf("unexpected variant: %+v", r)
			}
			out, err := json.Marshal(r)
			if err != nil {
				t.Fatal(err)
			}
			testutil.AssertJSONSubset(t, out, tt.body)
		})
	}
	if out, _ := json.Marshal(ConvertResponse{}); string(out) != "null" {
		t.Errorf("empty marshal = %s", out)
	}
}

func TestRequiresModel(t *testing.T) {
	client, ts := newTestServer(t, testutil.Respond(200, transport.ContentTypeJSON, `{}`))
	if _, err := client.Convert(context.Background(), &ConvertRequest{}); err == nil {
		t.Error("expected error without model id")
	}
	if _, err := client.Convert(context.Background(), nil); err == nil {
		t.Error("expected error for nil request")
	}
	if len(ts.All()) != 0 {
		t.Error("request sent despite validation error")
	}
}

func newTestServer(t *testing.T, h http.HandlerFunc, opts ...option.RequestOption) (*Service, *testutil.Server) {
	t.Helper()
	ts, o := testutil.NewServer(t, h, opts...)
	return NewService(o...), ts
}
