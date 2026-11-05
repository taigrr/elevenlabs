package elevenlabs

import (
	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
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
	AudioEffects              *AudioEffectsService
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
	return newClient(requestconfig.Build(opts...))
}

func newClient(cfg *config) *Client {
	return &Client{
		cfg:                       cfg,
		Assets:                    &AssetsService{cfg: cfg},
		AudioEffects:              &AudioEffectsService{cfg: cfg},
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
func (c *Client) BaseURL() string { return c.cfg.BaseURL }
