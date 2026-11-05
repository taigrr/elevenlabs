package studio

import (
	"context"
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/types"
)

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, []testutil.Endpoint[*Service]{
		{
			Name: "studio.createPodcast", Call: func(ctx context.Context, c *Service) (any, error) {
				return c.CreatePodcast(ctx, &PodcastCreateRequest{ModelID: "m", Mode: ConversationPodcast("h", "g"), Source: PodcastSources{URLSource("https://x")}, DurationScale: "short"})
			},
			Method: http.MethodPost, Path: "/v1/studio/podcasts",
			Body:     `{"model_id":"m","mode":{"type":"conversation","conversation":{"host_voice_id":"h","guest_voice_id":"g"}},"source":{"type":"url","url":"https://x"},"duration_scale":"short"}`,
			RespBody: `{"project":{"project_id":"p","name":"n","state":"creating","access_level":"admin"}}`,
			Check: func(t *testing.T, r any) {
				if p := r.(*ProjectResponse).Project; p.ProjectID != "p" || p.State != ProjectStateCreating {
					t.Errorf("project = %+v", p)
				}
			},
		},
		{
			Name:   "studio.projects.list",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Projects.List(ctx) },
			Method: http.MethodGet, Path: "/v1/studio/projects", RespBody: `{"projects":[{"project_id":"p"}]}`,
		},
		{
			Name: "studio.projects.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				f := testutil.File("b.epub", "epub")
				return c.Projects.Create(ctx, &ProjectCreateRequest{Name: "n", FromDocument: &f, Genres: []string{"a", "b"}, AutoConvert: types.Ptr(true),
					PronunciationDictionaryLocators: []string{types.PronunciationDictionaryLocator{PronunciationDictionaryID: "d"}.String()}})
			},
			Method: http.MethodPost, Path: "/v1/studio/projects",
			Form: map[string][]string{"name": {"n"}, "from_document": {"file:b.epub:epub"}, "genres": {"a", "b"}, "auto_convert": {"true"},
				"pronunciation_dictionary_locators": {`{"pronunciation_dictionary_id":"d"}`}},
			RespBody: `{"project":{"project_id":"p"}}`,
		},
		{
			Name:   "studio.projects.get",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Projects.Get(ctx, "p", "sh") },
			Method: http.MethodGet, Path: "/v1/studio/projects/p", Query: "share_id=sh",
			RespBody: `{"project_id":"p","quality_preset":"high","chapters":[{"chapter_id":"c","name":"one","state":"default","can_be_downloaded":true}],"voices":[],"assets":[{"type":"image"}]}`,
			Check: func(t *testing.T, r any) {
				p := r.(*ProjectExtended)
				if p.ProjectID != "p" || p.QualityPreset != QualityPresetHigh || p.Chapters[0].ChapterID != "c" || len(p.Assets) != 1 {
					t.Errorf("project = %+v", p)
				}
			},
		},
		{
			Name: "studio.projects.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Projects.Update(ctx, "p", &ProjectUpdateRequest{Name: "n", DefaultTitleVoiceID: "t", DefaultParagraphVoiceID: "q", Author: "a"})
			},
			Method: http.MethodPost, Path: "/v1/studio/projects/p",
			Body: `{"name":"n","default_title_voice_id":"t","default_paragraph_voice_id":"q","author":"a"}`, RespBody: `{"project":{"project_id":"p"}}`,
		},
		{
			Name:   "studio.projects.delete",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Projects.Delete(ctx, "p") },
			Method: http.MethodDelete, Path: "/v1/studio/projects/p", RespBody: `{"status":"ok"}`,
		},
		{
			Name:   "studio.projects.convert",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Projects.Convert(ctx, "p") },
			Method: http.MethodPost, Path: "/v1/studio/projects/p/convert", RespBody: `{"status":"ok"}`,
		},
		{
			Name:   "studio.projects.getMutedTracks",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Projects.GetMutedTracks(ctx, "p") },
			Method: http.MethodGet, Path: "/v1/studio/projects/p/muted-tracks", RespBody: `{"chapter_ids":["c"]}`,
		},
		{
			Name:   "studio.projects.chapters.list",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Projects.Chapters.List(ctx, "p") },
			Method: http.MethodGet, Path: "/v1/studio/projects/p/chapters", RespBody: `{"chapters":[]}`,
		},
		{
			Name: "studio.projects.chapters.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Projects.Chapters.Create(ctx, "p", &ChapterCreateRequest{Name: "c", FromURL: "https://x"})
			},
			Method: http.MethodPost, Path: "/v1/studio/projects/p/chapters",
			Body: `{"name":"c","from_url":"https://x"}`, RespBody: `{"chapter":{"chapter_id":"c","content":{"blocks":[]}}}`,
		},
		{
			Name: "studio.projects.chapters.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Projects.Chapters.Get(ctx, "p", "c")
			},
			Method: http.MethodGet, Path: "/v1/studio/projects/p/chapters/c",
			RespBody: `{"chapter_id":"c","name":"n","state":"converting","content":{"blocks":[{"block_id":"b","nodes":[{"type":"tts_node","text":"hi","voice_id":"v","project_voice_ref_id":"r"}]}]}}`,
			Check: func(t *testing.T, r any) {
				ch := r.(*ChapterWithContent)
				if ch.ChapterID != "c" || ch.State != ChapterStateConverting || ch.Content.Blocks[0].Nodes[0].Text != "hi" {
					t.Errorf("chapter = %+v", ch)
				}
			},
		},
		{
			Name: "studio.projects.chapters.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Projects.Chapters.Update(ctx, "p", "c", &ChapterUpdateRequest{Content: &ChapterContent{Blocks: []ChapterContentBlock{{SubType: "p", Nodes: []ChapterContentNode{TTSNode("hi", "v")}}}}})
			},
			Method: http.MethodPost, Path: "/v1/studio/projects/p/chapters/c",
			Body:     `{"content":{"blocks":[{"sub_type":"p","nodes":[{"type":"tts_node","text":"hi","voice_id":"v"}]}]}}`,
			RespBody: `{"chapter":{"chapter_id":"c"}}`,
		},
		{
			Name: "studio.projects.chapters.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Projects.Chapters.Delete(ctx, "p", "c")
			},
			Method: http.MethodDelete, Path: "/v1/studio/projects/p/chapters/c", RespBody: `{"status":"ok"}`,
		},
		{
			Name: "studio.projects.chapters.convert",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Projects.Chapters.Convert(ctx, "p", "c")
			},
			Method: http.MethodPost, Path: "/v1/studio/projects/p/chapters/c/convert", RespBody: `{"status":"ok"}`,
		},
		{
			Name: "studio.projects.chapters.snapshots.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Projects.Chapters.Snapshots.List(ctx, "p", "c")
			},
			Method: http.MethodGet, Path: "/v1/studio/projects/p/chapters/c/snapshots",
			RespBody: `{"snapshots":[{"chapter_snapshot_id":"s","project_id":"p","chapter_id":"c","created_at_unix":1,"name":"n"}]}`,
		},
		{
			Name: "studio.projects.chapters.snapshots.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Projects.Chapters.Snapshots.Get(ctx, "p", "c", "s")
			},
			Method: http.MethodGet, Path: "/v1/studio/projects/p/chapters/c/snapshots/s",
			RespBody: `{"chapter_snapshot_id":"s","character_alignments":[{"characters":["a"],"character_start_times_seconds":[0],"character_end_times_seconds":[1]}]}`,
			Check: func(t *testing.T, r any) {
				if s := r.(*ChapterSnapshotExtended); s.ChapterSnapshotID != "s" || s.CharacterAlignments[0].Characters[0] != "a" {
					t.Errorf("snapshot = %+v", s)
				}
			},
		},
		{
			Name: "studio.projects.chapters.snapshots.stream",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.ReadAll(c.Projects.Chapters.Snapshots.Stream(ctx, "p", "c", "s", &SnapshotStreamRequest{ConvertToMpeg: types.Ptr(true)}))
			},
			Method: http.MethodPost, Path: "/v1/studio/projects/p/chapters/c/snapshots/s/stream", Accept: transport.AcceptAny,
			Body: `{"convert_to_mpeg":true}`, RespType: "audio/mpeg", RespBody: "aud", Check: testutil.CheckString("aud"),
		},
		{
			Name: "studio.projects.content.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Projects.Content.Update(ctx, "p", &ProjectContentUpdateRequest{FromURL: "https://x", AutoConvert: types.Ptr(false)})
			},
			Method: http.MethodPost, Path: "/v1/studio/projects/p/content",
			Form: map[string][]string{"from_url": {"https://x"}, "auto_convert": {"false"}}, RespBody: `{"project":{"project_id":"p"}}`,
		},
		{
			Name: "studio.projects.pronunciationDictionaries.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Projects.PronunciationDictionaries.Create(ctx, "p", &ProjectPronunciationDictionariesRequest{PronunciationDictionaryLocators: []types.PronunciationDictionaryLocator{{PronunciationDictionaryID: "d", VersionID: "v"}}})
			},
			Method: http.MethodPost, Path: "/v1/studio/projects/p/pronunciation-dictionaries",
			Body: `{"pronunciation_dictionary_locators":[{"pronunciation_dictionary_id":"d","version_id":"v"}]}`, RespBody: `{"status":"ok"}`,
		},
		{
			Name:   "studio.projects.snapshots.list",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Projects.Snapshots.List(ctx, "p") },
			Method: http.MethodGet, Path: "/v1/studio/projects/p/snapshots", RespBody: `{"snapshots":[]}`,
		},
		{
			Name: "studio.projects.snapshots.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Projects.Snapshots.Get(ctx, "p", "s")
			},
			Method: http.MethodGet, Path: "/v1/studio/projects/p/snapshots/s",
			RespBody: `{"project_snapshot_id":"s","audio_duration_secs":12.5,"character_alignments":[]}`,
			Check: func(t *testing.T, r any) {
				if s := r.(*ProjectSnapshotExtended); s.ProjectSnapshotID != "s" || s.AudioDurationSecs != 12.5 {
					t.Errorf("snapshot = %+v", s)
				}
			},
		},
		{
			Name: "studio.projects.snapshots.stream",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.ReadAll(c.Projects.Snapshots.Stream(ctx, "p", "s", nil))
			},
			Method: http.MethodPost, Path: "/v1/studio/projects/p/snapshots/s/stream", Accept: transport.AcceptAny,
			Body: `{}`, RespType: "audio/mpeg", RespBody: "aud", Check: testutil.CheckString("aud"),
		},
		{
			Name: "studio.projects.snapshots.streamArchive",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.ReadAll(c.Projects.Snapshots.StreamArchive(ctx, "p", "s"))
			},
			Method: http.MethodPost, Path: "/v1/studio/projects/p/snapshots/s/archive", Accept: transport.AcceptAny,
			RespType: "application/zip", RespBody: "PK", Check: testutil.CheckString("PK"),
		},
	})
}

func TestPodcastSourcesMultiple(t *testing.T) {
	b, err := PodcastSources{TextSource("a"), URLSource("u")}.MarshalJSON()
	if err != nil || string(b) != `[{"type":"text","text":"a"},{"type":"url","url":"u"}]` {
		t.Fatalf("sources = %s, %v", b, err)
	}
	b, _ = BulletinPodcast("h").MarshalJSON()
	if string(b) != `{"bulletin":{"host_voice_id":"h"},"type":"bulletin"}` {
		t.Errorf("mode = %s", b)
	}
}
