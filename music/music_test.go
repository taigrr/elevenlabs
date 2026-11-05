package music

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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
			Name: "compose",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.ReadAll(c.Compose(ctx, &ComposeRequest{OutputFormat: types.OutputFormatMP3_44100_128, GenerationParams: GenerationParams{Prompt: "jazz", MusicLengthMs: types.Ptr(int64(10000)), ForceInstrumental: types.Ptr(true)}, SignWithC2PA: types.Ptr(false)}))
			},
			Method: http.MethodPost, Path: "/v1/music", Accept: transport.AcceptAny, RespType: "audio/mpeg", RespBody: "song",
			Query: "output_format=mp3_44100_128",
			Body:  `{"prompt":"jazz","music_length_ms":10000,"force_instrumental":true,"sign_with_c2pa":false}`,
			Check: testutil.CheckString("song"),
		},
		{
			Name: "stream",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.ReadAll(c.Stream(ctx, &StreamRequest{GenerationParams: GenerationParams{Prompt: "jazz"}}))
			},
			Method: http.MethodPost, Path: "/v1/music/stream", Accept: transport.AcceptAny, RespType: "audio/mpeg", RespBody: "song",
			Body: `{"prompt":"jazz"}`, Check: testutil.CheckString("song"),
		},
		{
			Name: "composeDetailed",
			Call: func(ctx context.Context, c *Service) (any, error) {
				resp, err := c.ComposeDetailed(ctx, &ComposeDetailedRequest{GenerationParams: GenerationParams{Prompt: "pop"}, WithTimestamps: types.Ptr(true)})
				if err != nil {
					return nil, err
				}
				defer resp.Close()
				audio, err := io.ReadAll(resp)
				if err != nil {
					return nil, err
				}
				if resp.Filename != "song.mp3" || resp.Metadata.SongMetadata.Title != "T" || resp.Metadata.CompositionPlan.Prompt == nil {
					return nil, io.ErrUnexpectedEOF
				}
				return string(audio), nil
			},
			Method: http.MethodPost, Path: "/v1/music/detailed", Accept: transport.AcceptAny,
			RespType: "multipart/mixed; boundary=b", RespBody: musicMultipartBody,
			Body: `{"prompt":"pop","with_timestamps":true}`, Check: testutil.CheckString("MP3DATA"),
		},
		{
			Name: "composeDetailedStream",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.DrainStream(c.ComposeDetailedStream(ctx, &ComposeDetailedStreamRequest{GenerationParams: GenerationParams{Prompt: "pop"}}))
			},
			Method: http.MethodPost, Path: "/v1/music/detailed/stream",
			Body: `{"prompt":"pop"}`, RespBody: "{\"a\":1}\n{\"b\":2}\n",
			Check: func(t *testing.T, r any) {
				if c := r.([]json.RawMessage); len(c) != 2 {
					t.Errorf("chunks = %s", c)
				}
			},
		},
		{
			Name: "videoToMusic",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.ReadAll(c.VideoToMusic(ctx, &VideoToMusicRequest{Videos: []types.File{testutil.File("a.mp4", "v1"), testutil.File("b.mp4", "v2")}, Description: "epic", Tags: []string{"x", "y"}, AddSFX: types.Ptr(true)}))
			},
			Method: http.MethodPost, Path: "/v1/music/video-to-music", Accept: transport.AcceptAny, RespType: "audio/mpeg", RespBody: "song",
			Form:  map[string][]string{"videos": {"file:a.mp4:v1", "file:b.mp4:v2"}, "description": {"epic"}, "tags": {"x", "y"}, "add_sfx": {"true"}},
			Check: testutil.CheckString("song"),
		},
		{
			Name: "upload",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Upload(ctx, &UploadRequest{File: testutil.File("s.mp3", "mp3"), WithTimestamps: types.Ptr(true)})
			},
			Method: http.MethodPost, Path: "/v1/music/upload",
			Form:     map[string][]string{"file": {"file:s.mp3:mp3"}, "with_timestamps": {"true"}},
			RespBody: `{"song_id":"s1","words_timestamps":[{"word":"la","start_ms":0,"end_ms":5}]}`,
			Check: func(t *testing.T, r any) {
				if u := r.(*Upload); u.SongID != "s1" || u.WordsTimestamps[0].EndMs != 5 {
					t.Errorf("upload = %+v", u)
				}
			},
		},
		{
			Name: "separateStems",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.ReadAll(c.SeparateStems(ctx, &SeparateStemsRequest{OutputFormat: types.OutputFormatMP3_44100_128, File: testutil.File("s.mp3", "mp3"), StemVariationID: StemVariation("two_stems_v1")}))
			},
			Method: http.MethodPost, Path: "/v1/music/stem-separation", Accept: transport.AcceptAny, RespType: "application/zip", RespBody: "zip",
			Query: "output_format=mp3_44100_128",
			Form:  map[string][]string{"file": {"file:s.mp3:mp3"}, "stem_variation_id": {"two_stems_v1"}},
			Check: testutil.CheckString("zip"),
		},
		{
			Name: "compositionPlan.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.CompositionPlan.Create(ctx, &CompositionPlanCreateRequest{Prompt: "pop", MusicLengthMs: types.Ptr(int64(5000))})
			},
			Method: http.MethodPost, Path: "/v1/music/plan",
			Body:     `{"prompt":"pop","music_length_ms":5000}`,
			RespBody: `{"chunks":[{"text":"a","duration_ms":1,"positive_styles":[]},{"song_id":"s","range":{"start_ms":0,"end_ms":1}}]}`,
			Check: func(t *testing.T, r any) {
				p := r.(*MusicCompositionPlan)
				if p.Chunks == nil || len(p.Chunks.Chunks) != 2 || p.Chunks.Chunks[0].Generation == nil || p.Chunks.Chunks[1].AudioRef == nil {
					t.Errorf("plan = %+v", p)
				}
			},
		},
		{
			Name: "finetunes.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Finetunes.List(ctx, &FinetunesListRequest{Cursor: "c", PageSize: types.Ptr(10)})
			},
			Method: http.MethodGet, Path: "/v1/music/finetunes", Query: "cursor=c&page_size=10",
			RespBody: `{"finetunes":[{"id":"f1","name":"n"}],"has_more":false}`,
			Check: func(t *testing.T, r any) {
				if p := r.(*FinetunePage); len(p.Finetunes) != 1 || p.Finetunes[0].ID != "f1" {
					t.Errorf("page = %+v", p)
				}
			},
		},
		{
			Name: "finetunes.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Finetunes.Create(ctx, &FinetuneCreateRequest{Name: "n", Files: []types.File{testutil.File("a.mp3", "x")}, Tags: []string{"t"}})
			},
			Method: http.MethodPost, Path: "/v1/music/finetunes",
			Form:     map[string][]string{"name": {"n"}, "files": {"file:a.mp3:x"}, "tags": {"t"}},
			RespBody: `{"id":"f1"}`,
		},
		{
			Name:   "finetunes.get",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Finetunes.Get(ctx, "f1") },
			Method: http.MethodGet, Path: "/v1/music/finetunes/f1", RespBody: `{"id":"f1"}`,
		},
		{
			Name:   "finetunes.delete",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Finetunes.Delete(ctx, "f1") },
			Method: http.MethodDelete, Path: "/v1/music/finetunes/f1", RespBody: `{"id":"f1"}`,
		},
		{
			Name: "finetunes.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Finetunes.Update(ctx, "f1", &FinetuneUpdateRequest{Name: types.Ptr("m"), Tags: []string{"a"}})
			},
			Method: http.MethodPatch, Path: "/v1/music/finetunes/f1",
			Body: `{"name":"m","tags":["a"]}`, RespBody: `{"id":"f1","name":"m"}`,
		},
	})
}

