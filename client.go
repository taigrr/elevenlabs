package elevenlabs

import (
	"github.com/taigrr/elevenlabs/v2/audioeffects"
	"github.com/taigrr/elevenlabs/v2/audioisolation"
	"github.com/taigrr/elevenlabs/v2/forcedalignment"
	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/models"
	"github.com/taigrr/elevenlabs/v2/music"
	"github.com/taigrr/elevenlabs/v2/soundeffects"
	"github.com/taigrr/elevenlabs/v2/speechtospeech"
	"github.com/taigrr/elevenlabs/v2/speechtotext"
	"github.com/taigrr/elevenlabs/v2/texttodialogue"
	"github.com/taigrr/elevenlabs/v2/texttospeech"
	"github.com/taigrr/elevenlabs/v2/texttovoice"
	"github.com/taigrr/elevenlabs/v2/tokens"
	"github.com/taigrr/elevenlabs/v2/usage"
	"github.com/taigrr/elevenlabs/v2/user"
)

// Client is the entry point to the ElevenLabs API. Endpoints are grouped into
// services mirroring the resources of the official TypeScript SDK, e.g.
// client.TextToSpeech.Convert or client.Voices.Settings.Get.
//
// A Client is safe for concurrent use.
type Client struct {
	cfg *config

	Assets *AssetsService
	// AudioEffects lists the presets usable in audio_effects.
	AudioEffects              *audioeffects.Service
	AudioIsolation            *audioisolation.Service
	AudioNative               *AudioNativeService
	ForcedAlignment           *forcedalignment.Service
	History                   *HistoryService
	Models                    *models.Service
	Music                     *music.Service
	PronunciationDictionaries *PronunciationDictionariesService
	Samples                   *SamplesService
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
	return newClient(requestconfig.Build(opts...))
}

func newClient(cfg *config) *Client {
	return &Client{
		cfg:                       cfg,
		Assets:                    &AssetsService{cfg: cfg},
		AudioEffects:              audioeffects.NewService(fromConfig(cfg)),
		AudioIsolation:            audioisolation.NewService(fromConfig(cfg)),
		AudioNative:               &AudioNativeService{cfg: cfg},
		ForcedAlignment:           forcedalignment.NewService(fromConfig(cfg)),
		History:                   &HistoryService{cfg: cfg},
		Models:                    models.NewService(fromConfig(cfg)),
		Music:                     music.NewService(fromConfig(cfg)),
		PronunciationDictionaries: newPronunciationDictionariesService(cfg),
		Samples:                   &SamplesService{cfg: cfg},
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
func (c *Client) BaseURL() string { return c.cfg.BaseURL }

// fromConfig makes a service share the client's configuration.
func fromConfig(cfg *config) Option {
	return func(c *requestconfig.Config) { *c = *cfg.Clone() }
}
