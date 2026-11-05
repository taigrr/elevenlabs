package conversations

import (
	"context"
	"encoding/json"
	"io"
	"iter"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/convai/agents"
	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/stream"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

// Service groups the /v1/convai/conversations endpoints.
type Service struct {
	cfg *requestconfig.Config

	// Analysis re-runs post-call analysis.
	Analysis *AnalysisService
	// Audio downloads conversation recordings.
	Audio *AudioService
	// Feedback rates conversations.
	Feedback *FeedbackService
	// Files uploads files into a running conversation.
	Files *FilesService
	// Messages searches conversation messages.
	Messages *MessagesService
	// Tags manages conversation tags.
	Tags *TagsService
	// Topics returns discovered conversation topics.
	Topics *TopicsService
}

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	cfg := requestconfig.Build(opts...)
	return &Service{
		cfg:      cfg,
		Analysis: &AnalysisService{cfg: cfg},
		Audio:    &AudioService{cfg: cfg},
		Feedback: &FeedbackService{cfg: cfg},
		Files:    &FilesService{cfg: cfg},
		Messages: &MessagesService{cfg: cfg},
		Tags:     &TagsService{cfg: cfg},
		Topics:   &TopicsService{cfg: cfg},
	}
}

// Status is the state of a conversation.
type Status string

// Status values.
const (
	StatusInitiated  Status = "initiated"
	StatusInProgress Status = "in-progress"
	StatusProcessing Status = "processing"
	StatusDone       Status = "done"
	StatusFailed     Status = "failed"
)

// EvaluationResult is the outcome of an evaluation, also used as a filter.
type EvaluationResult string

// EvaluationResult values.
const (
	EvaluationResultSuccess EvaluationResult = "success"
	EvaluationResultFailure EvaluationResult = "failure"
	EvaluationResultUnknown EvaluationResult = "unknown"
	EvaluationResultError   EvaluationResult = "error"
)

// Product is the product a conversation belongs to.
type Product string

// Product values.
const (
	ProductAgents       Product = "agents"
	ProductSpeechEngine Product = "speech_engine"
)

// SessionRequest holds the parameters of GetSignedURL and GetWebRTCToken.
type SessionRequest struct {
	// AgentID is required.
	AgentID string `query:"agent_id"`
	// IncludeConversationID is only used by GetSignedURL.
	IncludeConversationID *bool `query:"include_conversation_id"`
	// ParticipantName is only used by GetWebRTCToken.
	ParticipantName    string `query:"participant_name"`
	BranchID           string `query:"branch_id"`
	VersionID          string `query:"version_id"`
	Environment        string `query:"environment"`
	DebugEventsRequest *bool  `query:"debug_events_request"`
}

// WebRTCToken is a token for starting a WebRTC conversation.
type WebRTCToken struct {
	Token          string `json:"token"`
	ConversationID string `json:"conversation_id"`
}

// Filter holds the filters shared by conversation listing and
// message text search.
type Filter struct {
	AgentID                 string           `query:"agent_id"`
	VisitedAgentIDs         []string         `query:"visited_agent_ids"`
	VisitedAgentBranchIDs   []string         `query:"visited_agent_branch_ids"`
	TriggeredProcedureIDs   []string         `query:"triggered_procedure_ids"`
	CallSuccessful          EvaluationResult `query:"call_successful"`
	CallStartBeforeUnix     *int64           `query:"call_start_before_unix"`
	CallStartAfterUnix      *int64           `query:"call_start_after_unix"`
	CallDurationMinSecs     *int             `query:"call_duration_min_secs"`
	CallDurationMaxSecs     *int             `query:"call_duration_max_secs"`
	RatingMax               *int             `query:"rating_max"`
	RatingMin               *int             `query:"rating_min"`
	HasFeedbackComment      *bool            `query:"has_feedback_comment"`
	UserID                  string           `query:"user_id"`
	EvaluationParams        []string         `query:"evaluation_params"`
	DataCollectionParams    []string         `query:"data_collection_params"`
	DynamicVariableParams   []string         `query:"dynamic_variable_params"`
	ToolNames               []string         `query:"tool_names"`
	ToolNamesSuccessful     []string         `query:"tool_names_successful"`
	ToolNamesErrored        []string         `query:"tool_names_errored"`
	IncludeInvalidToolCalls *bool            `query:"include_invalid_tool_calls"`
	MainLanguages           []string         `query:"main_languages"`
	ExcludeStatuses         []Status         `query:"exclude_statuses"`
	TerminationReasons      []string         `query:"termination_reasons"`
	PageSize                *int             `query:"page_size"`
	// SummaryMode is "exclude" or "include".
	SummaryMode string `query:"summary_mode"`
	// ConversationInitiationSource is e.g. "widget", "twilio" or "sip_trunk".
	ConversationInitiationSource string   `query:"conversation_initiation_source"`
	TextOnly                     *bool    `query:"text_only"`
	ConversationProductType      Product  `query:"conversation_product_type"`
	BranchID                     string   `query:"branch_id"`
	VersionID                    string   `query:"version_id"`
	TopicIDs                     []string `query:"topic_ids"`
	Cursor                       string   `query:"cursor"`
}

