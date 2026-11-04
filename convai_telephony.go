package elevenlabs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
)

// TelephonyProvider identifies a phone number's telephony provider.
type TelephonyProvider string

// TelephonyProvider values.
const (
	TelephonyProviderTwilio   TelephonyProvider = "twilio"
	TelephonyProviderSIPTrunk TelephonyProvider = "sip_trunk"
	TelephonyProviderExotel   TelephonyProvider = "exotel"
)

// TelephonyDirection is the direction of a call.
type TelephonyDirection string

// TelephonyDirection values.
const (
	TelephonyDirectionInbound  TelephonyDirection = "inbound"
	TelephonyDirectionOutbound TelephonyDirection = "outbound"
)

// LivekitStackType selects the LiveKit stack used for SIP numbers.
type LivekitStackType string

// LivekitStackType values.
const (
	LivekitStackStandard LivekitStackType = "standard"
	LivekitStackStatic   LivekitStackType = "static"
)

// ConversationInitiationClientData customizes a conversation started by an
// outbound call or batch call.
type ConversationInitiationClientData struct {
	ConversationConfigOverride json.RawMessage `json:"conversation_config_override,omitempty"`
	CustomLLMExtraBody         map[string]any  `json:"custom_llm_extra_body,omitempty"`
	// UserID identifies the end user participating in the conversation.
	UserID                 string          `json:"user_id,omitempty"`
	SourceInfo             json.RawMessage `json:"source_info,omitempty"`
	BranchID               string          `json:"branch_id,omitempty"`
	Environment            string          `json:"environment,omitempty"`
	StartingWorkflowNodeID string          `json:"starting_workflow_node_id,omitempty"`
	ProcedureIDs           []string        `json:"procedure_ids,omitempty"`
	StartProcedureIDs      []string        `json:"start_procedure_ids,omitempty"`
	DynamicVariables       map[string]any  `json:"dynamic_variables,omitempty"`
}

// TelephonyCallConfig configures call-level telephony behaviour.
type TelephonyCallConfig struct {
	RingingTimeoutSecs         *int  `json:"ringing_timeout_secs,omitempty"`
	TwilioCallRecordingEnabled *bool `json:"twilio_call_recording_enabled,omitempty"`
	// TwilioMachineDetection configures Twilio answering machine detection (raw JSON).
	TwilioMachineDetection json.RawMessage `json:"twilio_machine_detection,omitempty"`
}

// OutboundCallRequest starts an outbound call through Twilio, a SIP trunk
// or Exotel.
type OutboundCallRequest struct {
	AgentID                          string                            `json:"agent_id"`
	AgentPhoneNumberID               string                            `json:"agent_phone_number_id"`
	ToNumber                         string                            `json:"to_number"`
	ConversationInitiationClientData *ConversationInitiationClientData `json:"conversation_initiation_client_data,omitempty"`
	TelephonyCallConfig              *TelephonyCallConfig              `json:"telephony_call_config,omitempty"`
}

// TwilioOutboundCallRequest starts an outbound Twilio call.
type TwilioOutboundCallRequest struct {
	OutboundCallRequest
	// CallRecordingEnabled lets Twilio record the call.
	CallRecordingEnabled *bool `json:"call_recording_enabled,omitempty"`
}

// OutboundCallResponse is the result of starting an outbound call. CallSID
// is set for Twilio and Exotel, SIPCallID for SIP trunks.
type OutboundCallResponse struct {
	Success        bool    `json:"success"`
	Message        string  `json:"message"`
	ConversationID *string `json:"conversation_id,omitempty"`
	CallSID        *string `json:"callSid,omitempty"`
	SIPCallID      *string `json:"sip_call_id,omitempty"`
}

// PhoneNumberAgentInfo is the agent assigned to a phone number.
type PhoneNumberAgentInfo struct {
	AgentID     string `json:"agent_id"`
	AgentName   string `json:"agent_name"`
	Environment string `json:"environment,omitempty"`
	BranchID    string `json:"branch_id,omitempty"`
}

