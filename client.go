package elevenlabs

import (
	"github.com/taigrr/elevenlabs/v2/assets"
	"github.com/taigrr/elevenlabs/v2/audioeffects"
	"github.com/taigrr/elevenlabs/v2/audioisolation"
	"github.com/taigrr/elevenlabs/v2/audionative"
	"github.com/taigrr/elevenlabs/v2/forcedalignment"
	"github.com/taigrr/elevenlabs/v2/history"
	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/models"
	"github.com/taigrr/elevenlabs/v2/music"
	"github.com/taigrr/elevenlabs/v2/pronunciation"
	"github.com/taigrr/elevenlabs/v2/samples"
	"github.com/taigrr/elevenlabs/v2/soundeffects"
	"github.com/taigrr/elevenlabs/v2/speechtospeech"
	"github.com/taigrr/elevenlabs/v2/speechtotext"
	"github.com/taigrr/elevenlabs/v2/texttodialogue"
	"github.com/taigrr/elevenlabs/v2/texttospeech"
	"github.com/taigrr/elevenlabs/v2/texttovoice"
	"github.com/taigrr/elevenlabs/v2/tokens"
	"github.com/taigrr/elevenlabs/v2/usage"
	"github.com/taigrr/elevenlabs/v2/user"
	"github.com/taigrr/elevenlabs/v2/voices"
)

// Client is the entry point to the ElevenLabs API. Endpoints are grouped into
// services mirroring the resources of the official TypeScript SDK, e.g.
// client.TextToSpeech.Convert or client.Voices.Settings.Get.
//
// A Client is safe for concurrent use.
type Client struct {
	cfg *config

	Assets *assets.Service
	// AudioEffects lists the presets usable in audio_effects.
	AudioEffects              *audioeffects.Service
	AudioIsolation            *audioisolation.Service
	AudioNative               *audionative.Service
	ForcedAlignment           *forcedalignment.Service
	History                   *history.Service
	Models                    *models.Service
	Music                     *music.Service
	PronunciationDictionaries *pronunciation.Service
	Samples                   *samples.Service
	SpeechToSpeech            *speechtospeech.Service
	SpeechToText              *speechtotext.Service
	TextToDialogue            *texttodialogue.Service
	TextToSoundEffects        *soundeffects.Service
	TextToSpeech              *texttospeech.Service
	TextToVoice               *texttovoice.Service
	Tokens                    *tokens.Service
	// Translation holds the realtime translation WebSocket.
	Translation *TranslationService
	Usage       *usage.Service
	User        *user.Service
	Voices      *voices.Service
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
	return newClient(requestconfig.Build(opts...))
}

func newClient(cfg *config) *Client {
	return &Client{
		cfg:                       cfg,
		Assets:                    assets.NewService(fromConfig(cfg)),
		AudioEffects:              audioeffects.NewService(fromConfig(cfg)),
		AudioIsolation:            audioisolation.NewService(fromConfig(cfg)),
		AudioNative:               audionative.NewService(fromConfig(cfg)),
		ForcedAlignment:           forcedalignment.NewService(fromConfig(cfg)),
		History:                   history.NewService(fromConfig(cfg)),
		Models:                    models.NewService(fromConfig(cfg)),
		Music:                     music.NewService(fromConfig(cfg)),
		PronunciationDictionaries: pronunciation.NewService(fromConfig(cfg)),
		Samples:                   samples.NewService(fromConfig(cfg)),
		SpeechToSpeech:            speechtospeech.NewService(fromConfig(cfg)),
		SpeechToText:              speechtotext.NewService(fromConfig(cfg)),
		TextToDialogue:            texttodialogue.NewService(fromConfig(cfg)),
		TextToSoundEffects:        soundeffects.NewService(fromConfig(cfg)),
		TextToSpeech:              texttospeech.NewService(fromConfig(cfg)),
		TextToVoice:               texttovoice.NewService(fromConfig(cfg)),
		Tokens:                    tokens.NewService(fromConfig(cfg)),
		Translation:               &TranslationService{cfg: cfg},
		Usage:                     usage.NewService(fromConfig(cfg)),
		User:                      user.NewService(fromConfig(cfg)),
		Voices:                    voices.NewService(fromConfig(cfg)),
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
func (c *Client) BaseURL() string { return c.cfg.BaseURL }

// fromConfig makes a service share the client's configuration.
func fromConfig(cfg *config) Option {
	return func(c *requestconfig.Config) { *c = *cfg.Clone() }
}
