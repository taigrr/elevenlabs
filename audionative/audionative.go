package audionative

import (
	"context"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

// Service groups the /v1/audio-native endpoints.
type Service struct{ cfg *requestconfig.Config }

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	return &Service{cfg: requestconfig.Build(opts...)}
}

// CreateRequest is the body of Service.Create.
type CreateRequest struct {
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
	File *types.File `form:"file"`
	// AutoConvert converts the project to audio automatically.
	AutoConvert *bool `form:"auto_convert"`
	// ApplyTextNormalization controls text normalization (auto, on, apply_english, off).
	ApplyTextNormalization types.TextNormalization `form:"apply_text_normalization"`
	// PronunciationDictionaryLocators are JSON-encoded PronunciationDictionaryLocator
	// values, one per dictionary (see PronunciationDictionaryLocator.String).
	PronunciationDictionaryLocators []string `form:"pronunciation_dictionary_locators"`
}

// CreateResponse is the response of Service.Create.
type CreateResponse struct {
	ProjectID   string `json:"project_id"`
	Converting  bool   `json:"converting"`
	HTMLSnippet string `json:"html_snippet"`
}

// Create creates an Audio Native enabled project.
func (s *Service) Create(ctx context.Context, req *CreateRequest, opts ...option.RequestOption) (*CreateResponse, error) {
	return transport.DoValue[CreateResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/audio-native", Form: transport.OrEmpty(req), Opts: opts})
}

// ProjectStatus is the processing state of an Audio Native project.
type ProjectStatus string

// ProjectStatus values.
const (
	ProjectStatusProcessing ProjectStatus = "processing"
	ProjectStatusReady      ProjectStatus = "ready"
)

// ProjectSettings are the player settings of an Audio Native project.
type ProjectSettings struct {
	Title           string         `json:"title"`
	Image           string         `json:"image"`
	Author          string         `json:"author"`
	Small           bool           `json:"small"`
	TextColor       string         `json:"text_color"`
	BackgroundColor string         `json:"background_color"`
	Sessionization  int64          `json:"sessionization"`
	AudioPath       *string        `json:"audio_path,omitempty"`
	AudioURL        *string        `json:"audio_url,omitempty"`
	Status          *ProjectStatus `json:"status,omitempty"`
}

// SettingsResponse is the response of Service.GetSettings.
type SettingsResponse struct {
	Enabled    bool             `json:"enabled"`
	SnapshotID *string          `json:"snapshot_id,omitempty"`
	Settings   *ProjectSettings `json:"settings,omitempty"`
}

// GetSettings returns the player settings of an Audio Native project.
func (s *Service) GetSettings(ctx context.Context, projectID string, opts ...option.RequestOption) (*SettingsResponse, error) {
	return transport.DoValue[SettingsResponse](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/audio-native/%s/settings", projectID), Opts: opts})
}

// UpdateRequest is the body of Service.Update.
type UpdateRequest struct {
	// File is a txt or HTML file with the new article content.
	File *types.File `form:"file"`
	// AutoConvert converts the project to audio automatically.
	AutoConvert *bool `form:"auto_convert"`
	// AutoPublish publishes the new snapshot after conversion.
	AutoPublish *bool `form:"auto_publish"`
}

// EditContentResponse is the response of the content update endpoints.
type EditContentResponse struct {
	ProjectID   string `json:"project_id"`
	Converting  bool   `json:"converting"`
	Publishing  bool   `json:"publishing"`
	HTMLSnippet string `json:"html_snippet"`
}

// Update replaces the content of an Audio Native project.
func (s *Service) Update(ctx context.Context, projectID string, req *UpdateRequest, opts ...option.RequestOption) (*EditContentResponse, error) {
	return transport.DoValue[EditContentResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/audio-native/%s/content", projectID), Form: transport.OrEmpty(req), Opts: opts})
}

// UpdateFromURLRequest is the body of Service.UpdateContentFromURL.
type UpdateFromURLRequest struct {
	// URL is the page to extract content from. Required.
	URL string `json:"url"`
	// Author overrides the player's default author.
	Author string `json:"author,omitempty"`
	// Title overrides the player's default title.
	Title string `json:"title,omitempty"`
}

// UpdateContentFromURL updates the Audio Native project associated with a
// page URL using the page's current content.
func (s *Service) UpdateContentFromURL(ctx context.Context, req *UpdateFromURLRequest, opts ...option.RequestOption) (*EditContentResponse, error) {
	return transport.DoValue[EditContentResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/audio-native/content", Body: transport.OrEmpty(req), Opts: opts})
}
