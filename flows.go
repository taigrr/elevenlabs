package elevenlabs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iter"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// FlowsService groups the /v1/flows endpoints: asynchronous image, video and
// speech generations and template runs. Each create endpoint has a
// CreateAndWait helper that polls until the work finishes.
type FlowsService struct {
	// Image generates images.
	Image *FlowsImageService
	// Video generates videos.
	Video *FlowsVideoService
	// TextToSpeech generates speech.
	TextToSpeech *FlowsTextToSpeechService
	// Templates lists templates and runs them.
	Templates *FlowsTemplatesService
}

func newFlowsService(cfg *config) *FlowsService {
	return &FlowsService{
		Image:        &FlowsImageService{cfg: cfg},
		Video:        &FlowsVideoService{cfg: cfg},
		TextToSpeech: &FlowsTextToSpeechService{cfg: cfg},
		Templates:    &FlowsTemplatesService{cfg: cfg, Runs: &FlowsTemplateRunsService{cfg: cfg}},
	}
}

// GenerationStatus is the state of a generation or template run.
type GenerationStatus string

// GenerationStatus values.
const (
	GenerationStatusPending    GenerationStatus = "pending"
	GenerationStatusGenerating GenerationStatus = "generating"
	GenerationStatusCompleted  GenerationStatus = "completed"
	GenerationStatusFailed     GenerationStatus = "failed"
)

// Terminal reports whether the status is completed or failed.
func (s GenerationStatus) Terminal() bool {
	return s == GenerationStatusCompleted || s == GenerationStatusFailed
}

// WebhookTarget selects which workspace webhooks are notified when a
// generation finishes. Use AllWebhooks or WebhookIDs.
type WebhookTarget struct {
	// Type is "all" or "ids".
	Type string   `json:"type"`
	IDs  []string `json:"ids,omitempty"`
}

// AllWebhooks notifies every workspace webhook subscribed to flows events.
func AllWebhooks() *WebhookTarget { return &WebhookTarget{Type: "all"} }

// WebhookIDs notifies only the given webhooks.
func WebhookIDs(ids ...string) *WebhookTarget {
	return &WebhookTarget{Type: "ids", IDs: nonNil(ids)}
}

// MediaReference references input media: an asset, a previous generation,
// inline base64 content or (for template inputs) a voice. Use AssetRef,
// GenerationRef, InlineRef or VoiceRef.
type MediaReference struct {
	// Type is "asset", "generation", "inline_base64" or "voice".
	Type          string `json:"type"`
	AssetID       string `json:"asset_id,omitempty"`
	GenerationID  string `json:"generation_id,omitempty"`
	ContentBase64 string `json:"content_base64,omitempty"`
	MimeType      string `json:"mime_type,omitempty"`
	VoiceID       string `json:"voice_id,omitempty"`
}

// AssetRef references an uploaded asset.
func AssetRef(assetID string) MediaReference {
	return MediaReference{Type: "asset", AssetID: assetID}
}

// GenerationRef references the output of a previous generation.
func GenerationRef(generationID string) MediaReference {
	return MediaReference{Type: "generation", GenerationID: generationID}
}

// InlineRef embeds base64 content with its MIME type.
func InlineRef(contentBase64, mimeType string) MediaReference {
	return MediaReference{Type: "inline_base64", ContentBase64: contentBase64, MimeType: mimeType}
}

// VoiceRef references a voice (template run inputs only).
func VoiceRef(voiceID string) MediaReference {
	return MediaReference{Type: "voice", VoiceID: voiceID}
}

// MediaGenerationCreateResponse is returned when a generation is accepted.
type MediaGenerationCreateResponse struct {
	ID     string           `json:"id"`
	Status GenerationStatus `json:"status"`
}

// MediaGeneration is the state of a generation. ContentURL and
// ContentMimeType are set when Status is completed; FailureReason and
// ErrorMessage when it is failed.
type MediaGeneration struct {
	ID              string           `json:"id"`
	Status          GenerationStatus `json:"status"`
	ContentURL      string           `json:"content_url,omitempty"`
	ContentMimeType string           `json:"content_mime_type,omitempty"`
	FailureReason   string           `json:"failure_reason,omitempty"`
	ErrorMessage    string           `json:"error_message,omitempty"`
}