// ListRequest holds the optional parameters of Service.List.
type ListRequest struct {
	Filter
	Search                string              `query:"search"`
	DataCollectionIDs     []string            `query:"data_collection_ids"`
	EvaluationCriteriaIDs []string            `query:"evaluation_criteria_ids"`
	ParentConversationID  string              `query:"parent_conversation_id"`
	TagIDs                []string            `query:"tag_ids"`
	WorkflowNodeEnteredID string              `query:"workflow_node_entered_id"`
	GuardrailTypes        []string            `query:"guardrail_types"`
	CustomGuardrailNames  []string            `query:"custom_guardrail_names"`
	SortDirection         types.SortDirection `query:"sort_direction"`
}

// Summary is a conversation in list responses.
type Summary struct {
	AgentID                      string           `json:"agent_id"`
	BranchID                     string           `json:"branch_id,omitempty"`
	VersionID                    string           `json:"version_id,omitempty"`
	AgentName                    string           `json:"agent_name,omitempty"`
	ConversationID               string           `json:"conversation_id"`
	StartTimeUnixSecs            int64            `json:"start_time_unix_secs"`
	CallDurationSecs             int              `json:"call_duration_secs"`
	MessageCount                 int              `json:"message_count"`
	Status                       Status           `json:"status"`
	TerminationReason            string           `json:"termination_reason,omitempty"`
	CallSuccessful               EvaluationResult `json:"call_successful"`
	CallSuccessScore             *float64         `json:"call_success_score,omitempty"`
	TranscriptSummary            string           `json:"transcript_summary,omitempty"`
	CallSummaryTitle             string           `json:"call_summary_title,omitempty"`
	MainLanguage                 string           `json:"main_language,omitempty"`
	ConversationInitiationSource string           `json:"conversation_initiation_source,omitempty"`
	ToolNames                    []string         `json:"tool_names,omitempty"`
	// Direction is "inbound" or "outbound".
	Direction                 string                     `json:"direction,omitempty"`
	Rating                    *float64                   `json:"rating,omitempty"`
	SentimentAnalysis         json.RawMessage            `json:"sentiment_analysis,omitempty"`
	DataCollectionResults     map[string]json.RawMessage `json:"data_collection_results,omitempty"`
	EvaluationCriteriaResults map[string]json.RawMessage `json:"evaluation_criteria_results,omitempty"`
	TagIDs                    []string                   `json:"tag_ids,omitempty"`
}

// Page is a page of conversations.
type Page struct {
	Conversations []Summary `json:"conversations"`
	NextCursor    *string   `json:"next_cursor,omitempty"`
	HasMore       bool      `json:"has_more"`
}

