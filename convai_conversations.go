package elevenlabs

import (
	"context"
	"encoding/json"
	"io"
	"iter"
	"net/http"
)

// ConversationsService groups the /v1/convai/conversations endpoints.
type ConversationsService struct {
	cfg *config

	// Analysis re-runs post-call analysis.
	Analysis *ConversationAnalysisService
	// Audio downloads conversation recordings.
	Audio *ConversationAudioService
	// Feedback rates conversations.
	Feedback *ConversationFeedbackService
	// Files uploads files into a running conversation.
	Files *ConversationFilesService
	// Messages searches conversation messages.
	Messages *ConversationMessagesService
	// Tags manages conversation tags.
	Tags *ConversationTagsService
	// Topics returns discovered conversation topics.
	Topics *ConversationTopicsService
}

func newConversationsService(cfg *config) *ConversationsService {
	return &ConversationsService{
		cfg:      cfg,
		Analysis: &ConversationAnalysisService{cfg: cfg},
		Audio:    &ConversationAudioService{cfg: cfg},
		Feedback: &ConversationFeedbackService{cfg: cfg},
		Files:    &ConversationFilesService{cfg: cfg},
		Messages: &ConversationMessagesService{cfg: cfg},
		Tags:     &ConversationTagsService{cfg: cfg},
		Topics:   &ConversationTopicsService{cfg: cfg},
	}
}

// ConversationStatus is the state of a conversation.
type ConversationStatus string

