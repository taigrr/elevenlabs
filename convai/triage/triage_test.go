package triage

import (
	"context"
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/types"
)

const triageTicketJSON = `{"agentqa_ticket_id":"t","workspace_id":"w","owner_user_id":"u","agent_id":"a","needs_clustering":false,"labels":[],"conversation_ids":["c"],"ticket_comments":[],"turn_comments":[{"turn_index":2,"comment":"x","created_at_unix_secs":3}],"status":"open","priority_changes":[],"status_changes":[],"source":"manual","created_at_unix_secs":1,"updated_at_unix_secs":2}`

func endpointCases() []testutil.Endpoint[*Service] {
	return []testutil.Endpoint[*Service]{
		{
			Name: "convai.triage_tickets.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.List(ctx, "a", &ListRequest{Status: StatusOpen, Sources: []Source{SourceQA, SourceAgent}})
			},
			Method: http.MethodGet, Path: "/v1/convai/agents/a/triage-tickets", Query: "sources=qa&sources=agent&status=open",
			RespBody: `{"agent_conversation_tickets":[` + triageTicketJSON + `],"has_more":false}`,
			Check: func(t *testing.T, r any) {
				if p := r.(*Page); len(p.Tickets) != 1 || p.Tickets[0].TurnComments[0].TurnIndex != 2 {
					t.Errorf("page = %+v", p)
				}
			},
		},
		{
			Name: "convai.triage_tickets.create_manual",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.CreateManual(ctx, "a", &CreateManualTicketRequest{QAComment: "fix", Priority: PriorityHigh})
			},
			Method: http.MethodPost, Path: "/v1/convai/agents/a/triage-tickets",
			Body: `{"qa_comment":"fix","priority":"high"}`, RespBody: triageTicketJSON,
		},
		{
			Name: "convai.triage_tickets.list_for_workspace",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.ListForWorkspace(ctx, &ListRequest{AgentID: "a", SortBy: SortByPriority, PageSize: types.Ptr(5)})
			},
			Method: http.MethodGet, Path: "/v1/convai/triage-tickets", Query: "agent_id=a&page_size=5&sort_by=priority",
			RespBody: `{"agent_conversation_tickets":[],"has_more":false}`,
		},
		{
			Name: "convai.triage_tickets.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Create(ctx, &CreateRequest{ConversationID: "c", TurnComments: []TurnCommentInput{{TurnIndex: 1, Comment: "bad"}}})
			},
			Method: http.MethodPost, Path: "/v1/convai/triage-tickets",
			Body: `{"conversation_id":"c","turn_comments":[{"turn_index":1,"comment":"bad"}]}`, RespBody: triageTicketJSON,
		},
		{
			Name: "convai.triage_tickets.list_assignable_users",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.ListAssignableUsers(ctx, "a")
			},
			Method: http.MethodGet, Path: "/v1/convai/agents/a/triage-tickets/assignable-users",
			RespBody: `[{"user_id":"u","email":"e","is_service_account":false,"has_access":true}]`,
			Check: func(t *testing.T, r any) {
				if u := r.([]AssignableUser); len(u) != 1 || !u[0].HasAccess {
					t.Errorf("users = %+v", u)
				}
			},
		},
		{
			Name: "convai.triage_tickets.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Get(ctx, "t")
			},
			Method: http.MethodGet, Path: "/v1/convai/triage-tickets/t", RespBody: triageTicketJSON,
		},
		{
			Name: "convai.triage_tickets.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return nil, c.Delete(ctx, "t")
			},
			Method: http.MethodDelete, Path: "/v1/convai/triage-tickets/t",
		},
		{
			Name: "convai.triage_tickets.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Update(ctx, "t", &UpdateRequest{Status: StatusResolved, AssigneeUserID: &NullString{Null: true}, Priority: &NullString{Value: "low"}})
			},
			Method: http.MethodPatch, Path: "/v1/convai/triage-tickets/t",
			Body: `{"status":"resolved","assignee_user_id":null,"priority":"low"}`, RespBody: triageTicketJSON,
		},
		{
			Name: "convai.triage_tickets.add_comment",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.AddComment(ctx, "t", "note")
			},
			Method: http.MethodPost, Path: "/v1/convai/triage-tickets/t/comments",
			Body: `{"comment":"note"}`, RespBody: triageTicketJSON,
		},
		{
			Name: "convai.triage_tickets.merge",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Merge(ctx, "t", []string{"t2", "t3"})
			},
			Method: http.MethodPost, Path: "/v1/convai/triage-tickets/t/merge",
			Body: `{"source_ticket_ids":["t2","t3"]}`, RespBody: triageTicketJSON,
		},
		{
			Name: "convai.triage_tickets.add_turn_comment",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.AddTurnComment(ctx, "t", TurnCommentInput{TurnIndex: 0, Comment: "x"})
			},
			Method: http.MethodPost, Path: "/v1/convai/triage-tickets/t/turn-comments",
			Body: `{"turn_index":0,"comment":"x"}`, RespBody: triageTicketJSON,
		},
	}
}

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, endpointCases())
}
