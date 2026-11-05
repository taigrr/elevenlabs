package elevenlabs

import (
	"context"
	"net/http"
)

// AudioNativeService groups the /v1/audio-native endpoints.
type AudioNativeService struct{ cfg *config }

// AudioNativeCreateRequest is the body of AudioNativeService.Create.
type AudioNativeCreateRequest struct {
	// Name is the project name. Required.
	Name string `form:"name"`
	// Image is the player image URL (deprecated upstream).
	Image string `form:"image"`
	// Author is shown in the player and prepended to the article.
	Author string `form:"author"`
	// Title is shown in the player and prepended to the article.
	Title string `form:"title"`
	// Small selects the small player (deprecated upstream).
	Small *bool `form:"small"`
	// TextColor is the player text color.
	TextColor string `form:"text_color"`
	// BackgroundColor is the player background color.
	BackgroundColor string `form:"background_color"`
	// Sessionization is how many minutes to persist sessions (deprecated upstream).
	Sessionization *int `form:"sessionization"`
	// VoiceID voices the content.
	VoiceID string `form:"voice_id"`
	// ModelID is the TTS model used.
	ModelID string `form:"model_id"`
	// File is a txt or HTML file with the article content.
	File *File `form:"file"`
	// AutoConvert converts the project to audio automatically.
	AutoConvert *bool `form:"auto_convert"`
	// ApplyTextNormalization controls text normalization (auto, on, apply_english, off).
	ApplyTextNormalization TextNormalization `form:"apply_text_normalization"`
	// PronunciationDictionaryLocators are JSON-encoded PronunciationDictionaryLocator
	// values, one per dictionary (see PronunciationDictionaryLocator.String).
	PronunciationDictionaryLocators []string `form:"pronunciation_dictionary_locators"`
}

// AudioNativeCreateResponse is the response of AudioNativeService.Create.
type AudioNativeCreateResponse struct {
	ProjectID   string `json:"project_id"`
	Converting  bool   `json:"converting"`
	HTMLSnippet string `json:"html_snippet"`
}

// Create creates an Audio Native enabled project.
func (s *AudioNativeService) Create(ctx context.Context, req *AudioNativeCreateRequest, opts ...Option) (*AudioNativeCreateResponse, error) {
	return doValue[AudioNativeCreateResponse](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/audio-native", form: orEmpty(req), opts: opts})
}

// AudioNativeProjectStatus is the processing state of an Audio Native project.
type AudioNativeProjectStatus string

// AudioNativeProjectStatus values.
const (
	AudioNativeProjectStatusProcessing AudioNativeProjectStatus = "processing"
	AudioNativeProjectStatusReady      AudioNativeProjectStatus = "ready"
)

// AudioNativeProjectSettings are the player settings of an Audio Native project.
type AudioNativeProjectSettings struct {
	Title           string                    `json:"title"`
	Image           string                    `json:"image"`
	Author          string                    `json:"author"`
	Small           bool                      `json:"small"`
	TextColor       string                    `json:"text_color"`
	BackgroundColor string                    `json:"background_color"`
	Sessionization  int64                     `json:"sessionization"`
	AudioPath       *string                   `json:"audio_path,omitempty"`
	AudioURL        *string                   `json:"audio_url,omitempty"`
	Status          *AudioNativeProjectStatus `json:"status,omitempty"`
}

// AudioNativeSettingsResponse is the response of AudioNativeService.GetSettings.
type AudioNativeSettingsResponse struct {
	Enabled    bool                        `json:"enabled"`
	SnapshotID *string                     `json:"snapshot_id,omitempty"`
	Settings   *AudioNativeProjectSettings `json:"settings,omitempty"`
}

// GetSettings returns the player settings of an Audio Native project.
func (s *AudioNativeService) GetSettings(ctx context.Context, projectID string, opts ...Option) (*AudioNativeSettingsResponse, error) {
	return doValue[AudioNativeSettingsResponse](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/audio-native/%s/settings", projectID), opts: opts})
}

// AudioNativeUpdateRequest is the body of AudioNativeService.Update.
type AudioNativeUpdateRequest struct {
	// File is a txt or HTML file with the new article content.
	File *File `form:"file"`
	// AutoConvert converts the project to audio automatically.
	AutoConvert *bool `form:"auto_convert"`
	// AutoPublish publishes the new snapshot after conversion.
	AutoPublish *bool `form:"auto_publish"`
}

// AudioNativeEditContentResponse is the response of the content update endpoints.
type AudioNativeEditContentResponse struct {
	ProjectID   string `json:"project_id"`
	Converting  bool   `json:"converting"`
	Publishing  bool   `json:"publishing"`
	HTMLSnippet string `json:"html_snippet"`
}

// Update replaces the content of an Audio Native project.
func (s *AudioNativeService) Update(ctx context.Context, projectID string, req *AudioNativeUpdateRequest, opts ...Option) (*AudioNativeEditContentResponse, error) {
	return doValue[AudioNativeEditContentResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/audio-native/%s/content", projectID), form: orEmpty(req), opts: opts})
}

// AudioNativeUpdateFromURLRequest is the body of AudioNativeService.UpdateContentFromURL.
type AudioNativeUpdateFromURLRequest struct {
	// URL is the page to extract content from. Required.
	URL string `json:"url"`
	// Author overrides the player's default author.
	Author string `json:"author,omitempty"`
	// Title overrides the player's default title.
	Title string `json:"title,omitempty"`
}

// UpdateContentFromURL updates the Audio Native project associated with a
// page URL using the page's current content.
func (s *AudioNativeService) UpdateContentFromURL(ctx context.Context, req *AudioNativeUpdateFromURLRequest, opts ...Option) (*AudioNativeEditContentResponse, error) {
	return doValue[AudioNativeEditContentResponse](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/audio-native/content", body: orEmpty(req), opts: opts})
}
