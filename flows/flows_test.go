package flows

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, []testutil.Endpoint[*Service]{
		{

			Name: "flows.image.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Image.List(ctx, &MediaGenerationListRequest{PageSize: types.Ptr(2), Status: GenerationStatusCompleted, ModelID: "gpt-image-2"})
			},
			Method: http.MethodGet, Path: "/v1/flows/image", Query: "model_id=gpt-image-2&page_size=2&status=completed",
			RespBody: `{"generations":[{"id":"g","status":"completed","content_url":"https://c","content_mime_type":"image/png"}],"has_more":false}`,
			Check: func(t *testing.T, r any) {
				if g := r.(*MediaGenerationPage).Generations[0]; g.Status != GenerationStatusCompleted || g.ContentURL != "https://c" {
					t.Errorf("generation = %+v", g)
				}
			},
		},
		{
			Name: "flows.image.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Image.Create(ctx, &ImageGenerationRequest{ModelID: "gpt-image-2", Prompt: "corgi", Images: []MediaReference{AssetRef("a")}, Webhook: AllWebhooks()})
			},
			Method: http.MethodPost, Path: "/v1/flows/image",
			Body:     `{"model_id":"gpt-image-2","prompt":"corgi","webhook":{"type":"all"},"images":[{"type":"asset","asset_id":"a"}]}`,
			RespBody: `{"id":"g","status":"pending"}`,
		},
		{
			Name:   "flows.image.get",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Image.Get(ctx, "g") },
			Method: http.MethodGet, Path: "/v1/flows/image/g", RespBody: `{"id":"g","status":"failed","failure_reason":"moderated","error_message":"no"}`,
			Check: func(t *testing.T, r any) {
				if g := r.(*MediaGeneration); g.FailureReason != "moderated" {
					t.Errorf("generation = %+v", g)
				}
			},
		},
		{
			Name: "flows.video.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Video.List(ctx, nil)
			},
			Method: http.MethodGet, Path: "/v1/flows/video", RespBody: `{"generations":[]}`,
		},
		{
			Name: "flows.video.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Video.Create(ctx, &VideoGenerationRequest{ModelID: "veo-3.1-generate-001", Prompt: "p", DurationSecs: types.Ptr(8.0), Images: []VeoImageReference{{Image: GenerationRef("g"), Role: "style"}}})
			},
			Method: http.MethodPost, Path: "/v1/flows/video",
			Body:     `{"model_id":"veo-3.1-generate-001","prompt":"p","duration_secs":8,"images":[{"image":{"type":"generation","generation_id":"g"},"role":"style"}]}`,
			RespBody: `{"id":"v","status":"pending"}`,
		},
		{
			Name:   "flows.video.get",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Video.Get(ctx, "v") },
			Method: http.MethodGet, Path: "/v1/flows/video/v", RespBody: `{"id":"v","status":"generating"}`,
		},
		{
			Name: "flows.textToSpeech.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.TextToSpeech.List(ctx, &MediaGenerationListRequest{Cursor: "c"})
			},
			Method: http.MethodGet, Path: "/v1/flows/text-to-speech", Query: "cursor=c", RespBody: `{"generations":[]}`,
		},
		{
			Name: "flows.textToSpeech.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.TextToSpeech.Create(ctx, &TextToSpeechGenerationRequest{ModelID: "eleven_v3", Text: "hi", Voice: "v", VoiceSettings: &VoiceSettings{Stability: types.Ptr(0.5)}})
			},
			Method: http.MethodPost, Path: "/v1/flows/text-to-speech",
			Body: `{"model_id":"eleven_v3","text":"hi","voice":"v","voice_settings":{"stability":0.5}}`, RespBody: `{"id":"t","status":"pending"}`,
		},
		{
			Name:   "flows.textToSpeech.get",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.TextToSpeech.Get(ctx, "t") },
			Method: http.MethodGet, Path: "/v1/flows/text-to-speech/t", RespBody: `{"id":"t","status":"pending"}`,
		},
		{
			Name: "flows.templates.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Templates.List(ctx, &TemplatesListRequest{VersionsPerTemplate: types.Ptr(1), Search: "s"})
			},
			Method: http.MethodGet, Path: "/v1/flows/templates", Query: "search=s&versions_per_template=1",
			RespBody: `{"templates":[{"id":"t","name":"n","versions":[{"version_id":"v","published_at_unix":1,"is_latest":true,"inputs":[{"id":"prompt","content_schema":{"type":"string"}}],"outputs":[]}],"has_more_versions":false}]}`,
			Check: func(t *testing.T, r any) {
				if tp := r.(*TemplatePage).Templates[0]; tp.Versions[0].Inputs[0].ID != "prompt" {
					t.Errorf("template = %+v", tp)
				}
			},
		},
		{
			Name: "flows.templates.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Templates.Get(ctx, "t", &TemplatesGetRequest{VersionsPerTemplate: types.Ptr(2)})
			},
			Method: http.MethodGet, Path: "/v1/flows/templates/t", Query: "versions_per_template=2", RespBody: `{"id":"t"}`,
		},
		{
			Name: "flows.templates.runs.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Templates.Runs.List(ctx, "t", &TemplateRunsListRequest{VersionID: "v"})
			},
			Method: http.MethodGet, Path: "/v1/flows/templates/t/runs", Query: "version_id=v", RespBody: `{"runs":[]}`,
		},
		{
			Name: "flows.templates.runs.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Templates.Runs.Create(ctx, "t", &TemplateRunCreateRequest{Inputs: map[string]any{"prompt": "corgi", "voice": VoiceRef("v")}})
			},
			Method: http.MethodPost, Path: "/v1/flows/templates/t/runs",
			Body:     `{"inputs":{"prompt":"corgi","voice":{"type":"voice","voice_id":"v"}}}`,
			RespBody: `{"id":"r","template_id":"t","version_id":"v","status":"pending","outputs":{}}`,
		},
		{
			Name:   "flows.templates.runs.get",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Templates.Runs.Get(ctx, "t", "r") },
			Method: http.MethodGet, Path: "/v1/flows/templates/t/runs/r",
			RespBody: `{"id":"r","status":"completed","outputs":{"out":{"type":"image","id":"o","status":"completed","content_url":"https://i"}}}`,
			Check: func(t *testing.T, r any) {
				if o := r.(*TemplateRun).Outputs["out"]; o.Type != "image" || *o.ContentURL != "https://i" {
					t.Errorf("output = %+v", o)
				}
			},
		},
	})
}

