package knowledgebase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"strings"

	"github.com/taigrr/elevenlabs/v2/convai/agents"
	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/stream"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

// Service groups the /v1/convai/knowledge-base endpoints.
type Service struct {
	cfg *requestconfig.Config

	// CrawlJobs crawls websites into the knowledge base.
	CrawlJobs *CrawlJobsService
	// Document refreshes, re-uploads and indexes single documents.
	Document *DocumentService
	// Documents creates, reads, updates and organizes documents.
	Documents *DocumentsService
}

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	cfg := requestconfig.Build(opts...)
	return &Service{
		cfg:       cfg,
		CrawlJobs: &CrawlJobsService{cfg: cfg},
		Document:  &DocumentService{cfg: cfg},
		Documents: &DocumentsService{
			cfg:       cfg,
			Chunk:     &ChunkService{cfg: cfg},
			Chunks:    &ChunksService{cfg: cfg},
			Summaries: &SummariesService{cfg: cfg},
		},
	}
}

// EmbeddingModel selects the RAG embedding model.
type EmbeddingModel string

// EmbeddingModel values.
const (
	EmbeddingModelE5Mistral7BInstruct         EmbeddingModel = "e5_mistral_7b_instruct"
	EmbeddingModelMultilingualE5LargeInstruct EmbeddingModel = "multilingual_e5_large_instruct"
)

// RAGIndexStatus is the state of a RAG index.
type RAGIndexStatus string

// RAGIndexStatus values.
const (
	RAGIndexStatusNew               RAGIndexStatus = "new"
	RAGIndexStatusCreated           RAGIndexStatus = "created"
	RAGIndexStatusProcessing        RAGIndexStatus = "processing"
	RAGIndexStatusFailed            RAGIndexStatus = "failed"
	RAGIndexStatusSucceeded         RAGIndexStatus = "succeeded"
	RAGIndexStatusRAGLimitExceeded  RAGIndexStatus = "rag_limit_exceeded"
	RAGIndexStatusDocumentTooSmall  RAGIndexStatus = "document_too_small"
	RAGIndexStatusCannotIndexFolder RAGIndexStatus = "cannot_index_folder"
)

// RAGIndex is a document's RAG index for one embedding model.
type RAGIndex struct {
	ID                      string         `json:"id"`
	Model                   EmbeddingModel `json:"model"`
	Status                  RAGIndexStatus `json:"status"`
	ProgressPercentage      float64        `json:"progress_percentage"`
	DocumentModelIndexUsage struct {
		UsedBytes int64 `json:"used_bytes"`
	} `json:"document_model_index_usage"`
}

// RAGIndexes lists a document's RAG indexes.
type RAGIndexes struct {
	Indexes []RAGIndex `json:"indexes"`
}

// RAGIndexOverview reports RAG storage usage across the workspace.
type RAGIndexOverview struct {
	TotalUsedBytes int64 `json:"total_used_bytes"`
	TotalMaxBytes  int64 `json:"total_max_bytes"`
	Models         []struct {
		Model     EmbeddingModel `json:"model"`
		UsedBytes int64          `json:"used_bytes"`
	} `json:"models"`
}

// FolderPathSegment is one folder on a document's path.
type FolderPathSegment struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

// CreateResponse identifies a new document or folder.
type CreateResponse struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	FolderPath []FolderPathSegment `json:"folder_path,omitempty"`
}

// DocumentMetadata holds document timestamps and size.
type DocumentMetadata struct {
	CreatedAtUnixSecs     int64 `json:"created_at_unix_secs"`
	LastUpdatedAtUnixSecs int64 `json:"last_updated_at_unix_secs"`
	SizeBytes             int64 `json:"size_bytes"`
}

