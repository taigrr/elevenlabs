package history

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
			Name: "history.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.List(ctx, &ListRequest{
					PageSize: types.Ptr(10), StartAfterHistoryItemID: "h0", VoiceID: "v", ModelID: "m",
					DateBeforeUnix: types.Ptr(int64(200)), DateAfterUnix: types.Ptr(int64(100)),
					SortDirection: types.SortDirectionAsc, Search: "hello", Source: SourceTTS,
				})
			},
			Method:   http.MethodGet,
			Path:     "/v1/history",
			Query:    "date_after_unix=100&date_before_unix=200&model_id=m&page_size=10&search=hello&sort_direction=asc&source=TTS&start_after_history_item_id=h0&voice_id=v",
			RespBody: `{"history":[{"history_item_id":"h1","date_unix":5,"source":"TTS","settings":{"a":1}}],"last_history_item_id":"h1","has_more":false}`,
			Check: func(t *testing.T, r any) {
				h := r.(*ListResponse)
				if len(h.History) != 1 || h.History[0].HistoryItemID != "h1" || *h.History[0].Source != SourceTTS || *h.LastHistoryItemID != "h1" {
					t.Errorf("history = %+v", h)
				}
			},
		},
		{
			Name:     "history.get",
			Call:     func(ctx context.Context, c *Service) (any, error) { return c.Get(ctx, "h1") },
			Method:   http.MethodGet,
			Path:     "/v1/history/h1",
			RespBody: `{"history_item_id":"h1","alignments":{"alignment":{"characters":["a"],"character_start_times_seconds":[0],"character_end_times_seconds":[1]},"normalized_alignment":{"characters":[],"character_start_times_seconds":[],"character_end_times_seconds":[]}}}`,
			Check: func(t *testing.T, r any) {
				if h := r.(*Item); h.Alignments.Alignment.Characters[0] != "a" {
					t.Errorf("item = %+v", h)
				}
			},
		},
		{
			Name:     "history.delete",
			Call:     func(ctx context.Context, c *Service) (any, error) { return c.Delete(ctx, "h1") },
			Method:   http.MethodDelete,
			Path:     "/v1/history/h1",
			RespBody: `{"status":"ok"}`,
			Check:    checkStatusOK,
		},
		{
			Name:     "history.getAudio",
			Call:     func(ctx context.Context, c *Service) (any, error) { return testutil.ReadAll(c.GetAudio(ctx, "h1")) },
			Method:   http.MethodGet,
			Path:     "/v1/history/h1/audio",
			Accept:   transport.AcceptAny,
			RespType: "audio/mpeg",
			RespBody: "mp3data",
			Check:    testutil.CheckString("mp3data"),
		},
		{
			Name: "history.download",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.ReadAll(c.Download(ctx, &DownloadRequest{HistoryItemIDs: []string{"a", "b"}, OutputFormat: "wav"}))
			},
			Method:   http.MethodPost,
			Path:     "/v1/history/download",
			Body:     `{"history_item_ids":["a","b"],"output_format":"wav"}`,
			Accept:   transport.AcceptAny,
			RespType: "application/zip",
			RespBody: "PK",
			Check:    testutil.CheckString("PK"),
		},
	})
}

func TestListAll(t *testing.T) {
	pages := map[string]string{
		"":   `{"history":[{"history_item_id":"h1"},{"history_item_id":"h2"}],"last_history_item_id":"h2","has_more":true}`,
		"h2": `{"history":[{"history_item_id":"h3"}],"last_history_item_id":"h3","has_more":false}`,
	}
	client, ts := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		body, ok := pages[r.URL.Query().Get("start_after_history_item_id")]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = io.WriteString(w, body)
	})
	var ids []string
	for item, err := range client.ListAll(context.Background(), &ListRequest{PageSize: types.Ptr(2), VoiceID: "v"}) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, item.HistoryItemID)
	}
	if strings.Join(ids, ",") != "h1,h2,h3" {
		t.Errorf("ids = %v", ids)
	}
	reqs := ts.All()
	if len(reqs) != 2 || reqs[1].RawQuery != "page_size=2&start_after_history_item_id=h2&voice_id=v" {
		t.Errorf("requests = %+v", reqs)
	}
}

func TestListAllYieldsError(t *testing.T) {
	client, _ := newTestServer(t, testutil.Respond(http.StatusUnauthorized, transport.ContentTypeJSON, `{}`))
	for _, err := range client.ListAll(context.Background(), nil) {
		if !errors.Is(err, apierror.ErrUnauthorized) {
			t.Fatalf("error = %v", err)
		}
		return
	}
	t.Fatal("iterator yielded nothing")
}

func checkStatusOK(t *testing.T, r any) {
	t.Helper()
	if s := r.(*types.StatusResponse); s.Status != "ok" {
		t.Errorf("status = %q", s.Status)
	}
}

func newTestServer(t *testing.T, h http.HandlerFunc, opts ...option.RequestOption) (*Service, *testutil.Server) {
	t.Helper()
	ts, o := testutil.NewServer(t, h, opts...)
	return NewService(o...), ts
}
