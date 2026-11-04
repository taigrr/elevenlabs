package elevenlabs

import (
	"context"
	"net/http"
	"testing"
)

func convaiKnowledgeBaseEndpointCases() []endpointCase {
	docJSON := `{"id":"d","name":"n","type":"url","metadata":{"created_at_unix_secs":1,"last_updated_at_unix_secs":2,"size_bytes":3},"supported_usages":["prompt"],"access_info":{"is_creator":true,"creator_name":"c","creator_email":"e","role":"admin"},"url":"https://x","extracted_inner_html":"<p/>"}`
	checkDoc := func(t *testing.T, r any) {
		if d := r.(*KnowledgeBaseDocument); d.Type != KnowledgeBaseDocumentTypeURL || d.URL != "https://x" || d.Metadata.SizeBytes != 3 {
			t.Errorf("document = %+v", d)
		}
	}
	created := `{"id":"d","name":"n","folder_path":[{"id":"f"}]}`
	ragIndex := `{"id":"r","model":"e5_mistral_7b_instruct","status":"succeeded","progress_percentage":100,"document_model_index_usage":{"used_bytes":5}}`
	return []endpointCase{
		{
			name: "convai.add_to_knowledge_base",
			call: func(ctx context.Context, c *Client) (any, error) {
				f := testFile("faq.txt", "faq")
				return c.ConversationalAI.AddToKnowledgeBase(ctx, &KnowledgeBaseAddRequest{AgentID: "a", Name: "FAQ", File: &f})
			},
			method: http.MethodPost, path: "/v1/convai/knowledge-base", query: "agent_id=a",
			form: map[string][]string{"name": {"FAQ"}, "file": {"file:faq.txt:faq"}}, respBody: created,
		},
		{
			name:   "convai.rag_index_overview",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.ConversationalAI.RAGIndexOverview(ctx) },
			method: http.MethodGet, path: "/v1/convai/knowledge-base/rag-index",
			respBody: `{"total_used_bytes":1,"total_max_bytes":10,"models":[{"model":"e5_mistral_7b_instruct","used_bytes":1}]}`,
			check: func(t *testing.T, r any) {
				if o := r.(*RAGIndexOverview); o.TotalMaxBytes != 10 || o.Models[0].Model != EmbeddingModelE5Mistral7BInstruct {
					t.Errorf("overview = %+v", o)
				}
			},
		},
		{
			name: "convai.get_document_rag_indexes",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.GetDocumentRAGIndexes(ctx, "d")
			},
			method: http.MethodGet, path: "/v1/convai/knowledge-base/d/rag-index", respBody: `{"indexes":[` + ragIndex + `]}`,
			check: func(t *testing.T, r any) {
				if i := r.(*RAGIndexes).Indexes[0]; i.Status != RAGIndexStatusSucceeded || i.DocumentModelIndexUsage.UsedBytes != 5 {
					t.Errorf("index = %+v", i)
				}
			},
		},
		{
			name: "convai.delete_document_rag_index",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.DeleteDocumentRAGIndex(ctx, "d", "r")
			},
			method: http.MethodDelete, path: "/v1/convai/knowledge-base/d/rag-index/r", respBody: ragIndex,
		},
		{
			name: "convai.knowledge_base.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.KnowledgeBase.List(ctx, &KnowledgeBaseListRequest{Types: []KnowledgeBaseDocumentType{KnowledgeBaseDocumentTypeFile, KnowledgeBaseDocumentTypeURL}, FoldersFirst: Ptr(true)})
			},
			method: http.MethodGet, path: "/v1/convai/knowledge-base", query: "folders_first=true&types=file&types=url",
			respBody: `{"documents":[{"id":"d","name":"n","type":"folder","children_count":2,"dependent_agents":[{"type":"unknown","id":"a"}]}],"has_more":false}`,
			check: func(t *testing.T, r any) {
				if d := r.(*KnowledgeBasePage).Documents[0]; *d.ChildrenCount != 2 || d.DependentAgents[0].Type != "unknown" {
					t.Errorf("document = %+v", d)
				}
			},
		},
		{
			name: "convai.knowledge_base.get_or_create_rag_indexes",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.KnowledgeBase.GetOrCreateRAGIndexes(ctx, []RAGIndexRequest{{DocumentID: "d", CreateIfMissing: true, Model: EmbeddingModelMultilingualE5LargeInstruct}})
			},
			method: http.MethodPost, path: "/v1/convai/knowledge-base/rag-index",
			body:     `{"items":[{"document_id":"d","create_if_missing":true,"model":"multilingual_e5_large_instruct"}]}`,
			respBody: `{"d":{"status":"success","data":` + ragIndex + `}}`,
			check: func(t *testing.T, r any) {
				if m := r.(map[string]BatchResult[RAGIndex]); !m["d"].OK() || m["d"].Data.ID != "r" {
					t.Errorf("result = %+v", m)
				}
			},
		},
		{
			name: "convai.knowledge_base.search",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.KnowledgeBase.Search(ctx, &KnowledgeBaseSearchRequest{Query: "pricing", PageSize: Ptr(5)})
			},
			method: http.MethodGet, path: "/v1/convai/knowledge-base/search", query: "page_size=5&query=pricing",
			respBody: `{"results":[{"document":{"id":"d","name":"n","type":"text"},"score":0.5}]}`,
		},
		// crawl jobs
		{
			name: "convai.knowledge_base.crawl_jobs.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.KnowledgeBase.CrawlJobs.List(ctx, &CrawlJobListRequest{IncludeJobIDs: []string{"j"}})
			},
			method: http.MethodGet, path: "/v1/convai/knowledge-base/crawl", query: "include_job_ids=j",
			respBody: `{"crawl_jobs":[{"id":"j","seed_url":"https://x","max_pages":10,"status":"processing","root_folder_id":"f","updated_at":2,"created_at":1}]}`,
			check: func(t *testing.T, r any) {
				if j := r.(*CrawlJobPage).CrawlJobs[0]; j.Status != CrawlStatusProcessing {
					t.Errorf("job = %+v", j)
				}
			},
		},
		{
			name: "convai.knowledge_base.crawl_jobs.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.KnowledgeBase.CrawlJobs.Create(ctx, &CrawlJobCreateRequest{URL: "https://x", MaxPages: Ptr(10)})
			},
			method: http.MethodPost, path: "/v1/convai/knowledge-base/crawl", body: `{"url":"https://x","max_pages":10}`,
			respBody: `{"id":"j","type":"discovery","root_folder_id":"f","status":"queued","created_at":1}`,
		},
		{
			name: "convai.knowledge_base.crawl_jobs.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.KnowledgeBase.CrawlJobs.Get(ctx, "j")
			},
			method: http.MethodGet, path: "/v1/convai/knowledge-base/crawl/j", respBody: `{"id":"j"}`,
		},
		{
			name: "convai.knowledge_base.crawl_jobs.cancel",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.KnowledgeBase.CrawlJobs.Cancel(ctx, "j")
			},
			method: http.MethodPost, path: "/v1/convai/knowledge-base/crawl/j/cancel", respBody: `{}`,
		},
		// document
		{
			name: "convai.knowledge_base.document.update_file",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.KnowledgeBase.Document.UpdateFile(ctx, "d", testFile("v2.pdf", "pdf"))
			},
			method: http.MethodPatch, path: "/v1/convai/knowledge-base/d/update-file", form: map[string][]string{"file": {"file:v2.pdf:pdf"}},
			respBody: docJSON, check: checkDoc,
		},
		{
			name: "convai.knowledge_base.document.refresh",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.KnowledgeBase.Document.Refresh(ctx, "d")
			},
			method: http.MethodPost, path: "/v1/convai/knowledge-base/d/refresh", respBody: docJSON, check: checkDoc,
		},
		{
			name: "convai.knowledge_base.document.compute_rag_index",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.KnowledgeBase.Document.ComputeRAGIndex(ctx, "d", EmbeddingModelE5Mistral7BInstruct)
			},
			method: http.MethodPost, path: "/v1/convai/knowledge-base/d/rag-index", body: `{"model":"e5_mistral_7b_instruct"}`, respBody: ragIndex,
		},
		// documents
		{
			name: "convai.knowledge_base.documents.create_from_url",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.KnowledgeBase.Documents.CreateFromURL(ctx, &DocumentFromURLRequest{URL: "https://x", AutoSyncOptions: AutoSyncOptions{EnableAutoSync: Ptr(true)}})
			},
			method: http.MethodPost, path: "/v1/convai/knowledge-base/url", body: `{"url":"https://x","enable_auto_sync":true}`, respBody: created,
		},
		{
			name: "convai.knowledge_base.documents.create_from_file",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.KnowledgeBase.Documents.CreateFromFile(ctx, &DocumentFromFileRequest{File: testFile("a.pdf", "pdf"), ParentFolderID: "f"})
			},
			method: http.MethodPost, path: "/v1/convai/knowledge-base/file",
			form: map[string][]string{"file": {"file:a.pdf:pdf"}, "parent_folder_id": {"f"}}, respBody: created,
			check: func(t *testing.T, r any) {
				if c := r.(*KnowledgeBaseCreateResponse); c.FolderPath[0].ID != "f" {
					t.Errorf("created = %+v", c)
				}
			},
		},
		{
			name: "convai.knowledge_base.documents.create_from_text",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.KnowledgeBase.Documents.CreateFromText(ctx, &DocumentFromTextRequest{Text: "hello", Name: "t"})
			},
			method: http.MethodPost, path: "/v1/convai/knowledge-base/text", body: `{"text":"hello","name":"t"}`, respBody: created,
		},
		{
			name: "convai.knowledge_base.documents.create_folder",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.KnowledgeBase.Documents.CreateFolder(ctx, &FolderCreateRequest{Name: "docs"})
			},
			method: http.MethodPost, path: "/v1/convai/knowledge-base/folder", body: `{"name":"docs"}`, respBody: created,
		},
		{
			name: "convai.knowledge_base.documents.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.KnowledgeBase.Documents.Get(ctx, "d", "a")
			},
			method: http.MethodGet, path: "/v1/convai/knowledge-base/d", query: "agent_id=a", respBody: docJSON, check: checkDoc,
		},
		{
			name: "convai.knowledge_base.documents.delete",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.KnowledgeBase.Documents.Delete(ctx, "d", &ForceRequest{Force: Ptr(true)})
			},
			method: http.MethodDelete, path: "/v1/convai/knowledge-base/d", query: "force=true", respBody: `{}`,
		},
		{
			name: "convai.knowledge_base.documents.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.KnowledgeBase.Documents.Update(ctx, "d", &DocumentUpdateRequest{Name: "new"})
			},
			method: http.MethodPatch, path: "/v1/convai/knowledge-base/d", body: `{"name":"new"}`, respBody: docJSON,
		},
		{
			name: "convai.knowledge_base.documents.get_agents",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.KnowledgeBase.Documents.GetAgents(ctx, "d", &DependentAgentsRequest{DependentType: DependentTypeDirect})
			},
			method: http.MethodGet, path: "/v1/convai/knowledge-base/d/dependent-agents", query: "dependent_type=direct",
			respBody: `{"agents":[{"type":"available","id":"a","name":"n","created_at_unix_secs":1,"access_level":"admin"}],"has_more":false}`,
			check: func(t *testing.T, r any) {
				if a := r.(*DependentAgentsPage).Agents[0]; a.Name != "n" || a.AccessLevel != "admin" {
					t.Errorf("agent = %+v", a)
				}
			},
		},
		{
			name: "convai.knowledge_base.documents.get_bulk_agents",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.KnowledgeBase.Documents.GetBulkAgents(ctx, []string{"d1", "d2"}, &DependentAgentsRequest{PageSize: Ptr(4)})
			},
			method: http.MethodPost, path: "/v1/convai/knowledge-base/dependent-agents", query: "page_size=4",
			body: `{"document_ids":["d1","d2"]}`, respBody: `{"agents":[],"has_more":false}`,
		},
		{
			name: "convai.knowledge_base.documents.get_content",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.KnowledgeBase.Documents.GetContent(ctx, "d")
			},
			method: http.MethodGet, path: "/v1/convai/knowledge-base/d/content", accept: acceptAny, respType: "text/html", respBody: "<h1>x</h1>",
			check: func(t *testing.T, r any) {
				if r.(string) != "<h1>x</h1>" {
					t.Errorf("content = %q", r)
				}
			},
		},
		{
			name: "convai.knowledge_base.documents.get_source_file_url",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.KnowledgeBase.Documents.GetSourceFileURL(ctx, "d")
			},
			method: http.MethodGet, path: "/v1/convai/knowledge-base/d/source-file-url", respBody: `{"signed_url":"https://s"}`,
		},
		{
			name: "convai.knowledge_base.documents.move",
			call: func(ctx context.Context, c *Client) (any, error) {
				return nil, c.ConversationalAI.KnowledgeBase.Documents.Move(ctx, "d", "f")
			},
			method: http.MethodPost, path: "/v1/convai/knowledge-base/d/move", body: `{"move_to":"f"}`,
		},
		{
			name: "convai.knowledge_base.documents.bulk_move",
			call: func(ctx context.Context, c *Client) (any, error) {
				return nil, c.ConversationalAI.KnowledgeBase.Documents.BulkMove(ctx, []string{"d"}, "")
			},
			method: http.MethodPost, path: "/v1/convai/knowledge-base/bulk-move", body: `{"document_ids":["d"]}`,
		},
		{
			name: "convai.knowledge_base.documents.bulk_delete",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.KnowledgeBase.Documents.BulkDelete(ctx, []string{"d"}, Ptr(true))
			},
			method: http.MethodPost, path: "/v1/convai/knowledge-base/bulk-delete", body: `{"document_ids":["d"],"force":true}`,
			respBody: `{"d":{"status":"success","data":{"id":"d"}}}`,
			check: func(t *testing.T, r any) {
				if m := r.(map[string]BatchResult[DeletedDocument]); m["d"].Data.ID != "d" {
					t.Errorf("result = %+v", m)
				}
			},
		},
		{
			name: "convai.knowledge_base.documents.chunk.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.KnowledgeBase.Documents.Chunk.Get(ctx, "d", "k", EmbeddingModelE5Mistral7BInstruct)
			},
			method: http.MethodGet, path: "/v1/convai/knowledge-base/d/chunk/k", query: "embedding_model=e5_mistral_7b_instruct",
			respBody: `{"id":"k","name":"n","content":"c"}`,
		},
		{
			name: "convai.knowledge_base.documents.chunks.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.KnowledgeBase.Documents.Chunks.List(ctx, "d", &ChunkListRequest{EmbeddingModel: EmbeddingModelE5Mistral7BInstruct, Cursor: "x"})
			},
			method: http.MethodGet, path: "/v1/convai/knowledge-base/d/chunks", query: "cursor=x&embedding_model=e5_mistral_7b_instruct",
			respBody: `{"chunks":[{"id":"k","name":"n","content":"c"}]}`,
		},
		{
			name: "convai.knowledge_base.documents.summaries.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.KnowledgeBase.Documents.Summaries.Get(ctx, []string{"d"})
			},
			method: http.MethodGet, path: "/v1/convai/knowledge-base/summaries", query: "document_ids=d",
			respBody: `{"d":{"status":"success","data":{"id":"d","name":"n","type":"text"}}}`,
		},
	}
}