// Document is a knowledge base entry: a file, URL, text
// document or folder (see Type). Fields not applicable to the type are
// zero. Summary listings set DependentAgents; full reads set
// ExtractedInnerHTML.
type Document struct {
	ID                 string                           `json:"id"`
	Name               string                           `json:"name"`
	Type               agents.KnowledgeBaseDocumentType `json:"type"`
	Metadata           DocumentMetadata                 `json:"metadata"`
	SupportedUsages    []string                         `json:"supported_usages"`
	AccessInfo         types.ResourceAccessInfo         `json:"access_info"`
	FolderParentID     string                           `json:"folder_parent_id,omitempty"`
	FolderPath         []FolderPathSegment              `json:"folder_path,omitempty"`
	DependentAgents    []agents.DependentAgent          `json:"dependent_agents,omitempty"`
	URL                string                           `json:"url,omitempty"`
	Filename           string                           `json:"filename,omitempty"`
	ExtractedInnerHTML string                           `json:"extracted_inner_html,omitempty"`
	ContentFormat      string                           `json:"content_format,omitempty"`
	ChildrenCount      *int                             `json:"children_count,omitempty"`
	DocumentCount      *int                             `json:"document_count,omitempty"`
	ExternalSyncInfo   json.RawMessage                  `json:"external_sync_info,omitempty"`
	AutoSyncInfo       json.RawMessage                  `json:"auto_sync_info,omitempty"`
	RefreshStatus      json.RawMessage                  `json:"refresh_status,omitempty"`
	IsFrozen           *bool                            `json:"is_frozen,omitempty"`
	ActiveSyncJob      json.RawMessage                  `json:"active_sync_job,omitempty"`
}

// AddRequest is the body of convai.Service.AddToKnowledgeBase.
// Set either URL or File.
type AddRequest struct {
	AgentID string      `form:"-" query:"agent_id"`
	Name    string      `form:"name"`
	URL     string      `form:"url"`
	File    *types.File `form:"file"`
}

// SortBy selects the document list sort field.
type SortBy string

// SortBy values.
const (
	SortByName      SortBy = "name"
	SortByCreatedAt SortBy = "created_at"
	SortByUpdatedAt SortBy = "updated_at"
	SortBySize      SortBy = "size"
)

// ListRequest holds the optional parameters of Service.List.
type ListRequest struct {
	PageSize               *int                               `query:"page_size"`
	Search                 string                             `query:"search"`
	ShowOnlyOwnedDocuments *bool                              `query:"show_only_owned_documents"`
	CreatedByUserID        string                             `query:"created_by_user_id"`
	Types                  []agents.KnowledgeBaseDocumentType `query:"types"`
	ParentFolderID         string                             `query:"parent_folder_id"`
	AncestorFolderID       string                             `query:"ancestor_folder_id"`
	FoldersFirst           *bool                              `query:"folders_first"`
	SortDirection          types.SortDirection                `query:"sort_direction"`
	SortBy                 SortBy                             `query:"sort_by"`
	Cursor                 string                             `query:"cursor"`
}

// Page is a page of knowledge base entries.
type Page struct {
	Documents  []Document `json:"documents"`
	NextCursor *string    `json:"next_cursor,omitempty"`
	HasMore    bool       `json:"has_more"`
}

// RAGIndexRequest asks for a document's RAG index.
type RAGIndexRequest struct {
	DocumentID      string         `json:"document_id"`
	CreateIfMissing bool           `json:"create_if_missing"`
	Model           EmbeddingModel `json:"model"`
}

// SearchRequest holds the parameters of Service.Search.
type SearchRequest struct {
	// Query is required.
	Query    string                             `query:"query"`
	PageSize *int                               `query:"page_size"`
	Types    []agents.KnowledgeBaseDocumentType `query:"types"`
	Cursor   string                             `query:"cursor"`
}

// SearchResult is a document matching a content search.
type SearchResult struct {
	Document      Document          `json:"document"`
	SearchSnippet []json.RawMessage `json:"search_snippet,omitempty"`
	Score         float64           `json:"score"`
}

// SearchPage is a page of content search results.
type SearchPage struct {
	Results    []SearchResult `json:"results"`
	NextCursor *string        `json:"next_cursor,omitempty"`
}

// List returns one page of knowledge base entries. req may be nil.
func (s *Service) List(ctx context.Context, req *ListRequest, opts ...option.RequestOption) (*Page, error) {
	return transport.GetValue[Page](ctx, s.cfg, "/v1/convai/knowledge-base", req, opts)
}

