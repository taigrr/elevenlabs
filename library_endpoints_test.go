package elevenlabs

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/taigrr/elevenlabs/v2/types"
	"github.com/taigrr/elevenlabs/v2/voices"
)

func voidCall(err error) (any, error) { return nil, err }

func libraryEndpointCases() []endpointCase {
	return []endpointCase{
		{
			name: "voices.replicateToIsolatedEnvironment",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Voices.ReplicateToIsolatedEnvironment(ctx, "v/1", &ReplicateVoiceRequest{TargetWorkspaceID: "w", PreserveVoiceID: Ptr(false)})
			},
			method: http.MethodPost, path: "/v1/voices/v%2F1/replicate-to-isolated-environment",
			body:     `{"target_workspace_id":"w","preserve_voice_id":false}`,
			respBody: `{"voice_id":"v2"}`,
			check: func(t *testing.T, r any) {
				if r.(*VoiceIDResponse).VoiceID != "v2" {
					t.Errorf("resp = %+v", r)
				}
			},
		},
		{
			name: "voices.share",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Voices.Share(ctx, "pu", "v", &ShareVoiceRequest{NewName: "John", Bookmarked: Ptr(true)})
			},
			method: http.MethodPost, path: "/v1/voices/add/pu/v",
			body: `{"new_name":"John","bookmarked":true}`, respBody: `{"voice_id":"v"}`,
		},
		{
			name: "voices.findSimilarVoices",
			call: func(ctx context.Context, c *Client) (any, error) {
				f := testFile("s.mp3", "mp3")
				return c.Voices.FindSimilarVoices(ctx, &FindSimilarVoicesRequest{AudioFile: &f, SimilarityThreshold: Ptr(1.5), TopK: Ptr(3)})
			},
			method: http.MethodPost, path: "/v1/similar-voices",
			form:     map[string][]string{"audio_file": {"file:s.mp3:mp3"}, "similarity_threshold": {"1.5"}, "top_k": {"3"}},
			respBody: `{"voices":[{"public_owner_id":"p","voice_id":"v","category":"professional","usage_character_count_1y":7}],"has_more":false}`,
			check: func(t *testing.T, r any) {
				v := r.(*GetLibraryVoicesResponse).Voices[0]
				if v.PublicOwnerID != "p" || v.Category != voices.CategoryProfessional || v.UsageCharacterCount1Y != 7 {
					t.Errorf("voice = %+v", v)
				}
			},
		},
		{
			name: "voices.getShared",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Voices.GetShared(ctx, &VoicesGetSharedRequest{PageSize: Ptr(5), Category: SharedVoiceCategoryFamous, UseCases: []string{"a", "b"}, Featured: Ptr(true), Sort: SharedVoiceSortTrending, Page: Ptr(2)})
			},
			method: http.MethodGet, path: "/v1/shared-voices",
			query:    "category=famous&featured=true&page=2&page_size=5&sort=trending&use_cases=a&use_cases=b",
			respBody: `{"voices":[],"has_more":true,"total_count":9}`,
		},
		{
			name: "voices.accents.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Voices.Accents.Get(ctx, &AccentsGetRequest{Language: "en", ModelID: "m"})
			},
			method: http.MethodGet, path: "/v1/voices/accents", query: "language=en&model_id=m",
			respBody: `{"accents":[{"accent":"american","language":"en","code":"en-american","name":"American"}]}`,
			check: func(t *testing.T, r any) {
				if a := r.(*GetVoiceAccentsResponse).Accents; len(a) != 1 || a[0].Code != "en-american" {
					t.Errorf("accents = %+v", a)
				}
			},
		},
		{
			name:   "voices.collections.list",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Voices.Collections.List(ctx) },
			method: http.MethodGet, path: "/v1/voices/collections",
			respBody: `{"collections":[{"collection_id":"c","title":"t","icon":"i","permission_on_resource":"admin"}]}`,
			check: func(t *testing.T, r any) {
				if c := r.(*GetVoiceCollectionsResponse).Collections; c[0].PermissionOnResource != ResourcePermissionAdmin {
					t.Errorf("collections = %+v", c)
				}
			},
		},
		{
			name: "voices.collections.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Voices.Collections.Create(ctx, &VoiceCollectionRequest{Title: "t", Icon: "i"})
			},
			method: http.MethodPost, path: "/v1/voices/collections", body: `{"title":"t","icon":"i"}`, respBody: `{}`,
		},
		{
			name:   "voices.collections.get",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Voices.Collections.Get(ctx, "c") },
			method: http.MethodGet, path: "/v1/voices/collections/c", respBody: `{"collection_id":"c"}`,
		},
		{
			name:   "voices.collections.delete",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Voices.Collections.Delete(ctx, "c") },
			method: http.MethodDelete, path: "/v1/voices/collections/c", respBody: `{"collection_id":"c"}`,
		},
		{
			name: "voices.collections.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Voices.Collections.Update(ctx, "c", &VoiceCollectionRequest{Title: "n"})
			},
			method: http.MethodPatch, path: "/v1/voices/collections/c", body: `{"title":"n"}`, respBody: `{}`,
		},
		{
			name:   "voices.collections.addVoice",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Voices.Collections.AddVoice(ctx, "c", "v") },
			method: http.MethodPost, path: "/v1/voices/collections/c/voice/v", respBody: `{"collection_id":"c","voice_id":"v"}`,
		},
		{
			name: "voices.collections.removeVoice",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Voices.Collections.RemoveVoice(ctx, "c", "v")
			},
			method: http.MethodDelete, path: "/v1/voices/collections/c/voice/v", respBody: `{"collection_id":"c","voice_id":"v"}`,
		},
		{
			name: "voices.pvc.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Voices.PVC.Create(ctx, &PVCCreateRequest{Name: "John", Language: "en", Labels: map[string]string{"accent": "us"}})
			},
			method: http.MethodPost, path: "/v1/voices/pvc",
			body: `{"name":"John","language":"en","labels":{"accent":"us"}}`, respBody: `{"voice_id":"v"}`,
		},
		{
			name: "voices.pvc.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Voices.PVC.Update(ctx, "v", &PVCUpdateRequest{Description: "d"})
			},
			method: http.MethodPost, path: "/v1/voices/pvc/v", body: `{"description":"d"}`, respBody: `{"voice_id":"v"}`,
		},
		{
			name: "voices.pvc.train",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Voices.PVC.Train(ctx, "v", &PVCTrainRequest{ModelID: "m"})
			},
			method: http.MethodPost, path: "/v1/voices/pvc/v/train", body: `{"model_id":"m"}`, respBody: `{"status":"ok"}`,
		},
		{
			name: "voices.pvc.samples.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Voices.PVC.Samples.Create(ctx, "v", &PVCAddSamplesRequest{Files: []File{testFile("a.wav", "A"), testFile("b.wav", "B")}, RemoveBackgroundNoise: Ptr(true)})
			},
			method: http.MethodPost, path: "/v1/voices/pvc/v/samples",
			form:     map[string][]string{"files": {"file:a.wav:A", "file:b.wav:B"}, "remove_background_noise": {"true"}},
			respBody: `[{"sample_id":"s1"},{"sample_id":"s2"}]`,
			check: func(t *testing.T, r any) {
				if s := r.([]voices.Sample); len(s) != 2 || *s[1].SampleID != "s2" {
					t.Errorf("samples = %+v", s)
				}
			},
		},
		{
			name: "voices.pvc.samples.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Voices.PVC.Samples.Update(ctx, "v", "s", &PVCUpdateSampleRequest{SelectedSpeakerIDs: []string{"sp"}, TrimStartTime: Ptr[int64](10), FileName: "f"})
			},
			method: http.MethodPost, path: "/v1/voices/pvc/v/samples/s",
			body: `{"selected_speaker_ids":["sp"],"trim_start_time":10,"file_name":"f"}`, respBody: `{"voice_id":"v"}`,
		},
		{
			name:   "voices.pvc.samples.delete",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Voices.PVC.Samples.Delete(ctx, "v", "s") },
			method: http.MethodDelete, path: "/v1/voices/pvc/v/samples/s", respBody: `{"status":"ok"}`,
		},
		{
			name: "voices.pvc.samples.audio.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Voices.PVC.Samples.Audio.Get(ctx, "v", "s", &PVCSampleAudioRequest{RemoveBackgroundNoise: Ptr(true)})
			},
			method: http.MethodGet, path: "/v1/voices/pvc/v/samples/s/audio", query: "remove_background_noise=true",
			respBody: `{"audio_base_64":"AA==","voice_id":"v","sample_id":"s","media_type":"audio/mpeg","duration_secs":1.5}`,
			check: func(t *testing.T, r any) {
				if p := r.(*VoiceSamplePreview); p.AudioBase64 != "AA==" || *p.DurationSecs != 1.5 {
					t.Errorf("preview = %+v", p)
				}
			},
		},
		{
			name: "voices.pvc.samples.speakers.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Voices.PVC.Samples.Speakers.Get(ctx, "v", "s")
			},
			method: http.MethodGet, path: "/v1/voices/pvc/v/samples/s/speakers",
			respBody: `{"voice_id":"v","sample_id":"s","status":"completed","speakers":{"a":{"speaker_id":"a","duration_secs":2,"utterances":[{"start":0,"end":1}]}}}`,
			check: func(t *testing.T, r any) {
				s := r.(*voices.SpeakerSeparation)
				if s.Status != voices.SpeakerSeparationStatusCompleted || s.Speakers["a"].Utterances[0].End != 1 {
					t.Errorf("separation = %+v", s)
				}
			},
		},
		{
			name: "voices.pvc.samples.speakers.separate",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Voices.PVC.Samples.Speakers.Separate(ctx, "v", "s")
			},
			method: http.MethodPost, path: "/v1/voices/pvc/v/samples/s/separate-speakers", respBody: `{"status":"ok"}`,
		},
		{
			name: "voices.pvc.samples.speakers.audio.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Voices.PVC.Samples.Speakers.Audio.Get(ctx, "v", "s", "sp")
			},
			method: http.MethodGet, path: "/v1/voices/pvc/v/samples/s/speakers/sp/audio",
			respBody: `{"audio_base_64":"AA==","media_type":"audio/mpeg","duration_secs":2}`,
		},
		{
			name: "voices.pvc.samples.waveform.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Voices.PVC.Samples.Waveform.Get(ctx, "v", "s")
			},
			method: http.MethodGet, path: "/v1/voices/pvc/v/samples/s/waveform",
			respBody: `{"sample_id":"s","visual_waveform":[0.1,0.2]}`,
			check: func(t *testing.T, r any) {
				if w := r.(*VoiceSampleWaveform); len(w.VisualWaveform) != 2 {
					t.Errorf("waveform = %+v", w)
				}
			},
		},
		{
			name: "voices.pvc.verification.request",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Voices.PVC.Verification.Request(ctx, "v", &PVCManualVerificationRequest{Files: []File{testFile("id.pdf", "P")}, ExtraText: "x"})
			},
			method: http.MethodPost, path: "/v1/voices/pvc/v/verification",
			form: map[string][]string{"files": {"file:id.pdf:P"}, "extra_text": {"x"}}, respBody: `{"status":"ok"}`,
		},
		{
			name: "voices.pvc.verification.captcha.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return voidCall(c.Voices.PVC.Verification.Captcha.Get(ctx, "v"))
			},
			method: http.MethodGet, path: "/v1/voices/pvc/v/captcha",
		},
		{
			name: "voices.pvc.verification.captcha.verify",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Voices.PVC.Verification.Captcha.Verify(ctx, "v", &PVCVerifyCaptchaRequest{Recording: testFile("r.wav", "R")})
			},
			method: http.MethodPost, path: "/v1/voices/pvc/v/captcha",
			form: map[string][]string{"recording": {"file:r.wav:R"}}, respBody: `{"status":"ok"}`,
		},
		{
			name: "pronunciationDictionaries.createFromFile",
			call: func(ctx context.Context, c *Client) (any, error) {
				f := testFile("d.pls", "PLS")
				return c.PronunciationDictionaries.CreateFromFile(ctx, &PronunciationDictionaryCreateFromFileRequest{Name: "n", File: &f, Description: "d", WorkspaceAccess: ResourcePermissionEditor})
			},
			method: http.MethodPost, path: "/v1/pronunciation-dictionaries/add-from-file",
			form:     map[string][]string{"name": {"n"}, "file": {"file:d.pls:PLS"}, "description": {"d"}, "workspace_access": {"editor"}},
			respBody: `{"id":"pd","name":"n","version_id":"v1","version_rules_num":3,"permission_on_resource":"admin"}`,
			check: func(t *testing.T, r any) {
				if p := r.(*AddPronunciationDictionaryResponse); p.ID != "pd" || p.VersionRulesNum != 3 || *p.PermissionOnResource != ResourcePermissionAdmin {
					t.Errorf("resp = %+v", p)
				}
			},
		},
		{
			name: "pronunciationDictionaries.createFromRules",
			call: func(ctx context.Context, c *Client) (any, error) {
				alias := AliasRule("Thailand", "tie-land")
				alias.CaseSensitive = Ptr(true)
				return c.PronunciationDictionaries.CreateFromRules(ctx, &PronunciationDictionaryCreateFromRulesRequest{
					Rules: []PronunciationRule{alias, PhonemeRule("tomato", "təˈmeɪtoʊ", "ipa")},
					Name:  "My Dictionary",
				})
			},
			method: http.MethodPost, path: "/v1/pronunciation-dictionaries/add-from-rules",
			body:     `{"rules":[{"type":"alias","string_to_replace":"Thailand","case_sensitive":true,"alias":"tie-land"},{"type":"phoneme","string_to_replace":"tomato","phoneme":"təˈmeɪtoʊ","alphabet":"ipa"}],"name":"My Dictionary"}`,
			respBody: `{"id":"pd"}`,
		},
		{
			name:   "pronunciationDictionaries.get",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.PronunciationDictionaries.Get(ctx, "pd") },
			method: http.MethodGet, path: "/v1/pronunciation-dictionaries/pd",
			respBody: `{"id":"pd","latest_version_id":"v","name":"n","rules":[{"type":"alias","string_to_replace":"a","alias":"b"},{"type":"phoneme","string_to_replace":"c","phoneme":"d","alphabet":"ipa"}]}`,
			check: func(t *testing.T, r any) {
				d := r.(*PronunciationDictionaryWithRules)
				if d.ID != "pd" || d.LatestVersionID != "v" || len(d.Rules) != 2 || d.Rules[1].Type != PronunciationRuleTypePhoneme || d.Rules[0].Alias != "b" {
					t.Errorf("dictionary = %+v", d)
				}
			},
		},
		{
			name: "pronunciationDictionaries.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.PronunciationDictionaries.Update(ctx, "pd", &PronunciationDictionaryUpdateRequest{Archived: Ptr(true)})
			},
			method: http.MethodPatch, path: "/v1/pronunciation-dictionaries/pd", body: `{"archived":true}`, respBody: `{"id":"pd"}`,
		},
		{
			name: "pronunciationDictionaries.download",
			call: func(ctx context.Context, c *Client) (any, error) {
				return readAll(c.PronunciationDictionaries.Download(ctx, "pd", "v"))
			},
			method: http.MethodGet, path: "/v1/pronunciation-dictionaries/pd/v/download", accept: acceptAny,
			respType: "application/pls+xml", respBody: "<lexicon/>",
			check: checkString("<lexicon/>"),
		},
		{
			name: "pronunciationDictionaries.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.PronunciationDictionaries.List(ctx, &PronunciationDictionariesListRequest{PageSize: Ptr(2), Sort: PronunciationDictionarySortName, IncludeArchived: Ptr(false)})
			},
			method: http.MethodGet, path: "/v1/pronunciation-dictionaries",
			query:    "include_archived=false&page_size=2&sort=name",
			respBody: `{"pronunciation_dictionaries":[{"id":"a"}],"has_more":false}`,
		},
		{
			name: "pronunciationDictionaries.rules.set",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.PronunciationDictionaries.Rules.Set(ctx, "pd", []PronunciationRule{AliasRule("a", "b")})
			},
			method: http.MethodPost, path: "/v1/pronunciation-dictionaries/pd/set-rules",
			body:     `{"rules":[{"type":"alias","string_to_replace":"a","alias":"b"}]}`,
			respBody: `{"id":"pd","version_id":"v2","version_rules_num":1}`,
			check: func(t *testing.T, r any) {
				if p := r.(*PronunciationDictionaryRulesResponse); p.VersionID != "v2" || p.VersionRulesNum != 1 {
					t.Errorf("resp = %+v", p)
				}
			},
		},
		{
			name: "pronunciationDictionaries.rules.add",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.PronunciationDictionaries.Rules.Add(ctx, "pd", nil)
			},
			method: http.MethodPost, path: "/v1/pronunciation-dictionaries/pd/add-rules", body: `{"rules":[]}`, respBody: `{}`,
		},
		{
			name: "pronunciationDictionaries.rules.remove",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.PronunciationDictionaries.Rules.Remove(ctx, "pd", []string{"a"})
			},
			method: http.MethodPost, path: "/v1/pronunciation-dictionaries/pd/remove-rules", body: `{"rule_strings":["a"]}`, respBody: `{}`,
		},
		{
			name: "audioNative.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				f := testFile("a.html", "<p>x</p>")
				return c.AudioNative.Create(ctx, &AudioNativeCreateRequest{
					Name: "n", VoiceID: "v", File: &f, AutoConvert: Ptr(true), ApplyTextNormalization: types.TextNormalizationApplyEnglish,
					PronunciationDictionaryLocators: []string{types.PronunciationDictionaryLocator{PronunciationDictionaryID: "pd", VersionID: "v1"}.String(), `{"pronunciation_dictionary_id":"pd2"}`},
				})
			},
			method: http.MethodPost, path: "/v1/audio-native",
			form: map[string][]string{
				"name": {"n"}, "voice_id": {"v"}, "file": {"file:a.html:<p>x</p>"}, "auto_convert": {"true"},
				"apply_text_normalization":          {"apply_english"},
				"pronunciation_dictionary_locators": {`{"pronunciation_dictionary_id":"pd","version_id":"v1"}`, `{"pronunciation_dictionary_id":"pd2"}`},
			},
			respBody: `{"project_id":"p","converting":true,"html_snippet":"<div/>"}`,
			check: func(t *testing.T, r any) {
				if p := r.(*AudioNativeCreateResponse); p.ProjectID != "p" || !p.Converting || p.HTMLSnippet != "<div/>" {
					t.Errorf("resp = %+v", p)
				}
			},
		},
		{
			name:   "audioNative.getSettings",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.AudioNative.GetSettings(ctx, "p") },
			method: http.MethodGet, path: "/v1/audio-native/p/settings",
			respBody: `{"enabled":true,"snapshot_id":"s","settings":{"title":"t","sessionization":5,"status":"ready"}}`,
			check: func(t *testing.T, r any) {
				s := r.(*AudioNativeSettingsResponse)
				if !s.Enabled || s.Settings.Sessionization != 5 || *s.Settings.Status != AudioNativeProjectStatusReady {
					t.Errorf("settings = %+v", s)
				}
			},
		},
		{
			name: "audioNative.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				f := testFile("a.txt", "T")
				return c.AudioNative.Update(ctx, "p", &AudioNativeUpdateRequest{File: &f, AutoPublish: Ptr(true)})
			},
			method: http.MethodPost, path: "/v1/audio-native/p/content",
			form:     map[string][]string{"file": {"file:a.txt:T"}, "auto_publish": {"true"}},
			respBody: `{"project_id":"p","publishing":true}`,
		},
		{
			name: "audioNative.updateContentFromUrl",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.AudioNative.UpdateContentFromURL(ctx, &AudioNativeUpdateFromURLRequest{URL: "https://x", Title: "t"})
			},
			method: http.MethodPost, path: "/v1/audio-native/content", body: `{"url":"https://x","title":"t"}`, respBody: `{}`,
		},
		{
			name: "assets.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Assets.List(ctx, &AssetsListRequest{PageSize: Ptr(3), Search: "s"})
			},
			method: http.MethodGet, path: "/v1/assets", query: "page_size=3&search=s",
			respBody: `{"assets":[{"asset_id":"a","name":"n","mime_type":"audio/mpeg","created_at_unix":1}],"next_cursor":null,"has_more":false}`,
		},
		{
			name: "assets.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.Assets.Create(ctx, &AssetCreateRequest{Asset: testFile("a.mp3", "M"), Name: "n"})
			},
			method: http.MethodPost, path: "/v1/assets",
			form: map[string][]string{"asset": {"file:a.mp3:M"}, "name": {"n"}}, respBody: `{"asset_id":"a","content_url":"https://c"}`,
			check: func(t *testing.T, r any) {
				if a := r.(*Asset); a.AssetID != "a" || *a.ContentURL != "https://c" {
					t.Errorf("asset = %+v", a)
				}
			},
		},
		{
			name:   "assets.get",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.Assets.Get(ctx, "a") },
			method: http.MethodGet, path: "/v1/assets/a", respBody: `{"asset_id":"a"}`,
		},
		{
			name:   "assets.delete",
			call:   func(ctx context.Context, c *Client) (any, error) { return voidCall(c.Assets.Delete(ctx, "a")) },
			method: http.MethodDelete, path: "/v1/assets/a",
		},
	}
}

