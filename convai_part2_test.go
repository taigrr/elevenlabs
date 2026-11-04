package elevenlabs

import (
	"context"
	"net/http"
	"testing"
)

const (
	batchCallJSON    = `{"id":"b","name":"n","agent_id":"a","created_at_unix":1,"scheduled_time_unix":2,"total_calls_dispatched":0,"total_calls_scheduled":3,"total_calls_finished":0,"last_updated_at_unix":1,"status":"pending","retry_count":0,"agent_name":"A","phone_provider":"twilio"}`
	mcpServerJSON    = `{"id":"m","config":{"name":"srv","url":"https://mcp","approval_policy":"auto_approve_all"},"metadata":{"created_at":5}}`
	triageTicketJSON = `{"agentqa_ticket_id":"t","workspace_id":"w","owner_user_id":"u","agent_id":"a","needs_clustering":false,"labels":[],"conversation_ids":["c"],"ticket_comments":[],"turn_comments":[{"turn_index":2,"comment":"x","created_at_unix_secs":3}],"status":"open","priority_changes":[],"status_changes":[],"source":"manual","created_at_unix_secs":1,"updated_at_unix_secs":2}`
	phoneNumberJSON  = `{"provider":"sip_trunk","phone_number":"+1","label":"l","phone_number_id":"p","livekit_stack":"standard","assigned_agent":{"agent_id":"a","agent_name":"A"}}`
)

