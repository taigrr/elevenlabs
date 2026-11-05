package speechengine

import (
	"context"
	"encoding/json"
	"iter"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/stream"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

// Service groups the /v1/speech-engine endpoints, which manage
// speech engines: voice pipelines (ASR, turn taking and TTS) that delegate
// response generation to your own WebSocket server.
type Service struct{ cfg *requestconfig.Config }

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	return &Service{cfg: requestconfig.Build(opts...)}
}

// Config points a speech engine at your WebSocket server.
type Config struct {
	// WSURL is the WebSocket URL ElevenLabs connects to. Required.
	WSURL string `json:"ws_url"`
	// RequestHeaders are sent on connect. Values are strings or secret /
	// dynamic variable locators.
	RequestHeaders map[string]any `json:"request_headers,omitempty"`
}

// AgentCallLimits caps concurrent and daily calls.
type AgentCallLimits struct {
	AgentConcurrencyLimit *int  `json:"agent_concurrency_limit,omitempty"`
	DailyLimit            *int  `json:"daily_limit,omitempty"`
	BurstingEnabled       *bool `json:"bursting_enabled,omitempty"`
}

// Overrides lists which settings clients may override.
type Overrides struct {
	FirstMessage *bool `json:"first_message,omitempty"`
}

// Summary is a speech engine in a listing.
type Summary struct {
	SpeechEngineID    string                   `json:"speech_engine_id"`
	Name              string                   `json:"name"`
	VoiceID           string                   `json:"voice_id"`
	CreatedAtUnixSecs int64                    `json:"created_at_unix_secs"`
	UpdatedAtUnixSecs int64                    `json:"updated_at_unix_secs"`
	Tags              []string                 `json:"tags"`
	AccessInfo        types.ResourceAccessInfo `json:"access_info"`
}

// Engine is a speech engine. The ASR, TTS, Turn, VAD, Conversation,
// Privacy and Metadata objects share their schema with conversational AI
// agents and are kept as raw JSON.
type Engine struct {
	SpeechEngineID        string                    `json:"speech_engine_id"`
	Name                  string                    `json:"name"`
	SpeechEngine          Config                    `json:"speech_engine"`
	ASR                   json.RawMessage           `json:"asr"`
	TTS                   json.RawMessage           `json:"tts"`
	Turn                  json.RawMessage           `json:"turn"`
	VAD                   json.RawMessage           `json:"vad"`
	Conversation          json.RawMessage           `json:"conversation"`
	Privacy               json.RawMessage           `json:"privacy"`
	CallLimits            AgentCallLimits           `json:"call_limits"`
	Language              string                    `json:"language"`
	CascadeTimeoutSeconds float64                   `json:"cascade_timeout_seconds"`
	Tags                  []string                  `json:"tags"`
	Overrides             Overrides                 `json:"overrides"`
	Metadata              json.RawMessage           `json:"metadata"`
	AccessInfo            *types.ResourceAccessInfo `json:"access_info,omitempty"`
}

// SortBy is the field speech engines are sorted by.
type SortBy string

// SortBy values.
const (
	SortByName        SortBy = "name"
	SortByCreatedAt   SortBy = "created_at"
	SortByUpdatedAt   SortBy = "updated_at"
	SortByCallCount7d SortBy = "call_count_7d"
)

// ListRequest holds the optional parameters of Service.List.
type ListRequest struct {
	PageSize      *int                `query:"page_size"`
	Search        string              `query:"search"`
	SortDirection types.SortDirection `query:"sort_direction"`
	SortBy        SortBy              `query:"sort_by"`
	Cursor        string              `query:"cursor"`
}

// Page is a page of speech engines.
type Page struct {
	SpeechEngines []Summary `json:"speech_engines"`
	NextCursor    *string   `json:"next_cursor,omitempty"`
	HasMore       bool      `json:"has_more"`
}

