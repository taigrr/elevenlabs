package elevenlabs

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/taigrr/elevenlabs/v2/types"
)

func endpointCases() []endpointCase {
	cases := append(append(coreEndpointCases(), audioEndpointCases()...), libraryEndpointCases()...)
	cases = append(cases, platformEndpointCases()...)
	cases = append(cases, convaiEndpointCases()...)
	return append(cases, productionEndpointCases()...)
}

func testFile(name, data string) File {
	return File{Reader: strings.NewReader(data), Name: name}
}

func drainStream[T any](s *Stream[T], err error) (any, error) {
	if err != nil {
		return nil, err
	}
	defer s.Close()
	var out []T
	for v, err := range s.All() {
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

const musicMultipartBody = "--b\r\nContent-Type: application/json\r\n\r\n" +
	`{"composition_plan":{"positive_global_styles":["pop"],"negative_global_styles":[],"sections":[]},"song_metadata":{"title":"T","description":"d","genres":["pop"],"languages":["en"],"is_explicit":false}}` +
	"\r\n--b\r\nContent-Type: audio/mpeg\r\nContent-Disposition: attachment; filename=\"song.mp3\"\r\n\r\nMP3DATA\r\n--b--\r\n"

func audioEndpointCases() []endpointCase {
	return []endpointCase{
		{
			name: "forcedAlignment.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ForcedAlignment.Create(ctx, &ForcedAlignmentRequest{File: testFile("a.mp3", "mp3"), Text: "hi"})
			},
			method: http.MethodPost, path: "/v1/forced-alignment",
			form:     map[string][]string{"file": {"file:a.mp3:mp3"}, "text": {"hi"}},
			respBody: `{"characters":[{"text":"h","start":0,"end":0.1}],"words":[{"text":"hi","start":0,"end":0.2,"loss":0.3}],"loss":0.3}`,
			check: func(t *testing.T, r any) {
				a := r.(*ForcedAlignment)
				if len(a.Characters) != 1 || a.Words[0].Loss != 0.3 || a.Loss != 0.3 {
					t.Errorf("alignment = %+v", a)
				}
			},
		},
		{
			name:   "models.list",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Models.List(ctx) },
			method: http.MethodGet, path: "/v1/models",
			respBody: `[{"model_id":"eleven_v3","can_do_text_to_speech":true,"languages":[{"language_id":"en","name":"English"}],"model_rates":{"character_cost_multiplier":1}}]`,
			check: func(t *testing.T, r any) {
				m := r.([]Model)
				if len(m) != 1 || m[0].ModelID != "eleven_v3" || !*m[0].CanDoTextToSpeech || m[0].Languages[0].LanguageID != "en" || m[0].ModelRates.CharacterCostMultiplier != 1 {
					t.Errorf("models = %+v", m)
				}
			},
		},
		{
			name: "tokens.singleUse.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Tokens.SingleUse.Create(ctx, SingleUseTokenTypeRealtimeScribe)
			},
			method: http.MethodPost, path: "/v1/single-use-token/realtime_scribe", respBody: `{"token":"tk"}`,
			check: func(t *testing.T, r any) {
				if r.(*SingleUseToken).Token != "tk" {
					t.Errorf("token = %+v", r)
				}
			},
		},
		{
			name: "usage.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Usage.Get(ctx, &UsageGetRequest{StartUnix: 1, EndUnix: 2, IncludeWorkspaceMetrics: Ptr(true), BreakdownType: BreakdownTypeVoice, AggregationInterval: UsageAggregationIntervalDay, AggregationBucketSize: Ptr(60), Metric: MetricTypeCredits})
			},
			method: http.MethodGet, path: "/v1/usage/character-stats",
			query:    "aggregation_bucket_size=60&aggregation_interval=day&breakdown_type=voice&end_unix=2&include_workspace_metrics=true&metric=credits&start_unix=1",
			respBody: `{"time":[1],"usage":{"All":[5]}}`,
			check: func(t *testing.T, r any) {
				if u := r.(*UsageCharacters); len(u.Time) != 1 || u.Usage["All"][0] != 5 {
					t.Errorf("usage = %+v", u)
				}
			},
		},
		{
			name: "music.compose",
			call: func(ctx context.Context, c *Client) (any, error) {
				return readAll(c.Music.Compose(ctx, &MusicComposeRequest{OutputFormat: types.OutputFormatMP3_44100_128, MusicGenerationParams: MusicGenerationParams{Prompt: "jazz", MusicLengthMs: Ptr(int64(10000)), ForceInstrumental: Ptr(true)}, SignWithC2PA: Ptr(false)}))
			},
			method: http.MethodPost, path: "/v1/music", accept: acceptAny, respType: "audio/mpeg", respBody: "song",
			query: "output_format=mp3_44100_128",
			body:  `{"prompt":"jazz","music_length_ms":10000,"force_instrumental":true,"sign_with_c2pa":false}`,
			check: checkString("song"),
		},
		{
			name: "music.stream",
			call: func(ctx context.Context, c *Client) (any, error) {
				return readAll(c.Music.Stream(ctx, &MusicStreamRequest{MusicGenerationParams: MusicGenerationParams{Prompt: "jazz"}}))
			},
			method: http.MethodPost, path: "/v1/music/stream", accept: acceptAny, respType: "audio/mpeg", respBody: "song",
			body: `{"prompt":"jazz"}`, check: checkString("song"),
		},
		{
			name: "music.composeDetailed",
			call: func(ctx context.Context, c *Client) (any, error) {
				resp, err := c.Music.ComposeDetailed(ctx, &MusicComposeDetailedRequest{MusicGenerationParams: MusicGenerationParams{Prompt: "pop"}, WithTimestamps: Ptr(true)})
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
			method: http.MethodPost, path: "/v1/music/detailed", accept: acceptAny,
			respType: "multipart/mixed; boundary=b", respBody: musicMultipartBody,
			body: `{"prompt":"pop","with_timestamps":true}`, check: checkString("MP3DATA"),
		},
		{
			name: "music.composeDetailedStream",
			call: func(ctx context.Context, c *Client) (any, error) {
				return drainStream(c.Music.ComposeDetailedStream(ctx, &MusicComposeDetailedStreamRequest{MusicGenerationParams: MusicGenerationParams{Prompt: "pop"}}))
			},
			method: http.MethodPost, path: "/v1/music/detailed/stream",
			body: `{"prompt":"pop"}`, respBody: "{\"a\":1}\n{\"b\":2}\n",
			check: func(t *testing.T, r any) {
				if c := r.([]json.RawMessage); len(c) != 2 {
					t.Errorf("chunks = %s", c)
				}
			},
		},
		{
			name: "music.videoToMusic",
			call: func(ctx context.Context, c *Client) (any, error) {
				return readAll(c.Music.VideoToMusic(ctx, &VideoToMusicRequest{Videos: []File{testFile("a.mp4", "v1"), testFile("b.mp4", "v2")}, Description: "epic", Tags: []string{"x", "y"}, AddSFX: Ptr(true)}))
			},
			method: http.MethodPost, path: "/v1/music/video-to-music", accept: acceptAny, respType: "audio/mpeg", respBody: "song",
			form:  map[string][]string{"videos": {"file:a.mp4:v1", "file:b.mp4:v2"}, "description": {"epic"}, "tags": {"x", "y"}, "add_sfx": {"true"}},
			check: checkString("song"),
		},
		{
			name: "music.upload",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Music.Upload(ctx, &MusicUploadRequest{File: testFile("s.mp3", "mp3"), WithTimestamps: Ptr(true)})
			},
			method: http.MethodPost, path: "/v1/music/upload",
			form:     map[string][]string{"file": {"file:s.mp3:mp3"}, "with_timestamps": {"true"}},
			respBody: `{"song_id":"s1","words_timestamps":[{"word":"la","start_ms":0,"end_ms":5}]}`,
			check: func(t *testing.T, r any) {
				if u := r.(*MusicUpload); u.SongID != "s1" || u.WordsTimestamps[0].EndMs != 5 {
					t.Errorf("upload = %+v", u)
				}
			},
		},
		{
			name: "music.separateStems",
			call: func(ctx context.Context, c *Client) (any, error) {
				return readAll(c.Music.SeparateStems(ctx, &SeparateStemsRequest{OutputFormat: types.OutputFormatMP3_44100_128, File: testFile("s.mp3", "mp3"), StemVariationID: StemVariation("two_stems_v1")}))
			},
			method: http.MethodPost, path: "/v1/music/stem-separation", accept: acceptAny, respType: "application/zip", respBody: "zip",
			query: "output_format=mp3_44100_128",
			form:  map[string][]string{"file": {"file:s.mp3:mp3"}, "stem_variation_id": {"two_stems_v1"}},
			check: checkString("zip"),
		},
		{
			name: "music.compositionPlan.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Music.CompositionPlan.Create(ctx, &CompositionPlanCreateRequest{Prompt: "pop", MusicLengthMs: Ptr(int64(5000))})
			},
			method: http.MethodPost, path: "/v1/music/plan",
			body:     `{"prompt":"pop","music_length_ms":5000}`,
			respBody: `{"chunks":[{"text":"a","duration_ms":1,"positive_styles":[]},{"song_id":"s","range":{"start_ms":0,"end_ms":1}}]}`,
			check: func(t *testing.T, r any) {
				p := r.(*MusicCompositionPlan)
				if p.Chunks == nil || len(p.Chunks.Chunks) != 2 || p.Chunks.Chunks[0].Generation == nil || p.Chunks.Chunks[1].AudioRef == nil {
					t.Errorf("plan = %+v", p)
				}
			},
		},
		{
			name: "music.finetunes.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Music.Finetunes.List(ctx, &MusicFinetunesListRequest{Cursor: "c", PageSize: Ptr(10)})
			},
			method: http.MethodGet, path: "/v1/music/finetunes", query: "cursor=c&page_size=10",
			respBody: `{"finetunes":[{"id":"f1","name":"n"}],"has_more":false}`,
			check: func(t *testing.T, r any) {
				if p := r.(*MusicFinetunePage); len(p.Finetunes) != 1 || p.Finetunes[0].ID != "f1" {
					t.Errorf("page = %+v", p)
				}
			},
		},
		{
			name: "music.finetunes.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Music.Finetunes.Create(ctx, &MusicFinetuneCreateRequest{Name: "n", Files: []File{testFile("a.mp3", "x")}, Tags: []string{"t"}})
			},
			method: http.MethodPost, path: "/v1/music/finetunes",
			form:     map[string][]string{"name": {"n"}, "files": {"file:a.mp3:x"}, "tags": {"t"}},
			respBody: `{"id":"f1"}`,
		},
		{
			name:   "music.finetunes.get",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Music.Finetunes.Get(ctx, "f1") },
			method: http.MethodGet, path: "/v1/music/finetunes/f1", respBody: `{"id":"f1"}`,
		},
		{
			name:   "music.finetunes.delete",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Music.Finetunes.Delete(ctx, "f1") },
			method: http.MethodDelete, path: "/v1/music/finetunes/f1", respBody: `{"id":"f1"}`,
		},
		{
			name: "music.finetunes.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Music.Finetunes.Update(ctx, "f1", &MusicFinetuneUpdateRequest{Name: Ptr("m"), Tags: []string{"a"}})
			},
			method: http.MethodPatch, path: "/v1/music/finetunes/f1",
			body: `{"name":"m","tags":["a"]}`, respBody: `{"id":"f1","name":"m"}`,
		},
	}
}