func convaiTelephonyEndpointCases() []endpointCase {
	callResp := `{"success":true,"message":"ok","conversation_id":"c","callSid":"CA1"}`
	outbound := OutboundCallRequest{AgentID: "a", AgentPhoneNumberID: "p", ToNumber: "+2"}
	return []endpointCase{
		{
			name: "convai.phone_numbers.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.PhoneNumbers.List(ctx, &PhoneNumberListRequest{Provider: TelephonyProviderTwilio, AgentID: "a"})
			},
			method: http.MethodGet, path: "/v1/convai/phone-numbers", query: "agent_id=a&provider=twilio",
			respBody: "[" + phoneNumberJSON + "]",
			check: func(t *testing.T, r any) {
				if p := r.([]PhoneNumber); len(p) != 1 || p[0].Provider != TelephonyProviderSIPTrunk || p[0].AssignedAgent.AgentName != "A" {
					t.Errorf("numbers = %+v", p)
				}
			},
		},
		{
			name: "convai.phone_numbers.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.PhoneNumbers.Create(ctx, &CreatePhoneNumberRequest{Provider: TelephonyProviderTwilio, PhoneNumber: "+1", Label: "l", SID: "AC", Token: "t"})
			},
			method: http.MethodPost, path: "/v1/convai/phone-numbers",
			body:     `{"provider":"twilio","phone_number":"+1","label":"l","sid":"AC","token":"t"}`,
			respBody: `{"phone_number_id":"p"}`,
			check: func(t *testing.T, r any) {
				if r.(string) != "p" {
					t.Errorf("id = %v", r)
				}
			},
		},
		{
			name: "convai.phone_numbers.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.PhoneNumbers.Get(ctx, "p")
			},
			method: http.MethodGet, path: "/v1/convai/phone-numbers/p", respBody: phoneNumberJSON,
		},
		{
			name: "convai.phone_numbers.delete",
			call: func(ctx context.Context, c *Client) (any, error) {
				return nil, c.ConversationalAI.PhoneNumbers.Delete(ctx, "p")
			},
			method: http.MethodDelete, path: "/v1/convai/phone-numbers/p", respBody: `{}`,
		},
		{
			name: "convai.phone_numbers.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.PhoneNumbers.Update(ctx, "p", &UpdatePhoneNumberRequest{AgentID: "a", OutboundTrunkConfig: &OutboundSIPTrunkConfig{Address: "sip.x"}, StoreSIPMessages: Ptr(true)})
			},
			method: http.MethodPatch, path: "/v1/convai/phone-numbers/p",
			body:     `{"agent_id":"a","outbound_trunk_config":{"address":"sip.x"},"store_sip_messages":true}`,
			respBody: phoneNumberJSON,
		},
		{
			name: "convai.phone_numbers.list_v2",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.PhoneNumbers.ListV2(ctx, &PhoneNumberListV2Request{PageSize: Ptr(5), SupportsOutbound: Ptr(true), SortBy: PhoneNumberSortByLabel, Cursor: "c"})
			},
			method: http.MethodGet, path: "/v1/convai/v2/phone-numbers", query: "cursor=c&page_size=5&sort_by=label&supports_outbound=true",
			respBody: `{"phone_numbers":[` + phoneNumberJSON + `],"next_cursor":"n","has_more":true}`,
		},
		{
			name: "convai.phone_numbers.get_sip_messages",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.PhoneNumbers.GetSIPMessages(ctx, "p", &PageRequest{PageSize: Ptr(2)})
			},
			method: http.MethodGet, path: "/v1/convai/phone-numbers/p/sip-messages", query: "page_size=2",
			respBody: `{"sip_messages":[{"call_id":"x","direction":"inbound"}]}`,
		},
		{
			name: "convai.twilio.outbound_call",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Twilio.OutboundCall(ctx, &TwilioOutboundCallRequest{OutboundCallRequest: outbound, CallRecordingEnabled: Ptr(true)})
			},
			method: http.MethodPost, path: "/v1/convai/twilio/outbound-call",
			body:     `{"agent_id":"a","agent_phone_number_id":"p","to_number":"+2","call_recording_enabled":true}`,
			respBody: callResp,
			check: func(t *testing.T, r any) {
				if o := r.(*OutboundCallResponse); !o.Success || *o.CallSID != "CA1" {
					t.Errorf("resp = %+v", o)
				}
			},
		},
		{
			name: "convai.twilio.register_call",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Twilio.RegisterCall(ctx, &TwilioRegisterCallRequest{AgentID: "a", FromNumber: "+1", ToNumber: "+2", Direction: TelephonyDirectionInbound})
			},
			method: http.MethodPost, path: "/v1/convai/twilio/register-call", accept: acceptAny,
			body:     `{"agent_id":"a","from_number":"+1","to_number":"+2","direction":"inbound"}`,
			respType: "application/xml", respBody: `<Response/>`,
			check: func(t *testing.T, r any) {
				if r.(string) != "<Response/>" {
					t.Errorf("twiml = %v", r)
				}
			},
		},
		{
			name: "convai.sip_trunk.outbound_call",
			call: func(ctx context.Context, c *Client) (any, error) {
				req := outbound
				req.ConversationInitiationClientData = &ConversationInitiationClientData{DynamicVariables: map[string]any{"name": "x"}}
				req.TelephonyCallConfig = &TelephonyCallConfig{RingingTimeoutSecs: Ptr(30)}
				return c.ConversationalAI.SIPTrunk.OutboundCall(ctx, &req)
			},
			method: http.MethodPost, path: "/v1/convai/sip-trunk/outbound-call",
			body:     `{"agent_id":"a","agent_phone_number_id":"p","to_number":"+2","conversation_initiation_client_data":{"dynamic_variables":{"name":"x"}},"telephony_call_config":{"ringing_timeout_secs":30}}`,
			respBody: `{"success":true,"message":"ok","sip_call_id":"s"}`,
		},
		{
			name: "convai.exotel.outbound_call",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Exotel.OutboundCall(ctx, &outbound)
			},
			method: http.MethodPost, path: "/v1/convai/exotel/outbound-call",
			body:     `{"agent_id":"a","agent_phone_number_id":"p","to_number":"+2"}`,
			respBody: callResp,
		},
		{
			name: "convai.whatsapp.outbound_call",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.WhatsApp.OutboundCall(ctx, &WhatsAppOutboundCallRequest{WhatsAppPhoneNumberID: "w", WhatsAppUserID: "u", WhatsAppCallPermissionRequestTemplateName: "t", WhatsAppCallPermissionRequestTemplateLanguageCode: "en", AgentID: "a"})
			},
			method: http.MethodPost, path: "/v1/convai/whatsapp/outbound-call",
			body:     `{"whatsapp_phone_number_id":"w","whatsapp_user_id":"u","whatsapp_call_permission_request_template_name":"t","whatsapp_call_permission_request_template_language_code":"en","agent_id":"a"}`,
			respBody: `{"success":true,"message":"ok"}`,
		},
		{
			name: "convai.whatsapp.outbound_message",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.WhatsApp.OutboundMessage(ctx, &WhatsAppOutboundMessageRequest{
					WhatsAppPhoneNumberID: "w", WhatsAppUserID: "u", TemplateName: "t", TemplateLanguageCode: "en", AgentID: "a",
					TemplateParams: []WhatsAppTemplateComponent{{Type: "body", Parameters: []WhatsAppTemplateParam{{Text: "hi"}}}},
				})
			},
			method: http.MethodPost, path: "/v1/convai/whatsapp/outbound-message",
			body:     `{"whatsapp_phone_number_id":"w","whatsapp_user_id":"u","template_name":"t","template_language_code":"en","template_params":[{"type":"body","parameters":[{"text":"hi"}]}],"agent_id":"a"}`,
			respBody: `{"conversation_id":"c"}`,
			check: func(t *testing.T, r any) {
				if r.(string) != "c" {
					t.Errorf("id = %v", r)
				}
			},
		},
		{
			name: "convai.whatsapp_accounts.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.WhatsAppAccounts.Get(ctx, "p")
			},
			method: http.MethodGet, path: "/v1/convai/whatsapp-accounts/p",
			respBody: `{"business_account_id":"b","phone_number_id":"p","business_account_name":"B","phone_number_name":"P","phone_number":"+1"}`,
		},
		{
			name: "convai.whatsapp_accounts.delete",
			call: func(ctx context.Context, c *Client) (any, error) {
				return nil, c.ConversationalAI.WhatsAppAccounts.Delete(ctx, "p")
			},
			method: http.MethodDelete, path: "/v1/convai/whatsapp-accounts/p", respBody: `{}`,
		},
		{
			name: "convai.whatsapp_accounts.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return nil, c.ConversationalAI.WhatsAppAccounts.Update(ctx, "p", &UpdateWhatsAppAccountRequest{AssignedAgentID: "a", EnableMessaging: Ptr(false)})
			},
			method: http.MethodPatch, path: "/v1/convai/whatsapp-accounts/p",
			body: `{"assigned_agent_id":"a","enable_messaging":false}`, respBody: `{}`,
		},
		{
			name: "convai.whatsapp_accounts.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.WhatsAppAccounts.List(ctx, &WhatsAppAccountListRequest{AgentID: "a"})
			},
			method: http.MethodGet, path: "/v1/convai/whatsapp-accounts", query: "agent_id=a",
			respBody: `{"items":[{"phone_number_id":"p"}]}`,
			check: func(t *testing.T, r any) {
				if a := r.([]WhatsAppAccount); len(a) != 1 || a[0].PhoneNumberID != "p" {
					t.Errorf("accounts = %+v", a)
				}
			},
		},
		{
			name: "convai.batch_calls.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.BatchCalls.Create(ctx, &CreateBatchCallRequest{CallName: "n", AgentID: "a", Recipients: []BatchCallRecipient{{PhoneNumber: "+1"}}, ScheduledTimeUnix: Ptr[int64](9), TargetConcurrencyLimit: Ptr(2)})
			},
			method: http.MethodPost, path: "/v1/convai/batch-calling/submit",
			body:     `{"call_name":"n","agent_id":"a","recipients":[{"phone_number":"+1"}],"scheduled_time_unix":9,"target_concurrency_limit":2}`,
			respBody: batchCallJSON,
			check: func(t *testing.T, r any) {
				if b := r.(*BatchCall); b.Status != BatchCallStatusPending || b.PhoneProvider != TelephonyProviderTwilio {
					t.Errorf("batch = %+v", b)
				}
			},
		},
		{
			name: "convai.batch_calls.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.BatchCalls.List(ctx, &BatchCallListRequest{Limit: Ptr(10), LastDoc: "d", AgentID: "a"})
			},
			method: http.MethodGet, path: "/v1/convai/batch-calling/workspace", query: "agent_id=a&last_doc=d&limit=10",
			respBody: `{"batch_calls":[` + batchCallJSON + `],"next_doc":"e","has_more":true}`,
		},
		{
			name:   "convai.batch_calls.get",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.ConversationalAI.BatchCalls.Get(ctx, "b") },
			method: http.MethodGet, path: "/v1/convai/batch-calling/b",
			respBody: batchCallJSON[:len(batchCallJSON)-1] + `,"recipients":[{"id":"r","status":"completed","created_at_unix":1,"updated_at_unix":2}]}`,
			check: func(t *testing.T, r any) {
				if d := r.(*BatchCallDetails); d.ID != "b" || d.Recipients[0].Status != BatchCallRecipientCompleted {
					t.Errorf("details = %+v", d)
				}
			},
		},
		{
			name: "convai.batch_calls.delete",
			call: func(ctx context.Context, c *Client) (any, error) {
				return nil, c.ConversationalAI.BatchCalls.Delete(ctx, "b")
			},
			method: http.MethodDelete, path: "/v1/convai/batch-calling/b",
		},
		{
			name: "convai.batch_calls.cancel",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.BatchCalls.Cancel(ctx, "b")
			},
			method: http.MethodPost, path: "/v1/convai/batch-calling/b/cancel", respBody: batchCallJSON,
		},
		{
			name: "convai.batch_calls.retry",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.BatchCalls.Retry(ctx, "b")
			},
			method: http.MethodPost, path: "/v1/convai/batch-calling/b/retry", respBody: batchCallJSON,
		},
		{
			name: "convai.batch_calls.export",
			call: func(ctx context.Context, c *Client) (any, error) {
				return readAll(c.ConversationalAI.BatchCalls.Export(ctx, "b", &BatchCallExportRequest{Limit: Ptr(3)}))
			},
			method: http.MethodGet, path: "/v1/convai/batch-calling/b/export", query: "limit=3", accept: acceptAny,
			respType: "text/csv", respBody: "id,phone\n",
			check: func(t *testing.T, r any) {
				if r.(string) != "id,phone\n" {
					t.Errorf("csv = %q", r)
				}
			},
		},
	}
}

