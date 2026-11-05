package elevenlabs

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func signWebhook(secret string, ts int64, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strconv.FormatInt(ts, 10) + "." + body))
	return "t=" + strconv.FormatInt(ts, 10) + ",v0=" + hex.EncodeToString(mac.Sum(nil))
}

func TestConstructWebhookEvent(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	clock := WithWebhookClock(func() time.Time { return now })
	body := `{"type":"post_call_transcription","event_timestamp":5,"data":{"agent_id":"a"}}`
	tests := []struct {
		name   string
		header string
		secret string
		want   error
	}{
		{"valid", signWebhook("sec", now.Unix(), body), "sec", nil},
		{"valid within tolerance", signWebhook("sec", now.Add(-29*time.Minute).Unix(), body), "sec", nil},
		{"missing header", "", "sec", ErrWebhookMissingSignature},
		{"missing secret", signWebhook("sec", now.Unix(), body), "", ErrWebhookMissingSecret},
		{"no scheme", "t=1", "sec", ErrWebhookInvalidHeader},
		{"bad timestamp", "t=abc,v0=00", "sec", ErrWebhookInvalidHeader},
		{"too old", signWebhook("sec", now.Add(-31*time.Minute).Unix(), body), "sec", ErrWebhookTimestamp},
		{"wrong secret", signWebhook("other", now.Unix(), body), "sec", ErrWebhookSignature},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev, err := ConstructWebhookEvent([]byte(body), tt.header, tt.secret, clock)
			if !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
			if tt.want == nil && (ev.Type != "post_call_transcription" || ev.EventTimestamp != 5 || string(ev.Data) != `{"agent_id":"a"}` || string(ev.Raw) != body) {
				t.Errorf("event = %+v", ev)
			}
		})
	}
}

func TestWebhooksConstructEventTamperedBody(t *testing.T) {
	c := NewClient(WithAPIKey("k"))
	ts := time.Now().Unix()
	header := signWebhook("sec", ts, `{"type":"a"}`)
	if _, err := c.Webhooks.ConstructEvent([]byte(`{"type":"b"}`), header, "sec"); !errors.Is(err, ErrWebhookSignature) {
		t.Fatalf("err = %v", err)
	}
	if _, err := c.Webhooks.ConstructEvent([]byte(`{"type":"a"}`), header, "sec", WithWebhookTolerance(time.Hour)); err != nil {
		t.Fatalf("err = %v", err)
	}
}

