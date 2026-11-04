package elevenlabs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"strings"
)

// KnowledgeBaseService groups the /v1/convai/knowledge-base endpoints.
type KnowledgeBaseService struct {
	cfg *config

	// CrawlJobs crawls websites into the knowledge base.
	CrawlJobs *KnowledgeBaseCrawlJobsService
	// Document refreshes, re-uploads and indexes single documents.
	Document *KnowledgeBaseDocumentService
	// Documents creates, reads, updates and organizes documents.
	Documents *KnowledgeBaseDocumentsService
}

func newKnowledgeBaseService(cfg *config) *KnowledgeBaseService {
	return &KnowledgeBaseService{
		cfg:       cfg,
		CrawlJobs: &KnowledgeBaseCrawlJobsService{cfg: cfg},
		Document:  &KnowledgeBaseDocumentService{cfg: cfg},
		Documents: &KnowledgeBaseDocumentsService{
			cfg:       cfg,
			Chunk:     &KnowledgeBaseChunkService{cfg: cfg},
			Chunks:    &KnowledgeBaseChunksService{cfg: cfg},
			Summaries: &KnowledgeBaseSummariesService{cfg: cfg},
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

// KnowledgeBaseCreateResponse identifies a new document or folder.
type KnowledgeBaseCreateResponse struct {
	ID         string              `json:"id"`
	Name       string              `json:"name"`
	FolderPath []FolderPathSegment `json:"folder_path,omitempty"`
}

// KnowledgeBaseDocumentMetadata holds document timestamps and size.
type KnowledgeBaseDocumentMetadata struct {
	CreatedAtUnixSecs     int64 `json:"created_at_unix_secs"`
	LastUpdatedAtUnixSecs int64 `json:"last_updated_at_unix_secs"`
	SizeBytes             int64 `json:"size_bytes"`
}

// KnowledgeBaseDocument is a knowledge base entry: a file, URL, text
// document or folder (see Type). Fields not applicable to the type are
// zero. Summary listings set DependentAgents; full reads set
// ExtractedInnerHTML.
type KnowledgeBaseDocument struct {
	ID                 string                        `json:"id"`
	Name               string                        `json:"name"`
	Type               KnowledgeBaseDocumentType     `json:"type"`
	Metadata           KnowledgeBaseDocumentMetadata `json:"metadata"`
	SupportedUsages    []string                      `json:"supported_usages"`
	AccessInfo         ResourceAccessInfo            `json:"access_info"`
	FolderParentID     string                        `json:"folder_parent_id,omitempty"`
	FolderPath         []FolderPathSegment           `json:"folder_path,omitempty"`
	DependentAgents    []DependentAgent              `json:"dependent_agents,omitempty"`
	URL                string                        `json:"url,omitempty"`
	Filename           string                        `json:"filename,omitempty"`
	ExtractedInnerHTML string                        `json:"extracted_inner_html,omitempty"`
	ContentFormat      string                        `json:"content_format,omitempty"`
	ChildrenCount      *int                          `json:"children_count,omitempty"`
	DocumentCount      *int                          `json:"document_count,omitempty"`
	ExternalSyncInfo   json.RawMessage               `json:"external_sync_info,omitempty"`
	AutoSyncInfo       json.RawMessage               `json:"auto_sync_info,omitempty"`
	RefreshStatus      json.RawMessage               `json:"refresh_status,omitempty"`
	IsFrozen           *bool                         `json:"is_frozen,omitempty"`
	ActiveSyncJob      json.RawMessage               `json:"active_sync_job,omitempty"`
}

// DependentAgent is an agent that uses a resource. Type is "available"
// (all fields set) or "unknown" (only ID set, the caller lacks access).
type DependentAgent struct {
	Type                  string   `json:"type"`
	ID                    string   `json:"id"`
	Name                  string   `json:"name,omitempty"`
	CreatedAtUnixSecs     int64    `json:"created_at_unix_secs,omitempty"`
	AccessLevel           string   `json:"access_level,omitempty"`
	ReferencedResourceIDs []string `json:"referenced_resource_ids,omitempty"`
}

// DependentBranch is an agent branch that uses a resource.
type DependentBranch struct {
	AgentID    string `json:"agent_id"`
	AgentName  string `json:"agent_name"`
	BranchID   string `json:"branch_id"`
	BranchName string `json:"branch_name"`
	IsMain     bool   `json:"is_main"`
}

// DependentAgentsPage is a page of agents depending on a resource.
type DependentAgentsPage struct {
	Agents     []DependentAgent  `json:"agents"`
	Branches   []DependentBranch `json:"branches,omitempty"`
	NextCursor *string           `json:"next_cursor,omitempty"`
	HasMore    bool              `json:"has_more"`
}

// KnowledgeBaseAddRequest is the body of ConversationalAIService.AddToKnowledgeBase.
// Set either URL or File.
type KnowledgeBaseAddRequest struct {
	AgentID string `form:"-" query:"agent_id"`
	Name    string `form:"name"`
	URL     string `form:"url"`
	File    *File  `form:"file"`
}

// AddToKnowledgeBase uploads a file or URL into the knowledge base.
func (s *ConversationalAIService) AddToKnowledgeBase(ctx context.Context, req *KnowledgeBaseAddRequest, opts ...RequestOption) (*KnowledgeBaseCreateResponse, error) {
	req = orEmpty(req)
	return doValue[KnowledgeBaseCreateResponse](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/convai/knowledge-base", query: req, form: req, opts: opts})
}

// RAGIndexOverview reports RAG index storage usage.
func (s *ConversationalAIService) RAGIndexOverview(ctx context.Context, opts ...RequestOption) (*RAGIndexOverview, error) {
	return getValue[RAGIndexOverview](ctx, s.cfg, "/v1/convai/knowledge-base/rag-index", nil, opts)
}

// GetDocumentRAGIndexes lists a document's RAG indexes.
func (s *ConversationalAIService) GetDocumentRAGIndexes(ctx context.Context, documentID string, opts ...RequestOption) (*RAGIndexes, error) {
	return getValue[RAGIndexes](ctx, s.cfg, pathf("/v1/convai/knowledge-base/%s/rag-index", documentID), nil, opts)
}

// DeleteDocumentRAGIndex deletes one RAG index of a document.
func (s *ConversationalAIService) DeleteDocumentRAGIndex(ctx context.Context, documentID, ragIndexID string, opts ...RequestOption) (*RAGIndex, error) {
	return sendValue[RAGIndex](ctx, s.cfg, http.MethodDelete, pathf("/v1/convai/knowledge-base/%s/rag-index/%s", documentID, ragIndexID), nil, nil, opts)
}

// KnowledgeBaseSortBy selects the document list sort field.
type KnowledgeBaseSortBy string

// KnowledgeBaseSortBy values.
const (
	KnowledgeBaseSortByName      KnowledgeBaseSortBy = "name"
	KnowledgeBaseSortByCreatedAt KnowledgeBaseSortBy = "created_at"
	KnowledgeBaseSortByUpdatedAt KnowledgeBaseSortBy = "updated_at"
	KnowledgeBaseSortBySize      KnowledgeBaseSortBy = "size"
)

// KnowledgeBaseListRequest holds the optional parameters of KnowledgeBaseService.List.
type KnowledgeBaseListRequest struct {
	PageSize               *int                        `query:"page_size"`
	Search                 string                      `query:"search"`
	ShowOnlyOwnedDocuments *bool                       `query:"show_only_owned_documents"`
	CreatedByUserID        string                      `query:"created_by_user_id"`
	Types                  []KnowledgeBaseDocumentType `query:"types"`
	ParentFolderID         string                      `query:"parent_folder_id"`
	AncestorFolderID       string                      `query:"ancestor_folder_id"`
	FoldersFirst           *bool                       `query:"folders_first"`
	SortDirection          SortDirection               `query:"sort_direction"`
	SortBy                 KnowledgeBaseSortBy         `query:"sort_by"`
	Cursor                 string                      `query:"cursor"`
}

// KnowledgeBasePage is a page of knowledge base entries.
type KnowledgeBasePage struct {
	Documents  []KnowledgeBaseDocument `json:"documents"`
	NextCursor *string                 `json:"next_cursor,omitempty"`
	HasMore    bool                    `json:"has_more"`
}

// RAGIndexRequest asks for a document's RAG index.
type RAGIndexRequest struct {
	DocumentID      string         `json:"document_id"`
	CreateIfMissing bool           `json:"create_if_missing"`
	Model           EmbeddingModel `json:"model"`
}

// KnowledgeBaseSearchRequest holds the parameters of KnowledgeBaseService.Search.
type KnowledgeBaseSearchRequest struct {
	// Query is required.
	Query    string                      `query:"query"`
	PageSize *int                        `query:"page_size"`
	Types    []KnowledgeBaseDocumentType `query:"types"`
	Cursor   string                      `query:"cursor"`
}

// KnowledgeBaseSearchResult is a document matching a content search.
type KnowledgeBaseSearchResult struct {
	Document      KnowledgeBaseDocument `json:"document"`
	SearchSnippet []json.RawMessage     `json:"search_snippet,omitempty"`
	Score         float64               `json:"score"`
}

// KnowledgeBaseSearchPage is a page of content search results.
type KnowledgeBaseSearchPage struct {
	Results    []KnowledgeBaseSearchResult `json:"results"`
	NextCursor *string                     `json:"next_cursor,omitempty"`
}

// List returns one page of knowledge base entries. req may be nil.
func (s *KnowledgeBaseService) List(ctx context.Context, req *KnowledgeBaseListRequest, opts ...RequestOption) (*KnowledgeBasePage, error) {
	return getValue[KnowledgeBasePage](ctx, s.cfg, "/v1/convai/knowledge-base", req, opts)
}

// ListAll iterates over every knowledge base entry.
func (s *KnowledgeBaseService) ListAll(ctx context.Context, req *KnowledgeBaseListRequest, opts ...RequestOption) iter.Seq2[KnowledgeBaseDocument, error] {
	params := *orEmpty(req)
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[KnowledgeBaseDocument], error) {
		params.Cursor = cursor
		p, err := s.List(ctx, &params, opts...)
		if err != nil {
			return page[KnowledgeBaseDocument]{}, err
		}
		return page[KnowledgeBaseDocument]{items: p.Documents, next: deref(p.NextCursor), hasMore: p.HasMore}, nil
	})
}

// GetOrCreateRAGIndexes returns, and optionally creates, RAG indexes for
// several documents, keyed by document ID.
func (s *KnowledgeBaseService) GetOrCreateRAGIndexes(ctx context.Context, items []RAGIndexRequest, opts ...RequestOption) (map[string]BatchResult[RAGIndex], error) {
	body := struct {
		Items []RAGIndexRequest `json:"items"`
	}{nonNil(items)}
	out, err := sendValue[map[string]BatchResult[RAGIndex]](ctx, s.cfg, http.MethodPost, "/v1/convai/knowledge-base/rag-index", nil, body, opts)
	if err != nil {
		return nil, err
	}
	return *out, nil
}

// Search searches knowledge base contents.
func (s *KnowledgeBaseService) Search(ctx context.Context, req *KnowledgeBaseSearchRequest, opts ...RequestOption) (*KnowledgeBaseSearchPage, error) {
	return getValue[KnowledgeBaseSearchPage](ctx, s.cfg, "/v1/convai/knowledge-base/search", req, opts)
}

// SearchAll iterates over every content search result.
func (s *KnowledgeBaseService) SearchAll(ctx context.Context, req *KnowledgeBaseSearchRequest, opts ...RequestOption) iter.Seq2[KnowledgeBaseSearchResult, error] {
	params := *orEmpty(req)
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[KnowledgeBaseSearchResult], error) {
		params.Cursor = cursor
		p, err := s.Search(ctx, &params, opts...)
		if err != nil {
			return page[KnowledgeBaseSearchResult]{}, err
		}
		next := deref(p.NextCursor)
		return page[KnowledgeBaseSearchResult]{items: p.Results, next: next, hasMore: next != ""}, nil
	})
}

