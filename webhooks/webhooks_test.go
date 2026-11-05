package webhooks

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, []testutil.Endpoint[*Service]{
		{
			Name: "webhooks.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.List(ctx, &ListRequest{IncludeUsages: types.Ptr(true)})
			},
			Method: http.MethodGet, Path: "/v1/workspace/webhooks", Query: "include_usages=true",
			RespBody: `{"webhooks":[{"name":"n","webhook_id":"w","webhook_url":"u","is_disabled":false,"is_auto_disabled":false,"created_at_unix":1,"auth_type":"hmac","events":["flows"]}]}`,
			Check: func(t *testing.T, r any) {
				w := r.(*ListResponse).Webhooks[0]
				if w.WebhookID != "w" || w.AuthType != AuthMethodHMAC || w.Events[0] != EventFlows {
					t.Errorf("webhook = %+v", w)
				}
			},
		},
		{
			Name: "webhooks.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Create(ctx, &CreateRequest{Settings: HMACSettings{Name: "n", WebhookURL: "https://x", RequestHeaders: map[string]string{"a": "b"}}})
			},
			Method: http.MethodPost, Path: "/v1/workspace/webhooks",
			Body:     `{"settings":{"auth_type":"hmac","name":"n","webhook_url":"https://x","request_headers":{"a":"b"}}}`,
			RespBody: `{"webhook_id":"w","webhook_secret":"s"}`,
			Check: func(t *testing.T, r any) {
				if w := r.(*CreateResponse); w.WebhookID != "w" || *w.WebhookSecret != "s" {
					t.Errorf("resp = %+v", w)
				}
			},
		},
		{
			Name:   "webhooks.delete",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Delete(ctx, "w/1") },
			Method: http.MethodDelete, Path: "/v1/workspace/webhooks/w%2F1", RespBody: `{"status":"ok"}`,
		},
		{
			Name: "webhooks.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Update(ctx, "w", &UpdateRequest{IsDisabled: true, Name: "n", RetryEnabled: types.Ptr(false), Events: []EventType{EventSpeechToText}})
			},
			Method: http.MethodPatch, Path: "/v1/workspace/webhooks/w",
			Body: `{"is_disabled":true,"name":"n","retry_enabled":false,"events":["speech_to_text"]}`, RespBody: `{"status":"ok"}`,
		},
	})
}

func signWebhook(secret string, ts int64, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strconv.FormatInt(ts, 10) + "." + body))
	return "t=" + strconv.FormatInt(ts, 10) + ",v0=" + hex.EncodeToString(mac.Sum(nil))
}

func TestConstructWebhookEvent(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	clock := WithClock(func() time.Time { return now })
	body := `{"type":"post_call_transcription","event_timestamp":5,"data":{"agent_id":"a"}}`
	tests := []struct {
		name   string
		header string
		secret string
		want   error
	}{
		{"valid", signWebhook("sec", now.Unix(), body), "sec", nil},
		{"valid within tolerance", signWebhook("sec", now.Add(-29*time.Minute).Unix(), body), "sec", nil},
		{"missing header", "", "sec", ErrMissingSignature},
		{"missing secret", signWebhook("sec", now.Unix(), body), "", ErrMissingSecret},
		{"no scheme", "t=1", "sec", ErrInvalidHeader},
		{"bad timestamp", "t=abc,v0=00", "sec", ErrInvalidHeader},
		{"too old", signWebhook("sec", now.Add(-31*time.Minute).Unix(), body), "sec", ErrTimestamp},
		{"wrong secret", signWebhook("other", now.Unix(), body), "sec", ErrSignature},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev, err := ConstructEvent([]byte(body), tt.header, tt.secret, clock)
			if !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
			if tt.want == nil && (ev.Type != "post_call_transcription" || ev.EventTimestamp != 5 || string(ev.Data) != `{"agent_id":"a"}` || string(ev.Raw) != body) {
				t.Errorf("event = %+v", ev)
			}
		})
	}
}

func TestConstructEventTamperedBody(t *testing.T) {
	c := NewService(option.WithAPIKey("k"))
	ts := time.Now().Unix()
	header := signWebhook("sec", ts, `{"type":"a"}`)
	if _, err := c.ConstructEvent([]byte(`{"type":"b"}`), header, "sec"); !errors.Is(err, ErrSignature) {
		t.Fatalf("err = %v", err)
	}
	if _, err := c.ConstructEvent([]byte(`{"type":"a"}`), header, "sec", WithTolerance(time.Hour)); err != nil {
		t.Fatalf("err = %v", err)
	}
}