// MediaGenerationPage is a page of generations.
type MediaGenerationPage struct {
	Generations []MediaGeneration `json:"generations"`
	NextCursor  *string           `json:"next_cursor,omitempty"`
	HasMore     *bool             `json:"has_more,omitempty"`
}

// MediaGenerationListRequest holds the optional parameters of the flows list endpoints.
type MediaGenerationListRequest struct {
	Cursor   string           `query:"cursor"`
	PageSize *int             `query:"page_size"`
	Status   GenerationStatus `query:"status"`
	ModelID  string           `query:"model_id"`
}

// ImageGenerationRequest is the body of FlowsImageService.Create. Which
// optional fields are accepted depends on ModelID.
type ImageGenerationRequest struct {
	// ModelID selects the image model, e.g. "gpt-image-2". Required.
	ModelID string `json:"model_id"`
	// Prompt describes the image. Required.
	Prompt      string           `json:"prompt"`
	Webhook     *WebhookTarget   `json:"webhook,omitempty"`
	Images      []MediaReference `json:"images,omitempty"`
	Mask        *MediaReference  `json:"mask,omitempty"`
	Quality     string           `json:"quality,omitempty"`
	Background  string           `json:"background,omitempty"`
	AspectRatio string           `json:"aspect_ratio,omitempty"`
	Resolution  string           `json:"resolution,omitempty"`
	Seed        *int             `json:"seed,omitempty"`
}

// VeoImageReference is a reference image for Veo models with its role
// ("subject" or "style").
type VeoImageReference struct {
	Image MediaReference `json:"image"`
	Role  string         `json:"role"`
}

// VideoGenerationRequest is the body of FlowsVideoService.Create. Which
// optional fields are accepted depends on ModelID.
type VideoGenerationRequest struct {
	// ModelID selects the video model, e.g. "veo-3.1-generate-001". Required.
	ModelID string `json:"model_id"`
	// Prompt describes the video. Required for every model except creatify-aurora.
	Prompt             string           `json:"prompt,omitempty"`
	Webhook            *WebhookTarget   `json:"webhook,omitempty"`
	NegativePrompt     string           `json:"negative_prompt,omitempty"`
	Seed               *int             `json:"seed,omitempty"`
	EnhancePrompt      *bool            `json:"enhance_prompt,omitempty"`
	DurationSecs       *float64         `json:"duration_secs,omitempty"`
	AspectRatio        string           `json:"aspect_ratio,omitempty"`
	Resolution         string           `json:"resolution,omitempty"`
	GenerateAudio      *bool            `json:"generate_audio,omitempty"`
	StartFrame         *MediaReference  `json:"start_frame,omitempty"`
	EndFrame           *MediaReference  `json:"end_frame,omitempty"`
	Videos             []MediaReference `json:"videos,omitempty"`
	Audios             []MediaReference `json:"audios,omitempty"`
	Image              *MediaReference  `json:"image,omitempty"`
	Audio              *MediaReference  `json:"audio,omitempty"`
	GuidanceScale      *float64         `json:"guidance_scale,omitempty"`
	AudioGuidanceScale *float64         `json:"audio_guidance_scale,omitempty"`
	// Images are reference images. Veo models take []VeoImageReference,
	// other models []MediaReference.
	Images any `json:"images,omitempty"`
}

// FlowsVoiceSettings tunes a flows speech generation.
type FlowsVoiceSettings struct {
	Stability       *float64 `json:"stability,omitempty"`
	SimilarityBoost *float64 `json:"similarity_boost,omitempty"`
	Style           *float64 `json:"style,omitempty"`
	UseSpeakerBoost *bool    `json:"use_speaker_boost,omitempty"`
	Speed           *float64 `json:"speed,omitempty"`
}

// TextToSpeechGenerationRequest is the body of FlowsTextToSpeechService.Create.
type TextToSpeechGenerationRequest struct {
	// ModelID is "eleven_v3", "eleven_multilingual_v2" or "eleven_flash_v2_5". Required.
	ModelID string `json:"model_id"`
	// Text to speak. Required.
	Text string `json:"text"`
	// Voice is the voice ID. Required.
	Voice                           string                           `json:"voice"`
	Webhook                         *WebhookTarget                   `json:"webhook,omitempty"`
	OutputFormat                    OutputFormat                     `json:"output_format,omitempty"`
	PronunciationDictionaryLocators []PronunciationDictionaryLocator `json:"pronunciation_dictionary_locators,omitempty"`
	LanguageCode                    string                           `json:"language_code,omitempty"`
	VoiceSettings                   *FlowsVoiceSettings              `json:"voice_settings,omitempty"`
}

