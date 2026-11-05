package studio

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
	"github.com/taigrr/elevenlabs/v2/voices"
)

// Service groups the /v1/studio endpoints.
type Service struct {
	cfg *requestconfig.Config

	// Projects manages Studio projects.
	Projects *ProjectsService
}

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	cfg := requestconfig.Build(opts...)
	return &Service{cfg: cfg, Projects: &ProjectsService{
		cfg: cfg,
		Chapters: &ChaptersService{
			cfg:       cfg,
			Snapshots: &ChapterSnapshotsService{cfg: cfg},
		},
		Content:                   &ContentService{cfg: cfg},
		PronunciationDictionaries: &PronunciationDictionariesService{cfg: cfg},
		Snapshots:                 &ProjectSnapshotsService{cfg: cfg},
	}}
}

// QualityPreset is the output quality of a Studio project.
type QualityPreset string

// QualityPreset values.
const (
	QualityPresetStandard      QualityPreset = "standard"
	QualityPresetHigh          QualityPreset = "high"
	QualityPresetUltra         QualityPreset = "ultra"
	QualityPresetUltraLossless QualityPreset = "ultra_lossless"
)

// ProjectState is the state of a Studio project.
type ProjectState string

// ProjectState values.
const (
	ProjectStateCreating   ProjectState = "creating"
	ProjectStateDefault    ProjectState = "default"
	ProjectStateConverting ProjectState = "converting"
	ProjectStateInQueue    ProjectState = "in_queue"
)

// ProjectCreationMeta reports creation progress of a project.
type ProjectCreationMeta struct {
	CreationProgress float64 `json:"creation_progress"`
	Status           string  `json:"status"`
	Type             string  `json:"type"`
}

// Project is a Studio project. Rarely used nested objects (caption styles,
// agent settings) are kept as raw JSON.
type Project struct {
	ProjectID                     string                     `json:"project_id"`
	Name                          string                     `json:"name"`
	CreateDateUnix                int64                      `json:"create_date_unix"`
	CreatedByUserID               *string                    `json:"created_by_user_id,omitempty"`
	DefaultTitleVoiceRefID        string                     `json:"default_title_voice_ref_id"`
	DefaultParagraphVoiceRefID    string                     `json:"default_paragraph_voice_ref_id"`
	DefaultTitleVoiceID           *string                    `json:"default_title_voice_id,omitempty"`
	DefaultParagraphVoiceID       *string                    `json:"default_paragraph_voice_id,omitempty"`
	DefaultModelID                string                     `json:"default_model_id"`
	LastConversionDateUnix        *int64                     `json:"last_conversion_date_unix,omitempty"`
	CanBeDownloaded               bool                       `json:"can_be_downloaded"`
	Title                         *string                    `json:"title,omitempty"`
	Author                        *string                    `json:"author,omitempty"`
	Description                   *string                    `json:"description,omitempty"`
	Genres                        []string                   `json:"genres,omitempty"`
	CoverImageURL                 *string                    `json:"cover_image_url,omitempty"`
	TargetAudience                *string                    `json:"target_audience,omitempty"`
	Language                      *string                    `json:"language,omitempty"`
	ContentType                   *string                    `json:"content_type,omitempty"`
	OriginalPublicationDate       *string                    `json:"original_publication_date,omitempty"`
	MatureContent                 *bool                      `json:"mature_content,omitempty"`
	ISBNNumber                    *string                    `json:"isbn_number,omitempty"`
	VolumeNormalization           bool                       `json:"volume_normalization"`
	State                         ProjectState               `json:"state"`
	AccessLevel                   types.ResourcePermission   `json:"access_level"`
	Fiction                       *string                    `json:"fiction,omitempty"`
	QualityCheckOn                bool                       `json:"quality_check_on"`
	QualityCheckOnWhenBulkConvert bool                       `json:"quality_check_on_when_bulk_convert"`
	CreationMeta                  *ProjectCreationMeta       `json:"creation_meta,omitempty"`
	SourceType                    *string                    `json:"source_type,omitempty"`
	ChaptersEnabled               *bool                      `json:"chapters_enabled,omitempty"`
	CaptionsEnabled               *bool                      `json:"captions_enabled,omitempty"`
	CaptionStyle                  json.RawMessage            `json:"caption_style,omitempty"`
	CaptionStyleTemplateOverrides map[string]json.RawMessage `json:"caption_style_template_overrides,omitempty"`
	PublicShareID                 *string                    `json:"public_share_id,omitempty"`
	AspectRatio                   *string                    `json:"aspect_ratio,omitempty"`
	AgentSettings                 json.RawMessage            `json:"agent_settings,omitempty"`
}