// Conversation is a conversation with its transcript and analysis.
// Metadata and initiation data are kept as raw JSON.
type Conversation struct {
	AgentID                          string                       `json:"agent_id"`
	AgentName                        string                       `json:"agent_name,omitempty"`
	ConversationProduct              string                       `json:"conversation_product,omitempty"`
	Status                           Status                       `json:"status"`
	UserID                           string                       `json:"user_id,omitempty"`
	BranchID                         string                       `json:"branch_id,omitempty"`
	VersionID                        string                       `json:"version_id,omitempty"`
	Metadata                         json.RawMessage              `json:"metadata"`
	Analysis                         *agents.ConversationAnalysis `json:"analysis,omitempty"`
	VisitedAgents                    []json.RawMessage            `json:"visited_agents,omitempty"`
	ConversationInitiationClientData json.RawMessage              `json:"conversation_initiation_client_data,omitempty"`
	Environment                      string                       `json:"environment,omitempty"`
	ConversationID                   string                       `json:"conversation_id"`
	HasAudio                         bool                         `json:"has_audio"`
	HasUserAudio                     bool                         `json:"has_user_audio"`
	HasResponseAudio                 bool                         `json:"has_response_audio"`
	HasAuxiliaryAudio                bool                         `json:"has_auxiliary_audio"`
	Transcript                       []agents.TranscriptMessage   `json:"transcript"`
	TagIDs                           []string                     `json:"tag_ids,omitempty"`
	OTLPTraces                       json.RawMessage              `json:"otlp_traces,omitempty"`
}

// GetFormat selects the conversation response format.
type GetFormat string

// GetFormat values.
const (
	GetFormatJSON          GetFormat = "json"
	GetFormatOpenTelemetry GetFormat = "opentelemetry"
)

// Brief is a compact summary of a conversation.
type Brief struct {
	ConversationID    string            `json:"conversation_id"`
	AgentID           string            `json:"agent_id"`
	Status            Status            `json:"status"`
	CallSummaryTitle  string            `json:"call_summary_title,omitempty"`
	TranscriptSummary string            `json:"transcript_summary,omitempty"`
	CallSuccessful    EvaluationResult  `json:"call_successful,omitempty"`
	MessageCount      int               `json:"message_count"`
	Messages          []json.RawMessage `json:"messages,omitempty"`
	MessagesOmitted   *bool             `json:"messages_omitted,omitempty"`
	Note              string            `json:"note"`
}

// GetSignedURL returns a signed URL for starting a WebSocket conversation
// with a private agent.
func (s *Service) GetSignedURL(ctx context.Context, req *SessionRequest, opts ...option.RequestOption) (*agents.SignedURL, error) {
	r := *transport.OrEmpty(req)
	r.ParticipantName = ""
	return transport.GetValue[agents.SignedURL](ctx, s.cfg, "/v1/convai/conversation/get-signed-url", &r, opts)
}

// GetWebRTCToken returns a token for starting a WebRTC conversation.
func (s *Service) GetWebRTCToken(ctx context.Context, req *SessionRequest, opts ...option.RequestOption) (*WebRTCToken, error) {
	r := *transport.OrEmpty(req)
	r.IncludeConversationID = nil
	return transport.GetValue[WebRTCToken](ctx, s.cfg, "/v1/convai/conversation/token", &r, opts)
}

// List returns one page of conversations. req may be nil.
func (s *Service) List(ctx context.Context, req *ListRequest, opts ...option.RequestOption) (*Page, error) {
	return transport.GetValue[Page](ctx, s.cfg, "/v1/convai/conversations", req, opts)
}

// ListAll iterates over every conversation matching req.
func (s *Service) ListAll(ctx context.Context, req *ListRequest, opts ...option.RequestOption) iter.Seq2[Summary, error] {
	params := *transport.OrEmpty(req)
	return stream.Paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (stream.Page[Summary], error) {
		params.Cursor = cursor
		p, err := s.List(ctx, &params, opts...)
		if err != nil {
			return stream.Page[Summary]{}, err
		}
		return stream.Page[Summary]{Items: p.Conversations, Next: transport.Deref(p.NextCursor), HasMore: p.HasMore}, nil
	})
}

// Resolve finds a conversation of an agent by an external reference such as
// a call SID.
func (s *Service) Resolve(ctx context.Context, agentID, reference string, opts ...option.RequestOption) (*Conversation, error) {
	query := struct {
		AgentID   string `query:"agent_id"`
		Reference string `query:"reference"`
	}{agentID, reference}
	return transport.GetValue[Conversation](ctx, s.cfg, "/v1/convai/conversations/resolve", query, opts)
}

// Get returns a conversation. format may be empty.
func (s *Service) Get(ctx context.Context, conversationID string, format GetFormat, opts ...option.RequestOption) (*Conversation, error) {
	query := struct {
		Format GetFormat `query:"format"`
	}{format}
	return transport.GetValue[Conversation](ctx, s.cfg, transport.Pathf("/v1/convai/conversations/%s", conversationID), query, opts)
}

