package voices

import (
	"context"
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
			Name: "getAll",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.GetAll(ctx, &GetAllRequest{ShowLegacy: types.Ptr(true)})
			},
			Method:   http.MethodGet,
			Path:     "/v1/voices",
			Query:    "show_legacy=true",
			RespBody: `{"voices":[{"voice_id":"v1","name":"Rachel","category":"premade","settings":{"stability":0.5}}]}`,
			Check: func(t *testing.T, r any) {
				v := r.(*GetAllResponse)
				if len(v.Voices) != 1 || v.Voices[0].VoiceID != "v1" || *v.Voices[0].Name != "Rachel" ||
					*v.Voices[0].Category != CategoryPremade || *v.Voices[0].Settings.Stability != 0.5 {
					t.Errorf("voices = %+v", v)
				}
			},
		},
		{
			Name: "search",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Search(ctx, &SearchRequest{NextPageToken: "tok", Sort: "name", SortDirection: "asc", VoiceType: "personal", Category: "cloned", UseCases: []string{"a", "b"}, HighQuality: types.Ptr(true), MinNoticePeriodDays: types.Ptr(3)})
			},
			Method:   http.MethodGet,
			Path:     "/v2/voices",
			Query:    "category=cloned&high_quality=true&min_notice_period_days=3&next_page_token=tok&sort=name&sort_direction=asc&use_cases=a&use_cases=b&voice_type=personal",
			RespBody: `{"voices":[{"voice_id":"v2"}],"has_more":true,"total_count":2,"next_page_token":"n"}`,
			Check: func(t *testing.T, r any) {
				v := r.(*SearchResponse)
				if !v.HasMore || v.TotalCount != 2 || *v.NextPageToken != "n" || v.Voices[0].VoiceID != "v2" {
					t.Errorf("search = %+v", v)
				}
			},
		},
		{
			Name: "get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Get(ctx, "voice/1", &GetRequest{WithSettings: types.Ptr(true)})
			},
			Method:   http.MethodGet,
			Path:     "/v1/voices/voice%2F1",
			Query:    "with_settings=true",
			RespBody: `{"voice_id":"voice/1","fine_tuning":{"state":{"eleven_multilingual_v2":"fine_tuned"}},"samples":[{"sample_id":"s1","size_bytes":10}]}`,
			Check: func(t *testing.T, r any) {
				v := r.(*Voice)
				if v.FineTuning.State["eleven_multilingual_v2"] != FineTuningStateFineTuned || *v.Samples[0].SampleID != "s1" {
					t.Errorf("voice = %+v", v)
				}
			},
		},
		{
			Name:     "delete",
			Call:     func(ctx context.Context, c *Service) (any, error) { return c.Delete(ctx, "v1") },
			Method:   http.MethodDelete,
			Path:     "/v1/voices/v1",
			RespBody: `{"status":"ok"}`,
			Check:    checkStatusOK,
		},
		{
			Name: "update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Update(ctx, "v1", &UpdateRequest{
					Name:                  "New",
					Files:                 []types.File{types.NewFile(strings.NewReader("abc"), "s.mp3")},
					RemoveBackgroundNoise: types.Ptr(true),
					Description:           "desc",
					Labels:                map[string]string{"k": "v"},
					ModerateMetadata:      types.Ptr(false),
				})
			},
			Method: http.MethodPost,
			Path:   "/v1/voices/v1/edit",
			Form: map[string][]string{
				"name": {"New"}, "files": {"file:s.mp3:abc"}, "remove_background_noise": {"true"},
				"description": {"desc"}, "labels": {`{"k":"v"}`}, "moderate_metadata": {"false"},
			},
			RespBody: `{"status":"ok"}`,
			Check:    checkStatusOK,
		},
		{
			Name: "ivc.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.IVC.Create(ctx, &IVCCreateRequest{
					Name:        "Clone",
					Files:       []types.File{types.NewFile(strings.NewReader("1"), "a.mp3"), types.NewFile(strings.NewReader("2"), "b.mp3")},
					Description: "d",
				})
			},
			Method:   http.MethodPost,
			Path:     "/v1/voices/add",
			Form:     map[string][]string{"name": {"Clone"}, "files": {"file:a.mp3:1", "file:b.mp3:2"}, "description": {"d"}},
			RespBody: `{"voice_id":"new","requires_verification":true}`,
			Check: func(t *testing.T, r any) {
				if v := r.(*AddVoiceIVCResponse); v.VoiceID != "new" || !v.RequiresVerification {
					t.Errorf("ivc = %+v", v)
				}
			},
		},
		{
			Name:     "settings.getDefault",
			Call:     func(ctx context.Context, c *Service) (any, error) { return c.Settings.GetDefault(ctx) },
			Method:   http.MethodGet,
			Path:     "/v1/voices/settings/default",
			RespBody: `{"stability":0.5,"similarity_boost":0.75,"use_speaker_boost":true,"speed":1}`,
			Check: func(t *testing.T, r any) {
				if s := r.(*types.VoiceSettings); *s.SimilarityBoost != 0.75 || !*s.UseSpeakerBoost || *s.Speed != 1 {
					t.Errorf("settings = %+v", s)
				}
			},
		},
		{
			Name:     "settings.get",
			Call:     func(ctx context.Context, c *Service) (any, error) { return c.Settings.Get(ctx, "v1") },
			Method:   http.MethodGet,
			Path:     "/v1/voices/v1/settings",
			RespBody: `{"stability":0.1}`,
			Check: func(t *testing.T, r any) {
				if s := r.(*types.VoiceSettings); *s.Stability != 0.1 || s.Style != nil {
					t.Errorf("settings = %+v", s)
				}
			},
		},
		{
			Name: "settings.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Settings.Update(ctx, "v1", &types.VoiceSettings{Stability: types.Ptr(0.0), Style: types.Ptr(0.3)})
			},
			Method:   http.MethodPost,
			Path:     "/v1/voices/v1/settings/edit",
			Body:     `{"stability":0,"style":0.3}`,
			RespBody: `{"status":"ok"}`,
			Check:    checkStatusOK,
		},
		{
			Name: "samples.audio.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.ReadAll(c.Samples.Audio.Get(ctx, "v1", "s1"))
			},
			Method:   http.MethodGet,
			Path:     "/v1/voices/v1/samples/s1/audio",
			Accept:   transport.AcceptAny,
			RespType: "audio/mpeg",
			RespBody: "ID3audio",
			Check:    testutil.CheckString("ID3audio"),
		},
		{
			Name: "replicateToIsolatedEnvironment",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.ReplicateToIsolatedEnvironment(ctx, "v/1", &ReplicateVoiceRequest{TargetWorkspaceID: "w", PreserveVoiceID: types.Ptr(false)})
			},
			Method: http.MethodPost, Path: "/v1/voices/v%2F1/replicate-to-isolated-environment",
			Body:     `{"target_workspace_id":"w","preserve_voice_id":false}`,
			RespBody: `{"voice_id":"v2"}`,
			Check: func(t *testing.T, r any) {
				if r.(*VoiceIDResponse).VoiceID != "v2" {
					t.Errorf("resp = %+v", r)
				}
			},
		},
		{
			Name: "share",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Share(ctx, "pu", "v", &ShareVoiceRequest{NewName: "John", Bookmarked: types.Ptr(true)})
			},
			Method: http.MethodPost, Path: "/v1/voices/add/pu/v",
			Body: `{"new_name":"John","bookmarked":true}`, RespBody: `{"voice_id":"v"}`,
		},
		{
			Name: "findSimilarVoices",
			Call: func(ctx context.Context, c *Service) (any, error) {
				f := testutil.File("s.mp3", "mp3")
				return c.FindSimilarVoices(ctx, &FindSimilarVoicesRequest{AudioFile: &f, SimilarityThreshold: types.Ptr(1.5), TopK: types.Ptr(3)})
			},
			Method: http.MethodPost, Path: "/v1/similar-voices",
			Form:     map[string][]string{"audio_file": {"file:s.mp3:mp3"}, "similarity_threshold": {"1.5"}, "top_k": {"3"}},
			RespBody: `{"voices":[{"public_owner_id":"p","voice_id":"v","category":"professional","usage_character_count_1y":7}],"has_more":false}`,
			Check: func(t *testing.T, r any) {
				v := r.(*LibraryVoicesResponse).Voices[0]
				if v.PublicOwnerID != "p" || v.Category != CategoryProfessional || v.UsageCharacterCount1Y != 7 {
					t.Errorf("voice = %+v", v)
				}
			},
		},
		{
			Name: "getShared",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.GetShared(ctx, &GetSharedRequest{PageSize: types.Ptr(5), Category: SharedVoiceCategoryFamous, UseCases: []string{"a", "b"}, Featured: types.Ptr(true), Sort: SharedVoiceSortTrending, Page: types.Ptr(2)})
			},
			Method: http.MethodGet, Path: "/v1/shared-voices",
			Query:    "category=famous&featured=true&page=2&page_size=5&sort=trending&use_cases=a&use_cases=b",
			RespBody: `{"voices":[],"has_more":true,"total_count":9}`,
		},
		{
			Name: "accents.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Accents.Get(ctx, &AccentsGetRequest{Language: "en", ModelID: "m"})
			},
			Method: http.MethodGet, Path: "/v1/voices/accents", Query: "language=en&model_id=m",
			RespBody: `{"accents":[{"accent":"american","language":"en","code":"en-american","name":"American"}]}`,
			Check: func(t *testing.T, r any) {
				if a := r.(*AccentsResponse).Accents; len(a) != 1 || a[0].Code != "en-american" {
					t.Errorf("accents = %+v", a)
				}
			},
		},
		{
			Name:   "collections.list",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Collections.List(ctx) },
			Method: http.MethodGet, Path: "/v1/voices/collections",
			RespBody: `{"collections":[{"collection_id":"c","title":"t","icon":"i","permission_on_resource":"admin"}]}`,
			Check: func(t *testing.T, r any) {
				if c := r.(*ListCollectionsResponse).Collections; c[0].PermissionOnResource != types.ResourcePermissionAdmin {
					t.Errorf("collections = %+v", c)
				}
			},
		},
		{
			Name: "collections.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Collections.Create(ctx, &CollectionRequest{Title: "t", Icon: "i"})
			},
			Method: http.MethodPost, Path: "/v1/voices/collections", Body: `{"title":"t","icon":"i"}`, RespBody: `{}`,
		},
		{
			Name:   "collections.get",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Collections.Get(ctx, "c") },
			Method: http.MethodGet, Path: "/v1/voices/collections/c", RespBody: `{"collection_id":"c"}`,
		},
		{
			Name:   "collections.delete",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Collections.Delete(ctx, "c") },
			Method: http.MethodDelete, Path: "/v1/voices/collections/c", RespBody: `{"collection_id":"c"}`,
		},
		{
			Name: "collections.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Collections.Update(ctx, "c", &CollectionRequest{Title: "n"})
			},
			Method: http.MethodPatch, Path: "/v1/voices/collections/c", Body: `{"title":"n"}`, RespBody: `{}`,
		},
		{
			Name:   "collections.addVoice",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Collections.AddVoice(ctx, "c", "v") },
			Method: http.MethodPost, Path: "/v1/voices/collections/c/voice/v", RespBody: `{"collection_id":"c","voice_id":"v"}`,
		},
		{
			Name: "collections.removeVoice",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Collections.RemoveVoice(ctx, "c", "v")
			},
			Method: http.MethodDelete, Path: "/v1/voices/collections/c/voice/v", RespBody: `{"collection_id":"c","voice_id":"v"}`,
		},
		{
			Name: "pvc.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.PVC.Create(ctx, &PVCCreateRequest{Name: "John", Language: "en", Labels: map[string]string{"accent": "us"}})
			},
			Method: http.MethodPost, Path: "/v1/voices/pvc",
			Body: `{"name":"John","language":"en","labels":{"accent":"us"}}`, RespBody: `{"voice_id":"v"}`,
		},
		{
			Name: "pvc.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.PVC.Update(ctx, "v", &PVCUpdateRequest{Description: "d"})
			},
			Method: http.MethodPost, Path: "/v1/voices/pvc/v", Body: `{"description":"d"}`, RespBody: `{"voice_id":"v"}`,
		},
		{
			Name: "pvc.train",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.PVC.Train(ctx, "v", &PVCTrainRequest{ModelID: "m"})
			},
			Method: http.MethodPost, Path: "/v1/voices/pvc/v/train", Body: `{"model_id":"m"}`, RespBody: `{"status":"ok"}`,
		},
		{
			Name: "pvc.samples.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.PVC.Samples.Create(ctx, "v", &PVCAddSamplesRequest{Files: []types.File{testutil.File("a.wav", "A"), testutil.File("b.wav", "B")}, RemoveBackgroundNoise: types.Ptr(true)})
			},
			Method: http.MethodPost, Path: "/v1/voices/pvc/v/samples",
			Form:     map[string][]string{"files": {"file:a.wav:A", "file:b.wav:B"}, "remove_background_noise": {"true"}},
			RespBody: `[{"sample_id":"s1"},{"sample_id":"s2"}]`,
			Check: func(t *testing.T, r any) {
				if s := r.([]Sample); len(s) != 2 || *s[1].SampleID != "s2" {
					t.Errorf("samples = %+v", s)
				}
			},
		},
		{
			Name: "pvc.samples.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.PVC.Samples.Update(ctx, "v", "s", &PVCUpdateSampleRequest{SelectedSpeakerIDs: []string{"sp"}, TrimStartTime: types.Ptr[int64](10), FileName: "f"})
			},
			Method: http.MethodPost, Path: "/v1/voices/pvc/v/samples/s",
			Body: `{"selected_speaker_ids":["sp"],"trim_start_time":10,"file_name":"f"}`, RespBody: `{"voice_id":"v"}`,
		},
		{
			Name:   "pvc.samples.delete",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.PVC.Samples.Delete(ctx, "v", "s") },
			Method: http.MethodDelete, Path: "/v1/voices/pvc/v/samples/s", RespBody: `{"status":"ok"}`,
		},
		{
			Name: "pvc.samples.audio.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.PVC.Samples.Audio.Get(ctx, "v", "s", &PVCSampleAudioRequest{RemoveBackgroundNoise: types.Ptr(true)})
			},
			Method: http.MethodGet, Path: "/v1/voices/pvc/v/samples/s/audio", Query: "remove_background_noise=true",
			RespBody: `{"audio_base_64":"AA==","voice_id":"v","sample_id":"s","media_type":"audio/mpeg","duration_secs":1.5}`,
			Check: func(t *testing.T, r any) {
				if p := r.(*SamplePreview); p.AudioBase64 != "AA==" || *p.DurationSecs != 1.5 {
					t.Errorf("preview = %+v", p)
				}
			},
		},
		{
			Name: "pvc.samples.speakers.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.PVC.Samples.Speakers.Get(ctx, "v", "s")
			},
			Method: http.MethodGet, Path: "/v1/voices/pvc/v/samples/s/speakers",
			RespBody: `{"voice_id":"v","sample_id":"s","status":"completed","speakers":{"a":{"speaker_id":"a","duration_secs":2,"utterances":[{"start":0,"end":1}]}}}`,
			Check: func(t *testing.T, r any) {
				s := r.(*SpeakerSeparation)
				if s.Status != SpeakerSeparationStatusCompleted || s.Speakers["a"].Utterances[0].End != 1 {
					t.Errorf("separation = %+v", s)
				}
			},
		},
		{
			Name: "pvc.samples.speakers.separate",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.PVC.Samples.Speakers.Separate(ctx, "v", "s")
			},
			Method: http.MethodPost, Path: "/v1/voices/pvc/v/samples/s/separate-speakers", RespBody: `{"status":"ok"}`,
		},
		{
			Name: "pvc.samples.speakers.audio.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.PVC.Samples.Speakers.Audio.Get(ctx, "v", "s", "sp")
			},
			Method: http.MethodGet, Path: "/v1/voices/pvc/v/samples/s/speakers/sp/audio",
			RespBody: `{"audio_base_64":"AA==","media_type":"audio/mpeg","duration_secs":2}`,
		},
		{
			Name: "pvc.samples.waveform.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.PVC.Samples.Waveform.Get(ctx, "v", "s")
			},
			Method: http.MethodGet, Path: "/v1/voices/pvc/v/samples/s/waveform",
			RespBody: `{"sample_id":"s","visual_waveform":[0.1,0.2]}`,
			Check: func(t *testing.T, r any) {
				if w := r.(*SampleWaveform); len(w.VisualWaveform) != 2 {
					t.Errorf("waveform = %+v", w)
				}
			},
		},
		{
			Name: "pvc.verification.request",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.PVC.Verification.Request(ctx, "v", &PVCManualVerificationRequest{Files: []types.File{testutil.File("id.pdf", "P")}, ExtraText: "x"})
			},
			Method: http.MethodPost, Path: "/v1/voices/pvc/v/verification",
			Form: map[string][]string{"files": {"file:id.pdf:P"}, "extra_text": {"x"}}, RespBody: `{"status":"ok"}`,
		},
		{
			Name: "pvc.verification.captcha.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return voidCall(c.PVC.Verification.Captcha.Get(ctx, "v"))
			},
			Method: http.MethodGet, Path: "/v1/voices/pvc/v/captcha",
		},
		{
			Name: "pvc.verification.captcha.verify",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.PVC.Verification.Captcha.Verify(ctx, "v", &PVCVerifyCaptchaRequest{Recording: testutil.File("r.wav", "R")})
			},
			Method: http.MethodPost, Path: "/v1/voices/pvc/v/captcha",
			Form: map[string][]string{"recording": {"file:r.wav:R"}}, RespBody: `{"status":"ok"}`,
		},
	})
}