// ListAll iterates over every knowledge base entry.
func (s *Service) ListAll(ctx context.Context, req *ListRequest, opts ...option.RequestOption) iter.Seq2[Document, error] {
	params := *transport.OrEmpty(req)
	return stream.Paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (stream.Page[Document], error) {
		params.Cursor = cursor
		p, err := s.List(ctx, &params, opts...)
		if err != nil {
			return stream.Page[Document]{}, err
		}
		return stream.Page[Document]{Items: p.Documents, Next: transport.Deref(p.NextCursor), HasMore: p.HasMore}, nil
	})
}

// GetOrCreateRAGIndexes returns, and optionally creates, RAG indexes for
// several documents, keyed by document ID.
func (s *Service) GetOrCreateRAGIndexes(ctx context.Context, items []RAGIndexRequest, opts ...option.RequestOption) (map[string]agents.BatchResult[RAGIndex], error) {
	body := struct {
		Items []RAGIndexRequest `json:"items"`
	}{transport.NonNil(items)}
	out, err := transport.SendValue[map[string]agents.BatchResult[RAGIndex]](ctx, s.cfg, http.MethodPost, "/v1/convai/knowledge-base/rag-index", nil, body, opts)
	if err != nil {
		return nil, err
	}
	return *out, nil
}

// Search searches knowledge base contents.
func (s *Service) Search(ctx context.Context, req *SearchRequest, opts ...option.RequestOption) (*SearchPage, error) {
	return transport.GetValue[SearchPage](ctx, s.cfg, "/v1/convai/knowledge-base/search", req, opts)
}

// SearchAll iterates over every content search result.
func (s *Service) SearchAll(ctx context.Context, req *SearchRequest, opts ...option.RequestOption) iter.Seq2[SearchResult, error] {
	params := *transport.OrEmpty(req)
	return stream.Paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (stream.Page[SearchResult], error) {
		params.Cursor = cursor
		p, err := s.Search(ctx, &params, opts...)
		if err != nil {
			return stream.Page[SearchResult]{}, err
		}
		next := transport.Deref(p.NextCursor)
		return stream.Page[SearchResult]{Items: p.Results, Next: next, HasMore: next != ""}, nil
	})
}

// CrawlJobsService groups the /v1/convai/knowledge-base/crawl endpoints.
type CrawlJobsService struct{ cfg *requestconfig.Config }

// CrawlStatus is the state of a crawl job.
type CrawlStatus string

// CrawlStatus values.
const (
	CrawlStatusQueued     CrawlStatus = "queued"
	CrawlStatusProcessing CrawlStatus = "processing"
	CrawlStatusSucceeded  CrawlStatus = "succeeded"
	CrawlStatusFailed     CrawlStatus = "failed"
	CrawlStatusSkipped    CrawlStatus = "skipped"
	CrawlStatusCancelled  CrawlStatus = "cancelled"
)

// CrawlJob is a website crawl.
type CrawlJob struct {
	ID string `json:"id"`
	// Type is "discovery" or "sitemap".
	Type            string      `json:"type,omitempty"`
	SeedURL         string      `json:"seed_url"`
	Pattern         string      `json:"pattern,omitempty"`
	ExcludePatterns []string    `json:"exclude_patterns,omitempty"`
	MaxPages        int         `json:"max_pages"`
	AutoDiscover    *bool       `json:"auto_discover,omitempty"`
	Status          CrawlStatus `json:"status,omitempty"`
	PagesIdentified *int        `json:"pages_identified,omitempty"`
	PagesScraped    *int        `json:"pages_scraped,omitempty"`
	PagesSkipped    *int        `json:"pages_skipped,omitempty"`
	PagesFailed     *int        `json:"pages_failed,omitempty"`
	RootFolderID    string      `json:"root_folder_id"`
	UpdatedAt       int64       `json:"updated_at"`
	CreatedAt       int64       `json:"created_at"`
}

// CrawlJobPage is a page of crawl jobs.
type CrawlJobPage struct {
	CrawlJobs  []CrawlJob `json:"crawl_jobs"`
	NextCursor *string    `json:"next_cursor,omitempty"`
}

// CrawlJobListRequest holds the optional parameters of CrawlJobsService.List.
type CrawlJobListRequest struct {
	IncludeJobIDs []string `query:"include_job_ids"`
	PageSize      *int     `query:"page_size"`
	Cursor        string   `query:"cursor"`
}