func convaiMCPEndpointCases() []endpointCase {
	return []endpointCase{
		{
			name:   "convai.mcp_servers.list",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.ConversationalAI.MCPServers.List(ctx) },
			method: http.MethodGet, path: "/v1/convai/mcp-servers", respBody: `{"mcp_servers":[` + mcpServerJSON + `]}`,
			check: func(t *testing.T, r any) {
				if m := r.([]MCPServer); len(m) != 1 || m[0].Config.ApprovalPolicy != MCPApprovalAutoApproveAll || m[0].Metadata.CreatedAt != 5 {
					t.Errorf("servers = %+v", m)
				}
			},
		},
		{
			name: "convai.mcp_servers.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.MCPServers.Create(ctx, &MCPServerConfig{Name: "srv", URL: "https://mcp", Transport: MCPServerTransportSSE, RequestHeaders: map[string]any{"X": map[string]string{"secret_id": "s"}}})
			},
			method: http.MethodPost, path: "/v1/convai/mcp-servers",
			body:     `{"config":{"name":"srv","url":"https://mcp","transport":"SSE","request_headers":{"X":{"secret_id":"s"}}}}`,
			respBody: mcpServerJSON,
		},
		{
			name:   "convai.mcp_servers.get",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.ConversationalAI.MCPServers.Get(ctx, "m") },
			method: http.MethodGet, path: "/v1/convai/mcp-servers/m", respBody: mcpServerJSON,
		},
		{
			name: "convai.mcp_servers.delete",
			call: func(ctx context.Context, c *Client) (any, error) {
				return nil, c.ConversationalAI.MCPServers.Delete(ctx, "m")
			},
			method: http.MethodDelete, path: "/v1/convai/mcp-servers/m", respBody: `{}`,
		},
		{
			name: "convai.mcp_servers.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.MCPServers.Update(ctx, "m", &UpdateMCPServerRequest{ExecutionMode: "async", ResponseTimeoutSecs: Ptr(30)})
			},
			method: http.MethodPatch, path: "/v1/convai/mcp-servers/m",
			body: `{"execution_mode":"async","response_timeout_secs":30}`, respBody: mcpServerJSON,
		},
		{
			name: "convai.mcp_servers.approval_policy.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.MCPServers.ApprovalPolicy.Update(ctx, "m", MCPApprovalRequireApprovalPerTool)
			},
			method: http.MethodPatch, path: "/v1/convai/mcp-servers/m/approval-policy",
			body: `{"approval_policy":"require_approval_per_tool"}`, respBody: mcpServerJSON,
		},
		{
			name: "convai.mcp_servers.tool_approvals.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.MCPServers.ToolApprovals.Create(ctx, "m", &MCPToolApprovalRequest{ToolName: "t", ToolDescription: "d", ApprovalPolicy: MCPToolAutoApproved})
			},
			method: http.MethodPost, path: "/v1/convai/mcp-servers/m/tool-approvals",
			body: `{"tool_name":"t","tool_description":"d","approval_policy":"auto_approved"}`, respBody: mcpServerJSON,
		},
		{
			name: "convai.mcp_servers.tool_approvals.delete",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.MCPServers.ToolApprovals.Delete(ctx, "m", "t x")
			},
			method: http.MethodDelete, path: "/v1/convai/mcp-servers/m/tool-approvals/t%20x", respBody: mcpServerJSON,
		},
		{
			name: "convai.mcp_servers.tool_configs.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.MCPServers.ToolConfigs.Create(ctx, "m", &MCPToolConfig{ToolName: "t", ToolCallSound: "off"}, &MCPEnvironmentRequest{Environment: "staging"})
			},
			method: http.MethodPost, path: "/v1/convai/mcp-servers/m/tool-configs", query: "environment=staging",
			body: `{"tool_name":"t","tool_call_sound":"off"}`, respBody: mcpServerJSON,
		},
		{
			name: "convai.mcp_servers.tool_configs.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.MCPServers.ToolConfigs.Get(ctx, "m", "t")
			},
			method: http.MethodGet, path: "/v1/convai/mcp-servers/m/tool-configs/t", respBody: `{"tool_name":"t","execution_mode":"immediate"}`,
			check: func(t *testing.T, r any) {
				if c := r.(*MCPToolConfig); c.ToolName != "t" || c.ExecutionMode != "immediate" {
					t.Errorf("config = %+v", c)
				}
			},
		},
		{
			name: "convai.mcp_servers.tool_configs.delete",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.MCPServers.ToolConfigs.Delete(ctx, "m", "t")
			},
			method: http.MethodDelete, path: "/v1/convai/mcp-servers/m/tool-configs/t", respBody: mcpServerJSON,
		},
		{
			name: "convai.mcp_servers.tool_configs.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.MCPServers.ToolConfigs.Update(ctx, "m", "t", &MCPToolConfig{DisableInterruptions: Ptr(true)}, nil)
			},
			method: http.MethodPatch, path: "/v1/convai/mcp-servers/m/tool-configs/t",
			body: `{"disable_interruptions":true}`, respBody: mcpServerJSON,
		},
		{
			name: "convai.mcp_servers.tools.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.MCPServers.Tools.List(ctx, "m", &MCPEnvironmentRequest{Environment: "prod"})
			},
			method: http.MethodGet, path: "/v1/convai/mcp-servers/m/tools", query: "environment=prod",
			respBody: `{"success":true,"tools":[{"name":"t","inputSchema":{"type":"object"},"_meta":{"k":1}}]}`,
			check: func(t *testing.T, r any) {
				if l := r.(*MCPToolList); !l.Success || l.Tools[0].InputSchema["type"] != "object" || l.Tools[0].Meta["k"] == nil {
					t.Errorf("tools = %+v", l)
				}
			},
		},
	}
}