// KnowledgeBaseCrawlJobsService groups the /v1/convai/knowledge-base/crawl endpoints.
type KnowledgeBaseCrawlJobsService struct{ cfg *config }

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

// CrawlJobListRequest holds the optional parameters of KnowledgeBaseCrawlJobsService.List.
type CrawlJobListRequest struct {
	IncludeJobIDs []string `query:"include_job_ids"`
	PageSize      *int     `query:"page_size"`
	Cursor        string   `query:"cursor"`
}

// CrawlJobCreateRequest is the body of KnowledgeBaseCrawlJobsService.Create.
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
func (s *KnowledgeBaseCrawlJobsService) List(ctx context.Context, req *CrawlJobListRequest, opts ...RequestOption) (*CrawlJobPage, error) {
	return getValue[CrawlJobPage](ctx, s.cfg, "/v1/convai/knowledge-base/crawl", req, opts)
}

// Create starts a crawl job.
func (s *KnowledgeBaseCrawlJobsService) Create(ctx context.Context, req *CrawlJobCreateRequest, opts ...RequestOption) (*CrawlJobCreateResponse, error) {
	return sendValue[CrawlJobCreateResponse](ctx, s.cfg, http.MethodPost, "/v1/convai/knowledge-base/crawl", nil, orEmpty(req), opts)
}