// PhoneNumber is a phone number registered with the platform. SIP-specific
// fields are only set when Provider is TelephonyProviderSIPTrunk.
type PhoneNumber struct {
	Provider         TelephonyProvider     `json:"provider"`
	PhoneNumber      string                `json:"phone_number"`
	Label            string                `json:"label"`
	SupportsInbound  *bool                 `json:"supports_inbound,omitempty"`
	SupportsOutbound *bool                 `json:"supports_outbound,omitempty"`
	PhoneNumberID    string                `json:"phone_number_id"`
	AssignedAgent    *PhoneNumberAgentInfo `json:"assigned_agent,omitempty"`
	ProviderConfig   json.RawMessage       `json:"provider_config,omitempty"`
	OutboundTrunk    json.RawMessage       `json:"outbound_trunk,omitempty"`
	InboundTrunk     json.RawMessage       `json:"inbound_trunk,omitempty"`
	LivekitStack     LivekitStackType      `json:"livekit_stack,omitempty"`
	StoreSIPMessages *bool                 `json:"store_sip_messages,omitempty"`
}

// SIPTrunkCredentials are digest authentication credentials.
type SIPTrunkCredentials struct {
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
}

// InboundSIPTrunkConfig configures an inbound SIP trunk.
type InboundSIPTrunkConfig struct {
	AllowedAddresses []string `json:"allowed_addresses,omitempty"`
	AllowedNumbers   []string `json:"allowed_numbers,omitempty"`
	// MediaEncryption is "disabled", "allowed" or "required".
	MediaEncryption     string               `json:"media_encryption,omitempty"`
	Credentials         *SIPTrunkCredentials `json:"credentials,omitempty"`
	RemoteDomains       []string             `json:"remote_domains,omitempty"`
	AttributesToHeaders map[string]string    `json:"attributes_to_headers,omitempty"`
}

// OutboundSIPTrunkConfig configures an outbound SIP trunk.
type OutboundSIPTrunkConfig struct {
	Address string `json:"address"`
	// Transport is "auto", "udp", "tcp" or "tls".
	Transport           string               `json:"transport,omitempty"`
	MediaEncryption     string               `json:"media_encryption,omitempty"`
	Headers             map[string]string    `json:"headers,omitempty"`
	AttributesToHeaders map[string]string    `json:"attributes_to_headers,omitempty"`
	Credentials         *SIPTrunkCredentials `json:"credentials,omitempty"`
	EnabledCodecs       []string             `json:"enabled_codecs,omitempty"`
}

// CreatePhoneNumberRequest imports a phone number. Set Provider and the
// provider-specific fields: SID and Token for Twilio; the Exotel fields for
// Exotel; trunk configs for SIP trunks.
type CreatePhoneNumberRequest struct {
	Provider         TelephonyProvider `json:"provider"`
	PhoneNumber      string            `json:"phone_number"`
	Label            string            `json:"label"`
	SupportsInbound  *bool             `json:"supports_inbound,omitempty"`
	SupportsOutbound *bool             `json:"supports_outbound,omitempty"`
	AgentID          string            `json:"agent_id,omitempty"`

	// Twilio.
	SID              string          `json:"sid,omitempty"`
	Token            string          `json:"token,omitempty"`
	AccountAuthToken string          `json:"account_auth_token,omitempty"`
	RegionConfig     json.RawMessage `json:"region_config,omitempty"`
	EnableSMS        *bool           `json:"enable_sms,omitempty"`

	// SIP trunk.
	InboundTrunkConfig  *InboundSIPTrunkConfig  `json:"inbound_trunk_config,omitempty"`
	OutboundTrunkConfig *OutboundSIPTrunkConfig `json:"outbound_trunk_config,omitempty"`

	// Exotel.
	AccountSID   string `json:"account_sid,omitempty"`
	APIKey       string `json:"api_key,omitempty"`
	APIToken     string `json:"api_token,omitempty"`
	APISubdomain string `json:"api_subdomain,omitempty"`
	AppID        string `json:"app_id,omitempty"`
	AppletURL    string `json:"applet_url,omitempty"`
}

