package convai

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/taigrr/elevenlabs/v2/convai/agents"
	"github.com/taigrr/elevenlabs/v2/convai/knowledgebase"
	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/types"
)

func TestEndpoints(t *testing.T) {
	cases := settingsEndpointCases()
	cases = append(cases, miscEndpointCases()...)
	cases = append(cases, knowledgeBaseEndpointCases()...)
	testutil.RunEndpoints(t, NewService, cases)
}

func TestListAll(t *testing.T) {
	ctx := context.Background()
	t.Run("agents", func(t *testing.T) {
		client, ts := newTestServer(t, testutil.Sequence(
			testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"agents":[{"agent_id":"a1"}],"next_cursor":"c2","has_more":true}`),
			testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"agents":[{"agent_id":"a2"}],"has_more":false}`),
		))
		var ids []string
		for agent, err := range client.Agents.ListAll(ctx, &agents.ListRequest{PageSize: types.Ptr(1)}) {
			if err != nil {
				t.Fatal(err)
			}
			ids = append(ids, agent.AgentID)
		}
		if strings.Join(ids, ",") != "a1,a2" {
			t.Errorf("ids = %v", ids)
		}
		if q := ts.All()[1].RawQuery; q != "cursor=c2&page_size=1" {
			t.Errorf("second query = %q", q)
		}
	})
	t.Run("chunks without has_more", func(t *testing.T) {
		client, ts := newTestServer(t, testutil.Sequence(
			testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"chunks":[{"id":"k1"}],"next_cursor":"n"}`),
			testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"chunks":[{"id":"k2"}]}`),
		))
		var ids []string
		for chunk, err := range client.KnowledgeBase.Documents.Chunks.ListAll(ctx, "d", &knowledgebase.ChunkListRequest{EmbeddingModel: knowledgebase.EmbeddingModelE5Mistral7BInstruct}) {
			if err != nil {
				t.Fatal(err)
			}
			ids = append(ids, chunk.ID)
		}
		if strings.Join(ids, ",") != "k1,k2" || len(ts.All()) != 2 {
			t.Errorf("ids = %v", ids)
		}
	})
	t.Run("error stops iteration", func(t *testing.T) {
		client, _ := newTestServer(t, testutil.Respond(http.StatusNotFound, transport.ContentTypeJSON, `{"detail":"missing"}`))
		n := 0
		for _, err := range client.Conversations.ListAll(ctx, nil) {
			n++
			if err == nil {
				t.Fatal("expected error")
			}
		}
		if n != 1 {
			t.Errorf("yields = %d", n)
		}
	})
}

func TestListAllTelephonyTriageUsers(t *testing.T) {
	ctx := context.Background()
	t.Run("batch calls", func(t *testing.T) {
		client, ts := newTestServer(t, testutil.Sequence(
			testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"batch_calls":[{"id":"b1"}],"next_doc":"d2","has_more":true}`),
			testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"batch_calls":[{"id":"b2"}],"has_more":false}`),
		))
		var ids []string
		for b, err := range client.BatchCalls.ListAll(ctx, nil) {
			if err != nil {
				t.Fatal(err)
			}
			ids = append(ids, b.ID)
		}
		if len(ids) != 2 || ids[1] != "b2" || ts.All()[1].RawQuery != "last_doc=d2" {
			t.Errorf("ids = %v, query = %q", ids, ts.All()[1].RawQuery)
		}
	})
	t.Run("triage tickets", func(t *testing.T) {
		client, ts := newTestServer(t, testutil.Sequence(
			testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"agent_conversation_tickets":[{"agentqa_ticket_id":"t1"}],"next_cursor":"c","has_more":true}`),
			testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"agent_conversation_tickets":[{"agentqa_ticket_id":"t2"}],"has_more":false}`),
		))
		n := 0
		for _, err := range client.TriageTickets.ListAll(ctx, "a", nil) {
			if err != nil {
				t.Fatal(err)
			}
			n++
		}
		if n != 2 || ts.All()[1].Path != "/v1/convai/agents/a/triage-tickets" {
			t.Errorf("n = %d", n)
		}
	})
	t.Run("phone numbers and users", func(t *testing.T) {
		client, _ := newTestServer(t, testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"phone_numbers":[{"phone_number_id":"p"}],"users":[{"user_id":"u"}],"agent_conversation_tickets":[],"has_more":false}`))
		for p, err := range client.PhoneNumbers.ListAll(ctx, nil) {
			if err != nil || p.PhoneNumberID != "p" {
				t.Fatal(p, err)
			}
		}
		for u, err := range client.Users.ListAll(ctx, nil) {
			if err != nil || u.UserID != "u" {
				t.Fatal(u, err)
			}
		}
		for _, err := range client.TriageTickets.ListAllForWorkspace(ctx, nil) {
			t.Fatal("unexpected item", err)
		}
	})
}

func newTestServer(t *testing.T, handler http.HandlerFunc) (*Service, *testutil.Server) {
	t.Helper()
	ts, opts := testutil.NewServer(t, handler)
	return NewService(opts...), ts
}