// FlowsWaitOptions controls CreateAndWait polling. A nil *FlowsWaitOptions
// uses the defaults.
type FlowsWaitOptions struct {
	// WaitTimeout bounds the total wait; zero waits until ctx is done.
	WaitTimeout time.Duration
	// MinPollInterval is the shortest wait between status checks and the
	// wait used when the server sends no Retry-After header. Defaults to 3s;
	// must be at least 500ms.
	MinPollInterval time.Duration
}

const (
	defaultFlowsPollInterval  = 3 * time.Second
	smallestFlowsPollInterval = 500 * time.Millisecond
	flowsPollJitter           = 0.2
)

// ErrFlowsWaitTimeout is matched by *FlowsWaitTimeoutError.
var ErrFlowsWaitTimeout = errors.New("elevenlabs: flows wait timed out")

// ErrInvalidWaitOptions is returned for out-of-range FlowsWaitOptions.
var ErrInvalidWaitOptions = errors.New("elevenlabs: invalid flows wait options")

// FlowsWaitTimeoutError is returned by CreateAndWait when the work is still
// unfinished after WaitTimeout. The work keeps running on the server; resume
// with Get using ID (and TemplateID for template runs).
type FlowsWaitTimeoutError struct {
	ID          string
	TemplateID  string
	WaitTimeout time.Duration
	LastStatus  GenerationStatus
}

// Error implements error.
func (e *FlowsWaitTimeoutError) Error() string {
	if e.TemplateID != "" {
		return fmt.Sprintf("elevenlabs: template run %s of template %s is still %s after %s", e.ID, e.TemplateID, e.LastStatus, e.WaitTimeout)
	}
	return fmt.Sprintf("elevenlabs: generation %s is still %s after %s", e.ID, e.LastStatus, e.WaitTimeout)
}

// Is makes errors.Is(err, ErrFlowsWaitTimeout) match.
func (e *FlowsWaitTimeoutError) Is(target error) bool { return target == ErrFlowsWaitTimeout }

// pollDelay is the server's Retry-After (never below minInterval) plus up to
// 20% positive jitter.
func (cfg *config) pollDelay(h http.Header, minInterval time.Duration) time.Duration {
	d := minInterval
	if ra := strings.TrimSpace(h.Get(headerRetryAfter)); ra != "" {
		if secs, err := strconv.Atoi(ra); err == nil && secs >= 0 {
			d = max(time.Duration(secs)*time.Second, minInterval)
		} else if at, err := http.ParseTime(ra); err == nil {
			d = max(at.Sub(cfg.now()), minInterval)
		}
	}
	return time.Duration(float64(d) * (1 + cfg.jitter()*flowsPollJitter))
}

