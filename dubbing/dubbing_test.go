package dubbing

import (
	"context"
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
	"github.com/taigrr/elevenlabs/v2/voices"
)

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, []testutil.Endpoint[*Service]{
		{
			Name: "dubbing.list", Call: func(ctx context.Context, c *Service) (any, error) {
				return c.List(ctx, &ListRequest{PageSize: types.Ptr(2), DubbingStatuses: []Status{StatusDubbed}, DubbingModels: []Model{ModelV2}, OrderDirection: "ASCENDING"})
			},
			Method: http.MethodGet, Path: "/v1/dubbing",
			Query:    "dubbing_models=dubbing_v2&dubbing_statuses=dubbed&order_direction=ASCENDING&page_size=2",
			RespBody: `{"dubs":[{"dubbing_id":"d","name":"n","status":"dubbed","target_languages":["es"],"created_at":"2026-01-02T03:04:05Z"}],"has_more":false}`,
			Check: func(t *testing.T, r any) {
				if d := r.(*Page).Dubs[0]; d.Status != StatusDubbed || d.TargetLanguages[0] != "es" {
					t.Errorf("dub = %+v", d)
				}
			},
		},
		{
			Name: "dubbing.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				f := testutil.File("a.mp3", "mp3")
				return c.Create(ctx, &CreateRequest{File: &f, TargetLang: "es", NumSpeakers: types.Ptr(2), Watermark: types.Ptr(true), Mode: "automatic"})
			},
			Method: http.MethodPost, Path: "/v1/dubbing",
			Form:     map[string][]string{"file": {"file:a.mp3:mp3"}, "target_lang": {"es"}, "num_speakers": {"2"}, "watermark": {"true"}, "mode": {"automatic"}},
			RespBody: `{"dubbing_id":"d","expected_duration_sec":12}`,
			Check: func(t *testing.T, r any) {
				if d := r.(*CreateResponse); d.DubbingID != "d" || d.ExpectedDurationSec != 12 {
					t.Errorf("resp = %+v", d)
				}
			},
		},
		{
			Name:   "dubbing.get",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Get(ctx, "d") },
			Method: http.MethodGet, Path: "/v1/dubbing/d", RespBody: `{"dubbing_id":"d","created_at":"2026-01-02T03:04:05Z"}`,
		},
		{
			Name:   "dubbing.delete",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Delete(ctx, "d") },
			Method: http.MethodDelete, Path: "/v1/dubbing/d", RespBody: `{"status":"ok"}`,
		},
		{
			Name: "dubbing.audio.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.ReadAll(c.Audio.Get(ctx, "d", "es"))
			},
			Method: http.MethodGet, Path: "/v1/dubbing/d/audio/es", Accept: transport.AcceptAny, RespType: "audio/mpeg", RespBody: "dub",
			Check: testutil.CheckString("dub"),
		},
		{
			Name: "dubbing.transcript.getTranscriptForDub",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Transcript.GetTranscriptForDub(ctx, "d", "es", TranscriptFormatSRT)
			},
			Method: http.MethodGet, Path: "/v1/dubbing/d/transcript/es", Query: "format_type=srt", Accept: transport.AcceptAny,
			RespType: "text/plain", RespBody: "1\n00:00:00,000 --> 00:00:01,000\nhola\n",
			Check: func(t *testing.T, r any) {
				if tr := r.(*DubTranscript); tr.JSON != nil || tr.Text != "1\n00:00:00,000 --> 00:00:01,000\nhola\n" {
					t.Errorf("transcript = %+v", tr)
				}
			},
		},
		{
			Name: "dubbing.transcripts.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Transcripts.Get(ctx, "d", "es", TranscriptFormatJSON)
			},
			Method: http.MethodGet, Path: "/v1/dubbing/d/transcripts/es/format/json",
			RespBody: `{"transcript_format":"json","json":{"language":"es","utterances":[{"text":"hola","speaker_id":"s"}]}}`,
			Check: func(t *testing.T, r any) {
				if tr := r.(*Transcripts); *tr.JSON.Utterances[0].Text != "hola" {
					t.Errorf("transcripts = %+v", tr)
				}
			},
		},
		{
			Name: "dubbing.project.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Project.List(ctx, &ProjectListRequest{Status: "ready", SortDirection: "DESCENDING"})
			},
			Method: http.MethodGet, Path: "/v1/dubbing/project", Query: "sort_direction=DESCENDING&status=ready",
			RespBody: `{"projects":[{"project_id":"p","status":"ready","revision":1,"created_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-02T03:04:05Z"}]}`,
		},
		{
			Name: "dubbing.project.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Project.Create(ctx, &ProjectCreateRequest{SourceURL: "https://v", ModelID: "dubbing_v2", Keyterms: []string{"a", "b"}, WebhookIDs: []string{"w"}})
			},
			Method: http.MethodPost, Path: "/v1/dubbing/project",
			Form:     map[string][]string{"source_url": {"https://v"}, "model_id": {"dubbing_v2"}, "keyterms": {"a", "b"}, "webhook_ids": {"w"}},
			RespBody: `{"project_id":"p","status":"queued","revision":0,"created_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-02T03:04:05Z"}`,
		},
		{
			Name:   "dubbing.project.get",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Project.Get(ctx, "p") },
			Method: http.MethodGet, Path: "/v1/dubbing/project/p",
			RespBody: `{"project_id":"p","status":"failed","error":{"message_type":"error","error":"boom"},"revision":0,"created_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-02T03:04:05Z"}`,
			Check: func(t *testing.T, r any) {
				if p := r.(*Project); p.Error.Error != "boom" {
					t.Errorf("project = %+v", p)
				}
			},
		},
		{
			Name: "dubbing.project.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.VoidCall(c.Project.Delete(ctx, "p"))
			},
			Method: http.MethodDelete, Path: "/v1/dubbing/project/p",
		},
		{
			Name: "dubbing.project.language.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Project.Language.List(ctx, "p", &LanguageListRequest{PageSize: types.Ptr(1)})
			},
			Method: http.MethodGet, Path: "/v1/dubbing/project/p/language", Query: "page_size=1", RespBody: `{"languages":[]}`,
		},
		{
			Name: "dubbing.project.language.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Project.Language.Create(ctx, "p", &LanguageCreateRequest{TargetLanguage: "es", VoiceSettings: &types.VoiceSettings{Stability: types.Ptr(0.4)}})
			},
			Method: http.MethodPost, Path: "/v1/dubbing/project/p/language",
			Body:     `{"target_language":"es","voice_settings":{"stability":0.4}}`,
			RespBody: `{"language_id":"l","project_id":"p","target_language":"es","status":"queued","revision":0,"created_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-02T03:04:05Z"}`,
		},
		{
			Name: "dubbing.project.language.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Project.Language.Get(ctx, "p", "l")
			},
			Method: http.MethodGet, Path: "/v1/dubbing/project/p/language/l",
			RespBody: `{"language_id":"l","status":"completed","outputs":{"lossless_audio":"u"},"revision":1,"created_at":"2026-01-02T03:04:05Z","updated_at":"2026-01-02T03:04:05Z"}`,
		},
		{
			Name: "dubbing.project.language.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.VoidCall(c.Project.Language.Delete(ctx, "p", "l"))
			},
			Method: http.MethodDelete, Path: "/v1/dubbing/project/p/language/l",
		},
		{
			Name: "dubbing.project.language.transcript.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Project.Language.Transcript.Get(ctx, "p", "l")
			},
			Method: http.MethodGet, Path: "/v1/dubbing/project/p/language/l/transcript",
			RespBody: `{"target_language":"es","segments":[{"id":"s","speaker_id":"a","start_s":0,"end_s":1,"source_text":"hi","translation":"hola"}],"revision":2}`,
		},
		{
			Name: "dubbing.project.language.transcript.updateSegment",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Project.Language.Transcript.UpdateSegment(ctx, "p", "l", "s", &TargetSegmentUpdate{Translation: types.Ptr("hola")})
			},
			Method: http.MethodPatch, Path: "/v1/dubbing/project/p/language/l/transcript/segment/s",
			Body: `{"translation":"hola"}`, RespBody: `{"segment":{"id":"s"},"revision":3}`,
		},
		{
			Name: "dubbing.project.language.transcript.updateSegments",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Project.Language.Transcript.UpdateSegments(ctx, "p", "l", map[string]TargetSegmentUpdate{"s": {Translation: types.Ptr("x")}})
			},
			Method: http.MethodPatch, Path: "/v1/dubbing/project/p/language/l/transcript/segments",
			Body: `{"segments":{"s":{"translation":"x"}}}`, RespBody: `{"segments":[],"revision":4}`,
		},
		{
			Name: "dubbing.project.language.transcript.regenerate",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Project.Language.Transcript.Regenerate(ctx, "p", "l")
			},
			Method: http.MethodPost, Path: "/v1/dubbing/project/p/language/l/transcript/regenerate",
			RespBody: `{"regenerated_segment_ids":["s"],"regenerated_seconds":1.5,"charged_seconds":0,"free_regeneration_seconds_remaining":10}`,
		},
		{
			Name:   "dubbing.project.transcript.get",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Project.Transcript.Get(ctx, "p") },
			Method: http.MethodGet, Path: "/v1/dubbing/project/p/transcript",
			RespBody: `{"language":"en","segments":[{"id":"s","text":"hi","speaker_id":"a","start_s":0,"end_s":1}],"revision":1}`,
		},
		{
			Name: "dubbing.project.transcript.deleteSegment",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Project.Transcript.DeleteSegment(ctx, "p", "s")
			},
			Method: http.MethodDelete, Path: "/v1/dubbing/project/p/transcript/segment/s", RespBody: `{"revision":2}`,
		},
		{
			Name: "dubbing.project.transcript.updateSegment",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Project.Transcript.UpdateSegment(ctx, "p", "s", &SegmentUpdate{Text: types.Ptr("hey"), EndS: types.Ptr(2.0)})
			},
			Method: http.MethodPatch, Path: "/v1/dubbing/project/p/transcript/segment/s",
			Body: `{"text":"hey","end_s":2}`, RespBody: `{"segment":{"id":"s"},"revision":3}`,
		},
		{
			Name: "dubbing.project.transcript.updateSegments",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Project.Transcript.UpdateSegments(ctx, "p", nil)
			},
			Method: http.MethodPatch, Path: "/v1/dubbing/project/p/transcript/segments",
			Body: `{"segments":{}}`, RespBody: `{"segments":[],"revision":4}`,
		},
		{
			Name: "dubbing.project.transcript.createSegment",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Project.Transcript.CreateSegment(ctx, "p", &SegmentCreate{Text: "t", SpeakerID: "a", StartS: 1, EndS: 2})
			},
			Method: http.MethodPost, Path: "/v1/dubbing/project/p/transcript/segment",
			Body: `{"text":"t","speaker_id":"a","start_s":1,"end_s":2}`, RespBody: `{"segment":{"id":"n"},"revision":5}`,
		},
		{
			Name:   "dubbing.resource.get",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Resource.Get(ctx, "d") },
			Method: http.MethodGet, Path: "/v1/dubbing/resource/d",
			RespBody: `{"id":"d","version":3,"source_language":"en","target_languages":["es"],"input":{"src":"s","content_type":"video/mp4","bucket_name":"b","random_path_slug":"r","duration_secs":5,"is_audio":false,"url":"u"},"speaker_tracks":{},"speaker_segments":{"s":{"id":"s","start_time":0,"end_time":1,"text":"hi","subtitles":[],"dubs":{"es":{"start_time":0,"end_time":1,"subtitles":[],"audio_stale":true}}}},"renders":{"r":{"id":"r","version":1,"type":"mp4","status":"complete"}}}`,
			Check: func(t *testing.T, r any) {
				res := r.(*Resource)
				if !res.SpeakerSegments["s"].Dubs["es"].AudioStale || *res.Renders["r"].Type != RenderTypeMP4 {
					t.Errorf("resource = %+v", res)
				}
			},
		},
		{
			Name: "dubbing.resource.migrateSegments",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Resource.MigrateSegments(ctx, "d", []string{"a"}, "sp")
			},
			Method: http.MethodPost, Path: "/v1/dubbing/resource/d/migrate-segments",
			Body: `{"segment_ids":["a"],"speaker_id":"sp"}`, RespBody: `{"version":4}`,
		},
		{
			Name: "dubbing.resource.transcribe",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Resource.Transcribe(ctx, "d", []string{"a"})
			},
			Method: http.MethodPost, Path: "/v1/dubbing/resource/d/transcribe", Body: `{"segments":["a"]}`, RespBody: `{"version":4}`,
		},
		{
			Name: "dubbing.resource.translate",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Resource.Translate(ctx, "d", []string{"a"}, []string{"es"})
			},
			Method: http.MethodPost, Path: "/v1/dubbing/resource/d/translate", Body: `{"segments":["a"],"languages":["es"]}`, RespBody: `{"version":4}`,
		},
		{
			Name:   "dubbing.resource.dub",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Resource.Dub(ctx, "d", nil, nil) },
			Method: http.MethodPost, Path: "/v1/dubbing/resource/d/dub", Body: `{"segments":[]}`, RespBody: `{"version":4}`,
		},
		{
			Name: "dubbing.resource.render",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Resource.Render(ctx, "d", "es", &RenderRequest{RenderType: RenderTypeMP3, NormalizeVolume: types.Ptr(true)})
			},
			Method: http.MethodPost, Path: "/v1/dubbing/resource/d/render/es",
			Body: `{"render_type":"mp3","normalize_volume":true}`, RespBody: `{"version":5,"render_id":"r"}`,
		},
		{
			Name: "dubbing.resource.language.add",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Resource.Language.Add(ctx, "d", "fr")
			},
			Method: http.MethodPost, Path: "/v1/dubbing/resource/d/language", Body: `{"language":"fr"}`, RespBody: `{"version":6}`,
		},
		{
			Name: "dubbing.resource.segment.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Resource.Segment.Update(ctx, "d", "s", "es", &SegmentUpdateRequest{Text: types.Ptr("hola"), StartTime: types.Ptr(0.5)})
			},
			Method: http.MethodPatch, Path: "/v1/dubbing/resource/d/segment/s/es",
			Body: `{"start_time":0.5,"text":"hola"}`, RespBody: `{"version":7}`,
		},
		{
			Name: "dubbing.resource.segment.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Resource.Segment.Delete(ctx, "d", "s")
			},
			Method: http.MethodDelete, Path: "/v1/dubbing/resource/d/segment/s", RespBody: `{"version":8}`,
		},
		{
			Name: "dubbing.resource.speaker.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Resource.Speaker.Update(ctx, "d", "sp", &SpeakerRequest{VoiceID: "track-clone", Languages: []string{"es"}})
			},
			Method: http.MethodPatch, Path: "/v1/dubbing/resource/d/speaker/sp",
			Body: `{"voice_id":"track-clone","languages":["es"]}`, RespBody: `{"version":9}`,
		},
		{
			Name: "dubbing.resource.speaker.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Resource.Speaker.Create(ctx, "d", &SpeakerRequest{SpeakerName: "Bob", VoiceStability: types.Ptr(0.3), Languages: []string{"x"}})
			},
			Method: http.MethodPost, Path: "/v1/dubbing/resource/d/speaker",
			Body: `{"speaker_name":"Bob","voice_stability":0.3}`, RespBody: `{"version":10,"speaker_id":"sp2"}`,
		},
		{
			Name: "dubbing.resource.speaker.findSimilarVoices",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Resource.Speaker.FindSimilarVoices(ctx, "d", "sp")
			},
			Method: http.MethodGet, Path: "/v1/dubbing/resource/d/speaker/sp/similar-voices",
			RespBody: `{"voices":[{"voice_id":"v","name":"n","category":"premade"}]}`,
			Check: func(t *testing.T, r any) {
				if v := r.(*SimilarVoicesResponse).Voices[0]; v.Category != voices.CategoryPremade {
					t.Errorf("voice = %+v", v)
				}
			},
		},
		{
			Name: "dubbing.resource.speaker.segment.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Resource.Speaker.Segment.Create(ctx, "d", "sp", &SegmentCreateRequest{StartTime: 1, EndTime: 2, Text: types.Ptr("t")})
			},
			Method: http.MethodPost, Path: "/v1/dubbing/resource/d/speaker/sp/segment",
			Body: `{"start_time":1,"end_time":2,"text":"t"}`, RespBody: `{"version":11,"new_segment":"s2"}`,
		},
	})
}