// CrawlJobCreateRequest is the body of CrawlJobsService.Create.
type CrawlJobCreateRequest struct {
	// URL is the seed URL. Required.
	URL                  string   `json:"url"`
	MaxDepth             *int     `json:"max_depth,omitempty"`
	MaxPages             *int     `json:"max_pages,omitempty"`
	Pattern              string   `json:"pattern,omitempty"`
	ExcludePatterns      []string `json:"exclude_patterns,omitempty"`
	SitemapURLs          []string `json:"sitemap_urls,omitempty"`
	ParentFolderID       string   `json:"parent_folder_id,omitempty"`
	EnableAutoSync       *bool    `json:"enable_auto_sync,omitempty"`
	AutoRemove           *bool    `json:"auto_remove,omitempty"`
	AutoDiscover         *bool    `json:"auto_discover,omitempty"`
	MinimumFrequencyDays *int     `json:"minimum_frequency_days,omitempty"`
}

// CrawlJobCreateResponse describes a newly created crawl job.
type CrawlJobCreateResponse struct {
	ID           string              `json:"id"`
	Type         string              `json:"type"`
	RootFolderID string              `json:"root_folder_id"`
	Status       string              `json:"status"`
	CreatedAt    int64               `json:"created_at"`
	FolderPath   []FolderPathSegment `json:"folder_path,omitempty"`
}

// List returns one page of crawl jobs. req may be nil.
func (s *CrawlJobsService) List(ctx context.Context, req *CrawlJobListRequest, opts ...option.RequestOption) (*CrawlJobPage, error) {
	return transport.GetValue[CrawlJobPage](ctx, s.cfg, "/v1/convai/knowledge-base/crawl", req, opts)
}

// Create starts a crawl job.
func (s *CrawlJobsService) Create(ctx context.Context, req *CrawlJobCreateRequest, opts ...option.RequestOption) (*CrawlJobCreateResponse, error) {
	return transport.SendValue[CrawlJobCreateResponse](ctx, s.cfg, http.MethodPost, "/v1/convai/knowledge-base/crawl", nil, transport.OrEmpty(req), opts)
}

// Get returns a crawl job.
func (s *CrawlJobsService) Get(ctx context.Context, crawlJobID string, opts ...option.RequestOption) (*CrawlJob, error) {
	return transport.GetValue[CrawlJob](ctx, s.cfg, transport.Pathf("/v1/convai/knowledge-base/crawl/%s", crawlJobID), nil, opts)
}

// Cancel cancels a crawl job.
func (s *CrawlJobsService) Cancel(ctx context.Context, crawlJobID string, opts ...option.RequestOption) (json.RawMessage, error) {
	return transport.SendRaw(ctx, s.cfg, http.MethodPost, transport.Pathf("/v1/convai/knowledge-base/crawl/%s/cancel", crawlJobID), nil, nil, opts)
}

// DocumentService groups single-document maintenance endpoints.
type DocumentService struct{ cfg *requestconfig.Config }

// UpdateFile replaces the file of a file document.
func (s *DocumentService) UpdateFile(ctx context.Context, documentID string, file types.File, opts ...option.RequestOption) (*Document, error) {
	form := struct {
		File types.File `form:"file"`
	}{file}
	return transport.DoValue[Document](ctx, s.cfg, &transport.Request{Method: http.MethodPatch, Path: transport.Pathf("/v1/convai/knowledge-base/%s/update-file", documentID), Form: &form, Opts: opts})
}

// Refresh re-fetches a URL document or re-syncs a file document.
func (s *DocumentService) Refresh(ctx context.Context, documentID string, opts ...option.RequestOption) (*Document, error) {
	return transport.SendValue[Document](ctx, s.cfg, http.MethodPost, transport.Pathf("/v1/convai/knowledge-base/%s/refresh", documentID), nil, nil, opts)
}