// ConversationStatus values.
const (
	ConversationStatusInitiated  ConversationStatus = "initiated"
	ConversationStatusInProgress ConversationStatus = "in-progress"
	ConversationStatusProcessing ConversationStatus = "processing"
	ConversationStatusDone       ConversationStatus = "done"
	ConversationStatusFailed     ConversationStatus = "failed"
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

// ConversationProduct is the product a conversation belongs to.
type ConversationProduct string

// ConversationProduct values.
const (
	ConversationProductAgents       ConversationProduct = "agents"
	ConversationProductSpeechEngine ConversationProduct = "speech_engine"
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

// SignedURL is a signed WebSocket URL for starting a conversation.
type SignedURL struct {
	SignedURL string `json:"signed_url"`
}

// WebRTCToken is a token for starting a WebRTC conversation.
type WebRTCToken struct {
	Token          string `json:"token"`
	ConversationID string `json:"conversation_id"`
}

// ConversationFilter holds the filters shared by conversation listing and
// message text search.
type ConversationFilter struct {
	AgentID                 string               `query:"agent_id"`
	VisitedAgentIDs         []string             `query:"visited_agent_ids"`
	VisitedAgentBranchIDs   []string             `query:"visited_agent_branch_ids"`
	TriggeredProcedureIDs   []string             `query:"triggered_procedure_ids"`
	CallSuccessful          EvaluationResult     `query:"call_successful"`
	CallStartBeforeUnix     *int64               `query:"call_start_before_unix"`
	CallStartAfterUnix      *int64               `query:"call_start_after_unix"`
	CallDurationMinSecs     *int                 `query:"call_duration_min_secs"`
	CallDurationMaxSecs     *int                 `query:"call_duration_max_secs"`
	RatingMax               *int                 `query:"rating_max"`
	RatingMin               *int                 `query:"rating_min"`
	HasFeedbackComment      *bool                `query:"has_feedback_comment"`
	UserID                  string               `query:"user_id"`
	EvaluationParams        []string             `query:"evaluation_params"`
	DataCollectionParams    []string             `query:"data_collection_params"`
	DynamicVariableParams   []string             `query:"dynamic_variable_params"`
	ToolNames               []string             `query:"tool_names"`
	ToolNamesSuccessful     []string             `query:"tool_names_successful"`
	ToolNamesErrored        []string             `query:"tool_names_errored"`
	IncludeInvalidToolCalls *bool                `query:"include_invalid_tool_calls"`
	MainLanguages           []string             `query:"main_languages"`
	ExcludeStatuses         []ConversationStatus `query:"exclude_statuses"`
	TerminationReasons      []string             `query:"termination_reasons"`
	PageSize                *int                 `query:"page_size"`
	// SummaryMode is "exclude" or "include".
	SummaryMode string `query:"summary_mode"`
	// ConversationInitiationSource is e.g. "widget", "twilio" or "sip_trunk".
	ConversationInitiationSource string              `query:"conversation_initiation_source"`
	TextOnly                     *bool               `query:"text_only"`
	ConversationProductType      ConversationProduct `query:"conversation_product_type"`
	BranchID                     string              `query:"branch_id"`
	VersionID                    string              `query:"version_id"`
	TopicIDs                     []string            `query:"topic_ids"`
	Cursor                       string              `query:"cursor"`
}

// ConversationListRequest holds the optional parameters of ConversationsService.List.
type ConversationListRequest struct {
	ConversationFilter
	Search                string        `query:"search"`
	DataCollectionIDs     []string      `query:"data_collection_ids"`
	EvaluationCriteriaIDs []string      `query:"evaluation_criteria_ids"`
	ParentConversationID  string        `query:"parent_conversation_id"`
	TagIDs                []string      `query:"tag_ids"`
	WorkflowNodeEnteredID string        `query:"workflow_node_entered_id"`
	GuardrailTypes        []string      `query:"guardrail_types"`
	CustomGuardrailNames  []string      `query:"custom_guardrail_names"`
	SortDirection         SortDirection `query:"sort_direction"`
}

// ConversationSummary is a conversation in list responses.
type ConversationSummary struct {
	AgentID                      string             `json:"agent_id"`
	BranchID                     string             `json:"branch_id,omitempty"`
	VersionID                    string             `json:"version_id,omitempty"`
	AgentName                    string             `json:"agent_name,omitempty"`
	ConversationID               string             `json:"conversation_id"`
	StartTimeUnixSecs            int64              `json:"start_time_unix_secs"`
	CallDurationSecs             int                `json:"call_duration_secs"`
	MessageCount                 int                `json:"message_count"`
	Status                       ConversationStatus `json:"status"`
	TerminationReason            string             `json:"termination_reason,omitempty"`
	CallSuccessful               EvaluationResult   `json:"call_successful"`
	CallSuccessScore             *float64           `json:"call_success_score,omitempty"`
	TranscriptSummary            string             `json:"transcript_summary,omitempty"`
	CallSummaryTitle             string             `json:"call_summary_title,omitempty"`
	MainLanguage                 string             `json:"main_language,omitempty"`
	ConversationInitiationSource string             `json:"conversation_initiation_source,omitempty"`
	ToolNames                    []string           `json:"tool_names,omitempty"`
	// Direction is "inbound" or "outbound".
	Direction                 string                     `json:"direction,omitempty"`
	Rating                    *float64                   `json:"rating,omitempty"`
	SentimentAnalysis         json.RawMessage            `json:"sentiment_analysis,omitempty"`
	DataCollectionResults     map[string]json.RawMessage `json:"data_collection_results,omitempty"`
	EvaluationCriteriaResults map[string]json.RawMessage `json:"evaluation_criteria_results,omitempty"`
	TagIDs                    []string                   `json:"tag_ids,omitempty"`
}

// ConversationsPage is a page of conversations.
type ConversationsPage struct {
	Conversations []ConversationSummary `json:"conversations"`
	NextCursor    *string               `json:"next_cursor,omitempty"`
	HasMore       bool                  `json:"has_more"`
}

// Conversation is a conversation with its transcript and analysis.
// Metadata and initiation data are kept as raw JSON.
type Conversation struct {
	AgentID                          string                `json:"agent_id"`
	AgentName                        string                `json:"agent_name,omitempty"`
	ConversationProduct              string                `json:"conversation_product,omitempty"`
	Status                           ConversationStatus    `json:"status"`
	UserID                           string                `json:"user_id,omitempty"`
	BranchID                         string                `json:"branch_id,omitempty"`
	VersionID                        string                `json:"version_id,omitempty"`
	Metadata                         json.RawMessage       `json:"metadata"`
	Analysis                         *ConversationAnalysis `json:"analysis,omitempty"`
	VisitedAgents                    []json.RawMessage     `json:"visited_agents,omitempty"`
	ConversationInitiationClientData json.RawMessage       `json:"conversation_initiation_client_data,omitempty"`
	Environment                      string                `json:"environment,omitempty"`
	ConversationID                   string                `json:"conversation_id"`
	HasAudio                         bool                  `json:"has_audio"`
	HasUserAudio                     bool                  `json:"has_user_audio"`
	HasResponseAudio                 bool                  `json:"has_response_audio"`
	HasAuxiliaryAudio                bool                  `json:"has_auxiliary_audio"`
	Transcript                       []TranscriptMessage   `json:"transcript"`
	TagIDs                           []string              `json:"tag_ids,omitempty"`
	OTLPTraces                       json.RawMessage       `json:"otlp_traces,omitempty"`
}

// ConversationGetFormat selects the conversation response format.
type ConversationGetFormat string

// ConversationGetFormat values.
const (
	ConversationGetFormatJSON          ConversationGetFormat = "json"
	ConversationGetFormatOpenTelemetry ConversationGetFormat = "opentelemetry"
)

// ConversationBrief is a compact summary of a conversation.
type ConversationBrief struct {
	ConversationID    string             `json:"conversation_id"`
	AgentID           string             `json:"agent_id"`
	Status            ConversationStatus `json:"status"`
	CallSummaryTitle  string             `json:"call_summary_title,omitempty"`
	TranscriptSummary string             `json:"transcript_summary,omitempty"`
	CallSuccessful    EvaluationResult   `json:"call_successful,omitempty"`
	MessageCount      int                `json:"message_count"`
	Messages          []json.RawMessage  `json:"messages,omitempty"`
	MessagesOmitted   *bool              `json:"messages_omitted,omitempty"`
	Note              string             `json:"note"`
}

// SIPMessage is a logged SIP message.
type SIPMessage struct {
	CallID        string   `json:"call_id"`
	PhoneNumbers  []string `json:"phone_numbers"`
	LocalAddress  string   `json:"local_address"`
	RemoteAddress string   `json:"remote_address"`
	Transport     string   `json:"transport"`
	RawMessage    string   `json:"raw_message"`
	ErrorMessage  string   `json:"error_message"`
	// Direction is "inbound" or "outbound".
	Direction          string `json:"direction"`
	CreatedAtUnixMicro int64  `json:"created_at_unix_micro"`
}

// SIPMessagesPage is a page of SIP messages.
type SIPMessagesPage struct {
	SIPMessages []SIPMessage `json:"sip_messages"`
	NextCursor  *string      `json:"next_cursor,omitempty"`
	HasMore     *bool        `json:"has_more,omitempty"`
}

// GetSignedURL returns a signed URL for starting a WebSocket conversation
// with a private agent.
func (s *ConversationsService) GetSignedURL(ctx context.Context, req *SessionRequest, opts ...Option) (*SignedURL, error) {
	r := *orEmpty(req)
	r.ParticipantName = ""
	return getValue[SignedURL](ctx, s.cfg, "/v1/convai/conversation/get-signed-url", &r, opts)
}

// GetWebRTCToken returns a token for starting a WebRTC conversation.
func (s *ConversationsService) GetWebRTCToken(ctx context.Context, req *SessionRequest, opts ...Option) (*WebRTCToken, error) {
	r := *orEmpty(req)
	r.IncludeConversationID = nil
	return getValue[WebRTCToken](ctx, s.cfg, "/v1/convai/conversation/token", &r, opts)
}

// List returns one page of conversations. req may be nil.
func (s *ConversationsService) List(ctx context.Context, req *ConversationListRequest, opts ...Option) (*ConversationsPage, error) {
	return getValue[ConversationsPage](ctx, s.cfg, "/v1/convai/conversations", req, opts)
}

// ListAll iterates over every conversation matching req.
func (s *ConversationsService) ListAll(ctx context.Context, req *ConversationListRequest, opts ...Option) iter.Seq2[ConversationSummary, error] {
	params := *orEmpty(req)
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[ConversationSummary], error) {
		params.Cursor = cursor
		p, err := s.List(ctx, &params, opts...)
		if err != nil {
			return page[ConversationSummary]{}, err
		}
		return page[ConversationSummary]{items: p.Conversations, next: deref(p.NextCursor), hasMore: p.HasMore}, nil
	})
}

