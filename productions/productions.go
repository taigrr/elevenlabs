package productions

import (
	"context"
	"net/http"
	"time"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

// Service groups the /v1/productions endpoints for ordering
// human-reviewed dubbing, subtitles and transcription.
type Service struct {
	// Orders manages production orders.
	Orders *OrdersService
}

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	cfg := requestconfig.Build(opts...)
	return &Service{Orders: &OrdersService{
		cfg:          cfg,
		Deliverables: &DeliverablesService{cfg: cfg},
		Items:        &ItemsService{cfg: cfg},
		Languages:    &LanguagesService{cfg: cfg},
		Media:        &MediaService{cfg: cfg},
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

// OrdersService groups the /v1/productions/orders endpoints.
type OrdersService struct {
	cfg *requestconfig.Config

	// Deliverables lists finished files.
	Deliverables *DeliverablesService
	// Items edits the items of an order.
	Items *ItemsService
	// Languages lists supported languages.
	Languages *LanguagesService
	// Media uploads and inspects source media.
	Media *MediaService
}

// OrdersListRequest holds the optional parameters of OrdersService.List.
type OrdersListRequest struct {
	PageSize *int `query:"page_size"`
	Offset   *int `query:"offset"`
	// Status filters by one or more states.
	Status    []OrderState `query:"status"`
	StartDate *time.Time   `query:"start_date"`
	EndDate   *time.Time   `query:"end_date"`
}

// OrderList is the response of OrdersService.List.
type OrderList struct {
	Orders []OrderSummary `json:"orders"`
}

// List returns orders. req may be nil.
func (s *OrdersService) List(ctx context.Context, req *OrdersListRequest, opts ...option.RequestOption) (*OrderList, error) {
	return transport.DoValue[OrderList](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/productions/orders", Query: req, Opts: opts})
}

// OrderCreateRequest is the body of OrdersService.Create.
type OrderCreateRequest struct {
	// Sandbox creates a test order that is never fulfilled or charged.
	Sandbox *bool `json:"sandbox,omitempty"`
}

// OrderCreateResponse is the response of OrdersService.Create.
type OrderCreateResponse struct {
	OrderID string `json:"order_id"`
	Sandbox *bool  `json:"sandbox,omitempty"`
}

// Create opens a new order. req may be nil.
func (s *OrdersService) Create(ctx context.Context, req *OrderCreateRequest, opts ...option.RequestOption) (*OrderCreateResponse, error) {
	return transport.DoValue[OrderCreateResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/productions/orders", Body: transport.OrEmpty(req), Opts: opts})
}

// Get returns an order.
func (s *OrdersService) Get(ctx context.Context, orderID string, opts ...option.RequestOption) (*Order, error) {
	return transport.DoValue[Order](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/productions/orders/%s", orderID), Opts: opts})
}

type nameBody struct {
	Name string `json:"name"`
}

type wrappedRequest[T any] struct {
	Request T `json:"request"`
}

// OrderUpdateResponse is the response of OrdersService.Update.
type OrderUpdateResponse struct {
	Name string `json:"name"`
}

// Update renames an order.
func (s *OrdersService) Update(ctx context.Context, orderID, name string, opts ...option.RequestOption) (*OrderUpdateResponse, error) {
	return transport.DoValue[OrderUpdateResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPatch, Path: transport.Pathf("/v1/productions/orders/%s", orderID), Body: wrappedRequest[nameBody]{nameBody{name}}, Opts: opts})
}

// OrderSubmitResponse is the response of OrdersService.Submit.
type OrderSubmitResponse struct {
	OrderID     string     `json:"order_id"`
	State       OrderState `json:"state"`
	SubmittedAt time.Time  `json:"submitted_at"`
}

// Submit submits an order for fulfilment.
func (s *OrdersService) Submit(ctx context.Context, orderID string, opts ...option.RequestOption) (*OrderSubmitResponse, error) {
	return transport.DoValue[OrderSubmitResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/productions/orders/%s/submit", orderID), Opts: opts})
}

// DeliverablesService groups the order deliverables endpoint.
type DeliverablesService struct{ cfg *requestconfig.Config }

// Deliverable is a finished file of an order.
type Deliverable struct {
	SignedURL   string `json:"signed_url"`
	ContentType string `json:"content_type"`
	Name        string `json:"name"`
	Version     *int   `json:"version,omitempty"`
}

// DeliverableList is the response of DeliverablesService.List.
type DeliverableList struct {
	Deliverables []Deliverable `json:"deliverables"`
}

// List returns the deliverables of an order.
func (s *DeliverablesService) List(ctx context.Context, orderID string, opts ...option.RequestOption) (*DeliverableList, error) {
	return transport.DoValue[DeliverableList](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/productions/orders/%s/deliverables", orderID), Opts: opts})
}

// ItemsService groups the order items endpoints.
type ItemsService struct{ cfg *requestconfig.Config }

// OrderItemUpsertRequest is the body of ItemsService.Upsert.
type OrderItemUpsertRequest struct {
	Item OrderItem `json:"item"`
	// ItemID replaces an existing item; empty adds a new one.
	ItemID string `json:"item_id,omitempty"`
}

// OrderItemUpsertResponse is the response of ItemsService.Upsert.
type OrderItemUpsertResponse struct {
	ItemID string `json:"item_id"`
	Quote  *Quote `json:"quote,omitempty"`
}

// Upsert adds or replaces an order item.
func (s *ItemsService) Upsert(ctx context.Context, orderID string, req *OrderItemUpsertRequest, opts ...option.RequestOption) (*OrderItemUpsertResponse, error) {
	return transport.DoValue[OrderItemUpsertResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/productions/orders/%s/items", orderID), Body: wrappedRequest[*OrderItemUpsertRequest]{transport.OrEmpty(req)}, Opts: opts})
}

// SuccessResponse reports whether an operation succeeded.
type SuccessResponse struct {
	Success bool `json:"success"`
}

// Remove deletes an order item.
func (s *ItemsService) Remove(ctx context.Context, orderID, itemID string, opts ...option.RequestOption) (*SuccessResponse, error) {
	return transport.DoValue[SuccessResponse](ctx, s.cfg, &transport.Request{Method: http.MethodDelete, Path: transport.Pathf("/v1/productions/orders/%s/items/%s", orderID, itemID), Opts: opts})
}

// LanguagesService groups the supported languages endpoint.
type LanguagesService struct{ cfg *requestconfig.Config }

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

// Languages lists supported languages: Languages for single
// language kinds (Kind "single"), LanguagePairs for translation kinds
// (Kind "pair").
type Languages struct {
	Kind          string         `json:"kind"`
	Languages     []LanguageInfo `json:"languages,omitempty"`
	LanguagePairs []LanguagePair `json:"language_pairs,omitempty"`
}

// List returns the languages supported for an item kind.
func (s *LanguagesService) List(ctx context.Context, kind OrderItemKind, opts ...option.RequestOption) (*Languages, error) {
	return transport.DoValue[Languages](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/productions/orders/languages/%s", string(kind)), Opts: opts})
}

// MediaService groups the order media endpoints.
type MediaService struct{ cfg *requestconfig.Config }

// MediaRegisterRequest is the body of MediaService.Register.
// Provide either Media or MediaURL.
type MediaRegisterRequest struct {
	// DeclaredLanguage is the spoken language of the media. Required.
	DeclaredLanguage string `form:"declared_language"`
	// Media is the uploaded file.
	Media *types.File `form:"media"`
	// MediaURL is a URL to fetch the media from.
	MediaURL            string `form:"media_url"`
	MediaURLFilename    string `form:"media_url_filename"`
	MediaURLContentType string `form:"media_url_content_type"`
}

// MediaRegisterResponse is the response of MediaService.Register.
type MediaRegisterResponse struct {
	MediaID string `json:"media_id"`
}

// Register uploads or links source media for an order.
func (s *MediaService) Register(ctx context.Context, orderID string, req *MediaRegisterRequest, opts ...option.RequestOption) (*MediaRegisterResponse, error) {
	return transport.DoValue[MediaRegisterResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/productions/orders/%s/media", orderID), Form: transport.OrEmpty(req), Opts: opts})
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
func (s *MediaService) Get(ctx context.Context, orderID, mediaID string, opts ...option.RequestOption) (*OrderMedia, error) {
	return transport.DoValue[OrderMedia](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/productions/orders/%s/media/%s", orderID, mediaID), Opts: opts})
}