// UpdatePhoneNumberRequest changes a phone number. Only set fields change.
type UpdatePhoneNumberRequest struct {
	AgentID             string                  `json:"agent_id,omitempty"`
	Label               string                  `json:"label,omitempty"`
	InboundTrunkConfig  *InboundSIPTrunkConfig  `json:"inbound_trunk_config,omitempty"`
	OutboundTrunkConfig *OutboundSIPTrunkConfig `json:"outbound_trunk_config,omitempty"`
	LivekitStack        LivekitStackType        `json:"livekit_stack,omitempty"`
	StoreSIPMessages    *bool                   `json:"store_sip_messages,omitempty"`
	Environment         string                  `json:"environment,omitempty"`
	BranchID            string                  `json:"branch_id,omitempty"`
}

// PhoneNumberListRequest holds the optional filters of PhoneNumbersService.List.
type PhoneNumberListRequest struct {
	Provider TelephonyProvider `query:"provider"`
	AgentID  string            `query:"agent_id"`
	BranchID string            `query:"branch_id"`
}

// PhoneNumberSortBy is the sort field of PhoneNumbersService.ListV2.
type PhoneNumberSortBy string

// PhoneNumberSortBy values.
const (
	PhoneNumberSortByLabel       PhoneNumberSortBy = "label"
	PhoneNumberSortByPhoneNumber PhoneNumberSortBy = "phone_number"
)

// PhoneNumberListV2Request holds the optional parameters of PhoneNumbersService.ListV2.
type PhoneNumberListV2Request struct {
	PageSize         *int              `query:"page_size"`
	Search           string            `query:"search"`
	Label            string            `query:"label"`
	PhoneNumber      string            `query:"phone_number"`
	Provider         TelephonyProvider `query:"provider"`
	SupportsOutbound *bool             `query:"supports_outbound"`
	AgentID          string            `query:"agent_id"`
	BranchID         string            `query:"branch_id"`
	SortBy           PhoneNumberSortBy `query:"sort_by"`
	SortDirection    SortDirection     `query:"sort_direction"`
	Cursor           string            `query:"cursor"`
}

// PhoneNumbersPage is a page of phone numbers.
type PhoneNumbersPage struct {
	PhoneNumbers []PhoneNumber `json:"phone_numbers"`
	NextCursor   *string       `json:"next_cursor,omitempty"`
	HasMore      *bool         `json:"has_more,omitempty"`
}

// PhoneNumbersService groups the /v1/convai/phone-numbers endpoints.
type PhoneNumbersService struct{ cfg *config }

// List returns all phone numbers. req may be nil.
func (s *PhoneNumbersService) List(ctx context.Context, req *PhoneNumberListRequest, opts ...RequestOption) ([]PhoneNumber, error) {
	out, err := getValue[[]PhoneNumber](ctx, s.cfg, "/v1/convai/phone-numbers", req, opts)
	if err != nil {
		return nil, err
	}
	return *out, nil
}

// Create imports a phone number and returns its ID.
func (s *PhoneNumbersService) Create(ctx context.Context, req *CreatePhoneNumberRequest, opts ...RequestOption) (string, error) {
	out, err := sendValue[struct {
		PhoneNumberID string `json:"phone_number_id"`
	}](ctx, s.cfg, http.MethodPost, "/v1/convai/phone-numbers", nil, orEmpty(req), opts)
	if err != nil {
		return "", err
	}
	return out.PhoneNumberID, nil
}

// Get returns a phone number.
func (s *PhoneNumbersService) Get(ctx context.Context, phoneNumberID string, opts ...RequestOption) (*PhoneNumber, error) {
	return getValue[PhoneNumber](ctx, s.cfg, pathf("/v1/convai/phone-numbers/%s", phoneNumberID), nil, opts)
}

// Delete deletes a phone number.
func (s *PhoneNumbersService) Delete(ctx context.Context, phoneNumberID string, opts ...RequestOption) error {
	return doNoContent(ctx, s.cfg, http.MethodDelete, pathf("/v1/convai/phone-numbers/%s", phoneNumberID), nil, nil, opts)
}

