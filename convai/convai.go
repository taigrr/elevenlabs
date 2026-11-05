package convai

import (
	"github.com/taigrr/elevenlabs/v2/convai/agents"
	"github.com/taigrr/elevenlabs/v2/convai/agenttests"
	"github.com/taigrr/elevenlabs/v2/convai/conversations"
	"github.com/taigrr/elevenlabs/v2/convai/knowledgebase"
	"github.com/taigrr/elevenlabs/v2/convai/mcp"
	"github.com/taigrr/elevenlabs/v2/convai/telephony"
	"github.com/taigrr/elevenlabs/v2/convai/tools"
	"github.com/taigrr/elevenlabs/v2/convai/triage"
	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/option"
)

// Service groups the /v1/convai endpoints of the
// ElevenLabs Agents platform. Endpoints are split into sub-services that
// mirror the conversationalAi resources of the official SDK, e.g.
// client.ConvAI.Agents.List or
// client.ConvAI.KnowledgeBase.Documents.CreateFromFile.
//
// Agent configuration objects are large and evolve quickly. Frequently used
// fields are typed; deeply nested or polymorphic sections are exposed as
// json.RawMessage so they round-trip unchanged.
type Service struct {
	cfg *requestconfig.Config

	// Agents manages agents and their branches, drafts, deployments and more.
	Agents *agents.Service
	// Conversations reads and manages conversations.
	Conversations *conversations.Service
	// KnowledgeBase manages knowledge base documents, folders and crawls.
	KnowledgeBase *knowledgebase.Service
	// Tools manages workspace tools.
	Tools *tools.Service
	// Tests manages agent tests, test folders and test invocations.
	Tests *agenttests.Service
	// Settings reads and updates workspace conversational AI settings.
	Settings *SettingsService
	// Secrets manages workspace secrets.
	Secrets *SecretsService
	// Dashboard holds the dashboard settings endpoints.
	Dashboard *DashboardService
	// PhoneNumbers manages imported phone numbers.
	PhoneNumbers *telephony.PhoneNumbersService
	// Twilio places and registers Twilio calls.
	Twilio *telephony.TwilioService
	// SIPTrunk places outbound SIP trunk calls.
	SIPTrunk *telephony.SIPTrunkService
	// Exotel places outbound Exotel calls.
	Exotel *telephony.ExotelService
	// WhatsApp places WhatsApp calls and sends template messages.
	WhatsApp *telephony.WhatsAppService
	// WhatsAppAccounts manages connected WhatsApp accounts.
	WhatsAppAccounts *telephony.WhatsAppAccountsService
	// BatchCalls submits and manages batch calls.
	BatchCalls *telephony.BatchCallsService
	// MCPServers manages MCP servers, tool approvals and tool configs.
	MCPServers *mcp.Service
	// TriageTickets manages agent conversation triage tickets.
	TriageTickets *triage.Service
	// Analytics reports conversation analytics.
	Analytics *AnalyticsService
	// LLM lists available LLMs.
	LLM *LLMService
	// LLMUsage estimates LLM costs.
	LLMUsage *LLMUsageService
	// Realtime creates realtime client secrets.
	Realtime *RealtimeService
	// Users lists end users of conversations.
	Users *UsersService
}

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	cfg := requestconfig.Build(opts...)
	shared := func(c *requestconfig.Config) { *c = *cfg.Clone() }
	phone := telephony.NewService(shared)
	return &Service{
		cfg:              cfg,
		Agents:           agents.NewService(shared),
		Conversations:    conversations.NewService(shared),
		KnowledgeBase:    knowledgebase.NewService(shared),
		Tools:            tools.NewService(shared),
		Tests:            agenttests.NewService(shared),
		Settings:         &SettingsService{cfg: cfg},
		Secrets:          &SecretsService{cfg: cfg},
		Dashboard:        &DashboardService{Settings: &DashboardSettingsService{cfg: cfg}},
		PhoneNumbers:     phone.PhoneNumbers,
		Twilio:           phone.Twilio,
		SIPTrunk:         phone.SIPTrunk,
		Exotel:           phone.Exotel,
		WhatsApp:         phone.WhatsApp,
		WhatsAppAccounts: phone.WhatsAppAccounts,
		BatchCalls:       phone.BatchCalls,
		MCPServers:       mcp.NewService(shared),
		TriageTickets:    triage.NewService(shared),
		Analytics:        &AnalyticsService{LiveCount: &LiveCountService{cfg: cfg}},
		LLM:              &LLMService{cfg: cfg},
		LLMUsage:         &LLMUsageService{cfg: cfg},
		Realtime:         &RealtimeService{cfg: cfg},
		Users:            &UsersService{cfg: cfg},
	}
}
