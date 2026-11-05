package elevenlabs

import (
	"context"
	"encoding/json"
	"iter"
	"net/http"
)

// SpeechEngineService groups the /v1/speech-engine endpoints, which manage
// speech engines: voice pipelines (ASR, turn taking and TTS) that delegate
// response generation to your own WebSocket server.
type SpeechEngineService struct{ cfg *config }

// ResourceAccessInfo describes the caller's access to a resource.
type ResourceAccessInfo struct {
	IsCreator                    bool          `json:"is_creator"`
	CreatorName                  string        `json:"creator_name"`
	CreatorEmail                 string        `json:"creator_email"`
	Role                         ResourceRole  `json:"role"`
	AnonymousAccessLevelOverride *ResourceRole `json:"anonymous_access_level_override,omitempty"`
	// AccessSource is "creator", "explicit", "workspace_admin" or "workspace_default".
	AccessSource *string `json:"access_source,omitempty"`
}

// SpeechEngineConfig points a speech engine at your WebSocket server.
type SpeechEngineConfig struct {
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

// SpeechEngineOverrides lists which settings clients may override.
type SpeechEngineOverrides struct {
	FirstMessage *bool `json:"first_message,omitempty"`
}

// SpeechEngineSummary is a speech engine in a listing.
type SpeechEngineSummary struct {
	SpeechEngineID    string             `json:"speech_engine_id"`
	Name              string             `json:"name"`
	VoiceID           string             `json:"voice_id"`
	CreatedAtUnixSecs int64              `json:"created_at_unix_secs"`
	UpdatedAtUnixSecs int64              `json:"updated_at_unix_secs"`
	Tags              []string           `json:"tags"`
	AccessInfo        ResourceAccessInfo `json:"access_info"`
}

// SpeechEngine is a speech engine. The ASR, TTS, Turn, VAD, Conversation,
// Privacy and Metadata objects share their schema with conversational AI
// agents and are kept as raw JSON.
type SpeechEngine struct {
	SpeechEngineID        string                `json:"speech_engine_id"`
	Name                  string                `json:"name"`
	SpeechEngine          SpeechEngineConfig    `json:"speech_engine"`
	ASR                   json.RawMessage       `json:"asr"`
	TTS                   json.RawMessage       `json:"tts"`
	Turn                  json.RawMessage       `json:"turn"`
	VAD                   json.RawMessage       `json:"vad"`
	Conversation          json.RawMessage       `json:"conversation"`
	Privacy               json.RawMessage       `json:"privacy"`
	CallLimits            AgentCallLimits       `json:"call_limits"`
	Language              string                `json:"language"`
	CascadeTimeoutSeconds float64               `json:"cascade_timeout_seconds"`
	Tags                  []string              `json:"tags"`
	Overrides             SpeechEngineOverrides `json:"overrides"`
	Metadata              json.RawMessage       `json:"metadata"`
	AccessInfo            *ResourceAccessInfo   `json:"access_info,omitempty"`
}

// SpeechEngineSortBy is the field speech engines are sorted by.
type SpeechEngineSortBy string

// SpeechEngineSortBy values.
const (
	SpeechEngineSortByName        SpeechEngineSortBy = "name"
	SpeechEngineSortByCreatedAt   SpeechEngineSortBy = "created_at"
	SpeechEngineSortByUpdatedAt   SpeechEngineSortBy = "updated_at"
	SpeechEngineSortByCallCount7d SpeechEngineSortBy = "call_count_7d"
)

// SpeechEngineListRequest holds the optional parameters of SpeechEngineService.List.
type SpeechEngineListRequest struct {
	PageSize      *int               `query:"page_size"`
	Search        string             `query:"search"`
	SortDirection SortDirection      `query:"sort_direction"`
	SortBy        SpeechEngineSortBy `query:"sort_by"`
	Cursor        string             `query:"cursor"`
}

// SpeechEnginePage is a page of speech engines.
type SpeechEnginePage struct {
	SpeechEngines []SpeechEngineSummary `json:"speech_engines"`
	NextCursor    *string               `json:"next_cursor,omitempty"`
	HasMore       bool                  `json:"has_more"`
}

// List returns one page of speech engines. req may be nil.
func (s *SpeechEngineService) List(ctx context.Context, req *SpeechEngineListRequest, opts ...Option) (*SpeechEnginePage, error) {
	return doValue[SpeechEnginePage](ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/speech-engine", query: req, opts: opts})
}

// ListAll iterates over every speech engine, fetching pages on demand.
func (s *SpeechEngineService) ListAll(ctx context.Context, req *SpeechEngineListRequest, opts ...Option) iter.Seq2[SpeechEngineSummary, error] {
	params := *orEmpty(req)
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[SpeechEngineSummary], error) {
		params.Cursor = cursor
		resp, err := s.List(ctx, &params, opts...)
		if err != nil {
			return page[SpeechEngineSummary]{}, err
		}
		return page[SpeechEngineSummary]{items: resp.SpeechEngines, next: deref(resp.NextCursor), hasMore: resp.HasMore}, nil
	})
}

