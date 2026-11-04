package elevenlabs

import (
	"context"
	"math/rand/v2"
	"net/http"
	"os"
	"strings"
	"time"
)

// Environment is the base URL of an ElevenLabs API deployment.
type Environment string

// Environments mirror the ElevenLabsEnvironment values of the official SDK.
const (
	EnvironmentProduction          Environment = "https://api.elevenlabs.io"
	EnvironmentProductionUS        Environment = "https://api.us.elevenlabs.io"
	EnvironmentProductionEU        Environment = "https://api.eu.residency.elevenlabs.io"
	EnvironmentProductionIndia     Environment = "https://api.in.residency.elevenlabs.io"
	EnvironmentProductionSingapore Environment = "https://api.sg.residency.elevenlabs.io"
)

const (
	// DefaultUserAgent is sent as the User-Agent header unless overridden with WithUserAgent.
	DefaultUserAgent = "github.com/taigrr/elevenlabs/v2"
	// DefaultTimeout is the per-attempt request timeout, matching the official SDK.
	DefaultTimeout = 240 * time.Second
	// DefaultMaxRetries is the number of retries for retryable failures, matching the official SDK.
	DefaultMaxRetries = 2
	// APIKeyEnvVar is the environment variable consulted by NewClient for a default API key.
	APIKeyEnvVar = "ELEVENLABS_API_KEY"
)

// config holds the client-wide settings shared by every service.
type config struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	headers    http.Header
	timeout    time.Duration
	maxRetries int
	userAgent  string

	sleep  func(ctx context.Context, d time.Duration) error
	now    func() time.Time
	jitter func() float64
}

// Option configures a Client.
type Option func(*config)

// WithAPIKey sets the xi-api-key used to authenticate requests.
// It overrides the ELEVENLABS_API_KEY environment variable.
func WithAPIKey(apiKey string) Option {
	return func(c *config) { c.apiKey = apiKey }
}

// WithBaseURL points the client at a custom base URL (for example a proxy or a test server).
func WithBaseURL(baseURL string) Option {
	return func(c *config) { c.baseURL = strings.TrimRight(baseURL, "/") }
}

// WithEnvironment selects one of the predefined ElevenLabs deployments, such as a data residency region.
func WithEnvironment(env Environment) Option {
	return WithBaseURL(string(env))
}

// WithHTTPClient sets the *http.Client used to send requests.
// A nil client restores the default.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *config) {
		if hc == nil {
			hc = &http.Client{}
		}
		c.httpClient = hc
	}
}

// WithTimeout sets the default per-attempt timeout. Zero or negative disables the timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *config) { c.timeout = d }
}

// WithMaxRetries sets the default number of retries for retryable responses
// (408, 409, 429 and 5xx). Negative values are treated as zero.
func WithMaxRetries(n int) Option {
	return func(c *config) { c.maxRetries = max(n, 0) }
}

// WithHeader adds a header sent with every request.
func WithHeader(key, value string) Option {
	return func(c *config) { c.headers.Add(key, value) }
}

// WithUserAgent overrides the User-Agent header.
func WithUserAgent(ua string) Option {
	return func(c *config) { c.userAgent = ua }
}

