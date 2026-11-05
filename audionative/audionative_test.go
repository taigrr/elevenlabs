package audionative

import (
	"context"
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/types"
)

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, []testutil.Endpoint[*Service]{
		{
			Name: "audioNative.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				f := testutil.File("a.html", "<p>x</p>")
				return c.Create(ctx, &CreateRequest{
					Name: "n", VoiceID: "v", File: &f, AutoConvert: types.Ptr(true), ApplyTextNormalization: types.TextNormalizationApplyEnglish,
					PronunciationDictionaryLocators: []string{types.PronunciationDictionaryLocator{PronunciationDictionaryID: "pd", VersionID: "v1"}.String(), `{"pronunciation_dictionary_id":"pd2"}`},
				})
			},
			Method: http.MethodPost, Path: "/v1/audio-native",
			Form: map[string][]string{
				"name": {"n"}, "voice_id": {"v"}, "file": {"file:a.html:<p>x</p>"}, "auto_convert": {"true"},
				"apply_text_normalization":          {"apply_english"},
				"pronunciation_dictionary_locators": {`{"pronunciation_dictionary_id":"pd","version_id":"v1"}`, `{"pronunciation_dictionary_id":"pd2"}`},
			},
			RespBody: `{"project_id":"p","converting":true,"html_snippet":"<div/>"}`,
			Check: func(t *testing.T, r any) {
				if p := r.(*CreateResponse); p.ProjectID != "p" || !p.Converting || p.HTMLSnippet != "<div/>" {
					t.Errorf("resp = %+v", p)
				}
			},
		},
		{
			Name:   "audioNative.getSettings",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.GetSettings(ctx, "p") },
			Method: http.MethodGet, Path: "/v1/audio-native/p/settings",
			RespBody: `{"enabled":true,"snapshot_id":"s","settings":{"title":"t","sessionization":5,"status":"ready"}}`,
			Check: func(t *testing.T, r any) {
				s := r.(*SettingsResponse)
				if !s.Enabled || s.Settings.Sessionization != 5 || *s.Settings.Status != ProjectStatusReady {
					t.Errorf("settings = %+v", s)
				}
			},
		},
		{
			Name: "audioNative.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				f := testutil.File("a.txt", "T")
				return c.Update(ctx, "p", &UpdateRequest{File: &f, AutoPublish: types.Ptr(true)})
			},
			Method: http.MethodPost, Path: "/v1/audio-native/p/content",
			Form:     map[string][]string{"file": {"file:a.txt:T"}, "auto_publish": {"true"}},
			RespBody: `{"project_id":"p","publishing":true}`,
		},
		{
			Name: "audioNative.updateContentFromUrl",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.UpdateContentFromURL(ctx, &UpdateFromURLRequest{URL: "https://x", Title: "t"})
			},
			Method: http.MethodPost, Path: "/v1/audio-native/content", Body: `{"url":"https://x","title":"t"}`, RespBody: `{}`,
		},
	})
}
