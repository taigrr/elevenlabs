package elevenlabs

import (
	"context"
	"encoding/json"
	"iter"
	"net/http"
)

// TriageTicketStatus is the state of a triage ticket.
type TriageTicketStatus string

// TriageTicketStatus values.
const (
	TriageTicketOpen       TriageTicketStatus = "open"
	TriageTicketInProgress TriageTicketStatus = "in_progress"
	TriageTicketResolved   TriageTicketStatus = "resolved"
	TriageTicketCancelled  TriageTicketStatus = "cancelled"
	TriageTicketMerged     TriageTicketStatus = "merged"
)

// TriageTicketPriority is the urgency of a triage ticket.
type TriageTicketPriority string

// TriageTicketPriority values.
const (
	TriageTicketPriorityLow    TriageTicketPriority = "low"
	TriageTicketPriorityMedium TriageTicketPriority = "medium"
	TriageTicketPriorityHigh   TriageTicketPriority = "high"
	TriageTicketPriorityUrgent TriageTicketPriority = "urgent"
)

// TriageTicketSource is how a triage ticket was raised.
type TriageTicketSource string

// TriageTicketSource values.
const (
	TriageTicketSourceQA     TriageTicketSource = "qa"
	TriageTicketSourceAgent  TriageTicketSource = "agent"
	TriageTicketSourceManual TriageTicketSource = "manual"
)

// TriageTicketSortBy orders triage ticket lists.
type TriageTicketSortBy string

// TriageTicketSortBy values.
const (
	TriageTicketSortByCreatedAt TriageTicketSortBy = "created_at"
	TriageTicketSortByPriority  TriageTicketSortBy = "priority"
)

// TriageTicketIssueType classifies a triage ticket.
type TriageTicketIssueType string