func convaiTriageEndpointCases() []endpointCase {
	return []endpointCase{
		{
			name: "convai.triage_tickets.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.TriageTickets.List(ctx, "a", &TriageTicketListRequest{Status: TriageTicketOpen, Sources: []TriageTicketSource{TriageTicketSourceQA, TriageTicketSourceAgent}})
			},
			method: http.MethodGet, path: "/v1/convai/agents/a/triage-tickets", query: "sources=qa&sources=agent&status=open",
			respBody: `{"agent_conversation_tickets":[` + triageTicketJSON + `],"has_more":false}`,
			check: func(t *testing.T, r any) {
				if p := r.(*TriageTicketsPage); len(p.Tickets) != 1 || p.Tickets[0].TurnComments[0].TurnIndex != 2 {
					t.Errorf("page = %+v", p)
				}
			},
		},
		{
			name: "convai.triage_tickets.create_manual",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.TriageTickets.CreateManual(ctx, "a", &CreateManualTicketRequest{QAComment: "fix", Priority: TriageTicketPriorityHigh})
			},
			method: http.MethodPost, path: "/v1/convai/agents/a/triage-tickets",
			body: `{"qa_comment":"fix","priority":"high"}`, respBody: triageTicketJSON,
		},
		{
			name: "convai.triage_tickets.list_for_workspace",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.TriageTickets.ListForWorkspace(ctx, &TriageTicketListRequest{AgentID: "a", SortBy: TriageTicketSortByPriority, PageSize: Ptr(5)})
			},
			method: http.MethodGet, path: "/v1/convai/triage-tickets", query: "agent_id=a&page_size=5&sort_by=priority",
			respBody: `{"agent_conversation_tickets":[],"has_more":false}`,
		},
		{
			name: "convai.triage_tickets.create",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.TriageTickets.Create(ctx, &CreateTriageTicketRequest{ConversationID: "c", TurnComments: []TurnCommentInput{{TurnIndex: 1, Comment: "bad"}}})
			},
			method: http.MethodPost, path: "/v1/convai/triage-tickets",
			body: `{"conversation_id":"c","turn_comments":[{"turn_index":1,"comment":"bad"}]}`, respBody: triageTicketJSON,
		},
		{
			name: "convai.triage_tickets.list_assignable_users",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.TriageTickets.ListAssignableUsers(ctx, "a")
			},
			method: http.MethodGet, path: "/v1/convai/agents/a/triage-tickets/assignable-users",
			respBody: `[{"user_id":"u","email":"e","is_service_account":false,"has_access":true}]`,
			check: func(t *testing.T, r any) {
				if u := r.([]AssignableUser); len(u) != 1 || !u[0].HasAccess {
					t.Errorf("users = %+v", u)
				}
			},
		},
		{
			name: "convai.triage_tickets.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.TriageTickets.Get(ctx, "t")
			},
			method: http.MethodGet, path: "/v1/convai/triage-tickets/t", respBody: triageTicketJSON,
		},
		{
			name: "convai.triage_tickets.delete",
			call: func(ctx context.Context, c *Client) (any, error) {
				return nil, c.ConversationalAI.TriageTickets.Delete(ctx, "t")
			},
			method: http.MethodDelete, path: "/v1/convai/triage-tickets/t",
		},
		{
			name: "convai.triage_tickets.update",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.TriageTickets.Update(ctx, "t", &UpdateTriageTicketRequest{Status: TriageTicketResolved, AssigneeUserID: &NullString{Null: true}, Priority: &NullString{Value: "low"}})
			},
			method: http.MethodPatch, path: "/v1/convai/triage-tickets/t",
			body: `{"status":"resolved","assignee_user_id":null,"priority":"low"}`, respBody: triageTicketJSON,
		},
		{
			name: "convai.triage_tickets.add_comment",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.TriageTickets.AddComment(ctx, "t", "note")
			},
			method: http.MethodPost, path: "/v1/convai/triage-tickets/t/comments",
			body: `{"comment":"note"}`, respBody: triageTicketJSON,
		},
		{
			name: "convai.triage_tickets.merge",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.TriageTickets.Merge(ctx, "t", []string{"t2", "t3"})
			},
			method: http.MethodPost, path: "/v1/convai/triage-tickets/t/merge",
			body: `{"source_ticket_ids":["t2","t3"]}`, respBody: triageTicketJSON,
		},
		{
			name: "convai.triage_tickets.add_turn_comment",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.TriageTickets.AddTurnComment(ctx, "t", TurnCommentInput{TurnIndex: 0, Comment: "x"})
			},
			method: http.MethodPost, path: "/v1/convai/triage-tickets/t/turn-comments",
			body: `{"turn_index":0,"comment":"x"}`, respBody: triageTicketJSON,
		},
	}
}