// ChapterState is the state of a chapter.
type ChapterState string

// ChapterState values.
const (
	ChapterStateDefault    ChapterState = "default"
	ChapterStateConverting ChapterState = "converting"
)

// VoiceStatistics are per-voice conversion statistics of a chapter.
type VoiceStatistics struct {
	ProjectVoiceRefID      string   `json:"project_voice_ref_id"`
	CharactersUnconverted  int      `json:"characters_unconverted"`
	CharactersConverted    int      `json:"characters_converted"`
	CreditsNeededToConvert *float64 `json:"credits_needed_to_convert,omitempty"`
	VoiceID                *string  `json:"voice_id,omitempty"`
}

// ChapterStatistics are conversion statistics of a chapter.
type ChapterStatistics struct {
	CharactersUnconverted  int               `json:"characters_unconverted"`
	CharactersConverted    int               `json:"characters_converted"`
	ParagraphsConverted    int               `json:"paragraphs_converted"`
	ParagraphsUnconverted  int               `json:"paragraphs_unconverted"`
	CreditsNeededToConvert *float64          `json:"credits_needed_to_convert,omitempty"`
	VoiceStatistics        []VoiceStatistics `json:"voice_statistics,omitempty"`
}

// Chapter is a chapter of a Studio project.
type Chapter struct {
	ChapterID              string             `json:"chapter_id"`
	Name                   string             `json:"name"`
	LastConversionDateUnix *int64             `json:"last_conversion_date_unix,omitempty"`
	ConversionProgress     *float64           `json:"conversion_progress,omitempty"`
	CanBeDownloaded        bool               `json:"can_be_downloaded"`
	State                  ChapterState       `json:"state"`
	HasVideo               *bool              `json:"has_video,omitempty"`
	HasVisualContent       *bool              `json:"has_visual_content,omitempty"`
	VoiceIDs               []string           `json:"voice_ids,omitempty"`
	Statistics             *ChapterStatistics `json:"statistics,omitempty"`
	LastConversionError    *string            `json:"last_conversion_error,omitempty"`
}

// ChapterContentNode is a node of a content block. TTS nodes ("tts_node")
// carry text and a voice; other node types only set Type.
type ChapterContentNode struct {
	Type              string `json:"type"`
	Text              string `json:"text,omitempty"`
	VoiceID           string `json:"voice_id,omitempty"`
	ProjectVoiceRefID string `json:"project_voice_ref_id,omitempty"`
}

// TTSNode is a text node read by voiceID.
func TTSNode(text, voiceID string) ChapterContentNode {
	return ChapterContentNode{Type: "tts_node", Text: text, VoiceID: voiceID}
}

// ChapterContentBlock is a paragraph or heading of a chapter.
type ChapterContentBlock struct {
	BlockID string `json:"block_id,omitempty"`
	// SubType is "p", "h1", "h2" or "h3" (input only).
	SubType string               `json:"sub_type,omitempty"`
	Nodes   []ChapterContentNode `json:"nodes"`
}

// ChapterContent is the content of a chapter.
type ChapterContent struct {
	Blocks []ChapterContentBlock `json:"blocks"`
}

// ChapterWithContent is a chapter with its content.
type ChapterWithContent struct {
	Chapter
	Content ChapterContent `json:"content"`
}

// ProjectVoice is a voice used in a project.
type ProjectVoice struct {
	ProjectVoiceRefID string  `json:"project_voice_ref_id"`
	VoiceID           string  `json:"voice_id"`
	Alias             string  `json:"alias"`
	Stability         float64 `json:"stability"`
	SimilarityBoost   float64 `json:"similarity_boost"`
	Style             float64 `json:"style"`
	IsPinned          bool    `json:"is_pinned"`
	UseSpeakerBoost   bool    `json:"use_speaker_boost"`
	VolumeGain        float64 `json:"volume_gain"`
	Speed             float64 `json:"speed"`
}