// Update changes a phone number.
func (s *PhoneNumbersService) Update(ctx context.Context, phoneNumberID string, req *UpdatePhoneNumberRequest, opts ...RequestOption) (*PhoneNumber, error) {
	return sendValue[PhoneNumber](ctx, s.cfg, http.MethodPatch, pathf("/v1/convai/phone-numbers/%s", phoneNumberID), nil, orEmpty(req), opts)
}

// ListV2 returns one page of phone numbers. req may be nil.
func (s *PhoneNumbersService) ListV2(ctx context.Context, req *PhoneNumberListV2Request, opts ...RequestOption) (*PhoneNumbersPage, error) {
	return getValue[PhoneNumbersPage](ctx, s.cfg, "/v1/convai/v2/phone-numbers", req, opts)
}

// ListAll iterates over every phone number using ListV2.
func (s *PhoneNumbersService) ListAll(ctx context.Context, req *PhoneNumberListV2Request, opts ...RequestOption) iter.Seq2[PhoneNumber, error] {
	params := *orEmpty(req)
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[PhoneNumber], error) {
		params.Cursor = cursor
		p, err := s.ListV2(ctx, &params, opts...)
		if err != nil {
			return page[PhoneNumber]{}, err
		}
		next := deref(p.NextCursor)
		return page[PhoneNumber]{items: p.PhoneNumbers, next: next, hasMore: next != "" && (p.HasMore == nil || *p.HasMore)}, nil
	})
}

// GetSIPMessages returns one page of stored SIP messages. req may be nil.
func (s *PhoneNumbersService) GetSIPMessages(ctx context.Context, phoneNumberID string, req *PageRequest, opts ...RequestOption) (*SIPMessagesPage, error) {
	return getValue[SIPMessagesPage](ctx, s.cfg, pathf("/v1/convai/phone-numbers/%s/sip-messages", phoneNumberID), req, opts)
}

// TwilioService groups the /v1/convai/twilio endpoints.
type TwilioService struct{ cfg *config }

// OutboundCall starts an outbound call via Twilio.
func (s *TwilioService) OutboundCall(ctx context.Context, req *TwilioOutboundCallRequest, opts ...RequestOption) (*OutboundCallResponse, error) {
	return sendValue[OutboundCallResponse](ctx, s.cfg, http.MethodPost, "/v1/convai/twilio/outbound-call", nil, orEmpty(req), opts)
}

// TwilioRegisterCallRequest registers a call placed with your own Twilio integration.
type TwilioRegisterCallRequest struct {
	AgentID                          string                            `json:"agent_id"`
	FromNumber                       string                            `json:"from_number"`
	ToNumber                         string                            `json:"to_number"`
	Direction                        TelephonyDirection                `json:"direction,omitempty"`
	ConversationInitiationClientData *ConversationInitiationClientData `json:"conversation_initiation_client_data,omitempty"`
}

// RegisterCall registers a Twilio call and returns the TwiML to answer it with.
func (s *TwilioService) RegisterCall(ctx context.Context, req *TwilioRegisterCallRequest, opts ...RequestOption) (string, error) {
	const path = "/v1/convai/twilio/register-call"
	body, err := doStream(ctx, s.cfg, &request{method: http.MethodPost, path: path, body: orEmpty(req), accept: acceptAny, opts: opts})
	if err != nil {
		return "", err
	}
	defer body.Close()
	b, err := io.ReadAll(body)
	if err != nil {
		return "", fmt.Errorf("elevenlabs: read POST %s response: %w", path, err)
	}
	return string(b), nil
}

// SIPTrunkService groups the /v1/convai/sip-trunk endpoints.
type SIPTrunkService struct{ cfg *config }

// OutboundCall starts an outbound call via a SIP trunk.
func (s *SIPTrunkService) OutboundCall(ctx context.Context, req *OutboundCallRequest, opts ...RequestOption) (*OutboundCallResponse, error) {
	return sendValue[OutboundCallResponse](ctx, s.cfg, http.MethodPost, "/v1/convai/sip-trunk/outbound-call", nil, orEmpty(req), opts)
}

// ExotelService groups the /v1/convai/exotel endpoints.
type ExotelService struct{ cfg *config }