// Client is the entry point to the ElevenLabs API. Endpoints are grouped into
// services mirroring the resources of the official TypeScript SDK, e.g.
// client.TextToSpeech.Convert or client.Voices.Settings.Get.
//
// A Client is safe for concurrent use.
type Client struct {
	cfg *config

	Assets                    *AssetsService
	AudioIsolation            *AudioIsolationService
	AudioNative               *AudioNativeService
	ForcedAlignment           *ForcedAlignmentService
	History                   *HistoryService
	Models                    *ModelsService
	Music                     *MusicService
	PronunciationDictionaries *PronunciationDictionariesService
	Samples                   *SamplesService
	SpeechToSpeech            *SpeechToSpeechService
	SpeechToText              *SpeechToTextService
	TextToDialogue            *TextToDialogueService
	TextToSoundEffects        *TextToSoundEffectsService
	TextToSpeech              *TextToSpeechService
	TextToVoice               *TextToVoiceService
	Tokens                    *TokensService
	// Translation holds the realtime translation WebSocket.
	Translation *TranslationService
	Usage       *UsageService
	User        *UserService
	Voices      *VoicesService
	// Webhooks manages workspace webhooks and verifies webhook deliveries.
	Webhooks *WebhooksService
	// ServiceAccounts manages service accounts and their API keys.
	ServiceAccounts *ServiceAccountsService
	// EnvironmentVariables manages conversational AI environment variables.
	EnvironmentVariables *EnvironmentVariablesService
	// Environments manages conversational AI environments.
	Environments *EnvironmentsService
	// SpeechEngine manages speech engines.
	SpeechEngine *SpeechEngineService
	// Workspace manages the current workspace.
	Workspace *WorkspaceService
	// Workspaces holds cross-workspace endpoints.
	Workspaces *WorkspacesService
	// Flows runs asynchronous image, video and speech generations and templates.
	Flows *FlowsService
	// Productions orders human-reviewed dubbing, subtitles and transcription.
	Productions *ProductionsService
	// Dubbing dubs audio and video and edits dubbing projects and resources.
	Dubbing *DubbingService
	// Studio manages Studio projects, chapters, snapshots and podcasts.
	Studio *StudioService
	// ConversationalAI manages agents, conversations, knowledge bases, tools,
	// tests, secrets and workspace settings.
	ConversationalAI *ConversationalAIService
}

// NewClient returns a Client configured by opts. Without WithAPIKey the key
// is read from the ELEVENLABS_API_KEY environment variable.
func NewClient(opts ...Option) *Client {
	cfg := &config{
		baseURL:    string(EnvironmentProduction),
		apiKey:     os.Getenv(APIKeyEnvVar),
		httpClient: &http.Client{},
		headers:    make(http.Header),
		timeout:    DefaultTimeout,
		maxRetries: DefaultMaxRetries,
		userAgent:  DefaultUserAgent,
		sleep:      sleepContext,
		now:        time.Now,
		jitter:     rand.Float64,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return newClient(cfg)
}

func newClient(cfg *config) *Client {
	return &Client{
		cfg:                       cfg,
		Assets:                    &AssetsService{cfg: cfg},
		AudioIsolation:            &AudioIsolationService{cfg: cfg},
		AudioNative:               &AudioNativeService{cfg: cfg},
		ForcedAlignment:           &ForcedAlignmentService{cfg: cfg},
		History:                   &HistoryService{cfg: cfg},
		Models:                    &ModelsService{cfg: cfg},
		Music:                     newMusicService(cfg),
		PronunciationDictionaries: newPronunciationDictionariesService(cfg),
		Samples:                   &SamplesService{cfg: cfg},
		SpeechToSpeech:            &SpeechToSpeechService{cfg: cfg},
		SpeechToText:              newSpeechToTextService(cfg),
		TextToDialogue:            &TextToDialogueService{cfg: cfg},
		TextToSoundEffects:        &TextToSoundEffectsService{cfg: cfg},
		TextToSpeech:              &TextToSpeechService{cfg: cfg},
		TextToVoice:               newTextToVoiceService(cfg),
		Tokens:                    newTokensService(cfg),
		Translation:               &TranslationService{cfg: cfg},
		Usage:                     &UsageService{cfg: cfg},
		User:                      newUserService(cfg),
		Voices:                    newVoicesService(cfg),
		Webhooks:                  &WebhooksService{cfg: cfg},
		ServiceAccounts:           newServiceAccountsService(cfg),
		EnvironmentVariables:      &EnvironmentVariablesService{cfg: cfg},
		Environments:              &EnvironmentsService{cfg: cfg},
		SpeechEngine:              &SpeechEngineService{cfg: cfg},
		Workspace:                 newWorkspaceService(cfg),
		Flows:                     newFlowsService(cfg),
		Studio:                    newStudioService(cfg),
		Dubbing:                   newDubbingService(cfg),
		Productions:               newProductionsService(cfg),
		ConversationalAI:          newConversationalAIService(cfg),
		Workspaces:                &WorkspacesService{APIKeys: &WorkspacesAPIKeysService{cfg: cfg}},
	}
}

// BaseURL returns the base URL requests are sent to.
func (c *Client) BaseURL() string { return c.cfg.baseURL }

func sleepContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// Ptr returns a pointer to v. It is a convenience for populating optional
// request fields, e.g. Seed: elevenlabs.Ptr(42).
func Ptr[T any](v T) *T { return &v }
