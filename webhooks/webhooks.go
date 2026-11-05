package webhooks

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

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

// Service groups the /v1/workspace/webhooks endpoints and webhook
// signature verification.
type Service struct{ cfg *requestconfig.Config }

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	return &Service{cfg: requestconfig.Build(opts...)}
}

// AuthMethod is how ElevenLabs authenticates webhook deliveries.
type AuthMethod string

// AuthMethod values.
const (
	AuthMethodHMAC   AuthMethod = "hmac"
	AuthMethodOAuth2 AuthMethod = "oauth2"
	AuthMethodMTLS   AuthMethod = "mtls"
)

// EventType is an event a workspace webhook can subscribe to.
type EventType string

// EventType values.
const (
	EventVoiceLibraryRemovalNotice EventType = "voice_library_removal_notice"
	EventSpeechToText              EventType = "speech_to_text"
	EventAgentQA                   EventType = "agent_qa"
	EventFlows                     EventType = "flows"
)

// Usage describes a feature that uses a webhook.
type Usage struct {
	UsageType string `json:"usage_type"`
}

// Webhook is a webhook configured in the workspace.
type Webhook struct {
	Name                       string      `json:"name"`
	WebhookID                  string      `json:"webhook_id"`
	WebhookURL                 string      `json:"webhook_url"`
	IsDisabled                 bool        `json:"is_disabled"`
	IsAutoDisabled             bool        `json:"is_auto_disabled"`
	CreatedAtUnix              int64       `json:"created_at_unix"`
	AuthType                   AuthMethod  `json:"auth_type"`
	Usage                      []Usage     `json:"usage,omitempty"`
	Events                     []EventType `json:"events,omitempty"`
	MostRecentFailureErrorCode *int        `json:"most_recent_failure_error_code,omitempty"`
	MostRecentFailureTimestamp *int64      `json:"most_recent_failure_timestamp,omitempty"`
}

// ListRequest holds the optional parameters of Service.List.
type ListRequest struct {
	// IncludeUsages includes active usages of each webhook (admins only).
	IncludeUsages *bool `query:"include_usages"`
}

// ListResponse is the response of Service.List.
type ListResponse struct {
	Webhooks []Webhook `json:"webhooks"`
}

