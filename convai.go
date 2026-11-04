package elevenlabs

import (
	"context"
	"encoding/json"
	"net/http"
)

// ConversationalAIService groups the /v1/convai endpoints of the
// ElevenLabs Agents platform. Endpoints are split into sub-services that
// mirror the conversationalAi resources of the official SDK, e.g.
// client.ConversationalAI.Agents.List or
// client.ConversationalAI.KnowledgeBase.Documents.CreateFromFile.
//
// Agent configuration objects are large and evolve quickly. Frequently used
// fields are typed; deeply nested or polymorphic sections are exposed as
// json.RawMessage so they round-trip unchanged.
type ConversationalAIService struct {
	cfg *config

	// Agents manages agents and their branches, drafts, deployments and more.
	Agents *AgentsService
}

func newConversationalAIService(cfg *config) *ConversationalAIService {
	return &ConversationalAIService{
		cfg:    cfg,
		Agents: newAgentsService(cfg),
	}
}

// LLM identifies a language model usable by agents, e.g. "gpt-4o" or
// "gemini-2.5-flash". Any model name accepted by the API may be used.
type LLM string

// Common LLM values.
const (
	LLMGPT4oMini       LLM = "gpt-4o-mini"
	LLMGPT4o           LLM = "gpt-4o"
	LLMGPT41           LLM = "gpt-4.1"
	LLMGPT41Mini       LLM = "gpt-4.1-mini"
	LLMGPT5            LLM = "gpt-5"
	LLMGPT5Mini        LLM = "gpt-5-mini"
	LLMGemini20Flash   LLM = "gemini-2.0-flash"
	LLMGemini25Flash   LLM = "gemini-2.5-flash"
	LLMGemini25FlashLt LLM = "gemini-2.5-flash-lite"
	LLMClaudeSonnet45  LLM = "claude-sonnet-4-5"
	LLMClaudeHaiku45   LLM = "claude-haiku-4-5"
	LLMCustom          LLM = "custom-llm"
)

// LLMServiceTier selects the LLM serving tier.
type LLMServiceTier string

// LLMServiceTier values.
const (
	LLMServiceTierStandard LLMServiceTier = "standard"
	LLMServiceTierFast     LLMServiceTier = "fast"
)

// ConversationalConfig is an agent's conversation configuration.
type ConversationalConfig struct {
	ASR          *ASRConfig          `json:"asr,omitempty"`
	Turn         *TurnConfig         `json:"turn,omitempty"`
	TTS          *AgentTTSConfig     `json:"tts,omitempty"`
	Conversation *ConversationConfig `json:"conversation,omitempty"`
	// LanguagePresets maps language codes to per-language overrides.
	LanguagePresets map[string]json.RawMessage `json:"language_presets,omitempty"`
	VAD             json.RawMessage            `json:"vad,omitempty"`
	Agent           *AgentConfig               `json:"agent,omitempty"`
}

// ASRConfig configures speech recognition.
type ASRConfig struct {
	// Quality is "high".
	Quality string `json:"quality,omitempty"`
	// Provider is "elevenlabs", "scribe_realtime" or "scribe_v2_turbo".
	Provider string `json:"provider,omitempty"`
	// UserInputAudioFormat is e.g. "pcm_16000" or "ulaw_8000".
	UserInputAudioFormat string   `json:"user_input_audio_format,omitempty"`
	Keywords             []string `json:"keywords,omitempty"`
}

// TurnConfig configures turn taking.
type TurnConfig struct {
	TurnTimeout           *float64 `json:"turn_timeout,omitempty"`
	InitialWaitTime       *float64 `json:"initial_wait_time,omitempty"`
	SilenceEndCallTimeout *float64 `json:"silence_end_call_timeout,omitempty"`
	// TurnEagerness is "patient", "normal" or "eager".
	TurnEagerness string `json:"turn_eagerness,omitempty"`
	// SpellingPatience is "auto" or "off".
	SpellingPatience          string `json:"spelling_patience,omitempty"`
	SpeculativeTurn           *bool  `json:"speculative_turn,omitempty"`
	RetranscribeOnTurnTimeout *bool  `json:"retranscribe_on_turn_timeout,omitempty"`
	// TurnModel is "turn_v2" or "turn_v3".
	TurnModel                         string          `json:"turn_model,omitempty"`
	InterruptionIgnoreTerms           []string        `json:"interruption_ignore_terms,omitempty"`
	InterruptionIgnoreTermLanguages   []string        `json:"interruption_ignore_term_languages,omitempty"`
	MergeWithDefaultIgnoreTerms       *bool           `json:"merge_with_default_ignore_terms,omitempty"`
	TranscribeOnDisabledInterruptions *bool           `json:"transcribe_on_disabled_interruptions,omitempty"`
	SoftTimeoutConfig                 json.RawMessage `json:"soft_timeout_config,omitempty"`
}