// SpeechEngineCreateRequest is the body of SpeechEngineService.Create.
type SpeechEngineCreateRequest struct {
	Name string `json:"name,omitempty"`
	// SpeechEngine is the WebSocket server configuration. Required.
	SpeechEngine          SpeechEngineConfig     `json:"speech_engine"`
	ASR                   any                    `json:"asr,omitempty"`
	TTS                   any                    `json:"tts,omitempty"`
	Turn                  any                    `json:"turn,omitempty"`
	VAD                   any                    `json:"vad,omitempty"`
	Conversation          any                    `json:"conversation,omitempty"`
	Privacy               any                    `json:"privacy,omitempty"`
	CallLimits            *AgentCallLimits       `json:"call_limits,omitempty"`
	Language              string                 `json:"language,omitempty"`
	CascadeTimeoutSeconds *float64               `json:"cascade_timeout_seconds,omitempty"`
	Tags                  []string               `json:"tags,omitempty"`
	Overrides             *SpeechEngineOverrides `json:"overrides,omitempty"`
}

// SpeechEngineUpdateRequest is the body of SpeechEngineService.Update. Nil
// fields are left unchanged.
type SpeechEngineUpdateRequest struct {
	Name                  string                 `json:"name,omitempty"`
	SpeechEngine          *SpeechEngineConfig    `json:"speech_engine,omitempty"`
	ASR                   any                    `json:"asr,omitempty"`
	TTS                   any                    `json:"tts,omitempty"`
	Turn                  any                    `json:"turn,omitempty"`
	VAD                   any                    `json:"vad,omitempty"`
	Conversation          any                    `json:"conversation,omitempty"`
	Privacy               any                    `json:"privacy,omitempty"`
	CallLimits            *AgentCallLimits       `json:"call_limits,omitempty"`
	Language              string                 `json:"language,omitempty"`
	CascadeTimeoutSeconds *float64               `json:"cascade_timeout_seconds,omitempty"`
	Tags                  []string               `json:"tags,omitempty"`
	Overrides             *SpeechEngineOverrides `json:"overrides,omitempty"`
}

// Create creates a speech engine.
func (s *SpeechEngineService) Create(ctx context.Context, req *SpeechEngineCreateRequest, opts ...Option) (*SpeechEngine, error) {
	return doValue[SpeechEngine](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/speech-engine", body: orEmpty(req), opts: opts})
}

// Get returns a speech engine.
func (s *SpeechEngineService) Get(ctx context.Context, speechEngineID string, opts ...Option) (*SpeechEngine, error) {
	return doValue[SpeechEngine](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/speech-engine/%s", speechEngineID), opts: opts})
}

// Delete deletes a speech engine.
func (s *SpeechEngineService) Delete(ctx context.Context, speechEngineID string, opts ...Option) error {
	return doJSON(ctx, s.cfg, &request{method: http.MethodDelete, path: pathf("/v1/speech-engine/%s", speechEngineID), opts: opts}, nil)
}

// Update edits a speech engine. req may be nil.
func (s *SpeechEngineService) Update(ctx context.Context, speechEngineID string, req *SpeechEngineUpdateRequest, opts ...Option) (*SpeechEngine, error) {
	return doValue[SpeechEngine](ctx, s.cfg, &request{method: http.MethodPatch, path: pathf("/v1/speech-engine/%s", speechEngineID), body: orEmpty(req), opts: opts})
}

// SpeechEngineDuplicateRequest is the body of SpeechEngineService.Duplicate.
type SpeechEngineDuplicateRequest struct {
	// Name of the copy; defaults to the original name.
	Name string `json:"name,omitempty"`
}

// Duplicate copies a speech engine. req may be nil.
func (s *SpeechEngineService) Duplicate(ctx context.Context, speechEngineID string, req *SpeechEngineDuplicateRequest, opts ...Option) (*SpeechEngine, error) {
	return doValue[SpeechEngine](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/speech-engine/%s/duplicate", speechEngineID), body: orEmpty(req), opts: opts})
}