// PronunciationDictionaryVersion describes a dictionary version attached to a project.
type PronunciationDictionaryVersion struct {
	VersionID                 string  `json:"version_id"`
	VersionRulesNum           int     `json:"version_rules_num"`
	PronunciationDictionaryID string  `json:"pronunciation_dictionary_id"`
	DictionaryName            string  `json:"dictionary_name"`
	VersionName               string  `json:"version_name"`
	PermissionOnResource      *string `json:"permission_on_resource,omitempty"`
	CreatedBy                 string  `json:"created_by"`
	CreationTimeUnix          int64   `json:"creation_time_unix"`
	ArchivedTimeUnix          *int64  `json:"archived_time_unix,omitempty"`
}

// ProjectExtended is a Studio project with chapters, voices and assets.
type ProjectExtended struct {
	Project
	QualityPreset                   QualityPreset                          `json:"quality_preset"`
	Chapters                        []Chapter                              `json:"chapters"`
	PronunciationDictionaryVersions []PronunciationDictionaryVersion       `json:"pronunciation_dictionary_versions"`
	PronunciationDictionaryLocators []types.PronunciationDictionaryLocator `json:"pronunciation_dictionary_locators"`
	ApplyTextNormalization          string                                 `json:"apply_text_normalization"`
	Experimental                    map[string]any                         `json:"experimental,omitempty"`
	Assets                          []json.RawMessage                      `json:"assets"`
	Voices                          []ProjectVoice                         `json:"voices"`
	BaseVoices                      []voices.Voice                         `json:"base_voices,omitempty"`
	PublishingRead                  json.RawMessage                        `json:"publishing_read,omitempty"`
}

// ProjectResponse wraps a project.
type ProjectResponse struct {
	Project Project `json:"project"`
}

// PodcastMode selects a conversation (host and guest) or bulletin (host
// only) podcast. Use ConversationPodcast or BulletinPodcast.
type PodcastMode struct {
	Type         string
	HostVoiceID  string
	GuestVoiceID string
}

// ConversationPodcast is a two-voice podcast.
func ConversationPodcast(hostVoiceID, guestVoiceID string) PodcastMode {
	return PodcastMode{Type: "conversation", HostVoiceID: hostVoiceID, GuestVoiceID: guestVoiceID}
}

// BulletinPodcast is a single-voice podcast.
func BulletinPodcast(hostVoiceID string) PodcastMode {
	return PodcastMode{Type: "bulletin", HostVoiceID: hostVoiceID}
}

// MarshalJSON implements json.Marshaler.
func (m PodcastMode) MarshalJSON() ([]byte, error) {
	type voices struct {
		HostVoiceID  string `json:"host_voice_id"`
		GuestVoiceID string `json:"guest_voice_id,omitempty"`
	}
	return json.Marshal(map[string]any{"type": m.Type, m.Type: voices{m.HostVoiceID, m.GuestVoiceID}})
}

// PodcastSource is text or a URL to build a podcast from. Use TextSource or URLSource.
type PodcastSource struct {
	// Type is "text" or "url".
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	URL  string `json:"url,omitempty"`
}

// TextSource is a podcast source made of text.
func TextSource(text string) PodcastSource { return PodcastSource{Type: "text", Text: text} }

// URLSource is a podcast source fetched from a URL.
func URLSource(url string) PodcastSource { return PodcastSource{Type: "url", URL: url} }

// PodcastSources is one or more podcast sources. A single source is sent
// as an object, several as an array.
type PodcastSources []PodcastSource

// MarshalJSON implements json.Marshaler.
func (p PodcastSources) MarshalJSON() ([]byte, error) {
	if len(p) == 1 {
		return json.Marshal(p[0])
	}
	return json.Marshal([]PodcastSource(transport.NonNil(p)))
}