// ComputeRAGIndex starts computing a document's RAG index for model.
func (s *DocumentService) ComputeRAGIndex(ctx context.Context, documentID string, model EmbeddingModel, opts ...option.RequestOption) (*RAGIndex, error) {
	body := struct {
		Model EmbeddingModel `json:"model"`
	}{model}
	return transport.SendValue[RAGIndex](ctx, s.cfg, http.MethodPost, transport.Pathf("/v1/convai/knowledge-base/%s/rag-index", documentID), nil, body, opts)
}

// DocumentsService groups the knowledge base document endpoints.
type DocumentsService struct {
	cfg *requestconfig.Config

	// Chunk reads a single RAG chunk.
	Chunk *ChunkService
	// Chunks lists RAG chunks.
	Chunks *ChunksService
	// Summaries returns document summaries in bulk.
	Summaries *SummariesService
}

// AutoSyncOptions control automatic re-syncing of URL documents and folders.
type AutoSyncOptions struct {
	EnableAutoSync       *bool `json:"enable_auto_sync,omitempty"`
	AutoRemove           *bool `json:"auto_remove,omitempty"`
	MinimumFrequencyDays *int  `json:"minimum_frequency_days,omitempty"`
}

// DocumentFromURLRequest is the body of DocumentsService.CreateFromURL.
type DocumentFromURLRequest struct {
	// URL is required.
	URL            string `json:"url"`
	Name           string `json:"name,omitempty"`
	ParentFolderID string `json:"parent_folder_id,omitempty"`
	AutoSyncOptions
}

// DocumentFromFileRequest is the form of DocumentsService.CreateFromFile.
type DocumentFromFileRequest struct {
	// File is required.
	File           types.File `form:"file"`
	Name           string     `form:"name"`
	ParentFolderID string     `form:"parent_folder_id"`
}

// DocumentFromTextRequest is the body of DocumentsService.CreateFromText.
type DocumentFromTextRequest struct {
	// Text is required.
	Text           string `json:"text"`
	Name           string `json:"name,omitempty"`
	ParentFolderID string `json:"parent_folder_id,omitempty"`
}

// FolderCreateRequest is the body of DocumentsService.CreateFolder.
type FolderCreateRequest struct {
	// Name is required.
	Name           string `json:"name"`
	ParentFolderID string `json:"parent_folder_id,omitempty"`
	AutoSyncOptions
}

// DocumentUpdateRequest is the body of DocumentsService.Update.
type DocumentUpdateRequest struct {
	Name    string `json:"name,omitempty"`
	Content string `json:"content,omitempty"`
}

// DependentType selects direct, transitive or all dependents.
type DependentType string

// DependentType values.
const (
	DependentTypeDirect     DependentType = "direct"
	DependentTypeTransitive DependentType = "transitive"
	DependentTypeAll        DependentType = "all"
)

// DependentAgentsRequest holds the optional parameters of the dependent-agents endpoints.
type DependentAgentsRequest struct {
	DependentType DependentType `query:"dependent_type"`
	PageSize      *int          `query:"page_size"`
	Cursor        string        `query:"cursor"`
}

// SignedFileURL is a temporary download URL.
type SignedFileURL struct {
	SignedURL string `json:"signed_url"`
}

// CreateFromURL scrapes a web page into a document.
func (s *DocumentsService) CreateFromURL(ctx context.Context, req *DocumentFromURLRequest, opts ...option.RequestOption) (*CreateResponse, error) {
	return transport.SendValue[CreateResponse](ctx, s.cfg, http.MethodPost, "/v1/convai/knowledge-base/url", nil, transport.OrEmpty(req), opts)
}

