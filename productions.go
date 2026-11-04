package elevenlabs

import (
	"context"
	"net/http"
	"time"
)

// ProductionsService groups the /v1/productions endpoints for ordering
// human-reviewed dubbing, subtitles and transcription.
type ProductionsService struct {
	// Orders manages production orders.
	Orders *ProductionOrdersService
}

func newProductionsService(cfg *config) *ProductionsService {
	return &ProductionsService{Orders: &ProductionOrdersService{
		cfg:          cfg,
		Deliverables: &ProductionDeliverablesService{cfg: cfg},
		Items:        &ProductionItemsService{cfg: cfg},
		Languages:    &ProductionLanguagesService{cfg: cfg},
		Media:        &ProductionMediaService{cfg: cfg},
	}}
}

// OrderState is the lifecycle state of a production order.
type OrderState string

// OrderState values.
const (
	OrderStateOpen       OrderState = "open"
	OrderStateSubmitted  OrderState = "submitted"
	OrderStatePaid       OrderState = "paid"
	OrderStateAccepted   OrderState = "accepted"
	OrderStateRejected   OrderState = "rejected"
	OrderStateDone       OrderState = "done"
	OrderStateCancelling OrderState = "cancelling"
	OrderStateCancelled  OrderState = "cancelled"
	OrderStateExpired    OrderState = "expired"
	OrderStateMerged     OrderState = "merged"
)

// OrderItemKind is the kind of work an order item requests.
type OrderItemKind string

// OrderItemKind values.
const (
	OrderItemKindDub           OrderItemKind = "dub"
	OrderItemKindSubtitles     OrderItemKind = "subtitles"
	OrderItemKindTranscription OrderItemKind = "transcription"
)

// CueOptions tune subtitle cue segmentation.
type CueOptions struct {
	MinDurationMs           *int     `json:"min_duration_ms,omitempty"`
	MaxDurationMs           *int     `json:"max_duration_ms,omitempty"`
	MaxLinesPerCue          *int     `json:"max_lines_per_cue,omitempty"`
	MaxCharsPerLine         *int     `json:"max_chars_per_line,omitempty"`
	MaxCharsPerS            *float64 `json:"max_chars_per_s,omitempty"`
	MinGapBetweenCuesFrames *int     `json:"min_gap_between_cues_frames,omitempty"`
}

// OrderItem is a unit of work in an order. Kind selects the variant:
//   - dub: MediaID, SourceLanguage, DestinationLanguages, IncludeCaptions,
//     IncludeSourceCaptions (required), CaptionsSDH;
//   - subtitles: MediaIDs, SourceLanguage, DestinationLanguages, CueOptions, SDH;
//   - transcription: MediaIDs, SourceLanguage, Verbatim.
type OrderItem struct {
	Kind                  OrderItemKind `json:"kind"`
	MediaID               string        `json:"media_id,omitempty"`
	MediaIDs              []string      `json:"media_ids,omitempty"`
	SourceLanguage        string        `json:"source_language"`
	DestinationLanguages  []string      `json:"destination_languages,omitempty"`
	IncludeCaptions       *bool         `json:"include_captions,omitempty"`
	IncludeSourceCaptions *bool         `json:"include_source_captions,omitempty"`
	CaptionsSDH           *bool         `json:"captions_sdh,omitempty"`
	CueOptions            *CueOptions   `json:"cue_options,omitempty"`
	SDH                   *bool         `json:"sdh,omitempty"`
	Verbatim              *bool         `json:"verbatim,omitempty"`
	Instructions          string        `json:"instructions,omitempty"`
}

// Quote is the price of an order item.
type Quote struct {
	AmountUSD float64 `json:"amount_usd"`
}

// OrderItemInfo is an item of an order with its quote.
type OrderItemInfo struct {
	ItemID string    `json:"item_id"`
	Item   OrderItem `json:"item"`
	Quote  *Quote    `json:"quote,omitempty"`
}

// OrderSummary is an order in a listing.
type OrderSummary struct {
	OrderID        string     `json:"order_id"`
	Name           string     `json:"name"`
	State          OrderState `json:"state"`
	TotalAmountUSD *float64   `json:"total_amount_usd,omitempty"`
	Sandbox        *bool      `json:"sandbox,omitempty"`
	SubmittedAt    *time.Time `json:"submitted_at,omitempty"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty"`
	CancelReason   *string    `json:"cancel_reason,omitempty"`
}

