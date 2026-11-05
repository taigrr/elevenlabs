package productions

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/types"
)

func TestEndpoints(t *testing.T) {
	start := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	testutil.RunEndpoints(t, NewService, []testutil.Endpoint[*Service]{
		{
			Name: "productions.orders.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Orders.List(ctx, &OrdersListRequest{PageSize: types.Ptr(5), Status: []OrderState{OrderStateOpen, OrderStateDone}, StartDate: &start})
			},
			Method: http.MethodGet, Path: "/v1/productions/orders",
			Query:    "page_size=5&start_date=2026-01-02T03%3A04%3A05Z&status=open&status=done",
			RespBody: `{"orders":[{"order_id":"o","name":"n","state":"open","submitted_at":"2026-01-02T03:04:05Z"}]}`,
			Check: func(t *testing.T, r any) {
				if o := r.(*OrderList).Orders[0]; o.State != OrderStateOpen || !o.SubmittedAt.Equal(start) {
					t.Errorf("order = %+v", o)
				}
			},
		},
		{
			Name: "productions.orders.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Orders.Create(ctx, &OrderCreateRequest{Sandbox: types.Ptr(true)})
			},
			Method: http.MethodPost, Path: "/v1/productions/orders", Body: `{"sandbox":true}`, RespBody: `{"order_id":"o"}`,
		},
		{
			Name:   "productions.orders.get",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Orders.Get(ctx, "o") },
			Method: http.MethodGet, Path: "/v1/productions/orders/o",
			RespBody: `{"order_id":"o","name":"n","state":"paid","items":[{"item_id":"i","item":{"kind":"dub","media_id":"m","source_language":"en","destination_languages":["es"],"include_captions":true,"include_source_captions":false}}],"created_at":"2026-01-02T03:04:05Z"}`,
			Check: func(t *testing.T, r any) {
				o := r.(*Order)
				if o.Items[0].Item.Kind != OrderItemKindDub || !*o.Items[0].Item.IncludeCaptions {
					t.Errorf("order = %+v", o)
				}
			},
		},
		{
			Name:   "productions.orders.update",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Orders.Update(ctx, "o", "new") },
			Method: http.MethodPatch, Path: "/v1/productions/orders/o", Body: `{"request":{"name":"new"}}`, RespBody: `{"name":"new"}`,
		},
		{
			Name:   "productions.orders.submit",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Orders.Submit(ctx, "o") },
			Method: http.MethodPost, Path: "/v1/productions/orders/o/submit",
			RespBody: `{"order_id":"o","state":"submitted","submitted_at":"2026-01-02T03:04:05Z"}`,
		},
		{
			Name: "productions.orders.deliverables.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Orders.Deliverables.List(ctx, "o")
			},
			Method: http.MethodGet, Path: "/v1/productions/orders/o/deliverables",
			RespBody: `{"deliverables":[{"signed_url":"u","content_type":"audio/mpeg","name":"n"}]}`,
		},
		{
			Name: "productions.orders.items.upsert",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Orders.Items.Upsert(ctx, "o", &OrderItemUpsertRequest{Item: OrderItem{Kind: OrderItemKindTranscription, MediaIDs: []string{"m"}, SourceLanguage: "en", Verbatim: types.Ptr(true)}})
			},
			Method: http.MethodPost, Path: "/v1/productions/orders/o/items",
			Body:     `{"request":{"item":{"kind":"transcription","media_ids":["m"],"source_language":"en","verbatim":true}}}`,
			RespBody: `{"item_id":"i","quote":{"amount_usd":12.5}}`,
			Check: func(t *testing.T, r any) {
				if q := r.(*OrderItemUpsertResponse).Quote; q.AmountUSD != 12.5 {
					t.Errorf("quote = %+v", q)
				}
			},
		},
		{
			Name: "productions.orders.items.remove",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Orders.Items.Remove(ctx, "o", "i")
			},
			Method: http.MethodDelete, Path: "/v1/productions/orders/o/items/i", RespBody: `{"success":true}`,
		},
		{
			Name: "productions.orders.languages.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Orders.Languages.List(ctx, OrderItemKindSubtitles)
			},
			Method: http.MethodGet, Path: "/v1/productions/orders/languages/subtitles",
			RespBody: `{"kind":"pair","language_pairs":[{"source_language":{"code":"en","label":"English"},"destination_languages":[{"code":"es","label":"Spanish"}]}]}`,
			Check: func(t *testing.T, r any) {
				if l := r.(*Languages); l.LanguagePairs[0].DestinationLanguages[0].Code != "es" {
					t.Errorf("languages = %+v", l)
				}
			},
		},
		{
			Name: "productions.orders.media.register",
			Call: func(ctx context.Context, c *Service) (any, error) {
				f := testutil.File("v.mp4", "mp4")
				return c.Orders.Media.Register(ctx, "o", &MediaRegisterRequest{DeclaredLanguage: "en", Media: &f})
			},
			Method: http.MethodPost, Path: "/v1/productions/orders/o/media",
			Form:     map[string][]string{"declared_language": {"en"}, "media": {"file:v.mp4:mp4"}},
			RespBody: `{"media_id":"m"}`,
		},
		{
			Name: "productions.orders.media.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Orders.Media.Get(ctx, "o", "m")
			},
			Method: http.MethodGet, Path: "/v1/productions/orders/o/media/m",
			RespBody: `{"media_id":"m","name":"n","content_type":"video/mp4","signed_url":"u"}`,
		},
	})
}