func TestFlowsCreateAndWait(t *testing.T) {
	client, ts := newTestServer(t, sequence(
		respond(http.StatusOK, contentTypeJSON, `{"id":"g","status":"pending"}`),
		func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Retry-After", "7")
			_, _ = io.WriteString(w, `{"id":"g","status":"generating"}`)
		},
		respond(http.StatusOK, contentTypeJSON, `{"id":"g","status":"completed","content_url":"https://c"}`),
	))
	g, err := client.Flows.Image.CreateAndWait(context.Background(), &ImageGenerationRequest{ModelID: "m", Prompt: "p"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if g.Status != GenerationStatusCompleted || g.ContentURL != "https://c" {
		t.Errorf("generation = %+v", g)
	}
	reqs := ts.all()
	if len(reqs) != 3 || reqs[0].Method != http.MethodPost || reqs[1].Path != "/v1/flows/image/g" || reqs[2].Method != http.MethodGet {
		t.Fatalf("requests = %+v", reqs)
	}
	// jitter is fixed at 0.5 → +10%.
	want := []time.Duration{3300 * time.Millisecond, 7700 * time.Millisecond}
	if len(ts.Sleeps()) != 2 || ts.Sleeps()[0] != want[0] || ts.Sleeps()[1] != want[1] {
		t.Errorf("sleeps = %v, want %v", ts.Sleeps(), want)
	}
}

func TestFlowsCreateAndWaitFailedIsReturned(t *testing.T) {
	client, _ := newTestServer(t, sequence(
		respond(http.StatusOK, contentTypeJSON, `{"id":"g","status":"pending"}`),
		respond(http.StatusOK, contentTypeJSON, `{"id":"g","status":"failed","failure_reason":"timeout"}`),
	))
	g, err := client.Flows.Video.CreateAndWait(context.Background(), &VideoGenerationRequest{ModelID: "m"}, &FlowsWaitOptions{MinPollInterval: time.Second})
	if err != nil || g.Status != GenerationStatusFailed || g.FailureReason != "timeout" {
		t.Fatalf("g = %+v, err = %v", g, err)
	}
}

func TestFlowsCreateAndWaitTimeout(t *testing.T) {
	client, _ := newTestServer(t, sequence(
		respond(http.StatusOK, contentTypeJSON, `{"id":"r","template_id":"t","status":"pending","outputs":{}}`),
		respond(http.StatusOK, contentTypeJSON, `{"id":"r","template_id":"t","status":"generating","outputs":{}}`),
	))
	now := time.Unix(0, 0)
	client.cfg.Now = func() time.Time { return now }
	client.cfg.Sleep = func(ctx context.Context, d time.Duration) error { now = now.Add(d); return nil }
	_, err := client.Flows.Templates.Runs.CreateAndWait(context.Background(), "t", &TemplateRunCreateRequest{}, &FlowsWaitOptions{WaitTimeout: 10 * time.Second})
	var te *FlowsWaitTimeoutError
	if !errors.Is(err, ErrFlowsWaitTimeout) || !errors.As(err, &te) || te.ID != "r" || te.TemplateID != "t" || te.LastStatus != GenerationStatusGenerating {
		t.Fatalf("err = %v (%+v)", err, te)
	}
}

func TestFlowsCreateAndWaitTemplateAlreadyDone(t *testing.T) {
	client, ts := newTestServer(t, respond(http.StatusOK, contentTypeJSON, `{"id":"r","status":"completed","outputs":{}}`))
	r, err := client.Flows.Templates.Runs.CreateAndWait(context.Background(), "t", nil, nil)
	if err != nil || r.Status != GenerationStatusCompleted || len(ts.all()) != 1 {
		t.Fatalf("r = %+v, err = %v, requests = %d", r, err, len(ts.all()))
	}
}

func TestFlowsCreateAndWaitInvalidOptions(t *testing.T) {
	client, ts := newTestServer(t, respond(http.StatusOK, contentTypeJSON, `{}`))
	for _, w := range []*FlowsWaitOptions{{MinPollInterval: time.Millisecond}, {WaitTimeout: -time.Second}} {
		if _, err := client.Flows.TextToSpeech.CreateAndWait(context.Background(), nil, w); !errors.Is(err, ErrInvalidWaitOptions) {
			t.Errorf("err = %v", err)
		}
	}
	if len(ts.all()) != 0 {
		t.Error("request sent despite invalid options")
	}
}

func TestPlatformListAll(t *testing.T) {
	client, ts := newTestServer(t, sequence(
		respond(http.StatusOK, contentTypeJSON, `{"speech_engines":[{"speech_engine_id":"a"}],"next_cursor":"c","has_more":true}`),
		respond(http.StatusOK, contentTypeJSON, `{"speech_engines":[{"speech_engine_id":"b"}],"has_more":false}`),
	))
	var ids []string
	for s, err := range client.SpeechEngine.ListAll(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, s.SpeechEngineID)
	}
	if len(ids) != 2 || ids[1] != "b" || ts.last(t).RawQuery != "cursor=c" {
		t.Errorf("ids = %v, query = %q", ids, ts.last(t).RawQuery)
	}
}

func TestClearableAndPermissionsJSON(t *testing.T) {
	var c Clearable[int]
	if err := c.UnmarshalJSON([]byte(`"clear"`)); err != nil || !c.IsClear() {
		t.Errorf("clear = %+v, %v", c, err)
	}
	if err := c.UnmarshalJSON([]byte(`4`)); err != nil {
		t.Fatal(err)
	}
	if v, ok := c.Value(); !ok || v != 4 {
		t.Errorf("value = %d, %v", v, ok)
	}
	var p APIKeyPermissions
	if err := p.UnmarshalJSON([]byte(`["voices_read"]`)); err != nil || p.All || p.Scopes[0] != PermissionVoicesRead {
		t.Errorf("perms = %+v, %v", p, err)
	}
	if b, _ := Permissions().MarshalJSON(); string(b) != "[]" {
		t.Errorf("empty perms = %s", b)
	}
}