// Order is a production order.
type Order struct {
	OrderID        string          `json:"order_id"`
	Name           string          `json:"name"`
	State          OrderState      `json:"state"`
	Items          []OrderItemInfo `json:"items"`
	TotalAmountUSD *float64        `json:"total_amount_usd,omitempty"`
	Sandbox        *bool           `json:"sandbox,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	SubmittedAt    *time.Time      `json:"submitted_at,omitempty"`
	PaidAt         *time.Time      `json:"paid_at,omitempty"`
	AcceptedAt     *time.Time      `json:"accepted_at,omitempty"`
	CompletedAt    *time.Time      `json:"completed_at,omitempty"`
	CancelReason   *string         `json:"cancel_reason,omitempty"`
}

// ProductionOrdersService groups the /v1/productions/orders endpoints.
type ProductionOrdersService struct {
	cfg *config

	// Deliverables lists finished files.
	Deliverables *ProductionDeliverablesService
	// Items edits the items of an order.
	Items *ProductionItemsService
	// Languages lists supported languages.
	Languages *ProductionLanguagesService
	// Media uploads and inspects source media.
	Media *ProductionMediaService
}

// OrdersListRequest holds the optional parameters of ProductionOrdersService.List.
type OrdersListRequest struct {
	PageSize *int `query:"page_size"`
	Offset   *int `query:"offset"`
	// Status filters by one or more states.
	Status    []OrderState `query:"status"`
	StartDate *time.Time   `query:"start_date"`
	EndDate   *time.Time   `query:"end_date"`
}

// OrderList is the response of ProductionOrdersService.List.
type OrderList struct {
	Orders []OrderSummary `json:"orders"`
}

// List returns orders. req may be nil.
func (s *ProductionOrdersService) List(ctx context.Context, req *OrdersListRequest, opts ...RequestOption) (*OrderList, error) {
	return doValue[OrderList](ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/productions/orders", query: req, opts: opts})
}

// OrderCreateRequest is the body of ProductionOrdersService.Create.
type OrderCreateRequest struct {
	// Sandbox creates a test order that is never fulfilled or charged.
	Sandbox *bool `json:"sandbox,omitempty"`
}

// OrderCreateResponse is the response of ProductionOrdersService.Create.
type OrderCreateResponse struct {
	OrderID string `json:"order_id"`
	Sandbox *bool  `json:"sandbox,omitempty"`
}

// Create opens a new order. req may be nil.
func (s *ProductionOrdersService) Create(ctx context.Context, req *OrderCreateRequest, opts ...RequestOption) (*OrderCreateResponse, error) {
	return doValue[OrderCreateResponse](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/productions/orders", body: orEmpty(req), opts: opts})
}

// Get returns an order.
func (s *ProductionOrdersService) Get(ctx context.Context, orderID string, opts ...RequestOption) (*Order, error) {
	return doValue[Order](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/productions/orders/%s", orderID), opts: opts})
}

type wrappedRequest[T any] struct {
	Request T `json:"request"`
}

// OrderUpdateResponse is the response of ProductionOrdersService.Update.
type OrderUpdateResponse struct {
	Name string `json:"name"`
}

// Update renames an order.
func (s *ProductionOrdersService) Update(ctx context.Context, orderID, name string, opts ...RequestOption) (*OrderUpdateResponse, error) {
	return doValue[OrderUpdateResponse](ctx, s.cfg, &request{method: http.MethodPatch, path: pathf("/v1/productions/orders/%s", orderID), body: wrappedRequest[nameBody]{nameBody{name}}, opts: opts})
}

// OrderSubmitResponse is the response of ProductionOrdersService.Submit.
type OrderSubmitResponse struct {
	OrderID     string     `json:"order_id"`
	State       OrderState `json:"state"`
	SubmittedAt time.Time  `json:"submitted_at"`
}

// Submit submits an order for fulfilment.
func (s *ProductionOrdersService) Submit(ctx context.Context, orderID string, opts ...RequestOption) (*OrderSubmitResponse, error) {
	return doValue[OrderSubmitResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/productions/orders/%s/submit", orderID), opts: opts})
}

// ProductionDeliverablesService groups the order deliverables endpoint.
type ProductionDeliverablesService struct{ cfg *config }

// Deliverable is a finished file of an order.
type Deliverable struct {
	SignedURL   string `json:"signed_url"`
	ContentType string `json:"content_type"`
	Name        string `json:"name"`
	Version     *int   `json:"version,omitempty"`
}

// DeliverableList is the response of ProductionDeliverablesService.List.
type DeliverableList struct {
	Deliverables []Deliverable `json:"deliverables"`
}

// List returns the deliverables of an order.
func (s *ProductionDeliverablesService) List(ctx context.Context, orderID string, opts ...RequestOption) (*DeliverableList, error) {
	return doValue[DeliverableList](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/productions/orders/%s/deliverables", orderID), opts: opts})
}

// ProductionItemsService groups the order items endpoints.
type ProductionItemsService struct{ cfg *config }

// OrderItemUpsertRequest is the body of ProductionItemsService.Upsert.
type OrderItemUpsertRequest struct {
	Item OrderItem `json:"item"`
	// ItemID replaces an existing item; empty adds a new one.
	ItemID string `json:"item_id,omitempty"`
}

// OrderItemUpsertResponse is the response of ProductionItemsService.Upsert.
type OrderItemUpsertResponse struct {
	ItemID string `json:"item_id"`
	Quote  *Quote `json:"quote,omitempty"`
}

// Upsert adds or replaces an order item.
func (s *ProductionItemsService) Upsert(ctx context.Context, orderID string, req *OrderItemUpsertRequest, opts ...RequestOption) (*OrderItemUpsertResponse, error) {
	return doValue[OrderItemUpsertResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/productions/orders/%s/items", orderID), body: wrappedRequest[*OrderItemUpsertRequest]{orEmpty(req)}, opts: opts})
}

// SuccessResponse reports whether an operation succeeded.
type SuccessResponse struct {
	Success bool `json:"success"`
}

// Remove deletes an order item.
func (s *ProductionItemsService) Remove(ctx context.Context, orderID, itemID string, opts ...RequestOption) (*SuccessResponse, error) {
	return doValue[SuccessResponse](ctx, s.cfg, &request{method: http.MethodDelete, path: pathf("/v1/productions/orders/%s/items/%s", orderID, itemID), opts: opts})
}

// ProductionLanguagesService groups the supported languages endpoint.
type ProductionLanguagesService struct{ cfg *config }

// LanguageInfo is a supported language.
type LanguageInfo struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

// LanguagePair is a source language with its supported targets.
type LanguagePair struct {
	SourceLanguage       LanguageInfo   `json:"source_language"`
	DestinationLanguages []LanguageInfo `json:"destination_languages"`
}

// ProductionLanguages lists supported languages: Languages for single
// language kinds (Kind "single"), LanguagePairs for translation kinds
// (Kind "pair").
type ProductionLanguages struct {
	Kind          string         `json:"kind"`
	Languages     []LanguageInfo `json:"languages,omitempty"`
	LanguagePairs []LanguagePair `json:"language_pairs,omitempty"`
}

// List returns the languages supported for an item kind.
func (s *ProductionLanguagesService) List(ctx context.Context, kind OrderItemKind, opts ...RequestOption) (*ProductionLanguages, error) {
	return doValue[ProductionLanguages](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/productions/orders/languages/%s", string(kind)), opts: opts})
}

// ProductionMediaService groups the order media endpoints.
type ProductionMediaService struct{ cfg *config }

// MediaRegisterRequest is the body of ProductionMediaService.Register.
// Provide either Media or MediaURL.
type MediaRegisterRequest struct {
	// DeclaredLanguage is the spoken language of the media. Required.
	DeclaredLanguage string `form:"declared_language"`
	// Media is the uploaded file.
	Media *File `form:"media"`
	// MediaURL is a URL to fetch the media from.
	MediaURL            string `form:"media_url"`
	MediaURLFilename    string `form:"media_url_filename"`
	MediaURLContentType string `form:"media_url_content_type"`
}

// MediaRegisterResponse is the response of ProductionMediaService.Register.
type MediaRegisterResponse struct {
	MediaID string `json:"media_id"`
}

// Register uploads or links source media for an order.
func (s *ProductionMediaService) Register(ctx context.Context, orderID string, req *MediaRegisterRequest, opts ...RequestOption) (*MediaRegisterResponse, error) {
	return doValue[MediaRegisterResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/productions/orders/%s/media", orderID), form: orEmpty(req), opts: opts})
}

// OrderMedia describes registered media.
type OrderMedia struct {
	MediaID     string  `json:"media_id"`
	Name        string  `json:"name"`
	ContentType string  `json:"content_type"`
	Language    *string `json:"language,omitempty"`
	SignedURL   string  `json:"signed_url"`
}

// Get returns registered media.
func (s *ProductionMediaService) Get(ctx context.Context, orderID, mediaID string, opts ...RequestOption) (*OrderMedia, error) {
	return doValue[OrderMedia](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/productions/orders/%s/media/%s", orderID, mediaID), opts: opts})
}