// OutboundCall starts an outbound call via Exotel.
func (s *ExotelService) OutboundCall(ctx context.Context, req *OutboundCallRequest, opts ...RequestOption) (*OutboundCallResponse, error) {
	return sendValue[OutboundCallResponse](ctx, s.cfg, http.MethodPost, "/v1/convai/exotel/outbound-call", nil, orEmpty(req), opts)
}

// WhatsAppService groups the /v1/convai/whatsapp endpoints.
type WhatsAppService struct{ cfg *config }

// WhatsAppOutboundCallRequest starts an outbound WhatsApp call.
type WhatsAppOutboundCallRequest struct {
	WhatsAppPhoneNumberID                             string                            `json:"whatsapp_phone_number_id"`
	WhatsAppUserID                                    string                            `json:"whatsapp_user_id"`
	WhatsAppCallPermissionRequestTemplateName         string                            `json:"whatsapp_call_permission_request_template_name"`
	WhatsAppCallPermissionRequestTemplateLanguageCode string                            `json:"whatsapp_call_permission_request_template_language_code"`
	AgentID                                           string                            `json:"agent_id"`
	ConversationInitiationClientData                  *ConversationInitiationClientData `json:"conversation_initiation_client_data,omitempty"`
}

// WhatsAppTemplateParam is a text parameter of a WhatsApp template component.
// Header components may also carry media parameters; use Raw for those.
type WhatsAppTemplateParam struct {
	ParameterName string `json:"parameter_name,omitempty"`
	// Type defaults to "text" on the server.
	Type string `json:"type,omitempty"`
	Text string `json:"text,omitempty"`
}

// WhatsAppTemplateComponent is a WhatsApp template component. Type is
// "body", "button" or "header"; Index and SubType apply to buttons.
type WhatsAppTemplateComponent struct {
	Type       string                  `json:"type"`
	Parameters []WhatsAppTemplateParam `json:"parameters"`
	Index      *int                    `json:"index,omitempty"`
	SubType    string                  `json:"sub_type,omitempty"`
}

// WhatsAppOutboundMessageRequest sends an outbound WhatsApp template message.
type WhatsAppOutboundMessageRequest struct {
	WhatsAppPhoneNumberID            string                            `json:"whatsapp_phone_number_id"`
	WhatsAppUserID                   string                            `json:"whatsapp_user_id"`
	TemplateName                     string                            `json:"template_name"`
	TemplateLanguageCode             string                            `json:"template_language_code"`
	TemplateParams                   []WhatsAppTemplateComponent       `json:"template_params"`
	AgentID                          string                            `json:"agent_id"`
	ConversationInitiationClientData *ConversationInitiationClientData `json:"conversation_initiation_client_data,omitempty"`
}

// OutboundCall starts an outbound call via WhatsApp.
func (s *WhatsAppService) OutboundCall(ctx context.Context, req *WhatsAppOutboundCallRequest, opts ...RequestOption) (*OutboundCallResponse, error) {
	return sendValue[OutboundCallResponse](ctx, s.cfg, http.MethodPost, "/v1/convai/whatsapp/outbound-call", nil, orEmpty(req), opts)
}

// OutboundMessage sends a WhatsApp template message and returns the
// conversation ID.
func (s *WhatsAppService) OutboundMessage(ctx context.Context, req *WhatsAppOutboundMessageRequest, opts ...RequestOption) (string, error) {
	out, err := sendValue[struct {
		ConversationID string `json:"conversation_id"`
	}](ctx, s.cfg, http.MethodPost, "/v1/convai/whatsapp/outbound-message", nil, orEmpty(req), opts)
	if err != nil {
		return "", err
	}
	return out.ConversationID, nil
}

// WhatsAppAccountsService groups the /v1/convai/whatsapp-accounts endpoints.
type WhatsAppAccountsService struct{ cfg *config }