// Get returns a crawl job.
func (s *KnowledgeBaseCrawlJobsService) Get(ctx context.Context, crawlJobID string, opts ...RequestOption) (*CrawlJob, error) {
	return getValue[CrawlJob](ctx, s.cfg, pathf("/v1/convai/knowledge-base/crawl/%s", crawlJobID), nil, opts)
}

// Cancel cancels a crawl job.
func (s *KnowledgeBaseCrawlJobsService) Cancel(ctx context.Context, crawlJobID string, opts ...RequestOption) (json.RawMessage, error) {
	return sendRaw(ctx, s.cfg, http.MethodPost, pathf("/v1/convai/knowledge-base/crawl/%s/cancel", crawlJobID), nil, nil, opts)
}

// KnowledgeBaseDocumentService groups single-document maintenance endpoints.
type KnowledgeBaseDocumentService struct{ cfg *config }

// UpdateFile replaces the file of a file document.
func (s *KnowledgeBaseDocumentService) UpdateFile(ctx context.Context, documentID string, file File, opts ...RequestOption) (*KnowledgeBaseDocument, error) {
	form := struct {
		File File `form:"file"`
	}{file}
	return doValue[KnowledgeBaseDocument](ctx, s.cfg, &request{method: http.MethodPatch, path: pathf("/v1/convai/knowledge-base/%s/update-file", documentID), form: &form, opts: opts})
}

