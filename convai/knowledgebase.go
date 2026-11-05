package convai

import (
	"context"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/convai/knowledgebase"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
)

// AddToKnowledgeBase uploads a file or URL into the knowledge base.
func (s *Service) AddToKnowledgeBase(ctx context.Context, req *knowledgebase.AddRequest, opts ...option.RequestOption) (*knowledgebase.CreateResponse, error) {
	req = transport.OrEmpty(req)
	return transport.DoValue[knowledgebase.CreateResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/convai/knowledge-base", Query: req, Form: req, Opts: opts})
}

// RAGIndexOverview reports RAG index storage usage.
func (s *Service) RAGIndexOverview(ctx context.Context, opts ...option.RequestOption) (*knowledgebase.RAGIndexOverview, error) {
	return transport.GetValue[knowledgebase.RAGIndexOverview](ctx, s.cfg, "/v1/convai/knowledge-base/rag-index", nil, opts)
}

// GetDocumentRAGIndexes lists a document's RAG indexes.
func (s *Service) GetDocumentRAGIndexes(ctx context.Context, documentID string, opts ...option.RequestOption) (*knowledgebase.RAGIndexes, error) {
	return transport.GetValue[knowledgebase.RAGIndexes](ctx, s.cfg, transport.Pathf("/v1/convai/knowledge-base/%s/rag-index", documentID), nil, opts)
}

// DeleteDocumentRAGIndex deletes one RAG index of a document.
func (s *Service) DeleteDocumentRAGIndex(ctx context.Context, documentID, ragIndexID string, opts ...option.RequestOption) (*knowledgebase.RAGIndex, error) {
	return transport.SendValue[knowledgebase.RAGIndex](ctx, s.cfg, http.MethodDelete, transport.Pathf("/v1/convai/knowledge-base/%s/rag-index/%s", documentID, ragIndexID), nil, nil, opts)
}