// Resolve finds a conversation of an agent by an external reference such as
// a call SID.
func (s *ConversationsService) Resolve(ctx context.Context, agentID, reference string, opts ...Option) (*Conversation, error) {
	query := struct {
		AgentID   string `query:"agent_id"`
		Reference string `query:"reference"`
	}{agentID, reference}
	return getValue[Conversation](ctx, s.cfg, "/v1/convai/conversations/resolve", query, opts)
}

// Get returns a conversation. format may be empty.
func (s *ConversationsService) Get(ctx context.Context, conversationID string, format ConversationGetFormat, opts ...Option) (*Conversation, error) {
	query := struct {
		Format ConversationGetFormat `query:"format"`
	}{format}
	return getValue[Conversation](ctx, s.cfg, pathf("/v1/convai/conversations/%s", conversationID), query, opts)
}

// Delete deletes a conversation.
func (s *ConversationsService) Delete(ctx context.Context, conversationID string, opts ...Option) (json.RawMessage, error) {
	return sendRaw(ctx, s.cfg, http.MethodDelete, pathf("/v1/convai/conversations/%s", conversationID), nil, nil, opts)
}

// GetSummary returns a compact summary of a conversation. maxMessages limits
// the messages included; nil uses the server default.
func (s *ConversationsService) GetSummary(ctx context.Context, conversationID string, maxMessages *int, opts ...Option) (*ConversationBrief, error) {
	query := struct {
		MaxMessages *int `query:"max_messages"`
	}{maxMessages}
	return getValue[ConversationBrief](ctx, s.cfg, pathf("/v1/convai/conversations/%s/summary", conversationID), query, opts)
}