// WhatsAppAccount is a connected WhatsApp business account.
type WhatsAppAccount struct {
	BusinessAccountID   string `json:"business_account_id"`
	PhoneNumberID       string `json:"phone_number_id"`
	BusinessAccountName string `json:"business_account_name"`
	PhoneNumberName     string `json:"phone_number_name"`
	PhoneNumber         string `json:"phone_number"`
	// AccountType is the Embedded Signup flow that produced the account.
	AccountType                string  `json:"account_type,omitempty"`
	AssignedAgentID            *string `json:"assigned_agent_id,omitempty"`
	EnableMessaging            *bool   `json:"enable_messaging,omitempty"`
	EnableAudioMessageResponse *bool   `json:"enable_audio_message_response,omitempty"`
	EnableTypingIndicator      *bool   `json:"enable_typing_indicator,omitempty"`
	AssignedAgentName          *string `json:"assigned_agent_name,omitempty"`
	IsTokenExpired             *bool   `json:"is_token_expired,omitempty"`
}

// UpdateWhatsAppAccountRequest changes a WhatsApp account.
type UpdateWhatsAppAccountRequest struct {
	AssignedAgentID            string `json:"assigned_agent_id,omitempty"`
	EnableMessaging            *bool  `json:"enable_messaging,omitempty"`
	EnableAudioMessageResponse *bool  `json:"enable_audio_message_response,omitempty"`
	EnableTypingIndicator      *bool  `json:"enable_typing_indicator,omitempty"`
}

// WhatsAppAccountListRequest holds the optional filters of WhatsAppAccountsService.List.
type WhatsAppAccountListRequest struct {
	AgentID string `query:"agent_id"`
}

// Get returns a WhatsApp account.
func (s *WhatsAppAccountsService) Get(ctx context.Context, phoneNumberID string, opts ...RequestOption) (*WhatsAppAccount, error) {
	return getValue[WhatsAppAccount](ctx, s.cfg, pathf("/v1/convai/whatsapp-accounts/%s", phoneNumberID), nil, opts)
}

// Delete disconnects a WhatsApp account.
func (s *WhatsAppAccountsService) Delete(ctx context.Context, phoneNumberID string, opts ...RequestOption) error {
	return doNoContent(ctx, s.cfg, http.MethodDelete, pathf("/v1/convai/whatsapp-accounts/%s", phoneNumberID), nil, nil, opts)
}

// Update changes a WhatsApp account.
func (s *WhatsAppAccountsService) Update(ctx context.Context, phoneNumberID string, req *UpdateWhatsAppAccountRequest, opts ...RequestOption) error {
	return doNoContent(ctx, s.cfg, http.MethodPatch, pathf("/v1/convai/whatsapp-accounts/%s", phoneNumberID), nil, orEmpty(req), opts)
}

// List returns the connected WhatsApp accounts. req may be nil.
func (s *WhatsAppAccountsService) List(ctx context.Context, req *WhatsAppAccountListRequest, opts ...RequestOption) ([]WhatsAppAccount, error) {
	out, err := getValue[struct {
		Items []WhatsAppAccount `json:"items"`
	}](ctx, s.cfg, "/v1/convai/whatsapp-accounts", req, opts)
	if err != nil {
		return nil, err
	}
	return out.Items, nil
}

// BatchCallsService groups the /v1/convai/batch-calling endpoints.
type BatchCallsService struct{ cfg *config }

// BatchCallStatus is the state of a batch call.
type BatchCallStatus string

// BatchCallStatus values.
const (
	BatchCallStatusPending    BatchCallStatus = "pending"
	BatchCallStatusInProgress BatchCallStatus = "in_progress"
	BatchCallStatusCompleted  BatchCallStatus = "completed"
	BatchCallStatusFailed     BatchCallStatus = "failed"
	BatchCallStatusCancelled  BatchCallStatus = "cancelled"
)

// BatchCallRecipientStatus is the state of one batch call recipient.
type BatchCallRecipientStatus string

// BatchCallRecipientStatus values.
const (
	BatchCallRecipientPending    BatchCallRecipientStatus = "pending"
	BatchCallRecipientDispatched BatchCallRecipientStatus = "dispatched"
	BatchCallRecipientInitiated  BatchCallRecipientStatus = "initiated"
	BatchCallRecipientInProgress BatchCallRecipientStatus = "in_progress"
	BatchCallRecipientCompleted  BatchCallRecipientStatus = "completed"
	BatchCallRecipientFailed     BatchCallRecipientStatus = "failed"
	BatchCallRecipientCancelled  BatchCallRecipientStatus = "cancelled"
	BatchCallRecipientVoicemail  BatchCallRecipientStatus = "voicemail"
)