// Delete deletes a conversation.
func (s *Service) Delete(ctx context.Context, conversationID string, opts ...option.RequestOption) (json.RawMessage, error) {
	return transport.SendRaw(ctx, s.cfg, http.MethodDelete, transport.Pathf("/v1/convai/conversations/%s", conversationID), nil, nil, opts)
}

// GetSummary returns a compact summary of a conversation. maxMessages limits
// the messages included; nil uses the server default.
func (s *Service) GetSummary(ctx context.Context, conversationID string, maxMessages *int, opts ...option.RequestOption) (*Brief, error) {
	query := struct {
		MaxMessages *int `query:"max_messages"`
	}{maxMessages}
	return transport.GetValue[Brief](ctx, s.cfg, transport.Pathf("/v1/convai/conversations/%s/summary", conversationID), query, opts)
}

// GetSIPMessages returns the SIP messages logged for a conversation. req may be nil.
func (s *Service) GetSIPMessages(ctx context.Context, conversationID string, req *agents.PageRequest, opts ...option.RequestOption) (*agents.SIPMessagesPage, error) {
	return transport.GetValue[agents.SIPMessagesPage](ctx, s.cfg, transport.Pathf("/v1/convai/conversations/%s/sip-messages", conversationID), req, opts)
}

// AnalysisService groups the conversation analysis endpoints.
type AnalysisService struct{ cfg *requestconfig.Config }

// Run re-runs the full analysis of a conversation.
func (s *AnalysisService) Run(ctx context.Context, conversationID string, opts ...option.RequestOption) (*Conversation, error) {
	return transport.SendValue[Conversation](ctx, s.cfg, http.MethodPost, transport.Pathf("/v1/convai/conversations/%s/analysis/run", conversationID), nil, nil, opts)
}

// RunEvaluation runs one evaluation criterion. scope is "conversation",
// "agent" or empty.
func (s *AnalysisService) RunEvaluation(ctx context.Context, conversationID, evaluationID, scope string, opts ...option.RequestOption) (*Conversation, error) {
	body := struct {
		EvaluationID string `json:"evaluation_id"`
		Scope        string `json:"scope,omitempty"`
	}{evaluationID, scope}
	return transport.SendValue[Conversation](ctx, s.cfg, http.MethodPost, transport.Pathf("/v1/convai/conversations/%s/analysis/evaluations/run", conversationID), nil, body, opts)
}

// AudioService groups the conversation audio endpoint.
type AudioService struct{ cfg *requestconfig.Config }