const musicMultipartBody = "--b\r\nContent-Type: application/json\r\n\r\n" +
	`{"composition_plan":{"positive_global_styles":["pop"],"negative_global_styles":[],"sections":[]},"song_metadata":{"title":"T","description":"d","genres":["pop"],"languages":["en"],"is_explicit":false}}` +
	"\r\n--b\r\nContent-Type: audio/mpeg\r\nContent-Disposition: attachment; filename=\"song.mp3\"\r\n\r\nMP3DATA\r\n--b--\r\n"

func TestDetailedBoundaryFallback(t *testing.T) {
	client, _ := newTestServer(t, testutil.Respond(http.StatusOK, "application/octet-stream", musicMultipartBody))
	resp, err := client.ComposeDetailed(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Close()
	audio, _ := io.ReadAll(resp)
	if !bytes.Equal(audio, []byte("MP3DATA")) || resp.Metadata.SongMetadata.Genres[0] != "pop" {
		t.Errorf("audio = %q, metadata = %+v", audio, resp.Metadata)
	}
}

func TestDetailedSongIDHeader(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("song-id", "s42")
		w.Header().Set(transport.HeaderContentType, "multipart/mixed; boundary=b")
		_, _ = io.WriteString(w, musicMultipartBody)
	})
	resp, err := client.ComposeDetailed(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Close()
	if resp.SongID != "s42" {
		t.Errorf("song id = %q", resp.SongID)
	}
}

func TestFinetunesListAll(t *testing.T) {
	client, ts := newTestServer(t, testutil.Sequence(
		testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"finetunes":[{"id":"a"}],"next_cursor":"c2","has_more":true}`),
		testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"finetunes":[{"id":"b"}],"has_more":false}`),
	))
	var ids []string
	for f, err := range client.Finetunes.ListAll(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, f.ID)
	}
	if strings.Join(ids, ",") != "a,b" {
		t.Errorf("ids = %v", ids)
	}
	if q := ts.Last(t).RawQuery; q != "cursor=c2" {
		t.Errorf("second query = %q", q)
	}
}

func newTestServer(t *testing.T, h http.HandlerFunc, opts ...option.RequestOption) (*Service, *testutil.Server) {
	t.Helper()
	ts, o := testutil.NewServer(t, h, opts...)
	return NewService(o...), ts
}