// waitTerminal polls get until status is terminal. create may return a nil
// result, in which case get is called even if the status is terminal.
func waitTerminal[T any](ctx context.Context, cfg *config, w *FlowsWaitOptions, templateID string,
	create func() (*T, string, GenerationStatus, http.Header, error),
	get func(id string) (*T, GenerationStatus, http.Header, error),
) (*T, error) {
	opts := FlowsWaitOptions{MinPollInterval: defaultFlowsPollInterval}
	if w != nil {
		opts.WaitTimeout = w.WaitTimeout
		if w.MinPollInterval != 0 {
			opts.MinPollInterval = w.MinPollInterval
		}
	}
	if opts.MinPollInterval < smallestFlowsPollInterval {
		return nil, fmt.Errorf("%w: MinPollInterval must be at least %s, got %s", ErrInvalidWaitOptions, smallestFlowsPollInterval, opts.MinPollInterval)
	}
	if opts.WaitTimeout < 0 {
		return nil, fmt.Errorf("%w: WaitTimeout must not be negative", ErrInvalidWaitOptions)
	}
	started := cfg.now()
	result, id, status, header, err := create()
	if err != nil {
		return nil, err
	}
	for result == nil || !status.Terminal() {
		if !status.Terminal() {
			delay := cfg.pollDelay(header, opts.MinPollInterval)
			if opts.WaitTimeout > 0 {
				remaining := opts.WaitTimeout - cfg.now().Sub(started)
				if remaining <= 0 {
					return nil, &FlowsWaitTimeoutError{ID: id, TemplateID: templateID, WaitTimeout: opts.WaitTimeout, LastStatus: status}
				}
				delay = min(delay, remaining)
			}
			if err := cfg.sleep(ctx, delay); err != nil {
				return nil, err
			}
		}
		if result, status, header, err = get(id); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func listGenerations(ctx context.Context, cfg *config, path string, req *MediaGenerationListRequest, opts []RequestOption) (*MediaGenerationPage, error) {
	return doValue[MediaGenerationPage](ctx, cfg, &request{method: http.MethodGet, path: path, query: req, opts: opts})
}

func allGenerations(ctx context.Context, cfg *config, path string, req *MediaGenerationListRequest, opts []RequestOption) iter.Seq2[MediaGeneration, error] {
	params := *orEmpty(req)
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[MediaGeneration], error) {
		params.Cursor = cursor
		resp, err := listGenerations(ctx, cfg, path, &params, opts)
		if err != nil {
			return page[MediaGeneration]{}, err
		}
		return page[MediaGeneration]{items: resp.Generations, next: deref(resp.NextCursor), hasMore: resp.HasMore == nil || *resp.HasMore}, nil
	})
}

func createGeneration(ctx context.Context, cfg *config, path string, body any, opts []RequestOption) (*MediaGenerationCreateResponse, http.Header, error) {
	return doValueHeader[MediaGenerationCreateResponse](ctx, cfg, &request{method: http.MethodPost, path: path, body: body, opts: opts})
}

func getGeneration(ctx context.Context, cfg *config, path, id string, opts []RequestOption) (*MediaGeneration, http.Header, error) {
	return doValueHeader[MediaGeneration](ctx, cfg, &request{method: http.MethodGet, path: path + "/" + pathf("%s", id), opts: opts})
}

func createGenerationAndWait(ctx context.Context, cfg *config, path string, body any, w *FlowsWaitOptions, opts []RequestOption) (*MediaGeneration, error) {
	return waitTerminal(ctx, cfg, w, "",
		func() (*MediaGeneration, string, GenerationStatus, http.Header, error) {
			r, h, err := createGeneration(ctx, cfg, path, body, opts)
			if err != nil {
				return nil, "", "", nil, err
			}
			return nil, r.ID, r.Status, h, nil
		},
		func(id string) (*MediaGeneration, GenerationStatus, http.Header, error) {
			g, h, err := getGeneration(ctx, cfg, path, id, opts)
			if err != nil {
				return nil, "", nil, err
			}
			return g, g.Status, h, nil
		})
}

const (
	flowsImagePath = "/v1/flows/image"
	flowsVideoPath = "/v1/flows/video"
	flowsTTSPath   = "/v1/flows/text-to-speech"
)

// FlowsImageService groups the /v1/flows/image endpoints.
type FlowsImageService struct{ cfg *config }

// List returns one page of image generations. req may be nil.
func (s *FlowsImageService) List(ctx context.Context, req *MediaGenerationListRequest, opts ...RequestOption) (*MediaGenerationPage, error) {
	return listGenerations(ctx, s.cfg, flowsImagePath, req, opts)
}

// ListAll iterates over every image generation.
func (s *FlowsImageService) ListAll(ctx context.Context, req *MediaGenerationListRequest, opts ...RequestOption) iter.Seq2[MediaGeneration, error] {
	return allGenerations(ctx, s.cfg, flowsImagePath, req, opts)
}

// Create starts an image generation.
func (s *FlowsImageService) Create(ctx context.Context, req *ImageGenerationRequest, opts ...RequestOption) (*MediaGenerationCreateResponse, error) {
	r, _, err := createGeneration(ctx, s.cfg, flowsImagePath, orEmpty(req), opts)
	return r, err
}

// Get returns an image generation.
func (s *FlowsImageService) Get(ctx context.Context, generationID string, opts ...RequestOption) (*MediaGeneration, error) {
	g, _, err := getGeneration(ctx, s.cfg, flowsImagePath, generationID, opts)
	return g, err
}