// BatchCallWhatsAppParams configures WhatsApp batch calls.
type BatchCallWhatsAppParams struct {
	WhatsAppPhoneNumberID                             string `json:"whatsapp_phone_number_id,omitempty"`
	WhatsAppCallPermissionRequestTemplateName         string `json:"whatsapp_call_permission_request_template_name"`
	WhatsAppCallPermissionRequestTemplateLanguageCode string `json:"whatsapp_call_permission_request_template_language_code"`
}

// BatchCallRecipient is one recipient of a batch call submission.
type BatchCallRecipient struct {
	ID                               string                            `json:"id,omitempty"`
	PhoneNumber                      string                            `json:"phone_number,omitempty"`
	WhatsAppUserID                   string                            `json:"whatsapp_user_id,omitempty"`
	ConversationInitiationClientData *ConversationInitiationClientData `json:"conversation_initiation_client_data,omitempty"`
}

// CreateBatchCallRequest submits a batch call.
type CreateBatchCallRequest struct {
	CallName               string                   `json:"call_name"`
	AgentID                string                   `json:"agent_id"`
	Recipients             []BatchCallRecipient     `json:"recipients"`
	ScheduledTimeUnix      *int64                   `json:"scheduled_time_unix,omitempty"`
	AgentPhoneNumberID     string                   `json:"agent_phone_number_id,omitempty"`
	WhatsAppParams         *BatchCallWhatsAppParams `json:"whatsapp_params,omitempty"`
	Timezone               string                   `json:"timezone,omitempty"`
	BranchID               string                   `json:"branch_id,omitempty"`
	Environment            string                   `json:"environment,omitempty"`
	TelephonyCallConfig    *TelephonyCallConfig     `json:"telephony_call_config,omitempty"`
	TargetConcurrencyLimit *int                     `json:"target_concurrency_limit,omitempty"`
}

// BatchCall summarizes a batch call.
type BatchCall struct {
	ID                     string                   `json:"id"`
	PhoneNumberID          *string                  `json:"phone_number_id,omitempty"`
	PhoneProvider          TelephonyProvider        `json:"phone_provider,omitempty"`
	WhatsAppParams         *BatchCallWhatsAppParams `json:"whatsapp_params,omitempty"`
	Name                   string                   `json:"name"`
	AgentID                string                   `json:"agent_id"`
	BranchID               *string                  `json:"branch_id,omitempty"`
	Environment            *string                  `json:"environment,omitempty"`
	CreatedAtUnix          int64                    `json:"created_at_unix"`
	ScheduledTimeUnix      int64                    `json:"scheduled_time_unix"`
	Timezone               *string                  `json:"timezone,omitempty"`
	TotalCallsDispatched   int                      `json:"total_calls_dispatched"`
	TotalCallsScheduled    int                      `json:"total_calls_scheduled"`
	TotalCallsFinished     int                      `json:"total_calls_finished"`
	LastUpdatedAtUnix      int64                    `json:"last_updated_at_unix"`
	Status                 BatchCallStatus          `json:"status"`
	RetryCount             int                      `json:"retry_count"`
	TelephonyCallConfig    *TelephonyCallConfig     `json:"telephony_call_config,omitempty"`
	TargetConcurrencyLimit *int                     `json:"target_concurrency_limit,omitempty"`
	AgentName              string                   `json:"agent_name"`
	BranchName             *string                  `json:"branch_name,omitempty"`
}

// BatchCallRecipientStatusInfo is the state of one recipient of a batch call.
type BatchCallRecipientStatusInfo struct {
	ID                               string                   `json:"id"`
	PhoneNumber                      *string                  `json:"phone_number,omitempty"`
	WhatsAppUserID                   *string                  `json:"whatsapp_user_id,omitempty"`
	Status                           BatchCallRecipientStatus `json:"status"`
	CreatedAtUnix                    int64                    `json:"created_at_unix"`
	UpdatedAtUnix                    int64                    `json:"updated_at_unix"`
	ConversationID                   *string                  `json:"conversation_id,omitempty"`
	ConversationInitiationClientData json.RawMessage          `json:"conversation_initiation_client_data,omitempty"`
}