// CreateFromFile uploads a file as a document.
func (s *DocumentsService) CreateFromFile(ctx context.Context, req *DocumentFromFileRequest, opts ...option.RequestOption) (*CreateResponse, error) {
	return transport.DoValue[CreateResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/convai/knowledge-base/file", Form: transport.OrEmpty(req), Opts: opts})
}

// CreateFromText creates a document from text.
func (s *DocumentsService) CreateFromText(ctx context.Context, req *DocumentFromTextRequest, opts ...option.RequestOption) (*CreateResponse, error) {
	return transport.SendValue[CreateResponse](ctx, s.cfg, http.MethodPost, "/v1/convai/knowledge-base/text", nil, transport.OrEmpty(req), opts)
}

// CreateFolder creates a folder.
func (s *DocumentsService) CreateFolder(ctx context.Context, req *FolderCreateRequest, opts ...option.RequestOption) (*CreateResponse, error) {
	return transport.SendValue[CreateResponse](ctx, s.cfg, http.MethodPost, "/v1/convai/knowledge-base/folder", nil, transport.OrEmpty(req), opts)
}

// Get returns a document. agentID may be empty.
func (s *DocumentsService) Get(ctx context.Context, documentID, agentID string, opts ...option.RequestOption) (*Document, error) {
	query := struct {
		AgentID string `query:"agent_id"`
	}{agentID}
	return transport.GetValue[Document](ctx, s.cfg, transport.Pathf("/v1/convai/knowledge-base/%s", documentID), query, opts)
}

// Delete deletes a document or folder. req may be nil.
func (s *DocumentsService) Delete(ctx context.Context, documentID string, req *agents.ForceRequest, opts ...option.RequestOption) (json.RawMessage, error) {
	return transport.SendRaw(ctx, s.cfg, http.MethodDelete, transport.Pathf("/v1/convai/knowledge-base/%s", documentID), req, nil, opts)
}

// Update renames a document or replaces a text document's content.
func (s *DocumentsService) Update(ctx context.Context, documentID string, req *DocumentUpdateRequest, opts ...option.RequestOption) (*Document, error) {
	return transport.SendValue[Document](ctx, s.cfg, http.MethodPatch, transport.Pathf("/v1/convai/knowledge-base/%s", documentID), nil, transport.OrEmpty(req), opts)
}

// GetAgents lists the agents that use a document. req may be nil.
func (s *DocumentsService) GetAgents(ctx context.Context, documentID string, req *DependentAgentsRequest, opts ...option.RequestOption) (*agents.DependentAgentsPage, error) {
	return transport.GetValue[agents.DependentAgentsPage](ctx, s.cfg, transport.Pathf("/v1/convai/knowledge-base/%s/dependent-agents", documentID), req, opts)
}

// GetBulkAgents lists the agents that use any of documentIDs. req may be nil.
func (s *DocumentsService) GetBulkAgents(ctx context.Context, documentIDs []string, req *DependentAgentsRequest, opts ...option.RequestOption) (*agents.DependentAgentsPage, error) {
	body := struct {
		DocumentIDs []string `json:"document_ids"`
	}{transport.NonNil(documentIDs)}
	return transport.SendValue[agents.DependentAgentsPage](ctx, s.cfg, http.MethodPost, "/v1/convai/knowledge-base/dependent-agents", req, body, opts)
}

// GetContent returns a document's extracted content as HTML.
func (s *DocumentsService) GetContent(ctx context.Context, documentID string, opts ...option.RequestOption) (string, error) {
	body, err := transport.DoStream(ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/convai/knowledge-base/%s/content", documentID), Accept: transport.AcceptAny, Opts: opts})
	if err != nil {
		return "", err
	}
	defer body.Close()
	var b strings.Builder
	if _, err := io.Copy(&b, body); err != nil {
		return "", fmt.Errorf("elevenlabs: read document content: %w", err)
	}
	return b.String(), nil
}

// GetSourceFileURL returns a signed URL for downloading a document's source file.
func (s *DocumentsService) GetSourceFileURL(ctx context.Context, documentID string, opts ...option.RequestOption) (*SignedFileURL, error) {
	return transport.GetValue[SignedFileURL](ctx, s.cfg, transport.Pathf("/v1/convai/knowledge-base/%s/source-file-url", documentID), nil, opts)
}

// Move moves a document or folder into folderID (the root when empty).
func (s *DocumentsService) Move(ctx context.Context, documentID, folderID string, opts ...option.RequestOption) error {
	body := struct {
		MoveTo string `json:"move_to,omitempty"`
	}{folderID}
	return transport.NoContent(ctx, s.cfg, http.MethodPost, transport.Pathf("/v1/convai/knowledge-base/%s/move", documentID), nil, body, opts)
}

// BulkMove moves documents into folderID (the root when empty).
func (s *DocumentsService) BulkMove(ctx context.Context, documentIDs []string, folderID string, opts ...option.RequestOption) error {
	body := struct {
		DocumentIDs []string `json:"document_ids"`
		MoveTo      string   `json:"move_to,omitempty"`
	}{transport.NonNil(documentIDs), folderID}
	return transport.NoContent(ctx, s.cfg, http.MethodPost, "/v1/convai/knowledge-base/bulk-move", nil, body, opts)
}

// DeletedDocument identifies a deleted document.
type DeletedDocument struct {
	ID string `json:"id"`
}

// BulkDelete deletes several documents, returning per-document results
// keyed by document ID. force deletes documents still used by agents.
func (s *DocumentsService) BulkDelete(ctx context.Context, documentIDs []string, force *bool, opts ...option.RequestOption) (map[string]agents.BatchResult[DeletedDocument], error) {
	body := struct {
		DocumentIDs []string `json:"document_ids"`
		Force       *bool    `json:"force,omitempty"`
	}{transport.NonNil(documentIDs), force}
	out, err := transport.SendValue[map[string]agents.BatchResult[DeletedDocument]](ctx, s.cfg, http.MethodPost, "/v1/convai/knowledge-base/bulk-delete", nil, body, opts)
	if err != nil {
		return nil, err
	}
	return *out, nil
}

// Chunk is a RAG chunk of a document.
type Chunk struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

// ChunkPage is a page of chunks.
type ChunkPage struct {
	Chunks     []Chunk `json:"chunks"`
	NextCursor *string `json:"next_cursor,omitempty"`
}

// ChunkService groups the single chunk endpoint.
type ChunkService struct{ cfg *requestconfig.Config }

// Get returns one chunk. model may be empty.
func (s *ChunkService) Get(ctx context.Context, documentID, chunkID string, model EmbeddingModel, opts ...option.RequestOption) (*Chunk, error) {
	query := struct {
		EmbeddingModel EmbeddingModel `query:"embedding_model"`
	}{model}
	return transport.GetValue[Chunk](ctx, s.cfg, transport.Pathf("/v1/convai/knowledge-base/%s/chunk/%s", documentID, chunkID), query, opts)
}

// ChunksService groups the chunk listing endpoint.
type ChunksService struct{ cfg *requestconfig.Config }

// ChunkListRequest holds the parameters of ChunksService.List.
type ChunkListRequest struct {
	// EmbeddingModel is required.
	EmbeddingModel EmbeddingModel `query:"embedding_model"`
	PageSize       *int           `query:"page_size"`
	Cursor         string         `query:"cursor"`
}

// List returns one page of a document's chunks.
func (s *ChunksService) List(ctx context.Context, documentID string, req *ChunkListRequest, opts ...option.RequestOption) (*ChunkPage, error) {
	return transport.GetValue[ChunkPage](ctx, s.cfg, transport.Pathf("/v1/convai/knowledge-base/%s/chunks", documentID), req, opts)
}

// ListAll iterates over every chunk of a document.
func (s *ChunksService) ListAll(ctx context.Context, documentID string, req *ChunkListRequest, opts ...option.RequestOption) iter.Seq2[Chunk, error] {
	params := *transport.OrEmpty(req)
	return stream.Paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (stream.Page[Chunk], error) {
		params.Cursor = cursor
		p, err := s.List(ctx, documentID, &params, opts...)
		if err != nil {
			return stream.Page[Chunk]{}, err
		}
		next := transport.Deref(p.NextCursor)
		return stream.Page[Chunk]{Items: p.Chunks, Next: next, HasMore: next != ""}, nil
	})
}

// SummariesService groups the document summaries endpoint.
type SummariesService struct{ cfg *requestconfig.Config }

// Get returns summaries of the given documents keyed by document ID.
func (s *SummariesService) Get(ctx context.Context, documentIDs []string, opts ...option.RequestOption) (map[string]agents.BatchResult[Document], error) {
	query := struct {
		DocumentIDs []string `query:"document_ids"`
	}{documentIDs}
	out, err := transport.GetValue[map[string]agents.BatchResult[Document]](ctx, s.cfg, "/v1/convai/knowledge-base/summaries", query, opts)
	if err != nil {
		return nil, err
	}
	return *out, nil
}
