package assets

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/taigrr/elevenlabs/v2/apierror"
	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, []testutil.Endpoint[*Service]{
		{
			Name: "assets.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.List(ctx, &ListRequest{PageSize: types.Ptr(3), Search: "s"})
			},
			Method: http.MethodGet, Path: "/v1/assets", Query: "page_size=3&search=s",
			RespBody: `{"assets":[{"asset_id":"a","name":"n","mime_type":"audio/mpeg","created_at_unix":1}],"next_cursor":null,"has_more":false}`,
		},
		{
			Name: "assets.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Create(ctx, &CreateRequest{Asset: testutil.File("a.mp3", "M"), Name: "n"})
			},
			Method: http.MethodPost, Path: "/v1/assets",
			Form: map[string][]string{"asset": {"file:a.mp3:M"}, "name": {"n"}}, RespBody: `{"asset_id":"a","content_url":"https://c"}`,
			Check: func(t *testing.T, r any) {
				if a := r.(*Asset); a.AssetID != "a" || *a.ContentURL != "https://c" {
					t.Errorf("asset = %+v", a)
				}
			},
		},
		{
			Name:   "assets.get",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Get(ctx, "a") },
			Method: http.MethodGet, Path: "/v1/assets/a", RespBody: `{"asset_id":"a"}`,
		},
		{
			Name:   "assets.delete",
			Call:   func(ctx context.Context, c *Service) (any, error) { return voidCall(c.Delete(ctx, "a")) },
			Method: http.MethodDelete, Path: "/v1/assets/a",
		},
	})
}

func TestListAll(t *testing.T) {
	client, ts := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("cursor") == "" {
			_, _ = io.WriteString(w, `{"assets":[{"asset_id":"a"}],"next_cursor":"c2"}`)
			return
		}
		_, _ = io.WriteString(w, `{"assets":[{"asset_id":"b"}],"next_cursor":null}`)
	})
	var ids []string
	for a, err := range client.ListAll(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, a.AssetID)
	}
	if strings.Join(ids, ",") != "a,b" || len(ts.All()) != 2 {
		t.Errorf("ids = %v", ids)
	}
}

func TestListAllYieldsError(t *testing.T) {
	client, _ := newTestServer(t, testutil.Respond(http.StatusForbidden, transport.ContentTypeJSON, `{}`))
	for _, err := range client.ListAll(context.Background(), nil) {
		if !errors.Is(err, apierror.ErrForbidden) {
			t.Fatalf("error = %v", err)
		}
		return
	}
	t.Fatal("iterator yielded nothing")
}

func voidCall(err error) (any, error) { return nil, err }

func newTestServer(t *testing.T, h http.HandlerFunc, opts ...option.RequestOption) (*Service, *testutil.Server) {
	t.Helper()
	ts, o := testutil.NewServer(t, h, opts...)
	return NewService(o...), ts
}
