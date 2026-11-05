package assets

import (
	"context"
	"iter"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/stream"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

// Service groups the /v1/assets endpoints.
type Service struct{ cfg *requestconfig.Config }

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	return &Service{cfg: requestconfig.Build(opts...)}
}

// Asset is an uploaded workspace asset.
type Asset struct {
	AssetID       string `json:"asset_id"`
	Name          string `json:"name"`
	MimeType      string `json:"mime_type"`
	CreatedAtUnix int64  `json:"created_at_unix"`
	// ContentURL is a signed URL valid for about an hour; nil while processing.
	ContentURL *string `json:"content_url,omitempty"`
}

// ListRequest holds the optional parameters of Service.List.
type ListRequest struct {
	// PageSize is the number of assets per page.
	PageSize *int `query:"page_size"`
	// Cursor continues a previous listing.
	Cursor string `query:"cursor"`
	// Search filters by asset name.
	Search string `query:"search"`
}

// ListResponse is one page of assets, most recent first.
type ListResponse struct {
	Assets     []Asset `json:"assets"`
	NextCursor *string `json:"next_cursor,omitempty"`
	HasMore    *bool   `json:"has_more,omitempty"`
}

// List returns one page of the workspace's assets. req may be nil.
func (s *Service) List(ctx context.Context, req *ListRequest, opts ...option.RequestOption) (*ListResponse, error) {
	return transport.DoValue[ListResponse](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/assets", Query: req, Opts: opts})
}

// ListAll iterates over every asset, fetching pages on demand.
func (s *Service) ListAll(ctx context.Context, req *ListRequest, opts ...option.RequestOption) iter.Seq2[Asset, error] {
	var params ListRequest
	if req != nil {
		params = *req
	}
	return stream.Paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (stream.Page[Asset], error) {
		params.Cursor = cursor
		resp, err := s.List(ctx, &params, opts...)
		if err != nil {
			return stream.Page[Asset]{}, err
		}
		next := ""
		if resp.NextCursor != nil {
			next = *resp.NextCursor
		}
		hasMore := next != ""
		if resp.HasMore != nil {
			hasMore = *resp.HasMore
		}
		return stream.Page[Asset]{Items: resp.Assets, Next: next, HasMore: hasMore}, nil
	})
}

// CreateRequest is the body of Service.Create.
type CreateRequest struct {
	// Asset is the file to upload. Required.
	Asset types.File `form:"asset"`
	// Name is the asset's display name. Required.
	Name string `form:"name"`
}

// Create uploads an asset.
func (s *Service) Create(ctx context.Context, req *CreateRequest, opts ...option.RequestOption) (*Asset, error) {
	return transport.DoValue[Asset](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/assets", Form: transport.OrEmpty(req), Opts: opts})
}

// Get returns an asset.
func (s *Service) Get(ctx context.Context, assetID string, opts ...option.RequestOption) (*Asset, error) {
	return transport.DoValue[Asset](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/assets/%s", assetID), Opts: opts})
}

// Delete deletes an asset.
func (s *Service) Delete(ctx context.Context, assetID string, opts ...option.RequestOption) error {
	return transport.DoJSON(ctx, s.cfg, &transport.Request{Method: http.MethodDelete, Path: transport.Pathf("/v1/assets/%s", assetID), Opts: opts}, nil)
}
