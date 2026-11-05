package audioisolation

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, []testutil.Endpoint[*Service]{
		{
			Name: "audioIsolation.convert",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.ReadAll(c.Convert(ctx, &ConvertRequest{Audio: testutil.File("a.wav", "wav"), FileFormat: types.AudioFileFormatPCMS16LE16, PreviewB64: "p"}))
			},
			Method: http.MethodPost, Path: "/v1/audio-isolation", Accept: transport.AcceptAny, RespType: "audio/mpeg", RespBody: "iso",
			Form:  map[string][]string{"audio": {"file:a.wav:wav"}, "file_format": {"pcm_s16le_16"}, "preview_b64": {"p"}},
			Check: testutil.CheckString("iso"),
		},
		{
			Name: "audioIsolation.stream",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.ReadAll(c.Stream(ctx, &StreamRequest{Audio: testutil.File("a.wav", "wav")}))
			},
			Method: http.MethodPost, Path: "/v1/audio-isolation/stream", Accept: transport.AcceptAny, RespType: "audio/mpeg", RespBody: "iso",
			Form:  map[string][]string{"audio": {"file:a.wav:wav"}},
			Check: testutil.CheckString("iso"),
		},
		{
			Name: "audioIsolation.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.List(ctx, &ListRequest{PageSize: types.Ptr(5), Page: types.Ptr(2), Search: "x"})
			},
			Method: http.MethodGet, Path: "/v1/audio-isolation/history", Query: "page=2&page_size=5&search=x",
			RespBody: `{"items":[{"id":"i1","created_at_unix":3,"format":"mp3","supports_video":true}],"has_more":false}`,
			Check: func(t *testing.T, r any) {
				h := r.(*History)
				if len(h.Items) != 1 || h.Items[0].ID != "i1" || !h.Items[0].SupportsVideo {
					t.Errorf("history = %+v", h)
				}
			},
		},
		{
			Name: "audioIsolation.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return nil, c.Delete(ctx, "h/1")
			},
			Method: http.MethodDelete, Path: "/v1/audio-isolation/history/h%2F1", RespBody: `{}`,
		},
	})
}

func TestListAll(t *testing.T) {
	client, ts := newTestServer(t, testutil.Sequence(
		testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"items":[{"id":"a"}],"has_more":true}`),
		testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"items":[{"id":"b"}],"has_more":false}`),
	))
	var ids []string
	for it, err := range client.ListAll(context.Background(), &ListRequest{Search: "q"}) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, it.ID)
	}
	if strings.Join(ids, ",") != "a,b" {
		t.Errorf("ids = %v", ids)
	}
	if q := ts.Last(t).RawQuery; q != "page=2&search=q" {
		t.Errorf("second query = %q", q)
	}
}

func newTestServer(t *testing.T, h http.HandlerFunc, opts ...option.RequestOption) (*Service, *testutil.Server) {
	t.Helper()
	ts, o := testutil.NewServer(t, h, opts...)
	return NewService(o...), ts
}