func TestCreateAndWait(t *testing.T) {
	client, ts := newTestServer(t, testutil.Sequence(
		testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"id":"g","status":"pending"}`),
		func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Retry-After", "7")
			_, _ = io.WriteString(w, `{"id":"g","status":"generating"}`)
		},
		testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"id":"g","status":"completed","content_url":"https://c"}`),
	))
	g, err := client.Image.CreateAndWait(context.Background(), &ImageGenerationRequest{ModelID: "m", Prompt: "p"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if g.Status != GenerationStatusCompleted || g.ContentURL != "https://c" {
		t.Errorf("generation = %+v", g)
	}
	reqs := ts.All()
	if len(reqs) != 3 || reqs[0].Method != http.MethodPost || reqs[1].Path != "/v1/flows/image/g" || reqs[2].Method != http.MethodGet {
		t.Fatalf("requests = %+v", reqs)
	}
	// jitter is fixed at 0.5 → +10%.
	want := []time.Duration{3300 * time.Millisecond, 7700 * time.Millisecond}
	if len(ts.Sleeps()) != 2 || ts.Sleeps()[0] != want[0] || ts.Sleeps()[1] != want[1] {
		t.Errorf("sleeps = %v, want %v", ts.Sleeps(), want)
	}
}

func TestCreateAndWaitFailedIsReturned(t *testing.T) {
	client, _ := newTestServer(t, testutil.Sequence(
		testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"id":"g","status":"pending"}`),
		testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"id":"g","status":"failed","failure_reason":"timeout"}`),
	))
	g, err := client.Video.CreateAndWait(context.Background(), &VideoGenerationRequest{ModelID: "m"}, &WaitOptions{MinPollInterval: time.Second})
	if err != nil || g.Status != GenerationStatusFailed || g.FailureReason != "timeout" {
		t.Fatalf("g = %+v, err = %v", g, err)
	}
}

func TestCreateAndWaitTimeout(t *testing.T) {
	now := time.Unix(0, 0)
	client, _ := newTestServer(t, testutil.Sequence(
		testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"id":"r","template_id":"t","status":"pending","outputs":{}}`),
		testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"id":"r","template_id":"t","status":"generating","outputs":{}}`),
	), testutil.WithClock(func() time.Time { return now }), testutil.WithSleep(func(ctx context.Context, d time.Duration) error { now = now.Add(d); return nil }))
	_, err := client.Templates.Runs.CreateAndWait(context.Background(), "t", &TemplateRunCreateRequest{}, &WaitOptions{WaitTimeout: 10 * time.Second})
	var te *WaitTimeoutError
	if !errors.Is(err, ErrWaitTimeout) || !errors.As(err, &te) || te.ID != "r" || te.TemplateID != "t" || te.LastStatus != GenerationStatusGenerating {
		t.Fatalf("err = %v (%+v)", err, te)
	}
}

func TestCreateAndWaitTemplateAlreadyDone(t *testing.T) {
	client, ts := newTestServer(t, testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"id":"r","status":"completed","outputs":{}}`))
	r, err := client.Templates.Runs.CreateAndWait(context.Background(), "t", nil, nil)
	if err != nil || r.Status != GenerationStatusCompleted || len(ts.All()) != 1 {
		t.Fatalf("r = %+v, err = %v, requests = %d", r, err, len(ts.All()))
	}
}

func TestCreateAndWaitInvalidOptions(t *testing.T) {
	client, ts := newTestServer(t, testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{}`))
	for _, w := range []*WaitOptions{{MinPollInterval: time.Millisecond}, {WaitTimeout: -time.Second}} {
		if _, err := client.TextToSpeech.CreateAndWait(context.Background(), nil, w); !errors.Is(err, ErrInvalidWaitOptions) {
			t.Errorf("err = %v", err)
		}
	}
	if len(ts.All()) != 0 {
		t.Error("request sent despite invalid options")
	}
}

func newTestServer(t *testing.T, h http.HandlerFunc, opts ...option.RequestOption) (*Service, *testutil.Server) {
	t.Helper()
	ts, o := testutil.NewServer(t, h, opts...)
	return NewService(o...), ts
}
