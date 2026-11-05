package triage

import (
	"context"
	"encoding/json"
	"iter"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/stream"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
)

// Status is the state of a triage ticket.
type Status string

// Status values.
const (
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusResolved   Status = "resolved"
	StatusCancelled  Status = "cancelled"
	StatusMerged     Status = "merged"
)

// Priority is the urgency of a triage ticket.
type Priority string

// Priority values.
const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

// Source is how a triage ticket was raised.
type Source string

// Source values.
const (
	SourceQA     Source = "qa"
	SourceAgent  Source = "agent"
	SourceManual Source = "manual"
)

// SortBy orders triage ticket lists.
type SortBy string

// SortBy values.
const (
	SortByCreatedAt SortBy = "created_at"
	SortByPriority  SortBy = "priority"
)

// IssueType classifies a triage ticket.
type IssueType string

// IssueType values.
const (
	IssueKnowledgeGap          IssueType = "knowledge_gap"
	IssueIncorrectInformation  IssueType = "incorrect_information"
	IssueDocumentationGap      IssueType = "documentation_gap"
	IssueProductFeedback       IssueType = "product_feedback"
	IssuePlatformBug           IssueType = "platform_bug"
	IssueToolIssue             IssueType = "tool_issue"
	IssueMissingTool           IssueType = "missing_tool"
	IssueUnnecessaryEscalation IssueType = "unnecessary_escalation"
	IssueWrongAction           IssueType = "wrong_action"
)

// TicketComment is a comment on a triage ticket.
type TicketComment struct {
	Comment           string  `json:"comment"`
	CreatedAtUnixSecs int64   `json:"created_at_unix_secs"`
	OwnerUserID       *string `json:"owner_user_id,omitempty"`
}

// TurnComment is a comment on one transcript turn.
type TurnComment struct {
	TurnIndex         int     `json:"turn_index"`
	Comment           string  `json:"comment"`
	CreatedAtUnixSecs int64   `json:"created_at_unix_secs,omitempty"`
	OwnerUserID       *string `json:"owner_user_id,omitempty"`
}

// Ticket is an agent conversation (QA) ticket.
type Ticket struct {
	AgentQATicketID    string          `json:"agentqa_ticket_id"`
	WorkspaceID        string          `json:"workspace_id"`
	OwnerUserID        string          `json:"owner_user_id"`
	AgentID            string          `json:"agent_id"`
	NeedsClustering    bool            `json:"needs_clustering"`
	Title              *string         `json:"title,omitempty"`
	IssueType          IssueType       `json:"issue_type,omitempty"`
	Labels             []string        `json:"labels"`
	ConversationIDs    []string        `json:"conversation_ids"`
	FirstSeenUnixSecs  *int64          `json:"first_seen_unix_secs,omitempty"`
	LastSeenUnixSecs   *int64          `json:"last_seen_unix_secs,omitempty"`
	QAComment          *string         `json:"qa_comment,omitempty"`
	TicketComments     []TicketComment `json:"ticket_comments"`
	TurnComments       []TurnComment   `json:"turn_comments"`
	Status             Status          `json:"status"`
	Priority           Priority        `json:"priority,omitempty"`
	PriorityChanges    json.RawMessage `json:"priority_changes,omitempty"`
	StatusChanges      json.RawMessage `json:"status_changes,omitempty"`
	MergedIntoTicketID *string         `json:"merged_into_ticket_id,omitempty"`
	Source             Source          `json:"source"`
	AssigneeUserID     *string         `json:"assignee_user_id,omitempty"`
	CreatedAtUnixSecs  int64           `json:"created_at_unix_secs"`
	UpdatedAtUnixSecs  int64           `json:"updated_at_unix_secs"`
	SearchMatch        json.RawMessage `json:"search_match,omitempty"`
}

// Page is a page of triage tickets.
type Page struct {
	Tickets    []Ticket `json:"agent_conversation_tickets"`
	NextCursor *string  `json:"next_cursor,omitempty"`
	HasMore    bool     `json:"has_more"`
}

