package elevenlabs

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/types"
	"github.com/taigrr/elevenlabs/v2/voices"
)

// StudioService groups the /v1/studio endpoints.
type StudioService struct {
	cfg *config

	// Projects manages Studio projects.
	Projects *StudioProjectsService
}

func newStudioService(cfg *config) *StudioService {
	return &StudioService{cfg: cfg, Projects: &StudioProjectsService{
		cfg: cfg,
		Chapters: &StudioChaptersService{
			cfg:       cfg,
			Snapshots: &StudioChapterSnapshotsService{cfg: cfg},
		},
		Content:                   &StudioContentService{cfg: cfg},
		PronunciationDictionaries: &StudioPronunciationDictionariesService{cfg: cfg},
		Snapshots:                 &StudioProjectSnapshotsService{cfg: cfg},
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
	AccessLevel                   ResourceRole               `json:"access_level"`
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
	return json.Marshal([]PodcastSource(nonNil(p)))
}

// PodcastCreateRequest is the body of StudioService.CreatePodcast.
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
func (s *StudioService) CreatePodcast(ctx context.Context, req *PodcastCreateRequest, opts ...Option) (*ProjectResponse, error) {
	return doValue[ProjectResponse](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/studio/podcasts", body: orEmpty(req), opts: opts})
}

// StudioProjectsService groups the /v1/studio/projects endpoints.
type StudioProjectsService struct {
	cfg *config

	// Chapters manages project chapters.
	Chapters *StudioChaptersService
	// Content replaces project content.
	Content *StudioContentService
	// PronunciationDictionaries attaches dictionaries to a project.
	PronunciationDictionaries *StudioPronunciationDictionariesService
	// Snapshots lists and streams project snapshots.
	Snapshots *StudioProjectSnapshotsService
}

// ProjectList is the response of StudioProjectsService.List.
type ProjectList struct {
	Projects []Project `json:"projects"`
}

// List returns the Studio projects.
func (s *StudioProjectsService) List(ctx context.Context, opts ...Option) (*ProjectList, error) {
	return doValue[ProjectList](ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/studio/projects", opts: opts})
}

// ProjectCreateRequest is the body of StudioProjectsService.Create.
type ProjectCreateRequest struct {
	// Name of the project. Required.
	Name                    string `form:"name"`
	DefaultTitleVoiceID     string `form:"default_title_voice_id"`
	DefaultParagraphVoiceID string `form:"default_paragraph_voice_id"`
	DefaultModelID          string `form:"default_model_id"`
	// FromURL imports content from a web page.
	FromURL string `form:"from_url"`
	// FromDocument imports content from an EPUB, PDF, TXT, HTML or DOCX file.
	FromDocument *File `form:"from_document"`
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
func (s *StudioProjectsService) Create(ctx context.Context, req *ProjectCreateRequest, opts ...Option) (*ProjectResponse, error) {
	return doValue[ProjectResponse](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/studio/projects", form: orEmpty(req), opts: opts})
}

type shareIDQuery struct {
	ShareID string `query:"share_id"`
}

// Get returns a project with its chapters. shareID may be empty.
func (s *StudioProjectsService) Get(ctx context.Context, projectID, shareID string, opts ...Option) (*ProjectExtended, error) {
	return doValue[ProjectExtended](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/studio/projects/%s", projectID), query: shareIDQuery{shareID}, opts: opts})
}

// ProjectUpdateRequest is the body of StudioProjectsService.Update.
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
func (s *StudioProjectsService) Update(ctx context.Context, projectID string, req *ProjectUpdateRequest, opts ...Option) (*ProjectResponse, error) {
	return doValue[ProjectResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/studio/projects/%s", projectID), body: orEmpty(req), opts: opts})
}

// Delete deletes a project.
func (s *StudioProjectsService) Delete(ctx context.Context, projectID string, opts ...Option) (*StatusResponse, error) {
	return doValue[StatusResponse](ctx, s.cfg, &request{method: http.MethodDelete, path: pathf("/v1/studio/projects/%s", projectID), opts: opts})
}

// Convert starts converting a whole project to audio.
func (s *StudioProjectsService) Convert(ctx context.Context, projectID string, opts ...Option) (*StatusResponse, error) {
	return doValue[StatusResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/studio/projects/%s/convert", projectID), opts: opts})
}

// ProjectMutedTracks lists chapters with muted tracks.
type ProjectMutedTracks struct {
	ChapterIDs []string `json:"chapter_ids"`
}

// GetMutedTracks returns the chapters that contain muted tracks.
func (s *StudioProjectsService) GetMutedTracks(ctx context.Context, projectID string, opts ...Option) (*ProjectMutedTracks, error) {
	return doValue[ProjectMutedTracks](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/studio/projects/%s/muted-tracks", projectID), opts: opts})
}

// StudioContentService groups the project content endpoint.
type StudioContentService struct{ cfg *config }

// ProjectContentUpdateRequest is the body of StudioContentService.Update.
type ProjectContentUpdateRequest struct {
	FromURL         string `form:"from_url"`
	FromDocument    *File  `form:"from_document"`
	FromContentJSON string `form:"from_content_json"`
	AutoConvert     *bool  `form:"auto_convert"`
}

// Update replaces the content of a project.
func (s *StudioContentService) Update(ctx context.Context, projectID string, req *ProjectContentUpdateRequest, opts ...Option) (*ProjectResponse, error) {
	return doValue[ProjectResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/studio/projects/%s/content", projectID), form: orEmpty(req), opts: opts})
}

// StudioPronunciationDictionariesService groups the project pronunciation dictionary endpoint.
type StudioPronunciationDictionariesService struct{ cfg *config }

// ProjectPronunciationDictionariesRequest is the body of StudioPronunciationDictionariesService.Create.
type ProjectPronunciationDictionariesRequest struct {
	PronunciationDictionaryLocators []types.PronunciationDictionaryLocator `json:"pronunciation_dictionary_locators"`
	// InvalidateAffectedText marks affected text for reconversion.
	InvalidateAffectedText *bool `json:"invalidate_affected_text,omitempty"`
}

// Create attaches pronunciation dictionaries to a project.
func (s *StudioPronunciationDictionariesService) Create(ctx context.Context, projectID string, req *ProjectPronunciationDictionariesRequest, opts ...Option) (*StatusResponse, error) {
	body := *orEmpty(req)
	body.PronunciationDictionaryLocators = nonNil(body.PronunciationDictionaryLocators)
	return doValue[StatusResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/studio/projects/%s/pronunciation-dictionaries", projectID), body: &body, opts: opts})
}

// StudioChaptersService groups the /v1/studio/projects/{id}/chapters endpoints.
type StudioChaptersService struct {
	cfg *config

	// Snapshots lists and streams chapter snapshots.
	Snapshots *StudioChapterSnapshotsService
}

// ChapterList is the response of StudioChaptersService.List.
type ChapterList struct {
	Chapters []Chapter `json:"chapters"`
}

// List returns the chapters of a project.
func (s *StudioChaptersService) List(ctx context.Context, projectID string, opts ...Option) (*ChapterList, error) {
	return doValue[ChapterList](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/studio/projects/%s/chapters", projectID), opts: opts})
}

// ChapterCreateRequest is the body of StudioChaptersService.Create.
type ChapterCreateRequest struct {
	Name    string `json:"name"`
	FromURL string `json:"from_url,omitempty"`
}

// ChapterResponse wraps a chapter with content.
type ChapterResponse struct {
	Chapter ChapterWithContent `json:"chapter"`
}

// Create adds a chapter to a project.
func (s *StudioChaptersService) Create(ctx context.Context, projectID string, req *ChapterCreateRequest, opts ...Option) (*ChapterResponse, error) {
	return doValue[ChapterResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/studio/projects/%s/chapters", projectID), body: orEmpty(req), opts: opts})
}

// Get returns a chapter with its content.
func (s *StudioChaptersService) Get(ctx context.Context, projectID, chapterID string, opts ...Option) (*ChapterWithContent, error) {
	return doValue[ChapterWithContent](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/studio/projects/%s/chapters/%s", projectID, chapterID), opts: opts})
}

// ChapterUpdateRequest is the body of StudioChaptersService.Update.
type ChapterUpdateRequest struct {
	Name    string          `json:"name,omitempty"`
	Content *ChapterContent `json:"content,omitempty"`
}

// Update edits a chapter. req may be nil.
func (s *StudioChaptersService) Update(ctx context.Context, projectID, chapterID string, req *ChapterUpdateRequest, opts ...Option) (*ChapterResponse, error) {
	return doValue[ChapterResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/studio/projects/%s/chapters/%s", projectID, chapterID), body: orEmpty(req), opts: opts})
}

// Delete deletes a chapter.
func (s *StudioChaptersService) Delete(ctx context.Context, projectID, chapterID string, opts ...Option) (*StatusResponse, error) {
	return doValue[StatusResponse](ctx, s.cfg, &request{method: http.MethodDelete, path: pathf("/v1/studio/projects/%s/chapters/%s", projectID, chapterID), opts: opts})
}

// Convert starts converting a chapter to audio.
func (s *StudioChaptersService) Convert(ctx context.Context, projectID, chapterID string, opts ...Option) (*StatusResponse, error) {
	return doValue[StatusResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/studio/projects/%s/chapters/%s/convert", projectID, chapterID), opts: opts})
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

// ChapterSnapshotList is the response of StudioChapterSnapshotsService.List.
type ChapterSnapshotList struct {
	Snapshots []ChapterSnapshot `json:"snapshots"`
}

// StudioChapterSnapshotsService groups the chapter snapshot endpoints.
type StudioChapterSnapshotsService struct{ cfg *config }

// List returns the snapshots of a chapter.
func (s *StudioChapterSnapshotsService) List(ctx context.Context, projectID, chapterID string, opts ...Option) (*ChapterSnapshotList, error) {
	return doValue[ChapterSnapshotList](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/studio/projects/%s/chapters/%s/snapshots", projectID, chapterID), opts: opts})
}

// Get returns a chapter snapshot.
func (s *StudioChapterSnapshotsService) Get(ctx context.Context, projectID, chapterID, snapshotID string, opts ...Option) (*ChapterSnapshotExtended, error) {
	return doValue[ChapterSnapshotExtended](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/studio/projects/%s/chapters/%s/snapshots/%s", projectID, chapterID, snapshotID), opts: opts})
}

// Stream streams the audio of a chapter snapshot. req may be nil. The caller must close the returned reader.
func (s *StudioChapterSnapshotsService) Stream(ctx context.Context, projectID, chapterID, snapshotID string, req *SnapshotStreamRequest, opts ...Option) (io.ReadCloser, error) {
	return doStream(ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/studio/projects/%s/chapters/%s/snapshots/%s/stream", projectID, chapterID, snapshotID), body: orEmpty(req), accept: acceptAny, opts: opts})
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

// ProjectSnapshotList is the response of StudioProjectSnapshotsService.List.
type ProjectSnapshotList struct {
	Snapshots []ProjectSnapshot `json:"snapshots"`
}

// StudioProjectSnapshotsService groups the project snapshot endpoints.
type StudioProjectSnapshotsService struct{ cfg *config }

// List returns the snapshots of a project.
func (s *StudioProjectSnapshotsService) List(ctx context.Context, projectID string, opts ...Option) (*ProjectSnapshotList, error) {
	return doValue[ProjectSnapshotList](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/studio/projects/%s/snapshots", projectID), opts: opts})
}

// Get returns a project snapshot.
func (s *StudioProjectSnapshotsService) Get(ctx context.Context, projectID, snapshotID string, opts ...Option) (*ProjectSnapshotExtended, error) {
	return doValue[ProjectSnapshotExtended](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/studio/projects/%s/snapshots/%s", projectID, snapshotID), opts: opts})
}

// Stream streams the audio of a project snapshot. req may be nil. The caller must close the returned reader.
func (s *StudioProjectSnapshotsService) Stream(ctx context.Context, projectID, snapshotID string, req *SnapshotStreamRequest, opts ...Option) (io.ReadCloser, error) {
	return doStream(ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/studio/projects/%s/snapshots/%s/stream", projectID, snapshotID), body: orEmpty(req), accept: acceptAny, opts: opts})
}

// StreamArchive streams a ZIP archive of a project snapshot's chapter audio. The caller must close the returned reader.
func (s *StudioProjectSnapshotsService) StreamArchive(ctx context.Context, projectID, snapshotID string, opts ...Option) (io.ReadCloser, error) {
	return doStream(ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/studio/projects/%s/snapshots/%s/archive", projectID, snapshotID), accept: acceptAny, opts: opts})
}