func TestTranscriptJSONVariant(t *testing.T) {
	client, _ := newTestServer(t, testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"language":"es","utterances":[{"text":"hola"}]}`))
	tr, err := client.Transcript.GetTranscriptForDub(context.Background(), "d", "es", TranscriptFormatJSON)
	if err != nil || tr.JSON == nil || tr.JSON.Language != "es" || tr.Text != "" {
		t.Fatalf("transcript = %+v, err = %v", tr, err)
	}
	client, _ = newTestServer(t, testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `"WEBVTT\n"`))
	if tr, err = client.Transcript.GetTranscriptForDub(context.Background(), "d", "es", TranscriptFormatWebVTT); err != nil || tr.Text != "WEBVTT\n" {
		t.Fatalf("transcript = %+v, err = %v", tr, err)
	}
}

func TestListAll(t *testing.T) {
	client, ts := newTestServer(t, testutil.Sequence(
		testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"projects":[{"project_id":"a"}],"next_cursor":"c"}`),
		testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"projects":[{"project_id":"b"}]}`),
	))
	var ids []string
	for p, err := range client.Project.ListAll(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, p.ProjectID)
	}
	if len(ids) != 2 || ts.Last(t).RawQuery != "cursor=c" {
		t.Errorf("ids = %v", ids)
	}
}

func newTestServer(t *testing.T, h http.HandlerFunc, opts ...option.RequestOption) (*Service, *testutil.Server) {
	t.Helper()
	ts, o := testutil.NewServer(t, h, opts...)
	return NewService(o...), ts
}
