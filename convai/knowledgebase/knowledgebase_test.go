package knowledgebase

import (
	"context"
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/convai/agents"
	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/types"
)

func endpointCases() []testutil.Endpoint[*Service] {
	docJSON := `{"id":"d","name":"n","type":"url","metadata":{"created_at_unix_secs":1,"last_updated_at_unix_secs":2,"size_bytes":3},"supported_usages":["prompt"],"access_info":{"is_creator":true,"creator_name":"c","creator_email":"e","role":"admin"},"url":"https://x","extracted_inner_html":"<p/>"}`
	checkDoc := func(t *testing.T, r any) {
		if d := r.(*Document); d.Type != agents.KnowledgeBaseDocumentTypeURL || d.URL != "https://x" || d.Metadata.SizeBytes != 3 {
			t.Errorf("document = %+v", d)
		}
	}
	created := `{"id":"d","name":"n","folder_path":[{"id":"f"}]}`
	ragIndex := `{"id":"r","model":"e5_mistral_7b_instruct","status":"succeeded","progress_percentage":100,"document_model_index_usage":{"used_bytes":5}}`
	return []testutil.Endpoint[*Service]{
		{
			Name: "convai.knowledge_base.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.List(ctx, &ListRequest{Types: []agents.KnowledgeBaseDocumentType{agents.KnowledgeBaseDocumentTypeFile, agents.KnowledgeBaseDocumentTypeURL}, FoldersFirst: types.Ptr(true)})
			},
			Method: http.MethodGet, Path: "/v1/convai/knowledge-base", Query: "folders_first=true&types=file&types=url",
			RespBody: `{"documents":[{"id":"d","name":"n","type":"folder","children_count":2,"dependent_agents":[{"type":"unknown","id":"a"}]}],"has_more":false}`,
			Check: func(t *testing.T, r any) {
				if d := r.(*Page).Documents[0]; *d.ChildrenCount != 2 || d.DependentAgents[0].Type != "unknown" {
					t.Errorf("document = %+v", d)
				}
			},
		},
		{
			Name: "convai.knowledge_base.get_or_create_rag_indexes",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.GetOrCreateRAGIndexes(ctx, []RAGIndexRequest{{DocumentID: "d", CreateIfMissing: true, Model: EmbeddingModelMultilingualE5LargeInstruct}})
			},
			Method: http.MethodPost, Path: "/v1/convai/knowledge-base/rag-index",
			Body:     `{"items":[{"document_id":"d","create_if_missing":true,"model":"multilingual_e5_large_instruct"}]}`,
			RespBody: `{"d":{"status":"success","data":` + ragIndex + `}}`,
			Check: func(t *testing.T, r any) {
				if m := r.(map[string]agents.BatchResult[RAGIndex]); !m["d"].OK() || m["d"].Data.ID != "r" {
					t.Errorf("result = %+v", m)
				}
			},
		},
		{
			Name: "convai.knowledge_base.search",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Search(ctx, &SearchRequest{Query: "pricing", PageSize: types.Ptr(5)})
			},
			Method: http.MethodGet, Path: "/v1/convai/knowledge-base/search", Query: "page_size=5&query=pricing",
			RespBody: `{"results":[{"document":{"id":"d","name":"n","type":"text"},"score":0.5}]}`,
		},
		// crawl jobs
		{
			Name: "convai.knowledge_base.crawl_jobs.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.CrawlJobs.List(ctx, &CrawlJobListRequest{IncludeJobIDs: []string{"j"}})
			},
			Method: http.MethodGet, Path: "/v1/convai/knowledge-base/crawl", Query: "include_job_ids=j",
			RespBody: `{"crawl_jobs":[{"id":"j","seed_url":"https://x","max_pages":10,"status":"processing","root_folder_id":"f","updated_at":2,"created_at":1}]}`,
			Check: func(t *testing.T, r any) {
				if j := r.(*CrawlJobPage).CrawlJobs[0]; j.Status != CrawlStatusProcessing {
					t.Errorf("job = %+v", j)
				}
			},
		},
		{
			Name: "convai.knowledge_base.crawl_jobs.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.CrawlJobs.Create(ctx, &CrawlJobCreateRequest{URL: "https://x", MaxPages: types.Ptr(10)})
			},
			Method: http.MethodPost, Path: "/v1/convai/knowledge-base/crawl", Body: `{"url":"https://x","max_pages":10}`,
			RespBody: `{"id":"j","type":"discovery","root_folder_id":"f","status":"queued","created_at":1}`,
		},
		{
			Name: "convai.knowledge_base.crawl_jobs.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.CrawlJobs.Get(ctx, "j")
			},
			Method: http.MethodGet, Path: "/v1/convai/knowledge-base/crawl/j", RespBody: `{"id":"j"}`,
		},
		{
			Name: "convai.knowledge_base.crawl_jobs.cancel",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.CrawlJobs.Cancel(ctx, "j")
			},
			Method: http.MethodPost, Path: "/v1/convai/knowledge-base/crawl/j/cancel", RespBody: `{}`,
		},
		// document
		{
			Name: "convai.knowledge_base.document.update_file",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Document.UpdateFile(ctx, "d", testutil.File("v2.pdf", "pdf"))
			},
			Method: http.MethodPatch, Path: "/v1/convai/knowledge-base/d/update-file", Form: map[string][]string{"file": {"file:v2.pdf:pdf"}},
			RespBody: docJSON, Check: checkDoc,
		},
		{
			Name: "convai.knowledge_base.document.refresh",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Document.Refresh(ctx, "d")
			},
			Method: http.MethodPost, Path: "/v1/convai/knowledge-base/d/refresh", RespBody: docJSON, Check: checkDoc,
		},
		{
			Name: "convai.knowledge_base.document.compute_rag_index",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Document.ComputeRAGIndex(ctx, "d", EmbeddingModelE5Mistral7BInstruct)
			},
			Method: http.MethodPost, Path: "/v1/convai/knowledge-base/d/rag-index", Body: `{"model":"e5_mistral_7b_instruct"}`, RespBody: ragIndex,
		},
		// documents
		{
			Name: "convai.knowledge_base.documents.create_from_url",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Documents.CreateFromURL(ctx, &DocumentFromURLRequest{URL: "https://x", AutoSyncOptions: AutoSyncOptions{EnableAutoSync: types.Ptr(true)}})
			},
			Method: http.MethodPost, Path: "/v1/convai/knowledge-base/url", Body: `{"url":"https://x","enable_auto_sync":true}`, RespBody: created,
		},
		{
			Name: "convai.knowledge_base.documents.create_from_file",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Documents.CreateFromFile(ctx, &DocumentFromFileRequest{File: testutil.File("a.pdf", "pdf"), ParentFolderID: "f"})
			},
			Method: http.MethodPost, Path: "/v1/convai/knowledge-base/file",
			Form: map[string][]string{"file": {"file:a.pdf:pdf"}, "parent_folder_id": {"f"}}, RespBody: created,
			Check: func(t *testing.T, r any) {
				if c := r.(*CreateResponse); c.FolderPath[0].ID != "f" {
					t.Errorf("created = %+v", c)
				}
			},
		},
		{
			Name: "convai.knowledge_base.documents.create_from_text",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Documents.CreateFromText(ctx, &DocumentFromTextRequest{Text: "hello", Name: "t"})
			},
			Method: http.MethodPost, Path: "/v1/convai/knowledge-base/text", Body: `{"text":"hello","name":"t"}`, RespBody: created,
		},
		{
			Name: "convai.knowledge_base.documents.create_folder",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Documents.CreateFolder(ctx, &FolderCreateRequest{Name: "docs"})
			},
			Method: http.MethodPost, Path: "/v1/convai/knowledge-base/folder", Body: `{"name":"docs"}`, RespBody: created,
		},
		{
			Name: "convai.knowledge_base.documents.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Documents.Get(ctx, "d", "a")
			},
			Method: http.MethodGet, Path: "/v1/convai/knowledge-base/d", Query: "agent_id=a", RespBody: docJSON, Check: checkDoc,
		},
		{
			Name: "convai.knowledge_base.documents.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Documents.Delete(ctx, "d", &agents.ForceRequest{Force: types.Ptr(true)})
			},
			Method: http.MethodDelete, Path: "/v1/convai/knowledge-base/d", Query: "force=true", RespBody: `{}`,
		},
		{
			Name: "convai.knowledge_base.documents.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Documents.Update(ctx, "d", &DocumentUpdateRequest{Name: "new"})
			},
			Method: http.MethodPatch, Path: "/v1/convai/knowledge-base/d", Body: `{"name":"new"}`, RespBody: docJSON,
		},
		{
			Name: "convai.knowledge_base.documents.get_agents",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Documents.GetAgents(ctx, "d", &DependentAgentsRequest{DependentType: DependentTypeDirect})
			},
			Method: http.MethodGet, Path: "/v1/convai/knowledge-base/d/dependent-agents", Query: "dependent_type=direct",
			RespBody: `{"agents":[{"type":"available","id":"a","name":"n","created_at_unix_secs":1,"access_level":"admin"}],"has_more":false}`,
			Check: func(t *testing.T, r any) {
				if a := r.(*agents.DependentAgentsPage).Agents[0]; a.Name != "n" || a.AccessLevel != "admin" {
					t.Errorf("agent = %+v", a)
				}
			},
		},
		{
			Name: "convai.knowledge_base.documents.get_bulk_agents",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Documents.GetBulkAgents(ctx, []string{"d1", "d2"}, &DependentAgentsRequest{PageSize: types.Ptr(4)})
			},
			Method: http.MethodPost, Path: "/v1/convai/knowledge-base/dependent-agents", Query: "page_size=4",
			Body: `{"document_ids":["d1","d2"]}`, RespBody: `{"agents":[],"has_more":false}`,
		},
		{
			Name: "convai.knowledge_base.documents.get_content",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Documents.GetContent(ctx, "d")
			},
			Method: http.MethodGet, Path: "/v1/convai/knowledge-base/d/content", Accept: transport.AcceptAny, RespType: "text/html", RespBody: "<h1>x</h1>",
			Check: func(t *testing.T, r any) {
				if r.(string) != "<h1>x</h1>" {
					t.Errorf("content = %q", r)
				}
			},
		},
		{
			Name: "convai.knowledge_base.documents.get_source_file_url",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Documents.GetSourceFileURL(ctx, "d")
			},
			Method: http.MethodGet, Path: "/v1/convai/knowledge-base/d/source-file-url", RespBody: `{"signed_url":"https://s"}`,
		},
		{
			Name: "convai.knowledge_base.documents.move",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return nil, c.Documents.Move(ctx, "d", "f")
			},
			Method: http.MethodPost, Path: "/v1/convai/knowledge-base/d/move", Body: `{"move_to":"f"}`,
		},
		{
			Name: "convai.knowledge_base.documents.bulk_move",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return nil, c.Documents.BulkMove(ctx, []string{"d"}, "")
			},
			Method: http.MethodPost, Path: "/v1/convai/knowledge-base/bulk-move", Body: `{"document_ids":["d"]}`,
		},
		{
			Name: "convai.knowledge_base.documents.bulk_delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Documents.BulkDelete(ctx, []string{"d"}, types.Ptr(true))
			},
			Method: http.MethodPost, Path: "/v1/convai/knowledge-base/bulk-delete", Body: `{"document_ids":["d"],"force":true}`,
			RespBody: `{"d":{"status":"success","data":{"id":"d"}}}`,
			Check: func(t *testing.T, r any) {
				if m := r.(map[string]agents.BatchResult[DeletedDocument]); m["d"].Data.ID != "d" {
					t.Errorf("result = %+v", m)
				}
			},
		},
		{
			Name: "convai.knowledge_base.documents.chunk.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Documents.Chunk.Get(ctx, "d", "k", EmbeddingModelE5Mistral7BInstruct)
			},
			Method: http.MethodGet, Path: "/v1/convai/knowledge-base/d/chunk/k", Query: "embedding_model=e5_mistral_7b_instruct",
			RespBody: `{"id":"k","name":"n","content":"c"}`,
		},
		{
			Name: "convai.knowledge_base.documents.chunks.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Documents.Chunks.List(ctx, "d", &ChunkListRequest{EmbeddingModel: EmbeddingModelE5Mistral7BInstruct, Cursor: "x"})
			},
			Method: http.MethodGet, Path: "/v1/convai/knowledge-base/d/chunks", Query: "cursor=x&embedding_model=e5_mistral_7b_instruct",
			RespBody: `{"chunks":[{"id":"k","name":"n","content":"c"}]}`,
		},
		{
			Name: "convai.knowledge_base.documents.summaries.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Documents.Summaries.Get(ctx, []string{"d"})
			},
			Method: http.MethodGet, Path: "/v1/convai/knowledge-base/summaries", Query: "document_ids=d",
			RespBody: `{"d":{"status":"success","data":{"id":"d","name":"n","type":"text"}}}`,
		},
	}
}

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, endpointCases())
}