// Get streams a conversation's recording. The caller must close it.
func (s *AudioService) Get(ctx context.Context, conversationID string, opts ...option.RequestOption) (io.ReadCloser, error) {
	return transport.DoStream(ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/convai/conversations/%s/audio", conversationID), Accept: transport.AcceptAny, Opts: opts})
}

// FeedbackService groups the conversation feedback endpoint.
type FeedbackService struct{ cfg *requestconfig.Config }

// FeedbackScore is a like or dislike.
type FeedbackScore string

// FeedbackScore values.
const (
	FeedbackLike    FeedbackScore = "like"
	FeedbackDislike FeedbackScore = "dislike"
)

// Create rates a conversation. An empty score clears the rating.
func (s *FeedbackService) Create(ctx context.Context, conversationID string, score FeedbackScore, opts ...option.RequestOption) (json.RawMessage, error) {
	body := struct {
		Feedback *FeedbackScore `json:"feedback"`
	}{}
	if score != "" {
		body.Feedback = &score
	}
	return transport.SendRaw(ctx, s.cfg, http.MethodPost, transport.Pathf("/v1/convai/conversations/%s/feedback", conversationID), nil, body, opts)
}

// FilesService groups the conversation file endpoints.
type FilesService struct{ cfg *requestconfig.Config }

// FileUploadResponse identifies an uploaded file.
type FileUploadResponse struct {
	FileID string `json:"file_id"`
}

// Create uploads a file into a conversation.
func (s *FilesService) Create(ctx context.Context, conversationID string, file types.File, opts ...option.RequestOption) (*FileUploadResponse, error) {
	form := struct {
		File types.File `form:"file"`
	}{file}
	return transport.DoValue[FileUploadResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/convai/conversations/%s/files", conversationID), Form: &form, Opts: opts})
}

// Delete removes an uploaded file from a conversation.
func (s *FilesService) Delete(ctx context.Context, conversationID, fileID string, opts ...option.RequestOption) (*FileUploadResponse, error) {
	return transport.SendValue[FileUploadResponse](ctx, s.cfg, http.MethodDelete, transport.Pathf("/v1/convai/conversations/%s/files/%s", conversationID, fileID), nil, nil, opts)
}

// MessagesService groups the message search endpoints.
type MessagesService struct{ cfg *requestconfig.Config }

// MessageSearchResult is a matching message chunk.
type MessageSearchResult struct {
	ConversationID                string            `json:"conversation_id"`
	AgentID                       string            `json:"agent_id"`
	AgentName                     string            `json:"agent_name,omitempty"`
	TranscriptIndex               int               `json:"transcript_index"`
	ChunkText                     string            `json:"chunk_text"`
	ChunkHighlights               []json.RawMessage `json:"chunk_highlights,omitempty"`
	Score                         float64           `json:"score"`
	ConversationStartTimeUnixSecs int64             `json:"conversation_start_time_unix_secs"`
}

// MessageSearchPage is a page of message search results.
type MessageSearchPage struct {
	Meta       *agents.ListMeta      `json:"meta,omitempty"`
	Results    []MessageSearchResult `json:"results"`
	NextCursor *string               `json:"next_cursor,omitempty"`
	HasMore    bool                  `json:"has_more"`
}

// MessageTextSearchRequest holds the parameters of MessagesService.TextSearch.
type MessageTextSearchRequest struct {
	// TextQuery is required.
	TextQuery string `query:"text_query"`
	Filter
	// SortBy is "search_score" or "created_at".
	SortBy string `query:"sort_by"`
}

// MessageSearchRequest holds the parameters of MessagesService.Search.
type MessageSearchRequest struct {
	// TextQuery is required.
	TextQuery string `query:"text_query"`
	AgentID   string `query:"agent_id"`
	BranchID  string `query:"branch_id"`
	PageSize  *int   `query:"page_size"`
	Cursor    string `query:"cursor"`
}

// TextSearch runs a full-text search over conversation messages.
func (s *MessagesService) TextSearch(ctx context.Context, req *MessageTextSearchRequest, opts ...option.RequestOption) (*MessageSearchPage, error) {
	return transport.GetValue[MessageSearchPage](ctx, s.cfg, "/v1/convai/conversations/messages/text-search", req, opts)
}

// Search runs a semantic search over conversation messages.
func (s *MessagesService) Search(ctx context.Context, req *MessageSearchRequest, opts ...option.RequestOption) (*MessageSearchPage, error) {
	return transport.GetValue[MessageSearchPage](ctx, s.cfg, "/v1/convai/conversations/messages/smart-search", req, opts)
}

// TagsService groups the conversation tag endpoints.
type TagsService struct{ cfg *requestconfig.Config }

// Tag is a workspace tag for conversations.
type Tag struct {
	TagID             string `json:"tag_id"`
	WorkspaceID       string `json:"workspace_id"`
	OwnerUserID       string `json:"owner_user_id"`
	Title             string `json:"title"`
	Description       string `json:"description,omitempty"`
	CreatedAtUnixSecs int64  `json:"created_at_unix_secs"`
}

// TagsPage is a page of tags.
type TagsPage struct {
	ConversationTags []Tag   `json:"conversation_tags"`
	NextCursor       *string `json:"next_cursor,omitempty"`
	HasMore          bool    `json:"has_more"`
}

// TagRequest is the body of tag create and update. Title is
// required when creating.
type TagRequest struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// Assign adds tags to a conversation.
func (s *TagsService) Assign(ctx context.Context, conversationID string, tagIDs []string, opts ...option.RequestOption) error {
	body := struct {
		TagIDs []string `json:"tag_ids"`
	}{transport.NonNil(tagIDs)}
	return transport.NoContent(ctx, s.cfg, http.MethodPost, transport.Pathf("/v1/convai/conversations/%s/tags", conversationID), nil, body, opts)
}

// Unassign removes a tag from a conversation.
func (s *TagsService) Unassign(ctx context.Context, conversationID, tagID string, opts ...option.RequestOption) error {
	return transport.NoContent(ctx, s.cfg, http.MethodDelete, transport.Pathf("/v1/convai/conversations/%s/tags/%s", conversationID, tagID), nil, nil, opts)
}

// List returns one page of tags. req may be nil.
func (s *TagsService) List(ctx context.Context, req *agents.PageRequest, opts ...option.RequestOption) (*TagsPage, error) {
	return transport.GetValue[TagsPage](ctx, s.cfg, "/v1/convai/tags", req, opts)
}

// ListAll iterates over every tag.
func (s *TagsService) ListAll(ctx context.Context, req *agents.PageRequest, opts ...option.RequestOption) iter.Seq2[Tag, error] {
	params := *transport.OrEmpty(req)
	return stream.Paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (stream.Page[Tag], error) {
		params.Cursor = cursor
		p, err := s.List(ctx, &params, opts...)
		if err != nil {
			return stream.Page[Tag]{}, err
		}
		return stream.Page[Tag]{Items: p.ConversationTags, Next: transport.Deref(p.NextCursor), HasMore: p.HasMore}, nil
	})
}

// Create creates a tag.
func (s *TagsService) Create(ctx context.Context, req *TagRequest, opts ...option.RequestOption) (*Tag, error) {
	return transport.SendValue[Tag](ctx, s.cfg, http.MethodPost, "/v1/convai/tags", nil, transport.OrEmpty(req), opts)
}

// Get returns a tag.
func (s *TagsService) Get(ctx context.Context, tagID string, opts ...option.RequestOption) (*Tag, error) {
	return transport.GetValue[Tag](ctx, s.cfg, transport.Pathf("/v1/convai/tags/%s", tagID), nil, opts)
}

// Delete deletes a tag.
func (s *TagsService) Delete(ctx context.Context, tagID string, opts ...option.RequestOption) error {
	return transport.NoContent(ctx, s.cfg, http.MethodDelete, transport.Pathf("/v1/convai/tags/%s", tagID), nil, nil, opts)
}

// Update changes a tag.
func (s *TagsService) Update(ctx context.Context, tagID string, req *TagRequest, opts ...option.RequestOption) (*Tag, error) {
	return transport.SendValue[Tag](ctx, s.cfg, http.MethodPatch, transport.Pathf("/v1/convai/tags/%s", tagID), nil, transport.OrEmpty(req), opts)
}

// TopicsService groups the agent topics endpoint.
type TopicsService struct{ cfg *requestconfig.Config }

// TopicsRequest holds the optional parameters of TopicsService.Get.
type TopicsRequest struct {
	PageSize *int `query:"page_size"`
	// SortBy is "conversations", "sentiment", "success_rate" or "frustration".
	SortBy                    string              `query:"sort_by"`
	SortDirection             types.SortDirection `query:"sort_direction"`
	FromUnixSecs              *int64              `query:"from_unix_secs"`
	ToUnixSecs                *int64              `query:"to_unix_secs"`
	IncludeEvaluationCriteria *bool               `query:"include_evaluation_criteria"`
	Cursor                    string              `query:"cursor"`
}

// AgentTopics lists the topics discovered in an agent's conversations.
// Topic entries are kept as raw JSON.
type AgentTopics struct {
	Topics              []json.RawMessage `json:"topics"`
	WindowStartUnixSecs int64             `json:"window_start_unix_secs"`
	WindowEndUnixSecs   int64             `json:"window_end_unix_secs"`
	AggregatedRunCount  *int              `json:"aggregated_run_count,omitempty"`
	HasMore             *bool             `json:"has_more,omitempty"`
	NextCursor          *string           `json:"next_cursor,omitempty"`
}

// Get returns an agent's conversation topics. req may be nil.
func (s *TopicsService) Get(ctx context.Context, agentID string, req *TopicsRequest, opts ...option.RequestOption) (*AgentTopics, error) {
	return transport.GetValue[AgentTopics](ctx, s.cfg, transport.Pathf("/v1/convai/agents/%s/topics", agentID), req, opts)
}