// BatchCallDetails is a batch call with all recipients.
type BatchCallDetails struct {
	BatchCall
	Recipients []BatchCallRecipientStatusInfo `json:"recipients"`
}

// BatchCallListRequest holds the optional parameters of BatchCallsService.List.
type BatchCallListRequest struct {
	Limit   *int   `query:"limit"`
	LastDoc string `query:"last_doc"`
	AgentID string `query:"agent_id"`
}

// BatchCallsPage is a page of batch calls.
type BatchCallsPage struct {
	BatchCalls []BatchCall `json:"batch_calls"`
	NextDoc    *string     `json:"next_doc,omitempty"`
	HasMore    *bool       `json:"has_more,omitempty"`
}

// BatchCallExportRequest holds the optional parameters of BatchCallsService.Export.
type BatchCallExportRequest struct {
	// Limit exports only the first N recipients.
	Limit *int `query:"limit"`
}

// Create submits a batch call.
func (s *BatchCallsService) Create(ctx context.Context, req *CreateBatchCallRequest, opts ...RequestOption) (*BatchCall, error) {
	return sendValue[BatchCall](ctx, s.cfg, http.MethodPost, "/v1/convai/batch-calling/submit", nil, orEmpty(req), opts)
}

// List returns one page of the workspace's batch calls. req may be nil.
func (s *BatchCallsService) List(ctx context.Context, req *BatchCallListRequest, opts ...RequestOption) (*BatchCallsPage, error) {
	return getValue[BatchCallsPage](ctx, s.cfg, "/v1/convai/batch-calling/workspace", req, opts)
}

// ListAll iterates over every batch call.
func (s *BatchCallsService) ListAll(ctx context.Context, req *BatchCallListRequest, opts ...RequestOption) iter.Seq2[BatchCall, error] {
	params := *orEmpty(req)
	return paginate(ctx, params.LastDoc, func(ctx context.Context, cursor string) (page[BatchCall], error) {
		params.LastDoc = cursor
		p, err := s.List(ctx, &params, opts...)
		if err != nil {
			return page[BatchCall]{}, err
		}
		next := deref(p.NextDoc)
		return page[BatchCall]{items: p.BatchCalls, next: next, hasMore: next != "" && (p.HasMore == nil || *p.HasMore)}, nil
	})
}

// Get returns a batch call with its recipients.
func (s *BatchCallsService) Get(ctx context.Context, batchID string, opts ...RequestOption) (*BatchCallDetails, error) {
	return getValue[BatchCallDetails](ctx, s.cfg, pathf("/v1/convai/batch-calling/%s", batchID), nil, opts)
}

// Delete deletes a batch call.
func (s *BatchCallsService) Delete(ctx context.Context, batchID string, opts ...RequestOption) error {
	return doNoContent(ctx, s.cfg, http.MethodDelete, pathf("/v1/convai/batch-calling/%s", batchID), nil, nil, opts)
}

// Cancel cancels a running batch call.
func (s *BatchCallsService) Cancel(ctx context.Context, batchID string, opts ...RequestOption) (*BatchCall, error) {
	return sendValue[BatchCall](ctx, s.cfg, http.MethodPost, pathf("/v1/convai/batch-calling/%s/cancel", batchID), nil, nil, opts)
}

// Retry retries the failed and unanswered calls of a batch.
func (s *BatchCallsService) Retry(ctx context.Context, batchID string, opts ...RequestOption) (*BatchCall, error) {
	return sendValue[BatchCall](ctx, s.cfg, http.MethodPost, pathf("/v1/convai/batch-calling/%s/retry", batchID), nil, nil, opts)
}

// Export streams a CSV export of a batch call's recipients. The caller
// must close it. req may be nil.
func (s *BatchCallsService) Export(ctx context.Context, batchID string, req *BatchCallExportRequest, opts ...RequestOption) (io.ReadCloser, error) {
	return doStream(ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/convai/batch-calling/%s/export", batchID), query: req, accept: acceptAny, opts: opts})
}