func TestVoicesGetSharedAll(t *testing.T) {
	client, ts := newTestServer(t, sequence(
		respond(http.StatusOK, contentTypeJSON, `{"voices":[{"voice_id":"a"}],"has_more":true}`),
		respond(http.StatusOK, contentTypeJSON, `{"voices":[{"voice_id":"b"}],"has_more":false}`),
	))
	var ids []string
	for v, err := range client.Voices.GetSharedAll(context.Background(), &VoicesGetSharedRequest{Search: "q"}) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, v.VoiceID)
	}
	if strings.Join(ids, ",") != "a,b" {
		t.Errorf("ids = %v", ids)
	}
	reqs := ts.all()
	if reqs[0].RawQuery != "page=0&search=q" || reqs[1].RawQuery != "page=1&search=q" {
		t.Errorf("queries = %q, %q", reqs[0].RawQuery, reqs[1].RawQuery)
	}
}

func TestPronunciationDictionariesListAll(t *testing.T) {
	client, ts := newTestServer(t, sequence(
		respond(http.StatusOK, contentTypeJSON, `{"pronunciation_dictionaries":[{"id":"a"}],"next_cursor":"c2","has_more":true}`),
		respond(http.StatusOK, contentTypeJSON, `{"pronunciation_dictionaries":[{"id":"b"}],"has_more":false}`),
	))
	var ids []string
	for d, err := range client.PronunciationDictionaries.ListAll(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, d.ID)
	}
	if strings.Join(ids, ",") != "a,b" || ts.last(t).RawQuery != "cursor=c2" {
		t.Errorf("ids = %v, last query = %q", ids, ts.last(t).RawQuery)
	}
}

func TestAssetsListAll(t *testing.T) {
	client, ts := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("cursor") == "" {
			_, _ = io.WriteString(w, `{"assets":[{"asset_id":"a"}],"next_cursor":"c2"}`)
			return
		}
		_, _ = io.WriteString(w, `{"assets":[{"asset_id":"b"}],"next_cursor":null}`)
	})
	var ids []string
	for a, err := range client.Assets.ListAll(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, a.AssetID)
	}
	if strings.Join(ids, ",") != "a,b" || len(ts.all()) != 2 {
		t.Errorf("ids = %v", ids)
	}
}

func TestAssetsListAllYieldsError(t *testing.T) {
	client, _ := newTestServer(t, respond(http.StatusForbidden, contentTypeJSON, `{}`))
	for _, err := range client.Assets.ListAll(context.Background(), nil) {
		if !errors.Is(err, ErrForbidden) {
			t.Fatalf("error = %v", err)
		}
		return
	}
	t.Fatal("iterator yielded nothing")
}