// CreateAndWait starts an image generation and polls until it is completed
// or failed. A failed generation is returned, not reported as an error. In
// production prefer a webhook. Returns *FlowsWaitTimeoutError if
// w.WaitTimeout elapses first. w may be nil.
func (s *FlowsImageService) CreateAndWait(ctx context.Context, req *ImageGenerationRequest, w *FlowsWaitOptions, opts ...RequestOption) (*MediaGeneration, error) {
	return createGenerationAndWait(ctx, s.cfg, flowsImagePath, orEmpty(req), w, opts)
}

// FlowsVideoService groups the /v1/flows/video endpoints.
type FlowsVideoService struct{ cfg *config }

// List returns one page of video generations. req may be nil.
func (s *FlowsVideoService) List(ctx context.Context, req *MediaGenerationListRequest, opts ...RequestOption) (*MediaGenerationPage, error) {
	return listGenerations(ctx, s.cfg, flowsVideoPath, req, opts)
}

// ListAll iterates over every video generation.
func (s *FlowsVideoService) ListAll(ctx context.Context, req *MediaGenerationListRequest, opts ...RequestOption) iter.Seq2[MediaGeneration, error] {
	return allGenerations(ctx, s.cfg, flowsVideoPath, req, opts)
}

// Create starts a video generation.
func (s *FlowsVideoService) Create(ctx context.Context, req *VideoGenerationRequest, opts ...RequestOption) (*MediaGenerationCreateResponse, error) {
	r, _, err := createGeneration(ctx, s.cfg, flowsVideoPath, orEmpty(req), opts)
	return r, err
}

// Get returns a video generation.
func (s *FlowsVideoService) Get(ctx context.Context, generationID string, opts ...RequestOption) (*MediaGeneration, error) {
	g, _, err := getGeneration(ctx, s.cfg, flowsVideoPath, generationID, opts)
	return g, err
}

// CreateAndWait starts a video generation and polls until it is completed
// or failed. See FlowsImageService.CreateAndWait.
func (s *FlowsVideoService) CreateAndWait(ctx context.Context, req *VideoGenerationRequest, w *FlowsWaitOptions, opts ...RequestOption) (*MediaGeneration, error) {
	return createGenerationAndWait(ctx, s.cfg, flowsVideoPath, orEmpty(req), w, opts)
}

// FlowsTextToSpeechService groups the /v1/flows/text-to-speech endpoints.
type FlowsTextToSpeechService struct{ cfg *config }

// List returns one page of speech generations. req may be nil.
func (s *FlowsTextToSpeechService) List(ctx context.Context, req *MediaGenerationListRequest, opts ...RequestOption) (*MediaGenerationPage, error) {
	return listGenerations(ctx, s.cfg, flowsTTSPath, req, opts)
}

// ListAll iterates over every speech generation.
func (s *FlowsTextToSpeechService) ListAll(ctx context.Context, req *MediaGenerationListRequest, opts ...RequestOption) iter.Seq2[MediaGeneration, error] {
	return allGenerations(ctx, s.cfg, flowsTTSPath, req, opts)
}

// Create starts a speech generation.
func (s *FlowsTextToSpeechService) Create(ctx context.Context, req *TextToSpeechGenerationRequest, opts ...RequestOption) (*MediaGenerationCreateResponse, error) {
	r, _, err := createGeneration(ctx, s.cfg, flowsTTSPath, orEmpty(req), opts)
	return r, err
}

// Get returns a speech generation.
func (s *FlowsTextToSpeechService) Get(ctx context.Context, generationID string, opts ...RequestOption) (*MediaGeneration, error) {
	g, _, err := getGeneration(ctx, s.cfg, flowsTTSPath, generationID, opts)
	return g, err
}

// CreateAndWait starts a speech generation and polls until it is completed
// or failed. See FlowsImageService.CreateAndWait.
func (s *FlowsTextToSpeechService) CreateAndWait(ctx context.Context, req *TextToSpeechGenerationRequest, w *FlowsWaitOptions, opts ...RequestOption) (*MediaGeneration, error) {
	return createGenerationAndWait(ctx, s.cfg, flowsTTSPath, orEmpty(req), w, opts)
}

