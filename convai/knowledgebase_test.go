package convai

import (
	"context"
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/convai/knowledgebase"
	"github.com/taigrr/elevenlabs/v2/internal/testutil"
)

func knowledgeBaseEndpointCases() []testutil.Endpoint[*Service] {
	created := `{"id":"d","name":"n","folder_path":[{"id":"f"}]}`
	ragIndex := `{"id":"r","model":"e5_mistral_7b_instruct","status":"succeeded","progress_percentage":100,"document_model_index_usage":{"used_bytes":5}}`
	return []testutil.Endpoint[*Service]{
		{
			Name: "convai.add_to_knowledge_base",
			Call: func(ctx context.Context, c *Service) (any, error) {
				f := testutil.File("faq.txt", "faq")
				return c.AddToKnowledgeBase(ctx, &knowledgebase.AddRequest{AgentID: "a", Name: "FAQ", File: &f})
			},
			Method: http.MethodPost, Path: "/v1/convai/knowledge-base", Query: "agent_id=a",
			Form: map[string][]string{"name": {"FAQ"}, "file": {"file:faq.txt:faq"}}, RespBody: created,
		},
		{
			Name:   "convai.rag_index_overview",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.RAGIndexOverview(ctx) },
			Method: http.MethodGet, Path: "/v1/convai/knowledge-base/rag-index",
			RespBody: `{"total_used_bytes":1,"total_max_bytes":10,"models":[{"model":"e5_mistral_7b_instruct","used_bytes":1}]}`,
			Check: func(t *testing.T, r any) {
				if o := r.(*knowledgebase.RAGIndexOverview); o.TotalMaxBytes != 10 || o.Models[0].Model != knowledgebase.EmbeddingModelE5Mistral7BInstruct {
					t.Errorf("overview = %+v", o)
				}
			},
		},
		{
			Name: "convai.get_document_rag_indexes",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.GetDocumentRAGIndexes(ctx, "d")
			},
			Method: http.MethodGet, Path: "/v1/convai/knowledge-base/d/rag-index", RespBody: `{"indexes":[` + ragIndex + `]}`,
			Check: func(t *testing.T, r any) {
				if i := r.(*knowledgebase.RAGIndexes).Indexes[0]; i.Status != knowledgebase.RAGIndexStatusSucceeded || i.DocumentModelIndexUsage.UsedBytes != 5 {
					t.Errorf("index = %+v", i)
				}
			},
		},
		{
			Name: "convai.delete_document_rag_index",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.DeleteDocumentRAGIndex(ctx, "d", "r")
			},
			Method: http.MethodDelete, Path: "/v1/convai/knowledge-base/d/rag-index/r", RespBody: ragIndex,
		},
	}
}