// List returns the workspace webhooks. req may be nil.
func (s *Service) List(ctx context.Context, req *ListRequest, opts ...option.RequestOption) (*ListResponse, error) {
	return transport.DoValue[ListResponse](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/workspace/webhooks", Query: req, Opts: opts})
}

// HMACSettings configures an HMAC-authenticated webhook.
type HMACSettings struct {
	// Name identifies the webhook. Required.
	Name string `json:"name"`
	// URL receives deliveries. Required.
	WebhookURL string `json:"webhook_url"`
	// RequestHeaders are extra headers sent with each delivery.
	RequestHeaders map[string]string `json:"request_headers,omitempty"`
}

// MarshalJSON adds the "hmac" auth_type discriminator.
func (w HMACSettings) MarshalJSON() ([]byte, error) {
	type plain HMACSettings
	return json.Marshal(struct {
		AuthType AuthMethod `json:"auth_type"`
		plain
	}{AuthMethodHMAC, plain(w)})
}

// CreateRequest is the body of Service.Create.
type CreateRequest struct {
	Settings HMACSettings `json:"settings"`
}

// CreateResponse is the response of Service.Create.
type CreateResponse struct {
	WebhookID string `json:"webhook_id"`
	// Secret is the HMAC secret used to sign deliveries.
	WebhookSecret *string `json:"webhook_secret,omitempty"`
}

// Create creates a workspace webhook.
func (s *Service) Create(ctx context.Context, req *CreateRequest, opts ...option.RequestOption) (*CreateResponse, error) {
	return transport.DoValue[CreateResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/workspace/webhooks", Body: transport.OrEmpty(req), Opts: opts})
}

// Delete deletes a workspace webhook.
func (s *Service) Delete(ctx context.Context, webhookID string, opts ...option.RequestOption) (*types.StatusResponse, error) {
	return transport.DoValue[types.StatusResponse](ctx, s.cfg, &transport.Request{Method: http.MethodDelete, Path: transport.Pathf("/v1/workspace/webhooks/%s", webhookID), Opts: opts})
}

// UpdateRequest is the body of Service.Update.
type UpdateRequest struct {
	// IsDisabled disables or enables the webhook. Required.
	IsDisabled bool `json:"is_disabled"`
	// Name renames the webhook. Required.
	Name string `json:"name"`
	// RetryEnabled toggles delivery retries.
	RetryEnabled *bool `json:"retry_enabled,omitempty"`
	// RequestHeaders replaces the extra delivery headers.
	RequestHeaders map[string]string `json:"request_headers,omitempty"`
	// Events replaces the subscribed events.
	Events []EventType `json:"events,omitempty"`
}

// Update edits a workspace webhook.
func (s *Service) Update(ctx context.Context, webhookID string, req *UpdateRequest, opts ...option.RequestOption) (*types.StatusResponse, error) {
	return transport.DoValue[types.StatusResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPatch, Path: transport.Pathf("/v1/workspace/webhooks/%s", webhookID), Body: transport.OrEmpty(req), Opts: opts})
}

// Webhook signature verification errors.
var (
	ErrMissingSignature = errors.New("elevenlabs: missing webhook signature header")
	ErrMissingSecret    = errors.New("elevenlabs: webhook secret not configured")
	ErrInvalidHeader    = errors.New("elevenlabs: no signature hash found with expected scheme v0")
	ErrTimestamp        = errors.New("elevenlabs: webhook timestamp outside the tolerance zone")
	ErrSignature        = errors.New("elevenlabs: webhook signature does not match payload")
)

const (
	// SignatureHeader is the HTTP header carrying the delivery signature.
	SignatureHeader = "ElevenLabs-Signature"
	// DefaultWebhookTolerance is how old a delivery timestamp may be, matching the official SDK.
	DefaultWebhookTolerance = 30 * time.Minute
)

// Event is a verified webhook delivery.
type Event struct {
	// Type is the event type, e.g. "post_call_transcription".
	Type string `json:"type"`
	// EventTimestamp is the unix time the event was emitted.
	EventTimestamp int64 `json:"event_timestamp"`
	// Data is the event payload; decode it into the type matching Type.
	Data json.RawMessage `json:"data"`
	// Raw is the complete verified payload.
	Raw json.RawMessage `json:"-"`
}

// VerifyOption customizes webhook signature verification.
type VerifyOption func(*verifyConfig)

type verifyConfig struct {
	tolerance time.Duration
	now       func() time.Time
}

// WithTolerance sets how old a delivery timestamp may be.
func WithTolerance(d time.Duration) VerifyOption {
	return func(c *verifyConfig) { c.tolerance = d }
}

// WithClock sets the clock used to check the delivery timestamp.
func WithClock(now func() time.Time) VerifyOption {
	return func(c *verifyConfig) {
		if now != nil {
			c.now = now
		}
	}
}

// VerifySignature checks that payload was signed by ElevenLabs.
// sigHeader is the value of the ElevenLabs-Signature header
// ("t=<unix>,v0=<hex hmac-sha256>") and payload must be the raw request body.
func VerifySignature(payload []byte, sigHeader, secret string, opts ...VerifyOption) error {
	cfg := verifyConfig{tolerance: DefaultWebhookTolerance, now: time.Now}
	for _, opt := range opts {
		opt(&cfg)
	}
	if sigHeader == "" {
		return ErrMissingSignature
	}
	if secret == "" {
		return ErrMissingSecret
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
		return ErrInvalidHeader
	}
	secs, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return fmt.Errorf("%w: invalid timestamp %q", ErrInvalidHeader, timestamp)
	}
	if time.Unix(secs, 0).Before(cfg.now().Add(-cfg.tolerance)) {
		return ErrTimestamp
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp))
	mac.Write([]byte("."))
	mac.Write(payload)
	expected := "v0=" + hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(signature), []byte(expected)) {
		return ErrSignature
	}
	return nil
}

// ConstructEvent verifies the signature of a webhook delivery and
// decodes it. It is the equivalent of the official SDK's
// webhooks.constructEvent.
func ConstructEvent(payload []byte, sigHeader, secret string, opts ...VerifyOption) (*Event, error) {
	if err := VerifySignature(payload, sigHeader, secret, opts...); err != nil {
		return nil, err
	}
	event := &Event{}
	if err := json.Unmarshal(payload, event); err != nil {
		return nil, fmt.Errorf("elevenlabs: decode webhook event: %w", err)
	}
	event.Raw = append(json.RawMessage(nil), payload...)
	return event, nil
}

// ConstructEvent is ConstructEvent, offered on the service for parity
// with the official SDK's client.webhooks.constructEvent.
func (s *Service) ConstructEvent(payload []byte, sigHeader, secret string, opts ...VerifyOption) (*Event, error) {
	return ConstructEvent(payload, sigHeader, secret, opts...)
}
