package elevenlabs

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// WebhooksService groups the /v1/workspace/webhooks endpoints and webhook
// signature verification.
type WebhooksService struct{ cfg *config }

// WebhookAuthMethod is how ElevenLabs authenticates webhook deliveries.
type WebhookAuthMethod string

// WebhookAuthMethod values.
const (
	WebhookAuthMethodHMAC   WebhookAuthMethod = "hmac"
	WebhookAuthMethodOAuth2 WebhookAuthMethod = "oauth2"
	WebhookAuthMethodMTLS   WebhookAuthMethod = "mtls"
)

// WebhookEventType is an event a workspace webhook can subscribe to.
type WebhookEventType string

// WebhookEventType values.
const (
	WebhookEventVoiceLibraryRemovalNotice WebhookEventType = "voice_library_removal_notice"
	WebhookEventSpeechToText              WebhookEventType = "speech_to_text"
	WebhookEventAgentQA                   WebhookEventType = "agent_qa"
	WebhookEventFlows                     WebhookEventType = "flows"
)

// WebhookUsage describes a feature that uses a webhook.
type WebhookUsage struct {
	UsageType string `json:"usage_type"`
}

// WorkspaceWebhook is a webhook configured in the workspace.
type WorkspaceWebhook struct {
	Name                       string             `json:"name"`
	WebhookID                  string             `json:"webhook_id"`
	WebhookURL                 string             `json:"webhook_url"`
	IsDisabled                 bool               `json:"is_disabled"`
	IsAutoDisabled             bool               `json:"is_auto_disabled"`
	CreatedAtUnix              int64              `json:"created_at_unix"`
	AuthType                   WebhookAuthMethod  `json:"auth_type"`
	Usage                      []WebhookUsage     `json:"usage,omitempty"`
	Events                     []WebhookEventType `json:"events,omitempty"`
	MostRecentFailureErrorCode *int               `json:"most_recent_failure_error_code,omitempty"`
	MostRecentFailureTimestamp *int64             `json:"most_recent_failure_timestamp,omitempty"`
}

// WebhooksListRequest holds the optional parameters of WebhooksService.List.
type WebhooksListRequest struct {
	// IncludeUsages includes active usages of each webhook (admins only).
	IncludeUsages *bool `query:"include_usages"`
}

// WorkspaceWebhookList is the response of WebhooksService.List.
type WorkspaceWebhookList struct {
	Webhooks []WorkspaceWebhook `json:"webhooks"`
}