// ListRequest holds the optional filters of the triage ticket
// list endpoints. AgentID applies only to ListForWorkspace and
// ConversationID only to List.
type ListRequest struct {
	PageSize           *int       `query:"page_size"`
	AgentID            string     `query:"agent_id"`
	ConversationID     string     `query:"conversation_id"`
	Status             Status     `query:"status"`
	Sources            []Source   `query:"sources"`
	Priorities         []Priority `query:"priorities"`
	SortBy             SortBy     `query:"sort_by"`
	OwnerUserID        string     `query:"owner_user_id"`
	AssigneeUserID     string     `query:"assignee_user_id"`
	IssueType          IssueType  `query:"issue_type"`
	Label              string     `query:"label"`
	MergedIntoTicketID string     `query:"merged_into_ticket_id"`
	Search             string     `query:"search"`
	Cursor             string     `query:"cursor"`
}

// CreateManualTicketRequest raises a ticket on an agent without a conversation.
type CreateManualTicketRequest struct {
	Title     string   `json:"title,omitempty"`
	QAComment string   `json:"qa_comment"`
	Priority  Priority `json:"priority,omitempty"`
}

// TurnCommentInput is a turn-level comment to add to a ticket.
type TurnCommentInput struct {
	TurnIndex int    `json:"turn_index"`
	Comment   string `json:"comment"`
}

// CreateRequest raises a ticket about a conversation.
type CreateRequest struct {
	ConversationID string             `json:"conversation_id"`
	Title          string             `json:"title,omitempty"`
	QAComment      string             `json:"qa_comment,omitempty"`
	TurnComments   []TurnCommentInput `json:"turn_comments,omitempty"`
	Priority       Priority           `json:"priority,omitempty"`
}

// UpdateRequest changes a ticket. Nil fields are left
// unchanged; use NullString to unassign or clear the priority.
type UpdateRequest struct {
	Title          *string     `json:"title,omitempty"`
	Status         Status      `json:"status,omitempty"`
	AssigneeUserID *NullString `json:"assignee_user_id,omitempty"`
	Priority       *NullString `json:"priority,omitempty"`
}

// NullString is a string that marshals to JSON null when Null is true.
type NullString struct {
	Value string
	Null  bool
}

// MarshalJSON implements json.Marshaler.
func (n NullString) MarshalJSON() ([]byte, error) {
	if n.Null {
		return []byte("null"), nil
	}
	return json.Marshal(n.Value)
}

// AssignableUser is a workspace member that can be assigned to tickets.
type AssignableUser struct {
	UserID           string  `json:"user_id"`
	Email            string  `json:"email"`
	FirstName        *string `json:"first_name,omitempty"`
	IsServiceAccount bool    `json:"is_service_account"`
	HasAccess        bool    `json:"has_access"`
}

// Service groups the triage ticket endpoints.
type Service struct{ cfg *requestconfig.Config }

// List returns one page of an agent's tickets. req may be nil.
func (s *Service) List(ctx context.Context, agentID string, req *ListRequest, opts ...option.RequestOption) (*Page, error) {
	return transport.GetValue[Page](ctx, s.cfg, transport.Pathf("/v1/convai/agents/%s/triage-tickets", agentID), req, opts)
}

// ListAll iterates over every ticket of an agent.
func (s *Service) ListAll(ctx context.Context, agentID string, req *ListRequest, opts ...option.RequestOption) iter.Seq2[Ticket, error] {
	return s.iterate(ctx, req, func(ctx context.Context, p *ListRequest) (*Page, error) {
		return s.List(ctx, agentID, p, opts...)
	})
}

// CreateManual raises a ticket on an agent.
func (s *Service) CreateManual(ctx context.Context, agentID string, req *CreateManualTicketRequest, opts ...option.RequestOption) (*Ticket, error) {
	return transport.SendValue[Ticket](ctx, s.cfg, http.MethodPost, transport.Pathf("/v1/convai/agents/%s/triage-tickets", agentID), nil, transport.OrEmpty(req), opts)
}

// ListForWorkspace returns one page of the workspace's tickets. req may be nil.
func (s *Service) ListForWorkspace(ctx context.Context, req *ListRequest, opts ...option.RequestOption) (*Page, error) {
	return transport.GetValue[Page](ctx, s.cfg, "/v1/convai/triage-tickets", req, opts)
}