func TestSearchAll(t *testing.T) {
	client, ts := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("next_page_token") == "" {
			_, _ = io.WriteString(w, `{"voices":[{"voice_id":"a"}],"has_more":true,"total_count":2,"next_page_token":"p2"}`)
			return
		}
		_, _ = io.WriteString(w, `{"voices":[{"voice_id":"b"}],"has_more":false,"total_count":2}`)
	})
	var ids []string
	for v, err := range client.SearchAll(context.Background(), &SearchRequest{Search: "x"}) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, v.VoiceID)
	}
	if strings.Join(ids, ",") != "a,b" {
		t.Errorf("ids = %v", ids)
	}
	if q := ts.Last(t).RawQuery; q != "next_page_token=p2&search=x" {
		t.Errorf("second page query = %q", q)
	}
}

func TestGetSharedAll(t *testing.T) {
	client, ts := newTestServer(t, testutil.Sequence(
		testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"voices":[{"voice_id":"a"}],"has_more":true}`),
		testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"voices":[{"voice_id":"b"}],"has_more":false}`),
	))
	var ids []string
	for v, err := range client.GetSharedAll(context.Background(), &GetSharedRequest{Search: "q"}) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, v.VoiceID)
	}
	if strings.Join(ids, ",") != "a,b" {
		t.Errorf("ids = %v", ids)
	}
	reqs := ts.All()
	if reqs[0].RawQuery != "page=0&search=q" || reqs[1].RawQuery != "page=1&search=q" {
		t.Errorf("queries = %q, %q", reqs[0].RawQuery, reqs[1].RawQuery)
	}
}

func checkStatusOK(t *testing.T, r any) {
	t.Helper()
	if s := r.(*types.StatusResponse); s.Status != "ok" {
		t.Errorf("status = %q", s.Status)
	}
}

func voidCall(err error) (any, error) { return nil, err }

func newTestServer(t *testing.T, h http.HandlerFunc, opts ...option.RequestOption) (*Service, *testutil.Server) {
	t.Helper()
	ts, o := testutil.NewServer(t, h, opts...)
	return NewService(o...), ts
}
