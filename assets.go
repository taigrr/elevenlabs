package elevenlabs

import (
	"context"
	"iter"
	"net/http"
)

// AssetsService groups the /v1/assets endpoints.
type AssetsService struct{ cfg *config }

// Asset is an uploaded workspace asset.
type Asset struct {
	AssetID       string `json:"asset_id"`
	Name          string `json:"name"`
	MimeType      string `json:"mime_type"`
	CreatedAtUnix int64  `json:"created_at_unix"`
	// ContentURL is a signed URL valid for about an hour; nil while processing.
	ContentURL *string `json:"content_url,omitempty"`
}

// AssetsListRequest holds the optional parameters of AssetsService.List.
type AssetsListRequest struct {
	// PageSize is the number of assets per page.
	PageSize *int `query:"page_size"`
	// Cursor continues a previous listing.
	Cursor string `query:"cursor"`
	// Search filters by asset name.
	Search string `query:"search"`
}

// AssetListResponse is one page of assets, most recent first.
type AssetListResponse struct {
	Assets     []Asset `json:"assets"`
	NextCursor *string `json:"next_cursor,omitempty"`
	HasMore    *bool   `json:"has_more,omitempty"`
}

// List returns one page of the workspace's assets. req may be nil.
func (s *AssetsService) List(ctx context.Context, req *AssetsListRequest, opts ...Option) (*AssetListResponse, error) {
	return doValue[AssetListResponse](ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/assets", query: req, opts: opts})
}

// ListAll iterates over every asset, fetching pages on demand.
func (s *AssetsService) ListAll(ctx context.Context, req *AssetsListRequest, opts ...Option) iter.Seq2[Asset, error] {
	var params AssetsListRequest
	if req != nil {
		params = *req
	}
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[Asset], error) {
		params.Cursor = cursor
		resp, err := s.List(ctx, &params, opts...)
		if err != nil {
			return page[Asset]{}, err
		}
		next := ""
		if resp.NextCursor != nil {
			next = *resp.NextCursor
		}
		hasMore := next != ""
		if resp.HasMore != nil {
			hasMore = *resp.HasMore
		}
		return page[Asset]{items: resp.Assets, next: next, hasMore: hasMore}, nil
	})
}

// AssetCreateRequest is the body of AssetsService.Create.
type AssetCreateRequest struct {
	// Asset is the file to upload. Required.
	Asset File `form:"asset"`
	// Name is the asset's display name. Required.
	Name string `form:"name"`
}

// Create uploads an asset.
func (s *AssetsService) Create(ctx context.Context, req *AssetCreateRequest, opts ...Option) (*Asset, error) {
	return doValue[Asset](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/assets", form: orEmpty(req), opts: opts})
}

// Get returns an asset.
func (s *AssetsService) Get(ctx context.Context, assetID string, opts ...Option) (*Asset, error) {
	return doValue[Asset](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/assets/%s", assetID), opts: opts})
}

// Delete deletes an asset.
func (s *AssetsService) Delete(ctx context.Context, assetID string, opts ...Option) error {
	return doJSON(ctx, s.cfg, &request{method: http.MethodDelete, path: pathf("/v1/assets/%s", assetID), opts: opts}, nil)
}