// FlowsTemplatesService groups the /v1/flows/templates endpoints.
type FlowsTemplatesService struct {
	cfg *config

	// Runs starts and inspects template runs.
	Runs *FlowsTemplateRunsService
}

// TemplatePort is a template input or output with its JSON content schema.
type TemplatePort struct {
	ID            string          `json:"id"`
	ContentSchema json.RawMessage `json:"content_schema"`
}

// TemplateVersion is a published version of a template.
type TemplateVersion struct {
	VersionID       string         `json:"version_id"`
	PublishedAtUnix int64          `json:"published_at_unix"`
	IsLatest        bool           `json:"is_latest"`
	Inputs          []TemplatePort `json:"inputs"`
	Outputs         []TemplatePort `json:"outputs"`
}

// TemplateSummary describes a flows template.
type TemplateSummary struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Description     *string           `json:"description,omitempty"`
	Versions        []TemplateVersion `json:"versions"`
	HasMoreVersions bool              `json:"has_more_versions"`
}

// TemplatesListRequest holds the optional parameters of FlowsTemplatesService.List.
type TemplatesListRequest struct {
	Cursor              string `query:"cursor"`
	PageSize            *int   `query:"page_size"`
	VersionsPerTemplate *int   `query:"versions_per_template"`
	Search              string `query:"search"`
}

// TemplatePage is a page of templates.
type TemplatePage struct {
	Templates  []TemplateSummary `json:"templates"`
	NextCursor *string           `json:"next_cursor,omitempty"`
	HasMore    *bool             `json:"has_more,omitempty"`
}

// List returns one page of templates. req may be nil.
func (s *FlowsTemplatesService) List(ctx context.Context, req *TemplatesListRequest, opts ...RequestOption) (*TemplatePage, error) {
	return doValue[TemplatePage](ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/flows/templates", query: req, opts: opts})
}

// ListAll iterates over every template.
func (s *FlowsTemplatesService) ListAll(ctx context.Context, req *TemplatesListRequest, opts ...RequestOption) iter.Seq2[TemplateSummary, error] {
	params := *orEmpty(req)
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[TemplateSummary], error) {
		params.Cursor = cursor
		resp, err := s.List(ctx, &params, opts...)
		if err != nil {
			return page[TemplateSummary]{}, err
		}
		return page[TemplateSummary]{items: resp.Templates, next: deref(resp.NextCursor), hasMore: resp.HasMore == nil || *resp.HasMore}, nil
	})
}

// TemplatesGetRequest holds the optional parameters of FlowsTemplatesService.Get.
type TemplatesGetRequest struct {
	VersionsPerTemplate *int `query:"versions_per_template"`
}

// Get returns a template. req may be nil.
func (s *FlowsTemplatesService) Get(ctx context.Context, templateID string, req *TemplatesGetRequest, opts ...RequestOption) (*TemplateSummary, error) {
	return doValue[TemplateSummary](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/flows/templates/%s", templateID), query: req, opts: opts})
}

// FlowsTemplateRunsService groups the /v1/flows/templates/{id}/runs endpoints.
type FlowsTemplateRunsService struct{ cfg *config }

// TemplateOutput is one output of a template run. Which content field is
// set depends on Type: ContentURL/ContentMimeType for audio, image and
// video; Content (raw JSON) for scalar, array and object outputs.
type TemplateOutput struct {
	// Type is "array", "audio", "boolean", "image", "integer", "number", "object", "string" or "video".
	Type            string           `json:"type"`
	ID              string           `json:"id"`
	Status          GenerationStatus `json:"status"`
	FailureReason   *string          `json:"failure_reason,omitempty"`
	ErrorMessage    *string          `json:"error_message,omitempty"`
	ContentURL      *string          `json:"content_url,omitempty"`
	ContentMimeType *string          `json:"content_mime_type,omitempty"`
	Content         json.RawMessage  `json:"content,omitempty"`
	NextCursor      *string          `json:"next_cursor,omitempty"`
	HasMore         *bool            `json:"has_more,omitempty"`
}

// TemplateRun is a run of a template.
type TemplateRun struct {
	ID         string                    `json:"id"`
	TemplateID string                    `json:"template_id"`
	VersionID  string                    `json:"version_id"`
	Status     GenerationStatus          `json:"status"`
	Outputs    map[string]TemplateOutput `json:"outputs"`
}

// TemplateRunsListRequest holds the optional parameters of FlowsTemplateRunsService.List.
type TemplateRunsListRequest struct {
	Cursor    string `query:"cursor"`
	PageSize  *int   `query:"page_size"`
	VersionID string `query:"version_id"`
}

// TemplateRunPage is a page of template runs.
type TemplateRunPage struct {
	Runs       []TemplateRun `json:"runs"`
	NextCursor *string       `json:"next_cursor,omitempty"`
	HasMore    *bool         `json:"has_more,omitempty"`
}

// List returns one page of runs of a template. req may be nil.
func (s *FlowsTemplateRunsService) List(ctx context.Context, templateID string, req *TemplateRunsListRequest, opts ...RequestOption) (*TemplateRunPage, error) {
	return doValue[TemplateRunPage](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/flows/templates/%s/runs", templateID), query: req, opts: opts})
}

// ListAll iterates over every run of a template.
func (s *FlowsTemplateRunsService) ListAll(ctx context.Context, templateID string, req *TemplateRunsListRequest, opts ...RequestOption) iter.Seq2[TemplateRun, error] {
	params := *orEmpty(req)
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[TemplateRun], error) {
		params.Cursor = cursor
		resp, err := s.List(ctx, templateID, &params, opts...)
		if err != nil {
			return page[TemplateRun]{}, err
		}
		return page[TemplateRun]{items: resp.Runs, next: deref(resp.NextCursor), hasMore: resp.HasMore == nil || *resp.HasMore}, nil
	})
}