// TriageTicketIssueType values.
const (
	TriageIssueKnowledgeGap          TriageTicketIssueType = "knowledge_gap"
	TriageIssueIncorrectInformation  TriageTicketIssueType = "incorrect_information"
	TriageIssueDocumentationGap      TriageTicketIssueType = "documentation_gap"
	TriageIssueProductFeedback       TriageTicketIssueType = "product_feedback"
	TriageIssuePlatformBug           TriageTicketIssueType = "platform_bug"
	TriageIssueToolIssue             TriageTicketIssueType = "tool_issue"
	TriageIssueMissingTool           TriageTicketIssueType = "missing_tool"
	TriageIssueUnnecessaryEscalation TriageTicketIssueType = "unnecessary_escalation"
	TriageIssueWrongAction           TriageTicketIssueType = "wrong_action"
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

// TriageTicket is an agent conversation (QA) ticket.
type TriageTicket struct {
	AgentQATicketID    string                `json:"agentqa_ticket_id"`
	WorkspaceID        string                `json:"workspace_id"`
	OwnerUserID        string                `json:"owner_user_id"`
	AgentID            string                `json:"agent_id"`
	NeedsClustering    bool                  `json:"needs_clustering"`
	Title              *string               `json:"title,omitempty"`
	IssueType          TriageTicketIssueType `json:"issue_type,omitempty"`
	Labels             []string              `json:"labels"`
	ConversationIDs    []string              `json:"conversation_ids"`
	FirstSeenUnixSecs  *int64                `json:"first_seen_unix_secs,omitempty"`
	LastSeenUnixSecs   *int64                `json:"last_seen_unix_secs,omitempty"`
	QAComment          *string               `json:"qa_comment,omitempty"`
	TicketComments     []TicketComment       `json:"ticket_comments"`
	TurnComments       []TurnComment         `json:"turn_comments"`
	Status             TriageTicketStatus    `json:"status"`
	Priority           TriageTicketPriority  `json:"priority,omitempty"`
	PriorityChanges    json.RawMessage       `json:"priority_changes,omitempty"`
	StatusChanges      json.RawMessage       `json:"status_changes,omitempty"`
	MergedIntoTicketID *string               `json:"merged_into_ticket_id,omitempty"`
	Source             TriageTicketSource    `json:"source"`
	AssigneeUserID     *string               `json:"assignee_user_id,omitempty"`
	CreatedAtUnixSecs  int64                 `json:"created_at_unix_secs"`
	UpdatedAtUnixSecs  int64                 `json:"updated_at_unix_secs"`
	SearchMatch        json.RawMessage       `json:"search_match,omitempty"`
}

// TriageTicketsPage is a page of triage tickets.
type TriageTicketsPage struct {
	Tickets    []TriageTicket `json:"agent_conversation_tickets"`
	NextCursor *string        `json:"next_cursor,omitempty"`
	HasMore    bool           `json:"has_more"`
}

// TriageTicketListRequest holds the optional filters of the triage ticket
// list endpoints. AgentID applies only to ListForWorkspace and
// ConversationID only to List.
type TriageTicketListRequest struct {
	PageSize           *int                   `query:"page_size"`
	AgentID            string                 `query:"agent_id"`
	ConversationID     string                 `query:"conversation_id"`
	Status             TriageTicketStatus     `query:"status"`
	Sources            []TriageTicketSource   `query:"sources"`
	Priorities         []TriageTicketPriority `query:"priorities"`
	SortBy             TriageTicketSortBy     `query:"sort_by"`
	OwnerUserID        string                 `query:"owner_user_id"`
	AssigneeUserID     string                 `query:"assignee_user_id"`
	IssueType          TriageTicketIssueType  `query:"issue_type"`
	Label              string                 `query:"label"`
	MergedIntoTicketID string                 `query:"merged_into_ticket_id"`
	Search             string                 `query:"search"`
	Cursor             string                 `query:"cursor"`
}

// CreateManualTicketRequest raises a ticket on an agent without a conversation.
type CreateManualTicketRequest struct {
	Title     string               `json:"title,omitempty"`
	QAComment string               `json:"qa_comment"`
	Priority  TriageTicketPriority `json:"priority,omitempty"`
}

// TurnCommentInput is a turn-level comment to add to a ticket.
type TurnCommentInput struct {
	TurnIndex int    `json:"turn_index"`
	Comment   string `json:"comment"`
}

// CreateTriageTicketRequest raises a ticket about a conversation.
type CreateTriageTicketRequest struct {
	ConversationID string               `json:"conversation_id"`
	Title          string               `json:"title,omitempty"`
	QAComment      string               `json:"qa_comment,omitempty"`
	TurnComments   []TurnCommentInput   `json:"turn_comments,omitempty"`
	Priority       TriageTicketPriority `json:"priority,omitempty"`
}

// UpdateTriageTicketRequest changes a ticket. Nil fields are left
// unchanged; use NullString to unassign or clear the priority.
type UpdateTriageTicketRequest struct {
	Title          *string            `json:"title,omitempty"`
	Status         TriageTicketStatus `json:"status,omitempty"`
	AssigneeUserID *NullString        `json:"assignee_user_id,omitempty"`
	Priority       *NullString        `json:"priority,omitempty"`
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

// TriageTicketsService groups the triage ticket endpoints.
type TriageTicketsService struct{ cfg *config }

// List returns one page of an agent's tickets. req may be nil.
func (s *TriageTicketsService) List(ctx context.Context, agentID string, req *TriageTicketListRequest, opts ...Option) (*TriageTicketsPage, error) {
	return getValue[TriageTicketsPage](ctx, s.cfg, pathf("/v1/convai/agents/%s/triage-tickets", agentID), req, opts)
}

// ListAll iterates over every ticket of an agent.
func (s *TriageTicketsService) ListAll(ctx context.Context, agentID string, req *TriageTicketListRequest, opts ...Option) iter.Seq2[TriageTicket, error] {
	return s.iterate(ctx, req, func(ctx context.Context, p *TriageTicketListRequest) (*TriageTicketsPage, error) {
		return s.List(ctx, agentID, p, opts...)
	})
}

// CreateManual raises a ticket on an agent.
func (s *TriageTicketsService) CreateManual(ctx context.Context, agentID string, req *CreateManualTicketRequest, opts ...Option) (*TriageTicket, error) {
	return sendValue[TriageTicket](ctx, s.cfg, http.MethodPost, pathf("/v1/convai/agents/%s/triage-tickets", agentID), nil, orEmpty(req), opts)
}

// ListForWorkspace returns one page of the workspace's tickets. req may be nil.
func (s *TriageTicketsService) ListForWorkspace(ctx context.Context, req *TriageTicketListRequest, opts ...Option) (*TriageTicketsPage, error) {
	return getValue[TriageTicketsPage](ctx, s.cfg, "/v1/convai/triage-tickets", req, opts)
}

// ListAllForWorkspace iterates over every ticket of the workspace.
func (s *TriageTicketsService) ListAllForWorkspace(ctx context.Context, req *TriageTicketListRequest, opts ...Option) iter.Seq2[TriageTicket, error] {
	return s.iterate(ctx, req, func(ctx context.Context, p *TriageTicketListRequest) (*TriageTicketsPage, error) {
		return s.ListForWorkspace(ctx, p, opts...)
	})
}

func (s *TriageTicketsService) iterate(ctx context.Context, req *TriageTicketListRequest, fetch func(context.Context, *TriageTicketListRequest) (*TriageTicketsPage, error)) iter.Seq2[TriageTicket, error] {
	params := *orEmpty(req)
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[TriageTicket], error) {
		params.Cursor = cursor
		p, err := fetch(ctx, &params)
		if err != nil {
			return page[TriageTicket]{}, err
		}
		next := deref(p.NextCursor)
		return page[TriageTicket]{items: p.Tickets, next: next, hasMore: p.HasMore && next != ""}, nil
	})
}