// PodcastCreateRequest is the body of Service.CreatePodcast.
type PodcastCreateRequest struct {
	// ModelID is the TTS model. Required.
	ModelID string `json:"model_id"`
	// Mode selects the podcast format. Required.
	Mode PodcastMode `json:"mode"`
	// Source is the content. Required.
	Source        PodcastSources `json:"source"`
	QualityPreset QualityPreset  `json:"quality_preset,omitempty"`
	// DurationScale is "short", "default" or "long".
	DurationScale      string   `json:"duration_scale,omitempty"`
	Language           string   `json:"language,omitempty"`
	Intro              string   `json:"intro,omitempty"`
	Outro              string   `json:"outro,omitempty"`
	InstructionsPrompt string   `json:"instructions_prompt,omitempty"`
	Highlights         []string `json:"highlights,omitempty"`
	CallbackURL        string   `json:"callback_url,omitempty"`
	// ApplyTextNormalization is "auto", "on", "off" or "apply_english".
	ApplyTextNormalization string `json:"apply_text_normalization,omitempty"`
}

// CreatePodcast creates a podcast Studio project.
func (s *Service) CreatePodcast(ctx context.Context, req *PodcastCreateRequest, opts ...option.RequestOption) (*ProjectResponse, error) {
	return transport.DoValue[ProjectResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/studio/podcasts", Body: transport.OrEmpty(req), Opts: opts})
}

// ProjectsService groups the /v1/studio/projects endpoints.
type ProjectsService struct {
	cfg *requestconfig.Config

	// Chapters manages project chapters.
	Chapters *ChaptersService
	// Content replaces project content.
	Content *ContentService
	// PronunciationDictionaries attaches dictionaries to a project.
	PronunciationDictionaries *PronunciationDictionariesService
	// Snapshots lists and streams project snapshots.
	Snapshots *ProjectSnapshotsService
}

// ProjectList is the response of ProjectsService.List.
type ProjectList struct {
	Projects []Project `json:"projects"`
}

// List returns the Studio projects.
func (s *ProjectsService) List(ctx context.Context, opts ...option.RequestOption) (*ProjectList, error) {
	return transport.DoValue[ProjectList](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/studio/projects", Opts: opts})
}

// ProjectCreateRequest is the body of ProjectsService.Create.
type ProjectCreateRequest struct {
	// Name of the project. Required.
	Name                    string `form:"name"`
	DefaultTitleVoiceID     string `form:"default_title_voice_id"`
	DefaultParagraphVoiceID string `form:"default_paragraph_voice_id"`
	DefaultModelID          string `form:"default_model_id"`
	// FromURL imports content from a web page.
	FromURL string `form:"from_url"`
	// FromDocument imports content from an EPUB, PDF, TXT, HTML or DOCX file.
	FromDocument *types.File `form:"from_document"`
	// FromContentJSON imports content from a JSON chapter structure.
	FromContentJSON         string        `form:"from_content_json"`
	QualityPreset           QualityPreset `form:"quality_preset"`
	Title                   string        `form:"title"`
	Author                  string        `form:"author"`
	Description             string        `form:"description"`
	Genres                  []string      `form:"genres"`
	TargetAudience          string        `form:"target_audience"`
	Language                string        `form:"language"`
	ContentType             string        `form:"content_type"`
	OriginalPublicationDate string        `form:"original_publication_date"`
	MatureContent           *bool         `form:"mature_content"`
	ISBNNumber              string        `form:"isbn_number"`
	ACXVolumeNormalization  *bool         `form:"acx_volume_normalization"`
	VolumeNormalization     *bool         `form:"volume_normalization"`
	// PronunciationDictionaryLocators are JSON-encoded locators; use
	// PronunciationDictionaryLocator.String.
	PronunciationDictionaryLocators []string `form:"pronunciation_dictionary_locators"`
	CallbackURL                     string   `form:"callback_url"`
	// Fiction is "fiction" or "non-fiction".
	Fiction                string `form:"fiction"`
	ApplyTextNormalization string `form:"apply_text_normalization"`
	AutoConvert            *bool  `form:"auto_convert"`
	AutoAssignVoices       *bool  `form:"auto_assign_voices"`
	// SourceType is e.g. "book", "article", "genfm", "video" or "screenplay".
	SourceType string `form:"source_type"`
	// VoiceSettings are JSON-encoded per-voice settings.
	VoiceSettings        []string `form:"voice_settings"`
	CreatePublishingRead *bool    `form:"create_publishing_read"`
}

// Create creates a Studio project.
func (s *ProjectsService) Create(ctx context.Context, req *ProjectCreateRequest, opts ...option.RequestOption) (*ProjectResponse, error) {
	return transport.DoValue[ProjectResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/studio/projects", Form: transport.OrEmpty(req), Opts: opts})
}

type shareIDQuery struct {
	ShareID string `query:"share_id"`
}

// Get returns a project with its chapters. shareID may be empty.
func (s *ProjectsService) Get(ctx context.Context, projectID, shareID string, opts ...option.RequestOption) (*ProjectExtended, error) {
	return transport.DoValue[ProjectExtended](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/studio/projects/%s", projectID), Query: shareIDQuery{shareID}, Opts: opts})
}

// ProjectUpdateRequest is the body of ProjectsService.Update.
type ProjectUpdateRequest struct {
	// Name, DefaultTitleVoiceID and DefaultParagraphVoiceID are required.
	Name                    string `json:"name"`
	DefaultTitleVoiceID     string `json:"default_title_voice_id"`
	DefaultParagraphVoiceID string `json:"default_paragraph_voice_id"`
	Title                   string `json:"title,omitempty"`
	Author                  string `json:"author,omitempty"`
	ISBNNumber              string `json:"isbn_number,omitempty"`
	VolumeNormalization     *bool  `json:"volume_normalization,omitempty"`
}

// Update edits a project's metadata.
func (s *ProjectsService) Update(ctx context.Context, projectID string, req *ProjectUpdateRequest, opts ...option.RequestOption) (*ProjectResponse, error) {
	return transport.DoValue[ProjectResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/studio/projects/%s", projectID), Body: transport.OrEmpty(req), Opts: opts})
}

// Delete deletes a project.
func (s *ProjectsService) Delete(ctx context.Context, projectID string, opts ...option.RequestOption) (*types.StatusResponse, error) {
	return transport.DoValue[types.StatusResponse](ctx, s.cfg, &transport.Request{Method: http.MethodDelete, Path: transport.Pathf("/v1/studio/projects/%s", projectID), Opts: opts})
}

// Convert starts converting a whole project to audio.
func (s *ProjectsService) Convert(ctx context.Context, projectID string, opts ...option.RequestOption) (*types.StatusResponse, error) {
	return transport.DoValue[types.StatusResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/studio/projects/%s/convert", projectID), Opts: opts})
}

// ProjectMutedTracks lists chapters with muted tracks.
type ProjectMutedTracks struct {
	ChapterIDs []string `json:"chapter_ids"`
}

// GetMutedTracks returns the chapters that contain muted tracks.
func (s *ProjectsService) GetMutedTracks(ctx context.Context, projectID string, opts ...option.RequestOption) (*ProjectMutedTracks, error) {
	return transport.DoValue[ProjectMutedTracks](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/studio/projects/%s/muted-tracks", projectID), Opts: opts})
}

// ContentService groups the project content endpoint.
type ContentService struct{ cfg *requestconfig.Config }

// ProjectContentUpdateRequest is the body of ContentService.Update.
type ProjectContentUpdateRequest struct {
	FromURL         string      `form:"from_url"`
	FromDocument    *types.File `form:"from_document"`
	FromContentJSON string      `form:"from_content_json"`
	AutoConvert     *bool       `form:"auto_convert"`
}

// Update replaces the content of a project.
func (s *ContentService) Update(ctx context.Context, projectID string, req *ProjectContentUpdateRequest, opts ...option.RequestOption) (*ProjectResponse, error) {
	return transport.DoValue[ProjectResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/studio/projects/%s/content", projectID), Form: transport.OrEmpty(req), Opts: opts})
}

// PronunciationDictionariesService groups the project pronunciation dictionary endpoint.
type PronunciationDictionariesService struct{ cfg *requestconfig.Config }

// ProjectPronunciationDictionariesRequest is the body of PronunciationDictionariesService.Create.
type ProjectPronunciationDictionariesRequest struct {
	PronunciationDictionaryLocators []types.PronunciationDictionaryLocator `json:"pronunciation_dictionary_locators"`
	// InvalidateAffectedText marks affected text for reconversion.
	InvalidateAffectedText *bool `json:"invalidate_affected_text,omitempty"`
}

// Create attaches pronunciation dictionaries to a project.
func (s *PronunciationDictionariesService) Create(ctx context.Context, projectID string, req *ProjectPronunciationDictionariesRequest, opts ...option.RequestOption) (*types.StatusResponse, error) {
	body := *transport.OrEmpty(req)
	body.PronunciationDictionaryLocators = transport.NonNil(body.PronunciationDictionaryLocators)
	return transport.DoValue[types.StatusResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/studio/projects/%s/pronunciation-dictionaries", projectID), Body: &body, Opts: opts})
}

// ChaptersService groups the /v1/studio/projects/{id}/chapters endpoints.
type ChaptersService struct {
	cfg *requestconfig.Config

	// Snapshots lists and streams chapter snapshots.
	Snapshots *ChapterSnapshotsService
}

// ChapterList is the response of ChaptersService.List.
type ChapterList struct {
	Chapters []Chapter `json:"chapters"`
}

// List returns the chapters of a project.
func (s *ChaptersService) List(ctx context.Context, projectID string, opts ...option.RequestOption) (*ChapterList, error) {
	return transport.DoValue[ChapterList](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/studio/projects/%s/chapters", projectID), Opts: opts})
}

// ChapterCreateRequest is the body of ChaptersService.Create.
type ChapterCreateRequest struct {
	Name    string `json:"name"`
	FromURL string `json:"from_url,omitempty"`
}

// ChapterResponse wraps a chapter with content.
type ChapterResponse struct {
	Chapter ChapterWithContent `json:"chapter"`
}

// Create adds a chapter to a project.
func (s *ChaptersService) Create(ctx context.Context, projectID string, req *ChapterCreateRequest, opts ...option.RequestOption) (*ChapterResponse, error) {
	return transport.DoValue[ChapterResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/studio/projects/%s/chapters", projectID), Body: transport.OrEmpty(req), Opts: opts})
}

// Get returns a chapter with its content.
func (s *ChaptersService) Get(ctx context.Context, projectID, chapterID string, opts ...option.RequestOption) (*ChapterWithContent, error) {
	return transport.DoValue[ChapterWithContent](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/studio/projects/%s/chapters/%s", projectID, chapterID), Opts: opts})
}

// ChapterUpdateRequest is the body of ChaptersService.Update.
type ChapterUpdateRequest struct {
	Name    string          `json:"name,omitempty"`
	Content *ChapterContent `json:"content,omitempty"`
}

// Update edits a chapter. req may be nil.
func (s *ChaptersService) Update(ctx context.Context, projectID, chapterID string, req *ChapterUpdateRequest, opts ...option.RequestOption) (*ChapterResponse, error) {
	return transport.DoValue[ChapterResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/studio/projects/%s/chapters/%s", projectID, chapterID), Body: transport.OrEmpty(req), Opts: opts})
}

// Delete deletes a chapter.
func (s *ChaptersService) Delete(ctx context.Context, projectID, chapterID string, opts ...option.RequestOption) (*types.StatusResponse, error) {
	return transport.DoValue[types.StatusResponse](ctx, s.cfg, &transport.Request{Method: http.MethodDelete, Path: transport.Pathf("/v1/studio/projects/%s/chapters/%s", projectID, chapterID), Opts: opts})
}

// Convert starts converting a chapter to audio.
func (s *ChaptersService) Convert(ctx context.Context, projectID, chapterID string, opts ...option.RequestOption) (*types.StatusResponse, error) {
	return transport.DoValue[types.StatusResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/studio/projects/%s/chapters/%s/convert", projectID, chapterID), Opts: opts})
}

// SnapshotStreamRequest holds the options of the snapshot stream endpoints.
type SnapshotStreamRequest struct {
	// ConvertToMpeg converts the audio to MP3.
	ConvertToMpeg *bool `json:"convert_to_mpeg,omitempty"`
}

// ChapterSnapshot is a converted version of a chapter.
type ChapterSnapshot struct {
	ChapterSnapshotID string `json:"chapter_snapshot_id"`
	ProjectID         string `json:"project_id"`
	ChapterID         string `json:"chapter_id"`
	CreatedAtUnix     int64  `json:"created_at_unix"`
	Name              string `json:"name"`
}

// ChapterSnapshotExtended is a chapter snapshot with character alignments.
type ChapterSnapshotExtended struct {
	ChapterSnapshot
	CharacterAlignments []types.CharacterAlignment `json:"character_alignments"`
}

// ChapterSnapshotList is the response of ChapterSnapshotsService.List.
type ChapterSnapshotList struct {
	Snapshots []ChapterSnapshot `json:"snapshots"`
}

// ChapterSnapshotsService groups the chapter snapshot endpoints.
type ChapterSnapshotsService struct{ cfg *requestconfig.Config }

// List returns the snapshots of a chapter.
func (s *ChapterSnapshotsService) List(ctx context.Context, projectID, chapterID string, opts ...option.RequestOption) (*ChapterSnapshotList, error) {
	return transport.DoValue[ChapterSnapshotList](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/studio/projects/%s/chapters/%s/snapshots", projectID, chapterID), Opts: opts})
}

// Get returns a chapter snapshot.
func (s *ChapterSnapshotsService) Get(ctx context.Context, projectID, chapterID, snapshotID string, opts ...option.RequestOption) (*ChapterSnapshotExtended, error) {
	return transport.DoValue[ChapterSnapshotExtended](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/studio/projects/%s/chapters/%s/snapshots/%s", projectID, chapterID, snapshotID), Opts: opts})
}

// Stream streams the audio of a chapter snapshot. req may be nil. The caller must close the returned reader.
func (s *ChapterSnapshotsService) Stream(ctx context.Context, projectID, chapterID, snapshotID string, req *SnapshotStreamRequest, opts ...option.RequestOption) (io.ReadCloser, error) {
	return transport.DoStream(ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/studio/projects/%s/chapters/%s/snapshots/%s/stream", projectID, chapterID, snapshotID), Body: transport.OrEmpty(req), Accept: transport.AcceptAny, Opts: opts})
}

// ProjectSnapshot is a converted version of a project.
type ProjectSnapshot struct {
	ProjectSnapshotID string         `json:"project_snapshot_id"`
	ProjectID         string         `json:"project_id"`
	CreatedAtUnix     int64          `json:"created_at_unix"`
	Name              string         `json:"name"`
	AudioUpload       map[string]any `json:"audio_upload,omitempty"`
	ZipUpload         map[string]any `json:"zip_upload,omitempty"`
}

// ProjectSnapshotExtended is a project snapshot with character alignments.
type ProjectSnapshotExtended struct {
	ProjectSnapshot
	CharacterAlignments []types.CharacterAlignment `json:"character_alignments"`
	AudioDurationSecs   float64                    `json:"audio_duration_secs"`
}

// ProjectSnapshotList is the response of ProjectSnapshotsService.List.
type ProjectSnapshotList struct {
	Snapshots []ProjectSnapshot `json:"snapshots"`
}

// ProjectSnapshotsService groups the project snapshot endpoints.
type ProjectSnapshotsService struct{ cfg *requestconfig.Config }

// List returns the snapshots of a project.
func (s *ProjectSnapshotsService) List(ctx context.Context, projectID string, opts ...option.RequestOption) (*ProjectSnapshotList, error) {
	return transport.DoValue[ProjectSnapshotList](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/studio/projects/%s/snapshots", projectID), Opts: opts})
}

// Get returns a project snapshot.
func (s *ProjectSnapshotsService) Get(ctx context.Context, projectID, snapshotID string, opts ...option.RequestOption) (*ProjectSnapshotExtended, error) {
	return transport.DoValue[ProjectSnapshotExtended](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/studio/projects/%s/snapshots/%s", projectID, snapshotID), Opts: opts})
}

// Stream streams the audio of a project snapshot. req may be nil. The caller must close the returned reader.
func (s *ProjectSnapshotsService) Stream(ctx context.Context, projectID, snapshotID string, req *SnapshotStreamRequest, opts ...option.RequestOption) (io.ReadCloser, error) {
	return transport.DoStream(ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/studio/projects/%s/snapshots/%s/stream", projectID, snapshotID), Body: transport.OrEmpty(req), Accept: transport.AcceptAny, Opts: opts})
}

// StreamArchive streams a ZIP archive of a project snapshot's chapter audio. The caller must close the returned reader.
func (s *ProjectSnapshotsService) StreamArchive(ctx context.Context, projectID, snapshotID string, opts ...option.RequestOption) (io.ReadCloser, error) {
	return transport.DoStream(ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/studio/projects/%s/snapshots/%s/archive", projectID, snapshotID), Accept: transport.AcceptAny, Opts: opts})
}