// AgentTTSConfig configures an agent's voice.
type AgentTTSConfig struct {
	// ModelID is e.g. "eleven_flash_v2_5" or "eleven_v3_conversational".
	ModelID         string          `json:"model_id,omitempty"`
	VoiceID         string          `json:"voice_id,omitempty"`
	SupportedVoices json.RawMessage `json:"supported_voices,omitempty"`
	ExpressiveMode  *bool           `json:"expressive_mode,omitempty"`
	// AgentOutputAudioFormat is e.g. "pcm_16000" or "ulaw_8000".
	AgentOutputAudioFormat          string                           `json:"agent_output_audio_format,omitempty"`
	OptimizeStreamingLatency        *int                             `json:"optimize_streaming_latency,omitempty"`
	Stability                       *float64                         `json:"stability,omitempty"`
	Speed                           *float64                         `json:"speed,omitempty"`
	SimilarityBoost                 *float64                         `json:"similarity_boost,omitempty"`
	TextNormalisationType           string                           `json:"text_normalisation_type,omitempty"`
	PronunciationDictionaryLocators []PronunciationDictionaryLocator `json:"pronunciation_dictionary_locators,omitempty"`
	EnablePhonemeTags               *bool                            `json:"enable_phoneme_tags,omitempty"`
	SuggestedAudioTags              json.RawMessage                  `json:"suggested_audio_tags,omitempty"`
	AudioEffects                    json.RawMessage                  `json:"audio_effects,omitempty"`
}

// ConversationConfig holds conversation-level settings.
type ConversationConfig struct {
	TextOnly           *bool           `json:"text_only,omitempty"`
	MaxDurationSeconds *int            `json:"max_duration_seconds,omitempty"`
	ClientEvents       []string        `json:"client_events,omitempty"`
	MonitoringEnabled  *bool           `json:"monitoring_enabled,omitempty"`
	MonitoringEvents   []string        `json:"monitoring_events,omitempty"`
	SourceAttribution  *bool           `json:"source_attribution,omitempty"`
	FileInput          json.RawMessage `json:"file_input,omitempty"`
	DTMFInputSettings  json.RawMessage `json:"dtmf_input_settings,omitempty"`
	BackgroundSound    json.RawMessage `json:"background_sound,omitempty"`
}

// AgentConfig configures the agent's behaviour.
type AgentConfig struct {
	FirstMessage                     string                     `json:"first_message,omitempty"`
	Language                         string                     `json:"language,omitempty"`
	HinglishMode                     *bool                      `json:"hinglish_mode,omitempty"`
	DynamicVariables                 json.RawMessage            `json:"dynamic_variables,omitempty"`
	DisableFirstMessageInterruptions *bool                      `json:"disable_first_message_interruptions,omitempty"`
	MaxConversationDurationMessage   string                     `json:"max_conversation_duration_message,omitempty"`
	TextBehaviorOverrides            map[string]json.RawMessage `json:"text_behavior_overrides,omitempty"`
	Prompt                           *AgentPrompt               `json:"prompt,omitempty"`
}

// AgentPrompt configures the agent's LLM.
type AgentPrompt struct {
	Prompt string `json:"prompt,omitempty"`
	LLM    LLM    `json:"llm,omitempty"`
	// ReasoningEffort is "none", "minimal", "low", "medium", "high", "xhigh" or "max".
	ReasoningEffort          string                 `json:"reasoning_effort,omitempty"`
	ServiceTier              LLMServiceTier         `json:"service_tier,omitempty"`
	ThinkingBudget           *int                   `json:"thinking_budget,omitempty"`
	EnableReasoningSummary   *bool                  `json:"enable_reasoning_summary,omitempty"`
	Temperature              *float64               `json:"temperature,omitempty"`
	MaxTokens                *int                   `json:"max_tokens,omitempty"`
	ToolIDs                  []string               `json:"tool_ids,omitempty"`
	ToolIDsStrict            *bool                  `json:"tool_ids_strict,omitempty"`
	BuiltInTools             json.RawMessage        `json:"built_in_tools,omitempty"`
	EnableParallelToolCalls  *bool                  `json:"enable_parallel_tool_calls,omitempty"`
	MCPServerIDs             []string               `json:"mcp_server_ids,omitempty"`
	NativeMCPServerIDs       []string               `json:"native_mcp_server_ids,omitempty"`
	KnowledgeBase            []KnowledgeBaseLocator `json:"knowledge_base,omitempty"`
	CustomLLM                json.RawMessage        `json:"custom_llm,omitempty"`
	IgnoreDefaultPersonality *bool                  `json:"ignore_default_personality,omitempty"`
	RAG                      json.RawMessage        `json:"rag,omitempty"`
	Timezone                 string                 `json:"timezone,omitempty"`
	BackupLLMConfig          json.RawMessage        `json:"backup_llm_config,omitempty"`
	CascadeTimeoutSeconds    *float64               `json:"cascade_timeout_seconds,omitempty"`
	// Tools are inline tool definitions.
	Tools json.RawMessage `json:"tools,omitempty"`
}