// List returns one page of speech engines. req may be nil.
func (s *Service) List(ctx context.Context, req *ListRequest, opts ...option.RequestOption) (*Page, error) {
	return transport.DoValue[Page](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/speech-engine", Query: req, Opts: opts})
}

// ListAll iterates over every speech engine, fetching pages on demand.
func (s *Service) ListAll(ctx context.Context, req *ListRequest, opts ...option.RequestOption) iter.Seq2[Summary, error] {
	params := *transport.OrEmpty(req)
	return stream.Paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (stream.Page[Summary], error) {
		params.Cursor = cursor
		resp, err := s.List(ctx, &params, opts...)
		if err != nil {
			return stream.Page[Summary]{}, err
		}
		return stream.Page[Summary]{Items: resp.SpeechEngines, Next: transport.Deref(resp.NextCursor), HasMore: resp.HasMore}, nil
	})
}

// CreateRequest is the body of Service.Create.
type CreateRequest struct {
	Name string `json:"name,omitempty"`
	// Engine is the WebSocket server configuration. Required.
	SpeechEngine          Config           `json:"speech_engine"`
	ASR                   any              `json:"asr,omitempty"`
	TTS                   any              `json:"tts,omitempty"`
	Turn                  any              `json:"turn,omitempty"`
	VAD                   any              `json:"vad,omitempty"`
	Conversation          any              `json:"conversation,omitempty"`
	Privacy               any              `json:"privacy,omitempty"`
	CallLimits            *AgentCallLimits `json:"call_limits,omitempty"`
	Language              string           `json:"language,omitempty"`
	CascadeTimeoutSeconds *float64         `json:"cascade_timeout_seconds,omitempty"`
	Tags                  []string         `json:"tags,omitempty"`
	Overrides             *Overrides       `json:"overrides,omitempty"`
}

// UpdateRequest is the body of Service.Update. Nil
// fields are left unchanged.
type UpdateRequest struct {
	Name                  string           `json:"name,omitempty"`
	SpeechEngine          *Config          `json:"speech_engine,omitempty"`
	ASR                   any              `json:"asr,omitempty"`
	TTS                   any              `json:"tts,omitempty"`
	Turn                  any              `json:"turn,omitempty"`
	VAD                   any              `json:"vad,omitempty"`
	Conversation          any              `json:"conversation,omitempty"`
	Privacy               any              `json:"privacy,omitempty"`
	CallLimits            *AgentCallLimits `json:"call_limits,omitempty"`
	Language              string           `json:"language,omitempty"`
	CascadeTimeoutSeconds *float64         `json:"cascade_timeout_seconds,omitempty"`
	Tags                  []string         `json:"tags,omitempty"`
	Overrides             *Overrides       `json:"overrides,omitempty"`
}

// Create creates a speech engine.
func (s *Service) Create(ctx context.Context, req *CreateRequest, opts ...option.RequestOption) (*Engine, error) {
	return transport.DoValue[Engine](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/speech-engine", Body: transport.OrEmpty(req), Opts: opts})
}

// Get returns a speech engine.
func (s *Service) Get(ctx context.Context, speechEngineID string, opts ...option.RequestOption) (*Engine, error) {
	return transport.DoValue[Engine](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/speech-engine/%s", speechEngineID), Opts: opts})
}

// Delete deletes a speech engine.
func (s *Service) Delete(ctx context.Context, speechEngineID string, opts ...option.RequestOption) error {
	return transport.DoJSON(ctx, s.cfg, &transport.Request{Method: http.MethodDelete, Path: transport.Pathf("/v1/speech-engine/%s", speechEngineID), Opts: opts}, nil)
}

// Update edits a speech engine. req may be nil.
func (s *Service) Update(ctx context.Context, speechEngineID string, req *UpdateRequest, opts ...option.RequestOption) (*Engine, error) {
	return transport.DoValue[Engine](ctx, s.cfg, &transport.Request{Method: http.MethodPatch, Path: transport.Pathf("/v1/speech-engine/%s", speechEngineID), Body: transport.OrEmpty(req), Opts: opts})
}

// DuplicateRequest is the body of Service.Duplicate.
type DuplicateRequest struct {
	// Name of the copy; defaults to the original name.
	Name string `json:"name,omitempty"`
}

// Duplicate copies a speech engine. req may be nil.
func (s *Service) Duplicate(ctx context.Context, speechEngineID string, req *DuplicateRequest, opts ...option.RequestOption) (*Engine, error) {
	return transport.DoValue[Engine](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/speech-engine/%s/duplicate", speechEngineID), Body: transport.OrEmpty(req), Opts: opts})
}
