package elevenlabs

import (
	"context"
	"net/http"
	"strings"
	"testing"
)

func convaiEndpointCases() []endpointCase {
	cases := convaiAgentEndpointCases()
	cases = append(cases, convaiConversationEndpointCases()...)
	cases = append(cases, convaiKnowledgeBaseEndpointCases()...)
	cases = append(cases, convaiToolEndpointCases()...)
	cases = append(cases, convaiSettingsEndpointCases()...)
	return cases
}

func TestConvAIListAll(t *testing.T) {
	ctx := context.Background()
	t.Run("agents", func(t *testing.T) {
		client, ts := newTestServer(t, sequence(
			respond(http.StatusOK, contentTypeJSON, `{"agents":[{"agent_id":"a1"}],"next_cursor":"c2","has_more":true}`),
			respond(http.StatusOK, contentTypeJSON, `{"agents":[{"agent_id":"a2"}],"has_more":false}`),
		))
		var ids []string
		for agent, err := range client.ConversationalAI.Agents.ListAll(ctx, &AgentsListRequest{PageSize: Ptr(1)}) {
			if err != nil {
				t.Fatal(err)
			}
			ids = append(ids, agent.AgentID)
		}
		if strings.Join(ids, ",") != "a1,a2" {
			t.Errorf("ids = %v", ids)
		}
		if q := ts.all()[1].RawQuery; q != "cursor=c2&page_size=1" {
			t.Errorf("second query = %q", q)
		}
	})
	t.Run("chunks without has_more", func(t *testing.T) {
		client, ts := newTestServer(t, sequence(
			respond(http.StatusOK, contentTypeJSON, `{"chunks":[{"id":"k1"}],"next_cursor":"n"}`),
			respond(http.StatusOK, contentTypeJSON, `{"chunks":[{"id":"k2"}]}`),
		))
		var ids []string
		for chunk, err := range client.ConversationalAI.KnowledgeBase.Documents.Chunks.ListAll(ctx, "d", &ChunkListRequest{EmbeddingModel: EmbeddingModelE5Mistral7BInstruct}) {
			if err != nil {
				t.Fatal(err)
			}
			ids = append(ids, chunk.ID)
		}
		if strings.Join(ids, ",") != "k1,k2" || len(ts.all()) != 2 {
			t.Errorf("ids = %v", ids)
		}
	})
	t.Run("error stops iteration", func(t *testing.T) {
		client, _ := newTestServer(t, respond(http.StatusNotFound, contentTypeJSON, `{"detail":"missing"}`))
		n := 0
		for _, err := range client.ConversationalAI.Conversations.ListAll(ctx, nil) {
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