func TestMusicDetailedBoundaryFallback(t *testing.T) {
	client, _ := newTestServer(t, respond(http.StatusOK, "application/octet-stream", musicMultipartBody))
	resp, err := client.Music.ComposeDetailed(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Close()
	audio, _ := io.ReadAll(resp)
	if !bytes.Equal(audio, []byte("MP3DATA")) || resp.Metadata.SongMetadata.Genres[0] != "pop" {
		t.Errorf("audio = %q, metadata = %+v", audio, resp.Metadata)
	}
}

func TestMusicDetailedSongIDHeader(t *testing.T) {
	client, _ := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("song-id", "s42")
		w.Header().Set(headerContentType, "multipart/mixed; boundary=b")
		_, _ = io.WriteString(w, musicMultipartBody)
	})
	resp, err := client.Music.ComposeDetailed(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Close()
	if resp.SongID != "s42" {
		t.Errorf("song id = %q", resp.SongID)
	}
}

func TestMusicFinetunesListAll(t *testing.T) {
	client, ts := newTestServer(t, sequence(
		respond(http.StatusOK, contentTypeJSON, `{"finetunes":[{"id":"a"}],"next_cursor":"c2","has_more":true}`),
		respond(http.StatusOK, contentTypeJSON, `{"finetunes":[{"id":"b"}],"has_more":false}`),
	))
	var ids []string
	for f, err := range client.Music.Finetunes.ListAll(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, f.ID)
	}
	if strings.Join(ids, ",") != "a,b" {
		t.Errorf("ids = %v", ids)
	}
	if q := ts.last(t).RawQuery; q != "cursor=c2" {
		t.Errorf("second query = %q", q)
	}
}