// Create raises a ticket about a conversation, or adds to the
// conversation's open ticket.
func (s *TriageTicketsService) Create(ctx context.Context, req *CreateTriageTicketRequest, opts ...Option) (*TriageTicket, error) {
	return sendValue[TriageTicket](ctx, s.cfg, http.MethodPost, "/v1/convai/triage-tickets", nil, orEmpty(req), opts)
}

// ListAssignableUsers lists the workspace members that can be assigned to
// an agent's tickets.
func (s *TriageTicketsService) ListAssignableUsers(ctx context.Context, agentID string, opts ...Option) ([]AssignableUser, error) {
	out, err := getValue[[]AssignableUser](ctx, s.cfg, pathf("/v1/convai/agents/%s/triage-tickets/assignable-users", agentID), nil, opts)
	if err != nil {
		return nil, err
	}
	return *out, nil
}

// Get returns a ticket.
func (s *TriageTicketsService) Get(ctx context.Context, ticketID string, opts ...Option) (*TriageTicket, error) {
	return getValue[TriageTicket](ctx, s.cfg, pathf("/v1/convai/triage-tickets/%s", ticketID), nil, opts)
}

// Delete deletes a ticket.
func (s *TriageTicketsService) Delete(ctx context.Context, ticketID string, opts ...Option) error {
	return doNoContent(ctx, s.cfg, http.MethodDelete, pathf("/v1/convai/triage-tickets/%s", ticketID), nil, nil, opts)
}

// Update changes a ticket.
func (s *TriageTicketsService) Update(ctx context.Context, ticketID string, req *UpdateTriageTicketRequest, opts ...Option) (*TriageTicket, error) {
	return sendValue[TriageTicket](ctx, s.cfg, http.MethodPatch, pathf("/v1/convai/triage-tickets/%s", ticketID), nil, orEmpty(req), opts)
}

// AddComment adds a ticket-level comment.
func (s *TriageTicketsService) AddComment(ctx context.Context, ticketID, comment string, opts ...Option) (*TriageTicket, error) {
	body := struct {
		Comment string `json:"comment"`
	}{comment}
	return sendValue[TriageTicket](ctx, s.cfg, http.MethodPost, pathf("/v1/convai/triage-tickets/%s/comments", ticketID), nil, body, opts)
}

// Merge folds the source tickets into this ticket.
func (s *TriageTicketsService) Merge(ctx context.Context, ticketID string, sourceTicketIDs []string, opts ...Option) (*TriageTicket, error) {
	body := struct {
		SourceTicketIDs []string `json:"source_ticket_ids"`
	}{sourceTicketIDs}
	return sendValue[TriageTicket](ctx, s.cfg, http.MethodPost, pathf("/v1/convai/triage-tickets/%s/merge", ticketID), nil, body, opts)
}

// AddTurnComment adds a comment on one transcript turn.
func (s *TriageTicketsService) AddTurnComment(ctx context.Context, ticketID string, req TurnCommentInput, opts ...Option) (*TriageTicket, error) {
	return sendValue[TriageTicket](ctx, s.cfg, http.MethodPost, pathf("/v1/convai/triage-tickets/%s/turn-comments", ticketID), nil, req, opts)
}