// ListAllForWorkspace iterates over every ticket of the workspace.
func (s *Service) ListAllForWorkspace(ctx context.Context, req *ListRequest, opts ...option.RequestOption) iter.Seq2[Ticket, error] {
	return s.iterate(ctx, req, func(ctx context.Context, p *ListRequest) (*Page, error) {
		return s.ListForWorkspace(ctx, p, opts...)
	})
}

func (s *Service) iterate(ctx context.Context, req *ListRequest, fetch func(context.Context, *ListRequest) (*Page, error)) iter.Seq2[Ticket, error] {
	params := *transport.OrEmpty(req)
	return stream.Paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (stream.Page[Ticket], error) {
		params.Cursor = cursor
		p, err := fetch(ctx, &params)
		if err != nil {
			return stream.Page[Ticket]{}, err
		}
		next := transport.Deref(p.NextCursor)
		return stream.Page[Ticket]{Items: p.Tickets, Next: next, HasMore: p.HasMore && next != ""}, nil
	})
}

// Create raises a ticket about a conversation, or adds to the
// conversation's open ticket.
func (s *Service) Create(ctx context.Context, req *CreateRequest, opts ...option.RequestOption) (*Ticket, error) {
	return transport.SendValue[Ticket](ctx, s.cfg, http.MethodPost, "/v1/convai/triage-tickets", nil, transport.OrEmpty(req), opts)
}

// ListAssignableUsers lists the workspace members that can be assigned to
// an agent's tickets.
func (s *Service) ListAssignableUsers(ctx context.Context, agentID string, opts ...option.RequestOption) ([]AssignableUser, error) {
	out, err := transport.GetValue[[]AssignableUser](ctx, s.cfg, transport.Pathf("/v1/convai/agents/%s/triage-tickets/assignable-users", agentID), nil, opts)
	if err != nil {
		return nil, err
	}
	return *out, nil
}

// Get returns a ticket.
func (s *Service) Get(ctx context.Context, ticketID string, opts ...option.RequestOption) (*Ticket, error) {
	return transport.GetValue[Ticket](ctx, s.cfg, transport.Pathf("/v1/convai/triage-tickets/%s", ticketID), nil, opts)
}

// Delete deletes a ticket.
func (s *Service) Delete(ctx context.Context, ticketID string, opts ...option.RequestOption) error {
	return transport.NoContent(ctx, s.cfg, http.MethodDelete, transport.Pathf("/v1/convai/triage-tickets/%s", ticketID), nil, nil, opts)
}

// Update changes a ticket.
func (s *Service) Update(ctx context.Context, ticketID string, req *UpdateRequest, opts ...option.RequestOption) (*Ticket, error) {
	return transport.SendValue[Ticket](ctx, s.cfg, http.MethodPatch, transport.Pathf("/v1/convai/triage-tickets/%s", ticketID), nil, transport.OrEmpty(req), opts)
}

// AddComment adds a ticket-level comment.
func (s *Service) AddComment(ctx context.Context, ticketID, comment string, opts ...option.RequestOption) (*Ticket, error) {
	body := struct {
		Comment string `json:"comment"`
	}{comment}
	return transport.SendValue[Ticket](ctx, s.cfg, http.MethodPost, transport.Pathf("/v1/convai/triage-tickets/%s/comments", ticketID), nil, body, opts)
}

// Merge folds the source tickets into this ticket.
func (s *Service) Merge(ctx context.Context, ticketID string, sourceTicketIDs []string, opts ...option.RequestOption) (*Ticket, error) {
	body := struct {
		SourceTicketIDs []string `json:"source_ticket_ids"`
	}{sourceTicketIDs}
	return transport.SendValue[Ticket](ctx, s.cfg, http.MethodPost, transport.Pathf("/v1/convai/triage-tickets/%s/merge", ticketID), nil, body, opts)
}

// AddTurnComment adds a comment on one transcript turn.
func (s *Service) AddTurnComment(ctx context.Context, ticketID string, req TurnCommentInput, opts ...option.RequestOption) (*Ticket, error) {
	return transport.SendValue[Ticket](ctx, s.cfg, http.MethodPost, transport.Pathf("/v1/convai/triage-tickets/%s/turn-comments", ticketID), nil, req, opts)
}

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	cfg := requestconfig.Build(opts...)
	return &Service{cfg: cfg}
}
