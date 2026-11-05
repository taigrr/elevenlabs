package elevenlabs

import (
	"github.com/taigrr/elevenlabs/v2/assets"
	"github.com/taigrr/elevenlabs/v2/audioeffects"
	"github.com/taigrr/elevenlabs/v2/audioisolation"
	"github.com/taigrr/elevenlabs/v2/audionative"
	"github.com/taigrr/elevenlabs/v2/dubbing"
	"github.com/taigrr/elevenlabs/v2/environments"
	"github.com/taigrr/elevenlabs/v2/envvars"
	"github.com/taigrr/elevenlabs/v2/flows"
	"github.com/taigrr/elevenlabs/v2/forcedalignment"
	"github.com/taigrr/elevenlabs/v2/history"
	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/models"
	"github.com/taigrr/elevenlabs/v2/music"
	"github.com/taigrr/elevenlabs/v2/productions"
	"github.com/taigrr/elevenlabs/v2/pronunciation"
	"github.com/taigrr/elevenlabs/v2/samples"
	"github.com/taigrr/elevenlabs/v2/serviceaccounts"
	"github.com/taigrr/elevenlabs/v2/soundeffects"
	"github.com/taigrr/elevenlabs/v2/speechengine"
	"github.com/taigrr/elevenlabs/v2/speechtospeech"
	"github.com/taigrr/elevenlabs/v2/speechtotext"
	"github.com/taigrr/elevenlabs/v2/studio"
	"github.com/taigrr/elevenlabs/v2/texttodialogue"
	"github.com/taigrr/elevenlabs/v2/texttospeech"
	"github.com/taigrr/elevenlabs/v2/texttovoice"
	"github.com/taigrr/elevenlabs/v2/tokens"
	"github.com/taigrr/elevenlabs/v2/translation"
	"github.com/taigrr/elevenlabs/v2/usage"
	"github.com/taigrr/elevenlabs/v2/user"
	"github.com/taigrr/elevenlabs/v2/voices"
	"github.com/taigrr/elevenlabs/v2/webhooks"
	"github.com/taigrr/elevenlabs/v2/workspace"
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
	Translation *translation.Service
	Usage       *usage.Service
	User        *user.Service
	Voices      *voices.Service
	// Webhooks manages workspace webhooks and verifies webhook deliveries.
	Webhooks *webhooks.Service
	// ServiceAccounts manages service accounts and their API keys.
	ServiceAccounts *serviceaccounts.Service
	// EnvironmentVariables manages conversational AI environment variables.
	EnvironmentVariables *envvars.Service
	// Environments manages conversational AI environments.
	Environments *environments.Service
	// SpeechEngine manages speech engines.
	SpeechEngine *speechengine.Service
	// Workspace manages the current workspace, including its auth connections
	// and API keys.
	Workspace *workspace.Service
	// Flows runs asynchronous image, video and speech generations and templates.
	Flows *flows.Service
	// Productions orders human-reviewed dubbing, subtitles and transcription.
	Productions *productions.Service
	// Dubbing dubs audio and video and edits dubbing projects and resources.
	Dubbing *dubbing.Service
	// Studio manages Studio projects, chapters, snapshots and podcasts.
	Studio *studio.Service
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
		Translation:               translation.NewService(fromConfig(cfg)),
		Usage:                     usage.NewService(fromConfig(cfg)),
		User:                      user.NewService(fromConfig(cfg)),
		Voices:                    voices.NewService(fromConfig(cfg)),
		Webhooks:                  webhooks.NewService(fromConfig(cfg)),
		ServiceAccounts:           serviceaccounts.NewService(fromConfig(cfg)),
		EnvironmentVariables:      envvars.NewService(fromConfig(cfg)),
		Environments:              environments.NewService(fromConfig(cfg)),
		SpeechEngine:              speechengine.NewService(fromConfig(cfg)),
		Workspace:                 workspace.NewService(fromConfig(cfg)),
		Flows:                     flows.NewService(fromConfig(cfg)),
		Studio:                    studio.NewService(fromConfig(cfg)),
		Dubbing:                   dubbing.NewService(fromConfig(cfg)),
		Productions:               productions.NewService(fromConfig(cfg)),
		ConversationalAI:          newConversationalAIService(cfg),
	}
}

// BaseURL returns the base URL requests are sent to.
func (c *Client) BaseURL() string { return c.cfg.BaseURL }

// fromConfig makes a service share the client's configuration.
func fromConfig(cfg *config) Option {
	return func(c *requestconfig.Config) { *c = *cfg.Clone() }
}
