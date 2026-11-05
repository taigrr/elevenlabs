package telephony

import (
	"context"
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/convai/agents"
	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/types"
)

const (
	batchCallJSON   = `{"id":"b","name":"n","agent_id":"a","created_at_unix":1,"scheduled_time_unix":2,"total_calls_dispatched":0,"total_calls_scheduled":3,"total_calls_finished":0,"last_updated_at_unix":1,"status":"pending","retry_count":0,"agent_name":"A","phone_provider":"twilio"}`
	phoneNumberJSON = `{"provider":"sip_trunk","phone_number":"+1","label":"l","phone_number_id":"p","livekit_stack":"standard","assigned_agent":{"agent_id":"a","agent_name":"A"}}`
)

func endpointCases() []testutil.Endpoint[*Service] {
	callResp := `{"success":true,"message":"ok","conversation_id":"c","callSid":"CA1"}`
	outbound := OutboundCallRequest{AgentID: "a", AgentPhoneNumberID: "p", ToNumber: "+2"}
	return []testutil.Endpoint[*Service]{
		{
			Name: "convai.phone_numbers.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.PhoneNumbers.List(ctx, &PhoneNumberListRequest{Provider: ProviderTwilio, AgentID: "a"})
			},
			Method: http.MethodGet, Path: "/v1/convai/phone-numbers", Query: "agent_id=a&provider=twilio",
			RespBody: "[" + phoneNumberJSON + "]",
			Check: func(t *testing.T, r any) {
				if p := r.([]PhoneNumber); len(p) != 1 || p[0].Provider != ProviderSIPTrunk || p[0].AssignedAgent.AgentName != "A" {
					t.Errorf("numbers = %+v", p)
				}
			},
		},
		{
			Name: "convai.phone_numbers.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.PhoneNumbers.Create(ctx, &CreatePhoneNumberRequest{Provider: ProviderTwilio, PhoneNumber: "+1", Label: "l", SID: "AC", Token: "t"})
			},
			Method: http.MethodPost, Path: "/v1/convai/phone-numbers",
			Body:     `{"provider":"twilio","phone_number":"+1","label":"l","sid":"AC","token":"t"}`,
			RespBody: `{"phone_number_id":"p"}`,
			Check: func(t *testing.T, r any) {
				if r.(string) != "p" {
					t.Errorf("id = %v", r)
				}
			},
		},
		{
			Name: "convai.phone_numbers.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.PhoneNumbers.Get(ctx, "p")
			},
			Method: http.MethodGet, Path: "/v1/convai/phone-numbers/p", RespBody: phoneNumberJSON,
		},
		{
			Name: "convai.phone_numbers.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return nil, c.PhoneNumbers.Delete(ctx, "p")
			},
			Method: http.MethodDelete, Path: "/v1/convai/phone-numbers/p", RespBody: `{}`,
		},
		{
			Name: "convai.phone_numbers.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.PhoneNumbers.Update(ctx, "p", &UpdatePhoneNumberRequest{AgentID: "a", OutboundTrunkConfig: &OutboundSIPTrunkConfig{Address: "sip.x"}, StoreSIPMessages: types.Ptr(true)})
			},
			Method: http.MethodPatch, Path: "/v1/convai/phone-numbers/p",
			Body:     `{"agent_id":"a","outbound_trunk_config":{"address":"sip.x"},"store_sip_messages":true}`,
			RespBody: phoneNumberJSON,
		},
		{
			Name: "convai.phone_numbers.list_v2",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.PhoneNumbers.ListV2(ctx, &PhoneNumberListV2Request{PageSize: types.Ptr(5), SupportsOutbound: types.Ptr(true), SortBy: PhoneNumberSortByLabel, Cursor: "c"})
			},
			Method: http.MethodGet, Path: "/v1/convai/v2/phone-numbers", Query: "cursor=c&page_size=5&sort_by=label&supports_outbound=true",
			RespBody: `{"phone_numbers":[` + phoneNumberJSON + `],"next_cursor":"n","has_more":true}`,
		},
		{
			Name: "convai.phone_numbers.get_sip_messages",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.PhoneNumbers.GetSIPMessages(ctx, "p", &agents.PageRequest{PageSize: types.Ptr(2)})
			},
			Method: http.MethodGet, Path: "/v1/convai/phone-numbers/p/sip-messages", Query: "page_size=2",
			RespBody: `{"sip_messages":[{"call_id":"x","direction":"inbound"}]}`,
		},
		{
			Name: "convai.twilio.outbound_call",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Twilio.OutboundCall(ctx, &TwilioOutboundCallRequest{OutboundCallRequest: outbound, CallRecordingEnabled: types.Ptr(true)})
			},
			Method: http.MethodPost, Path: "/v1/convai/twilio/outbound-call",
			Body:     `{"agent_id":"a","agent_phone_number_id":"p","to_number":"+2","call_recording_enabled":true}`,
			RespBody: callResp,
			Check: func(t *testing.T, r any) {
				if o := r.(*OutboundCallResponse); !o.Success || *o.CallSID != "CA1" {
					t.Errorf("resp = %+v", o)
				}
			},
		},
		{
			Name: "convai.twilio.register_call",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Twilio.RegisterCall(ctx, &TwilioRegisterCallRequest{AgentID: "a", FromNumber: "+1", ToNumber: "+2", Direction: DirectionInbound})
			},
			Method: http.MethodPost, Path: "/v1/convai/twilio/register-call", Accept: transport.AcceptAny,
			Body:     `{"agent_id":"a","from_number":"+1","to_number":"+2","direction":"inbound"}`,
			RespType: "application/xml", RespBody: `<Response/>`,
			Check: func(t *testing.T, r any) {
				if r.(string) != "<Response/>" {
					t.Errorf("twiml = %v", r)
				}
			},
		},
		{
			Name: "convai.sip_trunk.outbound_call",
			Call: func(ctx context.Context, c *Service) (any, error) {
				req := outbound
				req.ConversationInitiationClientData = &agents.ConversationInitiationClientData{DynamicVariables: map[string]any{"name": "x"}}
				req.TelephonyCallConfig = &CallConfig{RingingTimeoutSecs: types.Ptr(30)}
				return c.SIPTrunk.OutboundCall(ctx, &req)
			},
			Method: http.MethodPost, Path: "/v1/convai/sip-trunk/outbound-call",
			Body:     `{"agent_id":"a","agent_phone_number_id":"p","to_number":"+2","conversation_initiation_client_data":{"dynamic_variables":{"name":"x"}},"telephony_call_config":{"ringing_timeout_secs":30}}`,
			RespBody: `{"success":true,"message":"ok","sip_call_id":"s"}`,
		},
		{
			Name: "convai.exotel.outbound_call",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Exotel.OutboundCall(ctx, &outbound)
			},
			Method: http.MethodPost, Path: "/v1/convai/exotel/outbound-call",
			Body:     `{"agent_id":"a","agent_phone_number_id":"p","to_number":"+2"}`,
			RespBody: callResp,
		},
		{
			Name: "convai.whatsapp.outbound_call",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.WhatsApp.OutboundCall(ctx, &WhatsAppOutboundCallRequest{WhatsAppPhoneNumberID: "w", WhatsAppUserID: "u", WhatsAppCallPermissionRequestTemplateName: "t", WhatsAppCallPermissionRequestTemplateLanguageCode: "en", AgentID: "a"})
			},
			Method: http.MethodPost, Path: "/v1/convai/whatsapp/outbound-call",
			Body:     `{"whatsapp_phone_number_id":"w","whatsapp_user_id":"u","whatsapp_call_permission_request_template_name":"t","whatsapp_call_permission_request_template_language_code":"en","agent_id":"a"}`,
			RespBody: `{"success":true,"message":"ok"}`,
		},
		{
			Name: "convai.whatsapp.outbound_message",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.WhatsApp.OutboundMessage(ctx, &WhatsAppOutboundMessageRequest{
					WhatsAppPhoneNumberID: "w", WhatsAppUserID: "u", TemplateName: "t", TemplateLanguageCode: "en", AgentID: "a",
					TemplateParams: []WhatsAppTemplateComponent{{Type: "body", Parameters: []WhatsAppTemplateParam{{Text: "hi"}}}},
				})
			},
			Method: http.MethodPost, Path: "/v1/convai/whatsapp/outbound-message",
			Body:     `{"whatsapp_phone_number_id":"w","whatsapp_user_id":"u","template_name":"t","template_language_code":"en","template_params":[{"type":"body","parameters":[{"text":"hi"}]}],"agent_id":"a"}`,
			RespBody: `{"conversation_id":"c"}`,
			Check: func(t *testing.T, r any) {
				if r.(string) != "c" {
					t.Errorf("id = %v", r)
				}
			},
		},
		{
			Name: "convai.whatsapp_accounts.get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.WhatsAppAccounts.Get(ctx, "p")
			},
			Method: http.MethodGet, Path: "/v1/convai/whatsapp-accounts/p",
			RespBody: `{"business_account_id":"b","phone_number_id":"p","business_account_name":"B","phone_number_name":"P","phone_number":"+1"}`,
		},
		{
			Name: "convai.whatsapp_accounts.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return nil, c.WhatsAppAccounts.Delete(ctx, "p")
			},
			Method: http.MethodDelete, Path: "/v1/convai/whatsapp-accounts/p", RespBody: `{}`,
		},
		{
			Name: "convai.whatsapp_accounts.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return nil, c.WhatsAppAccounts.Update(ctx, "p", &UpdateWhatsAppAccountRequest{AssignedAgentID: "a", EnableMessaging: types.Ptr(false)})
			},
			Method: http.MethodPatch, Path: "/v1/convai/whatsapp-accounts/p",
			Body: `{"assigned_agent_id":"a","enable_messaging":false}`, RespBody: `{}`,
		},
		{
			Name: "convai.whatsapp_accounts.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.WhatsAppAccounts.List(ctx, &WhatsAppAccountListRequest{AgentID: "a"})
			},
			Method: http.MethodGet, Path: "/v1/convai/whatsapp-accounts", Query: "agent_id=a",
			RespBody: `{"items":[{"phone_number_id":"p"}]}`,
			Check: func(t *testing.T, r any) {
				if a := r.([]WhatsAppAccount); len(a) != 1 || a[0].PhoneNumberID != "p" {
					t.Errorf("accounts = %+v", a)
				}
			},
		},
		{
			Name: "convai.batch_calls.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.BatchCalls.Create(ctx, &CreateBatchCallRequest{CallName: "n", AgentID: "a", Recipients: []BatchCallRecipient{{PhoneNumber: "+1"}}, ScheduledTimeUnix: types.Ptr[int64](9), TargetConcurrencyLimit: types.Ptr(2)})
			},
			Method: http.MethodPost, Path: "/v1/convai/batch-calling/submit",
			Body:     `{"call_name":"n","agent_id":"a","recipients":[{"phone_number":"+1"}],"scheduled_time_unix":9,"target_concurrency_limit":2}`,
			RespBody: batchCallJSON,
			Check: func(t *testing.T, r any) {
				if b := r.(*BatchCall); b.Status != BatchCallStatusPending || b.PhoneProvider != ProviderTwilio {
					t.Errorf("batch = %+v", b)
				}
			},
		},
		{
			Name: "convai.batch_calls.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.BatchCalls.List(ctx, &BatchCallListRequest{Limit: types.Ptr(10), LastDoc: "d", AgentID: "a"})
			},
			Method: http.MethodGet, Path: "/v1/convai/batch-calling/workspace", Query: "agent_id=a&last_doc=d&limit=10",
			RespBody: `{"batch_calls":[` + batchCallJSON + `],"next_doc":"e","has_more":true}`,
		},
		{
			Name:   "convai.batch_calls.get",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.BatchCalls.Get(ctx, "b") },
			Method: http.MethodGet, Path: "/v1/convai/batch-calling/b",
			RespBody: batchCallJSON[:len(batchCallJSON)-1] + `,"recipients":[{"id":"r","status":"completed","created_at_unix":1,"updated_at_unix":2}]}`,
			Check: func(t *testing.T, r any) {
				if d := r.(*BatchCallDetails); d.ID != "b" || d.Recipients[0].Status != BatchCallRecipientCompleted {
					t.Errorf("details = %+v", d)
				}
			},
		},
		{
			Name: "convai.batch_calls.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return nil, c.BatchCalls.Delete(ctx, "b")
			},
			Method: http.MethodDelete, Path: "/v1/convai/batch-calling/b",
		},
		{
			Name: "convai.batch_calls.cancel",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.BatchCalls.Cancel(ctx, "b")
			},
			Method: http.MethodPost, Path: "/v1/convai/batch-calling/b/cancel", RespBody: batchCallJSON,
		},
		{
			Name: "convai.batch_calls.retry",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.BatchCalls.Retry(ctx, "b")
			},
			Method: http.MethodPost, Path: "/v1/convai/batch-calling/b/retry", RespBody: batchCallJSON,
		},
		{
			Name: "convai.batch_calls.export",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.ReadAll(c.BatchCalls.Export(ctx, "b", &BatchCallExportRequest{Limit: types.Ptr(3)}))
			},
			Method: http.MethodGet, Path: "/v1/convai/batch-calling/b/export", Query: "limit=3", Accept: transport.AcceptAny,
			RespType: "text/csv", RespBody: "id,phone\n",
			Check: func(t *testing.T, r any) {
				if r.(string) != "id,phone\n" {
					t.Errorf("csv = %q", r)
				}
			},
		},
	}
}

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, endpointCases())
}