// Refresh re-fetches a URL document or re-syncs a file document.
func (s *KnowledgeBaseDocumentService) Refresh(ctx context.Context, documentID string, opts ...RequestOption) (*KnowledgeBaseDocument, error) {
	return sendValue[KnowledgeBaseDocument](ctx, s.cfg, http.MethodPost, pathf("/v1/convai/knowledge-base/%s/refresh", documentID), nil, nil, opts)
}

// ComputeRAGIndex starts computing a document's RAG index for model.
func (s *KnowledgeBaseDocumentService) ComputeRAGIndex(ctx context.Context, documentID string, model EmbeddingModel, opts ...RequestOption) (*RAGIndex, error) {
	body := struct {
		Model EmbeddingModel `json:"model"`
	}{model}
	return sendValue[RAGIndex](ctx, s.cfg, http.MethodPost, pathf("/v1/convai/knowledge-base/%s/rag-index", documentID), nil, body, opts)
}

// KnowledgeBaseDocumentsService groups the knowledge base document endpoints.
type KnowledgeBaseDocumentsService struct {
	cfg *config

	// Chunk reads a single RAG chunk.
	Chunk *KnowledgeBaseChunkService
	// Chunks lists RAG chunks.
	Chunks *KnowledgeBaseChunksService
	// Summaries returns document summaries in bulk.
	Summaries *KnowledgeBaseSummariesService
}