// TemplateRunCreateRequest is the body of FlowsTemplateRunsService.Create.
type TemplateRunCreateRequest struct {
	// Inputs maps input port IDs to values: strings, booleans, numbers,
	// MediaReference values or slices of these. Required.
	Inputs map[string]any `json:"inputs"`
	// VersionID pins a template version; defaults to the latest.
	VersionID string         `json:"version_id,omitempty"`
	Webhook   *WebhookTarget `json:"webhook,omitempty"`
}

func (s *FlowsTemplateRunsService) create(ctx context.Context, templateID string, req *TemplateRunCreateRequest, opts []RequestOption) (*TemplateRun, http.Header, error) {
	body := *orEmpty(req)
	if body.Inputs == nil {
		body.Inputs = map[string]any{}
	}
	return doValueHeader[TemplateRun](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/flows/templates/%s/runs", templateID), body: &body, opts: opts})
}

func (s *FlowsTemplateRunsService) get(ctx context.Context, templateID, runID string, opts []RequestOption) (*TemplateRun, http.Header, error) {
	return doValueHeader[TemplateRun](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/flows/templates/%s/runs/%s", templateID, runID), opts: opts})
}

// Create starts a run of a template.
func (s *FlowsTemplateRunsService) Create(ctx context.Context, templateID string, req *TemplateRunCreateRequest, opts ...RequestOption) (*TemplateRun, error) {
	r, _, err := s.create(ctx, templateID, req, opts)
	return r, err
}

// Get returns a template run.
func (s *FlowsTemplateRunsService) Get(ctx context.Context, templateID, runID string, opts ...RequestOption) (*TemplateRun, error) {
	r, _, err := s.get(ctx, templateID, runID, opts)
	return r, err
}

// CreateAndWait starts a template run and polls until its status is
// completed or failed. A failed run is returned, not reported as an error;
// each output also has its own status. Returns *FlowsWaitTimeoutError if
// w.WaitTimeout elapses first. w may be nil.
func (s *FlowsTemplateRunsService) CreateAndWait(ctx context.Context, templateID string, req *TemplateRunCreateRequest, w *FlowsWaitOptions, opts ...RequestOption) (*TemplateRun, error) {
	return waitTerminal(ctx, s.cfg, w, templateID,
		func() (*TemplateRun, string, GenerationStatus, http.Header, error) {
			r, h, err := s.create(ctx, templateID, req, opts)
			if err != nil {
				return nil, "", "", nil, err
			}
			return r, r.ID, r.Status, h, nil
		},
		func(id string) (*TemplateRun, GenerationStatus, http.Header, error) {
			r, h, err := s.get(ctx, templateID, id, opts)
			if err != nil {
				return nil, "", nil, err
			}
			return r, r.Status, h, nil
		})
}