// List returns the workspace webhooks. req may be nil.
func (s *WebhooksService) List(ctx context.Context, req *WebhooksListRequest, opts ...RequestOption) (*WorkspaceWebhookList, error) {
	return doValue[WorkspaceWebhookList](ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/workspace/webhooks", query: req, opts: opts})
}

// WebhookHMACSettings configures an HMAC-authenticated webhook.
type WebhookHMACSettings struct {
	// Name identifies the webhook. Required.
	Name string `json:"name"`
	// WebhookURL receives deliveries. Required.
	WebhookURL string `json:"webhook_url"`
	// RequestHeaders are extra headers sent with each delivery.
	RequestHeaders map[string]string `json:"request_headers,omitempty"`
}

// MarshalJSON adds the "hmac" auth_type discriminator.
func (w WebhookHMACSettings) MarshalJSON() ([]byte, error) {
	type plain WebhookHMACSettings
	return json.Marshal(struct {
		AuthType WebhookAuthMethod `json:"auth_type"`
		plain
	}{WebhookAuthMethodHMAC, plain(w)})
}

// WebhookCreateRequest is the body of WebhooksService.Create.
type WebhookCreateRequest struct {
	Settings WebhookHMACSettings `json:"settings"`
}

// WebhookCreateResponse is the response of WebhooksService.Create.
type WebhookCreateResponse struct {
	WebhookID string `json:"webhook_id"`
	// WebhookSecret is the HMAC secret used to sign deliveries.
	WebhookSecret *string `json:"webhook_secret,omitempty"`
}

// Create creates a workspace webhook.
func (s *WebhooksService) Create(ctx context.Context, req *WebhookCreateRequest, opts ...RequestOption) (*WebhookCreateResponse, error) {
	return doValue[WebhookCreateResponse](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/workspace/webhooks", body: orEmpty(req), opts: opts})
}

// Delete deletes a workspace webhook.
func (s *WebhooksService) Delete(ctx context.Context, webhookID string, opts ...RequestOption) (*StatusResponse, error) {
	return doValue[StatusResponse](ctx, s.cfg, &request{method: http.MethodDelete, path: pathf("/v1/workspace/webhooks/%s", webhookID), opts: opts})
}

// WebhookUpdateRequest is the body of WebhooksService.Update.
type WebhookUpdateRequest struct {
	// IsDisabled disables or enables the webhook. Required.
	IsDisabled bool `json:"is_disabled"`
	// Name renames the webhook. Required.
	Name string `json:"name"`
	// RetryEnabled toggles delivery retries.
	RetryEnabled *bool `json:"retry_enabled,omitempty"`
	// RequestHeaders replaces the extra delivery headers.
	RequestHeaders map[string]string `json:"request_headers,omitempty"`
	// Events replaces the subscribed events.
	Events []WebhookEventType `json:"events,omitempty"`
}

// Update edits a workspace webhook.
func (s *WebhooksService) Update(ctx context.Context, webhookID string, req *WebhookUpdateRequest, opts ...RequestOption) (*StatusResponse, error) {
	return doValue[StatusResponse](ctx, s.cfg, &request{method: http.MethodPatch, path: pathf("/v1/workspace/webhooks/%s", webhookID), body: orEmpty(req), opts: opts})
}

// Webhook signature verification errors.
var (
	ErrWebhookMissingSignature = errors.New("elevenlabs: missing webhook signature header")
	ErrWebhookMissingSecret    = errors.New("elevenlabs: webhook secret not configured")
	ErrWebhookInvalidHeader    = errors.New("elevenlabs: no signature hash found with expected scheme v0")
	ErrWebhookTimestamp        = errors.New("elevenlabs: webhook timestamp outside the tolerance zone")
	ErrWebhookSignature        = errors.New("elevenlabs: webhook signature does not match payload")
)

const (
	// WebhookSignatureHeader is the HTTP header carrying the delivery signature.
	WebhookSignatureHeader = "ElevenLabs-Signature"
	// DefaultWebhookTolerance is how old a delivery timestamp may be, matching the official SDK.
	DefaultWebhookTolerance = 30 * time.Minute
)

// WebhookEvent is a verified webhook delivery.
type WebhookEvent struct {
	// Type is the event type, e.g. "post_call_transcription".
	Type string `json:"type"`
	// EventTimestamp is the unix time the event was emitted.
	EventTimestamp int64 `json:"event_timestamp"`
	// Data is the event payload; decode it into the type matching Type.
	Data json.RawMessage `json:"data"`
	// Raw is the complete verified payload.
	Raw json.RawMessage `json:"-"`
}

// WebhookVerifyOption customizes webhook signature verification.
type WebhookVerifyOption func(*webhookVerifyConfig)

type webhookVerifyConfig struct {
	tolerance time.Duration
	now       func() time.Time
}

// WithWebhookTolerance sets how old a delivery timestamp may be.
func WithWebhookTolerance(d time.Duration) WebhookVerifyOption {
	return func(c *webhookVerifyConfig) { c.tolerance = d }
}

// WithWebhookClock sets the clock used to check the delivery timestamp.
func WithWebhookClock(now func() time.Time) WebhookVerifyOption {
	return func(c *webhookVerifyConfig) {
		if now != nil {
			c.now = now
		}
	}
}

// VerifyWebhookSignature checks that payload was signed by ElevenLabs.
// sigHeader is the value of the ElevenLabs-Signature header
// ("t=<unix>,v0=<hex hmac-sha256>") and payload must be the raw request body.
func VerifyWebhookSignature(payload []byte, sigHeader, secret string, opts ...WebhookVerifyOption) error {
	cfg := webhookVerifyConfig{tolerance: DefaultWebhookTolerance, now: time.Now}
	for _, opt := range opts {
		opt(&cfg)
	}
	if sigHeader == "" {
		return ErrWebhookMissingSignature
	}
	if secret == "" {
		return ErrWebhookMissingSecret
	}
	var timestamp, signature string
	for part := range strings.SplitSeq(sigHeader, ",") {
		switch {
		case timestamp == "" && strings.HasPrefix(part, "t="):
			timestamp = part[2:]
		case signature == "" && strings.HasPrefix(part, "v0="):
			signature = part
		}
	}
	if timestamp == "" || signature == "" {
		return ErrWebhookInvalidHeader
	}
	secs, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("%w: invalid timestamp %q", ErrWebhookInvalidHeader, timestamp)
	}
	if time.Unix(secs, 0).Before(cfg.now().Add(-cfg.tolerance)) {
		return ErrWebhookTimestamp
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(payload)
	expected := "v0=" + hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(signature), []byte(expected)) {
		return ErrWebhookSignature
	}
	return nil
}

// ConstructWebhookEvent verifies the signature of a webhook delivery and
// decodes it. It is the equivalent of the official SDK's
// webhooks.constructEvent.
func ConstructWebhookEvent(payload []byte, sigHeader, secret string, opts ...WebhookVerifyOption) (*WebhookEvent, error) {
	if err := VerifyWebhookSignature(payload, sigHeader, secret, opts...); err != nil {
		return nil, err
	}
	event := &WebhookEvent{}
	if err := json.Unmarshal(payload, event); err != nil {
		return nil, fmt.Errorf("elevenlabs: decode webhook event: %w", err)
	}
	event.Raw = append(json.RawMessage(nil), payload...)
	return event, nil
}

// ConstructEvent is ConstructWebhookEvent, offered on the service for parity
// with the official SDK's client.webhooks.constructEvent.
func (s *WebhooksService) ConstructEvent(payload []byte, sigHeader, secret string, opts ...WebhookVerifyOption) (*WebhookEvent, error) {
	return ConstructWebhookEvent(payload, sigHeader, secret, opts...)
}