// AutoSyncOptions control automatic re-syncing of URL documents and folders.
type AutoSyncOptions struct {
	EnableAutoSync       *bool `json:"enable_auto_sync,omitempty"`
	AutoRemove           *bool `json:"auto_remove,omitempty"`
	MinimumFrequencyDays *int  `json:"minimum_frequency_days,omitempty"`
}

// DocumentFromURLRequest is the body of KnowledgeBaseDocumentsService.CreateFromURL.
type DocumentFromURLRequest struct {
	// URL is required.
	URL            string `json:"url"`
	Name           string `json:"name,omitempty"`
	ParentFolderID string `json:"parent_folder_id,omitempty"`
	AutoSyncOptions
}

// DocumentFromFileRequest is the form of KnowledgeBaseDocumentsService.CreateFromFile.
type DocumentFromFileRequest struct {
	// File is required.
	File           File   `form:"file"`
	Name           string `form:"name"`
	ParentFolderID string `form:"parent_folder_id"`
}

// DocumentFromTextRequest is the body of KnowledgeBaseDocumentsService.CreateFromText.
type DocumentFromTextRequest struct {
	// Text is required.
	Text           string `json:"text"`
	Name           string `json:"name,omitempty"`
	ParentFolderID string `json:"parent_folder_id,omitempty"`
}

// FolderCreateRequest is the body of KnowledgeBaseDocumentsService.CreateFolder.
type FolderCreateRequest struct {
	// Name is required.
	Name           string `json:"name"`
	ParentFolderID string `json:"parent_folder_id,omitempty"`
	AutoSyncOptions
}

// DocumentUpdateRequest is the body of KnowledgeBaseDocumentsService.Update.
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
func (s *KnowledgeBaseDocumentsService) CreateFromURL(ctx context.Context, req *DocumentFromURLRequest, opts ...RequestOption) (*KnowledgeBaseCreateResponse, error) {
	return sendValue[KnowledgeBaseCreateResponse](ctx, s.cfg, http.MethodPost, "/v1/convai/knowledge-base/url", nil, orEmpty(req), opts)
}