// GetSIPMessages returns the SIP messages logged for a conversation. req may be nil.
func (s *ConversationsService) GetSIPMessages(ctx context.Context, conversationID string, req *PageRequest, opts ...Option) (*SIPMessagesPage, error) {
	return getValue[SIPMessagesPage](ctx, s.cfg, pathf("/v1/convai/conversations/%s/sip-messages", conversationID), req, opts)
}

// ConversationAnalysisService groups the conversation analysis endpoints.
type ConversationAnalysisService struct{ cfg *config }

// Run re-runs the full analysis of a conversation.
func (s *ConversationAnalysisService) Run(ctx context.Context, conversationID string, opts ...Option) (*Conversation, error) {
	return sendValue[Conversation](ctx, s.cfg, http.MethodPost, pathf("/v1/convai/conversations/%s/analysis/run", conversationID), nil, nil, opts)
}

// RunEvaluation runs one evaluation criterion. scope is "conversation",
// "agent" or empty.
func (s *ConversationAnalysisService) RunEvaluation(ctx context.Context, conversationID, evaluationID, scope string, opts ...Option) (*Conversation, error) {
	body := struct {
		EvaluationID string `json:"evaluation_id"`
		Scope        string `json:"scope,omitempty"`
	}{evaluationID, scope}
	return sendValue[Conversation](ctx, s.cfg, http.MethodPost, pathf("/v1/convai/conversations/%s/analysis/evaluations/run", conversationID), nil, body, opts)
}

// ConversationAudioService groups the conversation audio endpoint.
type ConversationAudioService struct{ cfg *config }