func convaiMiscEndpointCases() []endpointCase {
	return []endpointCase{
		{
			name: "convai.analytics.live_count.get",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Analytics.LiveCount.Get(ctx, &LiveCountRequest{AgentIDs: []string{"a", "b"}})
			},
			method: http.MethodGet, path: "/v1/convai/analytics/live-count", query: "agent_ids=a&agent_ids=b",
			respBody: `{"count":4}`,
			check: func(t *testing.T, r any) {
				if r.(int) != 4 {
					t.Errorf("count = %v", r)
				}
			},
		},
		{
			name:   "convai.llm.list",
			call:   func(ctx context.Context, c *Client) (any, error) { return c.ConversationalAI.LLM.List(ctx) },
			method: http.MethodGet, path: "/v1/convai/llm/list",
			respBody: `{"llms":[{"llm":"gpt-4o","is_checkpoint":false,"max_tokens_limit":10,"max_context_limit":20,"supports_image_input":true,"supports_document_input":false,"supports_parallel_tool_calls":true,"available_service_tiers":["standard"]}],"default_deprecation_config":{}}`,
			check: func(t *testing.T, r any) {
				if l := r.(*LLMList); l.LLMs[0].LLM != LLMGPT4o || !l.LLMs[0].SupportsImageInput {
					t.Errorf("llms = %+v", l)
				}
			},
		},
		{
			name: "convai.llm_usage.calculate",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.LLMUsage.Calculate(ctx, &LLMUsageRequest{PromptLength: Ptr(100), NumberOfPages: Ptr(0), RAGEnabled: Ptr(false)})
			},
			method: http.MethodPost, path: "/v1/convai/llm-usage/calculate",
			body:     `{"prompt_length":100,"number_of_pages":0,"rag_enabled":false}`,
			respBody: `{"llm_prices":[{"llm":"gpt-4o","price_per_minute":0.1,"price_per_message":0.01}]}`,
		},
		{
			name: "convai.realtime.create_client_secret",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Realtime.CreateClientSecret(ctx, &CreateClientSecretRequest{ExpiresAfter: &ClientSecretExpiry{Seconds: 60}})
			},
			method: http.MethodPost, path: "/v1/convai/realtime/client_secrets",
			body: `{"expires_after":{"seconds":60}}`, respBody: `{"value":"ek_1","expires_at":99}`,
			check: func(t *testing.T, r any) {
				if s := r.(*ClientSecret); s.Value != "ek_1" || s.ExpiresAt != 99 {
					t.Errorf("secret = %+v", s)
				}
			},
		},
		{
			name: "convai.users.list",
			call: func(ctx context.Context, c *Client) (any, error) {
				return c.ConversationalAI.Users.List(ctx, &ConversationUsersListRequest{AgentID: "a", SortBy: ConversationUsersSortByConversationCount, SortDirection: SortDirectionDesc, CallStartAfterUnix: Ptr[int64](5)})
			},
			method: http.MethodGet, path: "/v1/convai/users", query: "agent_id=a&call_start_after_unix=5&sort_by=conversation_count&sort_direction=desc",
			respBody: `{"users":[{"user_id":"u","conversation_count":3,"last_contact_conversation_id":"c"}],"has_more":false}`,
		},
	}
}