// CreateFromFile uploads a file as a document.
func (s *KnowledgeBaseDocumentsService) CreateFromFile(ctx context.Context, req *DocumentFromFileRequest, opts ...RequestOption) (*KnowledgeBaseCreateResponse, error) {
	return doValue[KnowledgeBaseCreateResponse](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/convai/knowledge-base/file", form: orEmpty(req), opts: opts})
}

// CreateFromText creates a document from text.
func (s *KnowledgeBaseDocumentsService) CreateFromText(ctx context.Context, req *DocumentFromTextRequest, opts ...RequestOption) (*KnowledgeBaseCreateResponse, error) {
	return sendValue[KnowledgeBaseCreateResponse](ctx, s.cfg, http.MethodPost, "/v1/convai/knowledge-base/text", nil, orEmpty(req), opts)
}

// CreateFolder creates a folder.
func (s *KnowledgeBaseDocumentsService) CreateFolder(ctx context.Context, req *FolderCreateRequest, opts ...RequestOption) (*KnowledgeBaseCreateResponse, error) {
	return sendValue[KnowledgeBaseCreateResponse](ctx, s.cfg, http.MethodPost, "/v1/convai/knowledge-base/folder", nil, orEmpty(req), opts)
}

// Get returns a document. agentID may be empty.
func (s *KnowledgeBaseDocumentsService) Get(ctx context.Context, documentID, agentID string, opts ...RequestOption) (*KnowledgeBaseDocument, error) {
	query := struct {
		AgentID string `query:"agent_id"`
	}{agentID}
	return getValue[KnowledgeBaseDocument](ctx, s.cfg, pathf("/v1/convai/knowledge-base/%s", documentID), query, opts)
}

// Delete deletes a document or folder. req may be nil.
func (s *KnowledgeBaseDocumentsService) Delete(ctx context.Context, documentID string, req *ForceRequest, opts ...RequestOption) (json.RawMessage, error) {
	return sendRaw(ctx, s.cfg, http.MethodDelete, pathf("/v1/convai/knowledge-base/%s", documentID), req, nil, opts)
}

// Update renames a document or replaces a text document's content.
func (s *KnowledgeBaseDocumentsService) Update(ctx context.Context, documentID string, req *DocumentUpdateRequest, opts ...RequestOption) (*KnowledgeBaseDocument, error) {
	return sendValue[KnowledgeBaseDocument](ctx, s.cfg, http.MethodPatch, pathf("/v1/convai/knowledge-base/%s", documentID), nil, orEmpty(req), opts)
}

// GetAgents lists the agents that use a document. req may be nil.
func (s *KnowledgeBaseDocumentsService) GetAgents(ctx context.Context, documentID string, req *DependentAgentsRequest, opts ...RequestOption) (*DependentAgentsPage, error) {
	return getValue[DependentAgentsPage](ctx, s.cfg, pathf("/v1/convai/knowledge-base/%s/dependent-agents", documentID), req, opts)
}

// GetBulkAgents lists the agents that use any of documentIDs. req may be nil.
func (s *KnowledgeBaseDocumentsService) GetBulkAgents(ctx context.Context, documentIDs []string, req *DependentAgentsRequest, opts ...RequestOption) (*DependentAgentsPage, error) {
	body := struct {
		DocumentIDs []string `json:"document_ids"`
	}{nonNil(documentIDs)}
	return sendValue[DependentAgentsPage](ctx, s.cfg, http.MethodPost, "/v1/convai/knowledge-base/dependent-agents", req, body, opts)
}

// GetContent returns a document's extracted content as HTML.
func (s *KnowledgeBaseDocumentsService) GetContent(ctx context.Context, documentID string, opts ...RequestOption) (string, error) {
	body, err := doStream(ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/convai/knowledge-base/%s/content", documentID), accept: acceptAny, opts: opts})
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
func (s *KnowledgeBaseDocumentsService) GetSourceFileURL(ctx context.Context, documentID string, opts ...RequestOption) (*SignedFileURL, error) {
	return getValue[SignedFileURL](ctx, s.cfg, pathf("/v1/convai/knowledge-base/%s/source-file-url", documentID), nil, opts)
}

// Move moves a document or folder into folderID (the root when empty).
func (s *KnowledgeBaseDocumentsService) Move(ctx context.Context, documentID, folderID string, opts ...RequestOption) error {
	body := struct {
		MoveTo string `json:"move_to,omitempty"`
	}{folderID}
	return doNoContent(ctx, s.cfg, http.MethodPost, pathf("/v1/convai/knowledge-base/%s/move", documentID), nil, body, opts)
}

// BulkMove moves documents into folderID (the root when empty).
func (s *KnowledgeBaseDocumentsService) BulkMove(ctx context.Context, documentIDs []string, folderID string, opts ...RequestOption) error {
	body := struct {
		DocumentIDs []string `json:"document_ids"`
		MoveTo      string   `json:"move_to,omitempty"`
	}{nonNil(documentIDs), folderID}
	return doNoContent(ctx, s.cfg, http.MethodPost, "/v1/convai/knowledge-base/bulk-move", nil, body, opts)
}

// DeletedDocument identifies a deleted document.
type DeletedDocument struct {
	ID string `json:"id"`
}

// BulkDelete deletes several documents, returning per-document results
// keyed by document ID. force deletes documents still used by agents.
func (s *KnowledgeBaseDocumentsService) BulkDelete(ctx context.Context, documentIDs []string, force *bool, opts ...RequestOption) (map[string]BatchResult[DeletedDocument], error) {
	body := struct {
		DocumentIDs []string `json:"document_ids"`
		Force       *bool    `json:"force,omitempty"`
	}{nonNil(documentIDs), force}
	out, err := sendValue[map[string]BatchResult[DeletedDocument]](ctx, s.cfg, http.MethodPost, "/v1/convai/knowledge-base/bulk-delete", nil, body, opts)
	if err != nil {
		return nil, err
	}
	return *out, nil
}

// KnowledgeBaseChunk is a RAG chunk of a document.
type KnowledgeBaseChunk struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Content string `json:"content"`
}

// KnowledgeBaseChunkPage is a page of chunks.
type KnowledgeBaseChunkPage struct {
	Chunks     []KnowledgeBaseChunk `json:"chunks"`
	NextCursor *string              `json:"next_cursor,omitempty"`
}

// KnowledgeBaseChunkService groups the single chunk endpoint.
type KnowledgeBaseChunkService struct{ cfg *config }

// Get returns one chunk. model may be empty.
func (s *KnowledgeBaseChunkService) Get(ctx context.Context, documentID, chunkID string, model EmbeddingModel, opts ...RequestOption) (*KnowledgeBaseChunk, error) {
	query := struct {
		EmbeddingModel EmbeddingModel `query:"embedding_model"`
	}{model}
	return getValue[KnowledgeBaseChunk](ctx, s.cfg, pathf("/v1/convai/knowledge-base/%s/chunk/%s", documentID, chunkID), query, opts)
}

// KnowledgeBaseChunksService groups the chunk listing endpoint.
type KnowledgeBaseChunksService struct{ cfg *config }

// ChunkListRequest holds the parameters of KnowledgeBaseChunksService.List.
type ChunkListRequest struct {
	// EmbeddingModel is required.
	EmbeddingModel EmbeddingModel `query:"embedding_model"`
	PageSize       *int           `query:"page_size"`
	Cursor         string         `query:"cursor"`
}

// List returns one page of a document's chunks.
func (s *KnowledgeBaseChunksService) List(ctx context.Context, documentID string, req *ChunkListRequest, opts ...RequestOption) (*KnowledgeBaseChunkPage, error) {
	return getValue[KnowledgeBaseChunkPage](ctx, s.cfg, pathf("/v1/convai/knowledge-base/%s/chunks", documentID), req, opts)
}

// ListAll iterates over every chunk of a document.
func (s *KnowledgeBaseChunksService) ListAll(ctx context.Context, documentID string, req *ChunkListRequest, opts ...RequestOption) iter.Seq2[KnowledgeBaseChunk, error] {
	params := *orEmpty(req)
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[KnowledgeBaseChunk], error) {
		params.Cursor = cursor
		p, err := s.List(ctx, documentID, &params, opts...)
		if err != nil {
			return page[KnowledgeBaseChunk]{}, err
		}
		next := deref(p.NextCursor)
		return page[KnowledgeBaseChunk]{items: p.Chunks, next: next, hasMore: next != ""}, nil
	})
}

// KnowledgeBaseSummariesService groups the document summaries endpoint.
type KnowledgeBaseSummariesService struct{ cfg *config }

// Get returns summaries of the given documents keyed by document ID.
func (s *KnowledgeBaseSummariesService) Get(ctx context.Context, documentIDs []string, opts ...RequestOption) (map[string]BatchResult[KnowledgeBaseDocument], error) {
	query := struct {
		DocumentIDs []string `query:"document_ids"`
	}{documentIDs}
	out, err := getValue[map[string]BatchResult[KnowledgeBaseDocument]](ctx, s.cfg, "/v1/convai/knowledge-base/summaries", query, opts)
	if err != nil {
		return nil, err
	}
	return *out, nil
}