// KnowledgeBaseDocumentType is the kind of a knowledge base entry.
type KnowledgeBaseDocumentType string

// KnowledgeBaseDocumentType values.
const (
	KnowledgeBaseDocumentTypeFile   KnowledgeBaseDocumentType = "file"
	KnowledgeBaseDocumentTypeURL    KnowledgeBaseDocumentType = "url"
	KnowledgeBaseDocumentTypeText   KnowledgeBaseDocumentType = "text"
	KnowledgeBaseDocumentTypeFolder KnowledgeBaseDocumentType = "folder"
)

// KnowledgeBaseLocator attaches a knowledge base document to an agent.
type KnowledgeBaseLocator struct {
	Type KnowledgeBaseDocumentType `json:"type"`
	Name string                    `json:"name"`
	ID   string                    `json:"id"`
	// UsageMode is "prompt" or "auto".
	UsageMode string `json:"usage_mode,omitempty"`
}

// AgentPlatformSettings holds an agent's platform settings. Nested sections
// are kept as raw JSON.
type AgentPlatformSettings struct {
	Evaluation                           json.RawMessage `json:"evaluation,omitempty"`
	Widget                               json.RawMessage `json:"widget,omitempty"`
	DataCollection                       json.RawMessage `json:"data_collection,omitempty"`
	DataCollectionScopes                 json.RawMessage `json:"data_collection_scopes,omitempty"`
	AnalysisItems                        json.RawMessage `json:"analysis_items,omitempty"`
	Overrides                            json.RawMessage `json:"overrides,omitempty"`
	WorkspaceOverrides                   json.RawMessage `json:"workspace_overrides,omitempty"`
	Testing                              json.RawMessage `json:"testing,omitempty"`
	Archived                             *bool           `json:"archived,omitempty"`
	Guardrails                           json.RawMessage `json:"guardrails,omitempty"`
	SummaryLanguage                      string          `json:"summary_language,omitempty"`
	AutoTranslateTranscriptToAppLanguage *bool           `json:"auto_translate_transcript_to_app_language,omitempty"`
	Auth                                 json.RawMessage `json:"auth,omitempty"`
	CallLimits                           json.RawMessage `json:"call_limits,omitempty"`
	QueueingConfig                       json.RawMessage `json:"queueing_config,omitempty"`
	Privacy                              json.RawMessage `json:"privacy,omitempty"`
	TrustContext                         json.RawMessage `json:"trust_context,omitempty"`
	AnalysisLLM                          LLM             `json:"analysis_llm,omitempty"`
	TopicDiscovery                       json.RawMessage `json:"topic_discovery,omitempty"`
	SentimentAnalysis                    json.RawMessage `json:"sentiment_analysis,omitempty"`
	Alerting                             json.RawMessage `json:"alerting,omitempty"`
	// Safety is only present in responses.
	Safety json.RawMessage `json:"safety,omitempty"`
}

// AgentWorkflow is an agent's workflow graph. Nodes and edges are kept as
// raw JSON keyed by ID.
type AgentWorkflow struct {
	Edges                map[string]json.RawMessage `json:"edges,omitempty"`
	Nodes                map[string]json.RawMessage `json:"nodes,omitempty"`
	PreventSubagentLoops *bool                      `json:"prevent_subagent_loops,omitempty"`
}

// ListMeta is the paging metadata of page-numbered list responses.
type ListMeta struct {
	Total    *int `json:"total,omitempty"`
	Page     *int `json:"page,omitempty"`
	PageSize *int `json:"page_size,omitempty"`
}

// PageRequest holds cursor paging parameters shared by many list endpoints.
type PageRequest struct {
	PageSize *int   `query:"page_size"`
	Cursor   string `query:"cursor"`
}

// ForceRequest holds the optional force flag of delete endpoints.
type ForceRequest struct {
	Force *bool `query:"force"`
}

func doNoContent(ctx context.Context, cfg *config, method, path string, query, body any, opts []RequestOption) error {
	return doJSON(ctx, cfg, &request{method: method, path: path, query: query, body: body, opts: opts}, nil)
}

func getValue[T any](ctx context.Context, cfg *config, path string, query any, opts []RequestOption) (*T, error) {
	return doValue[T](ctx, cfg, &request{method: http.MethodGet, path: path, query: query, opts: opts})
}

func sendValue[T any](ctx context.Context, cfg *config, method, path string, query, body any, opts []RequestOption) (*T, error) {
	return doValue[T](ctx, cfg, &request{method: method, path: path, query: query, body: body, opts: opts})
}

func sendRaw(ctx context.Context, cfg *config, method, path string, query, body any, opts []RequestOption) (json.RawMessage, error) {
	return doRaw(ctx, cfg, &request{method: method, path: path, query: query, body: body, opts: opts})
}