// Get streams a conversation's recording. The caller must close it.
func (s *ConversationAudioService) Get(ctx context.Context, conversationID string, opts ...Option) (io.ReadCloser, error) {
	return doStream(ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/convai/conversations/%s/audio", conversationID), accept: acceptAny, opts: opts})
}

// ConversationFeedbackService groups the conversation feedback endpoint.
type ConversationFeedbackService struct{ cfg *config }

// FeedbackScore is a like or dislike.
type FeedbackScore string

// FeedbackScore values.
const (
	FeedbackLike    FeedbackScore = "like"
	FeedbackDislike FeedbackScore = "dislike"
)

// Create rates a conversation. An empty score clears the rating.
func (s *ConversationFeedbackService) Create(ctx context.Context, conversationID string, score FeedbackScore, opts ...Option) (json.RawMessage, error) {
	body := struct {
		Feedback *FeedbackScore `json:"feedback"`
	}{}
	if score != "" {
		body.Feedback = &score
	}
	return sendRaw(ctx, s.cfg, http.MethodPost, pathf("/v1/convai/conversations/%s/feedback", conversationID), nil, body, opts)
}

// ConversationFilesService groups the conversation file endpoints.
type ConversationFilesService struct{ cfg *config }

// FileUploadResponse identifies an uploaded file.
type FileUploadResponse struct {
	FileID string `json:"file_id"`
}

// Create uploads a file into a conversation.
func (s *ConversationFilesService) Create(ctx context.Context, conversationID string, file File, opts ...Option) (*FileUploadResponse, error) {
	form := struct {
		File File `form:"file"`
	}{file}
	return doValue[FileUploadResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/convai/conversations/%s/files", conversationID), form: &form, opts: opts})
}

// Delete removes an uploaded file from a conversation.
func (s *ConversationFilesService) Delete(ctx context.Context, conversationID, fileID string, opts ...Option) (*FileUploadResponse, error) {
	return sendValue[FileUploadResponse](ctx, s.cfg, http.MethodDelete, pathf("/v1/convai/conversations/%s/files/%s", conversationID, fileID), nil, nil, opts)
}

// ConversationMessagesService groups the message search endpoints.
type ConversationMessagesService struct{ cfg *config }

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
	Meta       *ListMeta             `json:"meta,omitempty"`
	Results    []MessageSearchResult `json:"results"`
	NextCursor *string               `json:"next_cursor,omitempty"`
	HasMore    bool                  `json:"has_more"`
}

// MessageTextSearchRequest holds the parameters of ConversationMessagesService.TextSearch.
type MessageTextSearchRequest struct {
	// TextQuery is required.
	TextQuery string `query:"text_query"`
	ConversationFilter
	// SortBy is "search_score" or "created_at".
	SortBy string `query:"sort_by"`
}

// MessageSearchRequest holds the parameters of ConversationMessagesService.Search.
type MessageSearchRequest struct {
	// TextQuery is required.
	TextQuery string `query:"text_query"`
	AgentID   string `query:"agent_id"`
	BranchID  string `query:"branch_id"`
	PageSize  *int   `query:"page_size"`
	Cursor    string `query:"cursor"`
}

// TextSearch runs a full-text search over conversation messages.
func (s *ConversationMessagesService) TextSearch(ctx context.Context, req *MessageTextSearchRequest, opts ...Option) (*MessageSearchPage, error) {
	return getValue[MessageSearchPage](ctx, s.cfg, "/v1/convai/conversations/messages/text-search", req, opts)
}

// Search runs a semantic search over conversation messages.
func (s *ConversationMessagesService) Search(ctx context.Context, req *MessageSearchRequest, opts ...Option) (*MessageSearchPage, error) {
	return getValue[MessageSearchPage](ctx, s.cfg, "/v1/convai/conversations/messages/smart-search", req, opts)
}

// ConversationTagsService groups the conversation tag endpoints.
type ConversationTagsService struct{ cfg *config }

// ConversationTag is a workspace tag for conversations.
type ConversationTag struct {
	TagID             string `json:"tag_id"`
	WorkspaceID       string `json:"workspace_id"`
	OwnerUserID       string `json:"owner_user_id"`
	Title             string `json:"title"`
	Description       string `json:"description,omitempty"`
	CreatedAtUnixSecs int64  `json:"created_at_unix_secs"`
}

// ConversationTagsPage is a page of tags.
type ConversationTagsPage struct {
	ConversationTags []ConversationTag `json:"conversation_tags"`
	NextCursor       *string           `json:"next_cursor,omitempty"`
	HasMore          bool              `json:"has_more"`
}

// ConversationTagRequest is the body of tag create and update. Title is
// required when creating.
type ConversationTagRequest struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// Assign adds tags to a conversation.
func (s *ConversationTagsService) Assign(ctx context.Context, conversationID string, tagIDs []string, opts ...Option) error {
	body := struct {
		TagIDs []string `json:"tag_ids"`
	}{nonNil(tagIDs)}
	return doNoContent(ctx, s.cfg, http.MethodPost, pathf("/v1/convai/conversations/%s/tags", conversationID), nil, body, opts)
}

// Unassign removes a tag from a conversation.
func (s *ConversationTagsService) Unassign(ctx context.Context, conversationID, tagID string, opts ...Option) error {
	return doNoContent(ctx, s.cfg, http.MethodDelete, pathf("/v1/convai/conversations/%s/tags/%s", conversationID, tagID), nil, nil, opts)
}

// List returns one page of tags. req may be nil.
func (s *ConversationTagsService) List(ctx context.Context, req *PageRequest, opts ...Option) (*ConversationTagsPage, error) {
	return getValue[ConversationTagsPage](ctx, s.cfg, "/v1/convai/tags", req, opts)
}

// ListAll iterates over every tag.
func (s *ConversationTagsService) ListAll(ctx context.Context, req *PageRequest, opts ...Option) iter.Seq2[ConversationTag, error] {
	params := *orEmpty(req)
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[ConversationTag], error) {
		params.Cursor = cursor
		p, err := s.List(ctx, &params, opts...)
		if err != nil {
			return page[ConversationTag]{}, err
		}
		return page[ConversationTag]{items: p.ConversationTags, next: deref(p.NextCursor), hasMore: p.HasMore}, nil
	})
}