func TestConvAIPart2ListAll(t *testing.T) {
	ctx := context.Background()
	t.Run("batch calls", func(t *testing.T) {
		client, ts := newTestServer(t, sequence(
			respond(http.StatusOK, contentTypeJSON, `{"batch_calls":[{"id":"b1"}],"next_doc":"d2","has_more":true}`),
			respond(http.StatusOK, contentTypeJSON, `{"batch_calls":[{"id":"b2"}],"has_more":false}`),
		))
		var ids []string
		for b, err := range client.ConversationalAI.BatchCalls.ListAll(ctx, nil) {
			if err != nil {
				t.Fatal(err)
			}
			ids = append(ids, b.ID)
		}
		if len(ids) != 2 || ids[1] != "b2" || ts.all()[1].RawQuery != "last_doc=d2" {
			t.Errorf("ids = %v, query = %q", ids, ts.all()[1].RawQuery)
		}
	})
	t.Run("triage tickets", func(t *testing.T) {
		client, ts := newTestServer(t, sequence(
			respond(http.StatusOK, contentTypeJSON, `{"agent_conversation_tickets":[{"agentqa_ticket_id":"t1"}],"next_cursor":"c","has_more":true}`),
			respond(http.StatusOK, contentTypeJSON, `{"agent_conversation_tickets":[{"agentqa_ticket_id":"t2"}],"has_more":false}`),
		))
		n := 0
		for _, err := range client.ConversationalAI.TriageTickets.ListAll(ctx, "a", nil) {
			if err != nil {
				t.Fatal(err)
			}
			n++
		}
		if n != 2 || ts.all()[1].Path != "/v1/convai/agents/a/triage-tickets" {
			t.Errorf("n = %d", n)
		}
	})
	t.Run("phone numbers and users", func(t *testing.T) {
		client, _ := newTestServer(t, respond(http.StatusOK, contentTypeJSON, `{"phone_numbers":[{"phone_number_id":"p"}],"users":[{"user_id":"u"}],"agent_conversation_tickets":[],"has_more":false}`))
		for p, err := range client.ConversationalAI.PhoneNumbers.ListAll(ctx, nil) {
			if err != nil || p.PhoneNumberID != "p" {
				t.Fatal(p, err)
			}
		}
		for u, err := range client.ConversationalAI.Users.ListAll(ctx, nil) {
			if err != nil || u.UserID != "u" {
				t.Fatal(u, err)
			}
		}
		for _, err := range client.ConversationalAI.TriageTickets.ListAllForWorkspace(ctx, nil) {
			t.Fatal("unexpected item", err)
		}
	})
}