// Create creates a tag.
func (s *ConversationTagsService) Create(ctx context.Context, req *ConversationTagRequest, opts ...Option) (*ConversationTag, error) {
	return sendValue[ConversationTag](ctx, s.cfg, http.MethodPost, "/v1/convai/tags", nil, orEmpty(req), opts)
}

// Get returns a tag.
func (s *ConversationTagsService) Get(ctx context.Context, tagID string, opts ...Option) (*ConversationTag, error) {
	return getValue[ConversationTag](ctx, s.cfg, pathf("/v1/convai/tags/%s", tagID), nil, opts)
}

// Delete deletes a tag.
func (s *ConversationTagsService) Delete(ctx context.Context, tagID string, opts ...Option) error {
	return doNoContent(ctx, s.cfg, http.MethodDelete, pathf("/v1/convai/tags/%s", tagID), nil, nil, opts)
}

// Update changes a tag.
func (s *ConversationTagsService) Update(ctx context.Context, tagID string, req *ConversationTagRequest, opts ...Option) (*ConversationTag, error) {
	return sendValue[ConversationTag](ctx, s.cfg, http.MethodPatch, pathf("/v1/convai/tags/%s", tagID), nil, orEmpty(req), opts)
}

// ConversationTopicsService groups the agent topics endpoint.
type ConversationTopicsService struct{ cfg *config }

// TopicsRequest holds the optional parameters of ConversationTopicsService.Get.
type TopicsRequest struct {
	PageSize *int `query:"page_size"`
	// SortBy is "conversations", "sentiment", "success_rate" or "frustration".
	SortBy                    string        `query:"sort_by"`
	SortDirection             SortDirection `query:"sort_direction"`
	FromUnixSecs              *int64        `query:"from_unix_secs"`
	ToUnixSecs                *int64        `query:"to_unix_secs"`
	IncludeEvaluationCriteria *bool         `query:"include_evaluation_criteria"`
	Cursor                    string        `query:"cursor"`
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
func (s *ConversationTopicsService) Get(ctx context.Context, agentID string, req *TopicsRequest, opts ...Option) (*AgentTopics, error) {
	return getValue[AgentTopics](ctx, s.cfg, pathf("/v1/convai/agents/%s/topics", agentID), req, opts)
}
