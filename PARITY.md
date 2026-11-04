# ElevenLabs Go parity checklist

Generated from elevenlabs-js fe015f4 (Mon Nov 2 17:14:23 2026 +0100) v2.75.0, `src/api/resources/**/client/Client.ts`. `[x]` = implemented on the `parity` branch (package `github.com/taigrr/elevenlabs/v2`).

**Totals:** 365/419 endpoints covered.  assets: 4/4; audioEffects: 0/1; audioIsolation: 4/4; audioNative: 4/4; conversationalAi: 133/186; dubbing: 37/37; environmentVariables: 5/5; environments: 3/3; flows: 14/14; forcedAlignment: 1/1; history: 5/5; models: 1/1; music: 13/13; productions: 11/11; pronunciationDictionaries: 9/9; samples: 1/1; serviceAccounts: 7/7; speechEngine: 6/6; speechToSpeech: 2/2; speechToText: 3/3; studio: 23/23; textToDialogue: 4/4; textToSoundEffects: 1/1; textToSpeech: 4/4; textToVoice: 5/5; tokens: 1/1; usage: 1/1; user: 2/2; voices: 36/36; webhooks: 4/4; workspace: 20/20; workspaces: 1/1

## core (beyond REST, not counted in totals)
- [x] Typed `*APIError` + sentinels (400/401/403/404/409/422/425/429/5xx) and `ErrTimeout` (ElevenLabsError/ElevenLabsTimeoutError)
- [x] Retries with exponential backoff + jitter (maxRetries, Retry-After, X-RateLimit-Reset)
- [x] Per-request options (headers, query params, body params, timeout, retries, API key)
- [x] Base URL environments (production, US, EU, India, Singapore residency)
- [x] Pagination iterators (`iter.Seq2`)
- [x] Streaming multipart uploads and streaming binary / NDJSON responses
- [ ] WebSocket realtime: TTS stream-input, multi-stream-input, dialogue stream-input, STT realtime, dubbing/translation realtime
- [ ] Conversational AI websocket
- [ ] Speech engine WebSocket server / session helpers (src/wrapper/speech-engine; deferred to the realtime batch, needs a WebSocket dependency)
- [x] Webhooks signature verification (src/wrapper/webhooks.ts)
- [x] Flows `createAndWait` polling helpers (src/wrapper/flows.ts)
- [x] Music multipart helper (src/wrapper/music.ts)
- [x] speechToText wrapper helpers (src/wrapper/speechToText.ts)

## assets (4)
- [x] `GET /v1/assets` — assets.list
- [x] `POST /v1/assets` — assets.create
- [x] `GET /v1/assets/{asset_id}` — assets.get
- [x] `DELETE /v1/assets/{asset_id}` — assets.delete

## audioEffects (1)
- [ ] `GET /v1/audio-effects/catalog` — audioEffects.getCatalog

## audioIsolation (4)
- [x] `POST /v1/audio-isolation` — audioIsolation.convert (stream/binary)
- [x] `GET /v1/audio-isolation/history` — audioIsolation.list
- [x] `DELETE /v1/audio-isolation/history/{history_item_id}` — audioIsolation.delete
- [x] `POST /v1/audio-isolation/stream` — audioIsolation.stream (stream/binary)

## audioNative (4)
- [x] `POST /v1/audio-native` — audioNative.create
- [x] `GET /v1/audio-native/{project_id}/settings` — audioNative.getSettings
- [x] `POST /v1/audio-native/{project_id}/content` — audioNative.update
- [x] `POST /v1/audio-native/content` — audioNative.updateContentFromUrl

## conversationalAi (186)
- [x] `POST /v1/convai/knowledge-base` — conversationalAi.addToKnowledgeBase
- [x] `GET /v1/convai/knowledge-base/rag-index` — conversationalAi.ragIndexOverview
- [x] `GET /v1/convai/knowledge-base/{documentation_id}/rag-index` — conversationalAi.getDocumentRagIndexes
- [x] `DELETE /v1/convai/knowledge-base/{documentation_id}/rag-index/{rag_index_id}` — conversationalAi.deleteDocumentRagIndex
- [x] `POST /v1/convai/agents/create` — conversationalAi.agents.create
- [x] `GET /v1/convai/agents/{agent_id}` — conversationalAi.agents.get
- [x] `DELETE /v1/convai/agents/{agent_id}` — conversationalAi.agents.delete
- [x] `PATCH /v1/convai/agents/{agent_id}` — conversationalAi.agents.update
- [x] `GET /v1/convai/agents` — conversationalAi.agents.list
- [x] `POST /v1/convai/agents/{agent_id}/duplicate` — conversationalAi.agents.duplicate
- [x] `POST /v1/convai/agents/{agent_id}/simulate-conversation` — conversationalAi.agents.simulateConversation
- [x] `POST /v1/convai/agents/{agent_id}/simulate-conversation/stream` — conversationalAi.agents.simulateConversationStream (stream/binary)
- [x] `POST /v1/convai/agents/{agent_id}/run-tests` — conversationalAi.agents.runTests
- [x] `GET /v1/convai/agents/{agent_id}/branches` — conversationalAi.agents.branches.list
- [x] `POST /v1/convai/agents/{agent_id}/branches` — conversationalAi.agents.branches.create
- [x] `GET /v1/convai/agents/{agent_id}/branches/{branch_id}` — conversationalAi.agents.branches.get
- [x] `PATCH /v1/convai/agents/{agent_id}/branches/{branch_id}` — conversationalAi.agents.branches.update
- [x] `GET /v1/convai/agents/{agent_id}/branches/{source_branch_id}/merge-preview` — conversationalAi.agents.branches.previewMerge
- [x] `POST /v1/convai/agents/{agent_id}/branches/{source_branch_id}/merge` — conversationalAi.agents.branches.merge
- [x] `GET /v1/convai/agents/{agent_id}/branches/{branch_id}/rebase-preview` — conversationalAi.agents.branches.previewRebase
- [x] `POST /v1/convai/agents/{agent_id}/branches/{branch_id}/rebase` — conversationalAi.agents.branches.rebase
- [x] `GET /v1/convai/agents/{agent_id}/deployments` — conversationalAi.agents.deployments.list
- [x] `POST /v1/convai/agents/{agent_id}/deployments` — conversationalAi.agents.deployments.create
- [x] `POST /v1/convai/agents/{agent_id}/drafts` — conversationalAi.agents.drafts.create
- [x] `DELETE /v1/convai/agents/{agent_id}/drafts` — conversationalAi.agents.drafts.delete
- [x] `POST /v1/convai/agents/{agent_id}/hold-audio` — conversationalAi.agents.holdAudio.create
- [x] `DELETE /v1/convai/agents/{agent_id}/hold-audio` — conversationalAi.agents.holdAudio.delete
- [x] `GET /v1/convai/agent/{agent_id}/knowledge-base/size` — conversationalAi.agents.knowledgeBase.size
- [x] `GET /v1/convai/agents/{agent_id}/link` — conversationalAi.agents.link.get
- [x] `POST /v1/convai/agent/{agent_id}/llm-usage/calculate` — conversationalAi.agents.llmUsage.calculate
- [x] `GET /v1/convai/agents/{agent_id}/merge-proposals` — conversationalAi.agents.mergeProposals.list
- [x] `POST /v1/convai/agents/{agent_id}/merge-proposals` — conversationalAi.agents.mergeProposals.create
- [x] `GET /v1/convai/agents/{agent_id}/merge-proposals/{merge_proposal_id}` — conversationalAi.agents.mergeProposals.get
- [x] `PATCH /v1/convai/agents/{agent_id}/merge-proposals/{merge_proposal_id}` — conversationalAi.agents.mergeProposals.update
- [x] `POST /v1/convai/agents/{agent_id}/merge-proposals/{merge_proposal_id}/reviews` — conversationalAi.agents.mergeProposals.submitReview
- [x] `POST /v1/convai/agents/{agent_id}/merge-proposals/{merge_proposal_id}/comments` — conversationalAi.agents.mergeProposals.addComment
- [x] `POST /v1/convai/agents/{agent_id}/merge-proposals/{merge_proposal_id}/merge` — conversationalAi.agents.mergeProposals.merge
- [x] `GET /v1/convai/agents/{agent_id}/branches/{branch_id}/procedures` — conversationalAi.agents.procedures.list
- [x] `POST /v1/convai/agents/{agent_id}/branches/{branch_id}/procedures` — conversationalAi.agents.procedures.create
- [x] `POST /v1/convai/agents/{agent_id}/branches/{branch_id}/procedures/compile` — conversationalAi.agents.procedures.compile
- [x] `GET /v1/convai/agents/{agent_id}/branches/{branch_id}/procedures/{procedure_id}` — conversationalAi.agents.procedures.get
- [x] `DELETE /v1/convai/agents/{agent_id}/branches/{branch_id}/procedures/{procedure_id}` — conversationalAi.agents.procedures.remove
- [x] `GET /v1/convai/agents/{agent_id}/branches/{branch_id}/procedures/{procedure_id}/draft` — conversationalAi.agents.procedures.drafts.get
- [x] `DELETE /v1/convai/agents/{agent_id}/branches/{branch_id}/procedures/{procedure_id}/draft` — conversationalAi.agents.procedures.drafts.delete
- [x] `PATCH /v1/convai/agents/{agent_id}/branches/{branch_id}/procedures/{procedure_id}/draft` — conversationalAi.agents.procedures.drafts.update
- [x] `GET /v1/convai/agents/summaries` — conversationalAi.agents.summaries.get
- [x] `GET /v1/convai/agents/{agent_id}/versions/{version_id}` — conversationalAi.agents.versions.get
- [x] `GET /v1/convai/agents/{agent_id}/widget` — conversationalAi.agents.widget.get
- [x] `POST /v1/convai/agents/{agent_id}/avatar` — conversationalAi.agents.widget.avatar.create
- [ ] `GET /v1/convai/analytics/live-count` — conversationalAi.analytics.liveCount.get
- [ ] `POST /v1/convai/batch-calling/submit` — conversationalAi.batchCalls.create
- [ ] `GET /v1/convai/batch-calling/workspace` — conversationalAi.batchCalls.list
- [ ] `GET /v1/convai/batch-calling/{batch_id}` — conversationalAi.batchCalls.get
- [ ] `DELETE /v1/convai/batch-calling/{batch_id}` — conversationalAi.batchCalls.delete
- [ ] `POST /v1/convai/batch-calling/{batch_id}/cancel` — conversationalAi.batchCalls.cancel
- [ ] `POST /v1/convai/batch-calling/{batch_id}/retry` — conversationalAi.batchCalls.retry
- [ ] `GET /v1/convai/batch-calling/{batch_id}/export` — conversationalAi.batchCalls.export (stream/binary)
- [x] `GET /v1/convai/conversation/get-signed-url` — conversationalAi.conversations.getSignedUrl
- [x] `GET /v1/convai/conversation/token` — conversationalAi.conversations.getWebrtcToken
- [x] `GET /v1/convai/conversations` — conversationalAi.conversations.list
- [x] `GET /v1/convai/conversations/resolve` — conversationalAi.conversations.resolve
- [x] `GET /v1/convai/conversations/{conversation_id}` — conversationalAi.conversations.get
- [x] `DELETE /v1/convai/conversations/{conversation_id}` — conversationalAi.conversations.delete
- [x] `GET /v1/convai/conversations/{conversation_id}/summary` — conversationalAi.conversations.getSummary
- [x] `GET /v1/convai/conversations/{conversation_id}/sip-messages` — conversationalAi.conversations.getSipMessages
- [x] `POST /v1/convai/conversations/{conversation_id}/analysis/run` — conversationalAi.conversations.analysis.run
- [x] `POST /v1/convai/conversations/{conversation_id}/analysis/evaluations/run` — conversationalAi.conversations.analysis.runEvaluation
- [x] `GET /v1/convai/conversations/{conversation_id}/audio` — conversationalAi.conversations.audio.get (stream/binary)
- [x] `POST /v1/convai/conversations/{conversation_id}/feedback` — conversationalAi.conversations.feedback.create
- [x] `POST /v1/convai/conversations/{conversation_id}/files` — conversationalAi.conversations.files.create
- [x] `DELETE /v1/convai/conversations/{conversation_id}/files/{file_id}` — conversationalAi.conversations.files.delete
- [x] `GET /v1/convai/conversations/messages/text-search` — conversationalAi.conversations.messages.textSearch
- [x] `GET /v1/convai/conversations/messages/smart-search` — conversationalAi.conversations.messages.search
- [x] `POST /v1/convai/conversations/{conversation_id}/tags` — conversationalAi.conversations.tags.assign
- [x] `DELETE /v1/convai/conversations/{conversation_id}/tags/{tag_id}` — conversationalAi.conversations.tags.unassign
- [x] `GET /v1/convai/tags` — conversationalAi.conversations.tags.list
- [x] `POST /v1/convai/tags` — conversationalAi.conversations.tags.create
- [x] `GET /v1/convai/tags/{tag_id}` — conversationalAi.conversations.tags.get
- [x] `DELETE /v1/convai/tags/{tag_id}` — conversationalAi.conversations.tags.delete
- [x] `PATCH /v1/convai/tags/{tag_id}` — conversationalAi.conversations.tags.update
- [x] `GET /v1/convai/agents/{agent_id}/topics` — conversationalAi.conversations.topics.get
- [x] `GET /v1/convai/settings/dashboard` — conversationalAi.dashboard.settings.get
- [x] `PATCH /v1/convai/settings/dashboard` — conversationalAi.dashboard.settings.update
- [ ] `POST /v1/convai/exotel/outbound-call` — conversationalAi.exotel.outboundCall
- [x] `GET /v1/convai/knowledge-base` — conversationalAi.knowledgeBase.list
- [x] `POST /v1/convai/knowledge-base/rag-index` — conversationalAi.knowledgeBase.getOrCreateRagIndexes
- [x] `GET /v1/convai/knowledge-base/search` — conversationalAi.knowledgeBase.search
- [x] `GET /v1/convai/knowledge-base/crawl` — conversationalAi.knowledgeBase.crawlJobs.list
- [x] `POST /v1/convai/knowledge-base/crawl` — conversationalAi.knowledgeBase.crawlJobs.create
- [x] `GET /v1/convai/knowledge-base/crawl/{crawl_job_id}` — conversationalAi.knowledgeBase.crawlJobs.get
- [x] `POST /v1/convai/knowledge-base/crawl/{crawl_job_id}/cancel` — conversationalAi.knowledgeBase.crawlJobs.cancel
- [x] `PATCH /v1/convai/knowledge-base/{documentation_id}/update-file` — conversationalAi.knowledgeBase.document.updateFile
- [x] `POST /v1/convai/knowledge-base/{documentation_id}/refresh` — conversationalAi.knowledgeBase.document.refresh
- [x] `POST /v1/convai/knowledge-base/{documentation_id}/rag-index` — conversationalAi.knowledgeBase.document.computeRagIndex
- [x] `POST /v1/convai/knowledge-base/url` — conversationalAi.knowledgeBase.documents.createFromUrl
- [x] `POST /v1/convai/knowledge-base/file` — conversationalAi.knowledgeBase.documents.createFromFile
- [x] `POST /v1/convai/knowledge-base/text` — conversationalAi.knowledgeBase.documents.createFromText
- [x] `POST /v1/convai/knowledge-base/folder` — conversationalAi.knowledgeBase.documents.createFolder
- [x] `GET /v1/convai/knowledge-base/{documentation_id}` — conversationalAi.knowledgeBase.documents.get
- [x] `DELETE /v1/convai/knowledge-base/{documentation_id}` — conversationalAi.knowledgeBase.documents.delete
- [x] `PATCH /v1/convai/knowledge-base/{documentation_id}` — conversationalAi.knowledgeBase.documents.update
- [x] `GET /v1/convai/knowledge-base/{documentation_id}/dependent-agents` — conversationalAi.knowledgeBase.documents.getAgents
- [x] `POST /v1/convai/knowledge-base/dependent-agents` — conversationalAi.knowledgeBase.documents.getBulkAgents
- [x] `GET /v1/convai/knowledge-base/{documentation_id}/content` — conversationalAi.knowledgeBase.documents.getContent
- [x] `GET /v1/convai/knowledge-base/{documentation_id}/source-file-url` — conversationalAi.knowledgeBase.documents.getSourceFileUrl
- [x] `POST /v1/convai/knowledge-base/{document_id}/move` — conversationalAi.knowledgeBase.documents.move
- [x] `POST /v1/convai/knowledge-base/bulk-move` — conversationalAi.knowledgeBase.documents.bulkMove
- [x] `POST /v1/convai/knowledge-base/bulk-delete` — conversationalAi.knowledgeBase.documents.bulkDelete
- [x] `GET /v1/convai/knowledge-base/{documentation_id}/chunk/{chunk_id}` — conversationalAi.knowledgeBase.documents.chunk.get
- [x] `GET /v1/convai/knowledge-base/{documentation_id}/chunks` — conversationalAi.knowledgeBase.documents.chunks.list
- [x] `GET /v1/convai/knowledge-base/summaries` — conversationalAi.knowledgeBase.documents.summaries.get
- [ ] `GET /v1/convai/llm/list` — conversationalAi.llm.list
- [ ] `POST /v1/convai/llm-usage/calculate` — conversationalAi.llmUsage.calculate
- [ ] `GET /v1/convai/mcp-servers` — conversationalAi.mcpServers.list
- [ ] `POST /v1/convai/mcp-servers` — conversationalAi.mcpServers.create
- [ ] `GET /v1/convai/mcp-servers/{mcp_server_id}` — conversationalAi.mcpServers.get
- [ ] `DELETE /v1/convai/mcp-servers/{mcp_server_id}` — conversationalAi.mcpServers.delete
- [ ] `PATCH /v1/convai/mcp-servers/{mcp_server_id}` — conversationalAi.mcpServers.update
- [ ] `PATCH /v1/convai/mcp-servers/{mcp_server_id}/approval-policy` — conversationalAi.mcpServers.approvalPolicy.update
- [ ] `POST /v1/convai/mcp-servers/{mcp_server_id}/tool-approvals` — conversationalAi.mcpServers.toolApprovals.create
- [ ] `DELETE /v1/convai/mcp-servers/{mcp_server_id}/tool-approvals/{tool_name}` — conversationalAi.mcpServers.toolApprovals.delete
- [ ] `POST /v1/convai/mcp-servers/{mcp_server_id}/tool-configs` — conversationalAi.mcpServers.toolConfigs.create
- [ ] `GET /v1/convai/mcp-servers/{mcp_server_id}/tool-configs/{tool_name}` — conversationalAi.mcpServers.toolConfigs.get
- [ ] `DELETE /v1/convai/mcp-servers/{mcp_server_id}/tool-configs/{tool_name}` — conversationalAi.mcpServers.toolConfigs.delete
- [ ] `PATCH /v1/convai/mcp-servers/{mcp_server_id}/tool-configs/{tool_name}` — conversationalAi.mcpServers.toolConfigs.update
- [ ] `GET /v1/convai/mcp-servers/{mcp_server_id}/tools` — conversationalAi.mcpServers.tools.list
- [ ] `GET /v1/convai/phone-numbers` — conversationalAi.phoneNumbers.list
- [ ] `POST /v1/convai/phone-numbers` — conversationalAi.phoneNumbers.create
- [ ] `GET /v1/convai/phone-numbers/{phone_number_id}` — conversationalAi.phoneNumbers.get
- [ ] `DELETE /v1/convai/phone-numbers/{phone_number_id}` — conversationalAi.phoneNumbers.delete
- [ ] `PATCH /v1/convai/phone-numbers/{phone_number_id}` — conversationalAi.phoneNumbers.update
- [ ] `GET /v1/convai/v2/phone-numbers` — conversationalAi.phoneNumbers.listV2
- [ ] `GET /v1/convai/phone-numbers/{phone_number_id}/sip-messages` — conversationalAi.phoneNumbers.getSipMessages
- [ ] `POST /v1/convai/realtime/client_secrets` — conversationalAi.realtime.createClientSecret
- [x] `GET /v1/convai/secrets` — conversationalAi.secrets.list
- [x] `POST /v1/convai/secrets` — conversationalAi.secrets.create
- [x] `GET /v1/convai/secrets/{secret_id}` — conversationalAi.secrets.get
- [x] `DELETE /v1/convai/secrets/{secret_id}` — conversationalAi.secrets.delete
- [x] `PATCH /v1/convai/secrets/{secret_id}` — conversationalAi.secrets.update
- [x] `GET /v1/convai/secrets/{secret_id}/dependencies/${core.url.encodePathParam(serializers.SecretDependencyResourceType.jsonOrThrow(resource_type))}` — conversationalAi.secrets.getDependencies
- [x] `GET /v1/convai/settings` — conversationalAi.settings.get
- [x] `PATCH /v1/convai/settings` — conversationalAi.settings.update
- [ ] `POST /v1/convai/sip-trunk/outbound-call` — conversationalAi.sipTrunk.outboundCall
- [x] `POST /v1/convai/agent-testing/create` — conversationalAi.tests.create
- [x] `POST /v1/convai/agent-testing/bulk-move` — conversationalAi.tests.move
- [x] `GET /v1/convai/agent-testing/{test_id}` — conversationalAi.tests.get
- [x] `PUT /v1/convai/agent-testing/{test_id}` — conversationalAi.tests.update
- [x] `DELETE /v1/convai/agent-testing/{test_id}` — conversationalAi.tests.delete
- [x] `POST /v1/convai/agent-testing/summaries` — conversationalAi.tests.summaries
- [x] `GET /v1/convai/agent-testing` — conversationalAi.tests.list
- [x] `POST /v1/convai/agent-testing/folders` — conversationalAi.tests.folders.create
- [x] `GET /v1/convai/agent-testing/folders/{folder_id}` — conversationalAi.tests.folders.get
- [x] `DELETE /v1/convai/agent-testing/folders/{folder_id}` — conversationalAi.tests.folders.delete
- [x] `PATCH /v1/convai/agent-testing/folders/{folder_id}` — conversationalAi.tests.folders.update
- [x] `GET /v1/convai/test-invocations` — conversationalAi.tests.invocations.list
- [x] `GET /v1/convai/test-invocations/{test_invocation_id}` — conversationalAi.tests.invocations.get
- [x] `PATCH /v1/convai/test-invocations/{test_invocation_id}` — conversationalAi.tests.invocations.update
- [x] `POST /v1/convai/test-invocations/{test_invocation_id}/cancel` — conversationalAi.tests.invocations.cancel
- [x] `POST /v1/convai/test-invocations/{test_invocation_id}/resubmit` — conversationalAi.tests.invocations.resubmit
- [x] `GET /v1/convai/tools` — conversationalAi.tools.list
- [x] `POST /v1/convai/tools` — conversationalAi.tools.create
- [x] `GET /v1/convai/tools/{tool_id}` — conversationalAi.tools.get
- [x] `DELETE /v1/convai/tools/{tool_id}` — conversationalAi.tools.delete
- [x] `PATCH /v1/convai/tools/{tool_id}` — conversationalAi.tools.update
- [x] `GET /v1/convai/tools/{tool_id}/dependent-agents` — conversationalAi.tools.getDependentAgents
- [x] `GET /v1/convai/tools/{tool_id}/executions` — conversationalAi.tools.executions.get
- [ ] `GET /v1/convai/agents/{agent_id}/triage-tickets` — conversationalAi.triageTickets.list
- [ ] `POST /v1/convai/agents/{agent_id}/triage-tickets` — conversationalAi.triageTickets.createManual
- [ ] `GET /v1/convai/triage-tickets` — conversationalAi.triageTickets.listForWorkspace
- [ ] `POST /v1/convai/triage-tickets` — conversationalAi.triageTickets.create
- [ ] `GET /v1/convai/agents/{agent_id}/triage-tickets/assignable-users` — conversationalAi.triageTickets.listAssignableUsers
- [ ] `GET /v1/convai/triage-tickets/{agentqa_ticket_id}` — conversationalAi.triageTickets.get
- [ ] `DELETE /v1/convai/triage-tickets/{agentqa_ticket_id}` — conversationalAi.triageTickets.delete
- [ ] `PATCH /v1/convai/triage-tickets/{agentqa_ticket_id}` — conversationalAi.triageTickets.update
- [ ] `POST /v1/convai/triage-tickets/{agentqa_ticket_id}/comments` — conversationalAi.triageTickets.addComment
- [ ] `POST /v1/convai/triage-tickets/{agentqa_ticket_id}/merge` — conversationalAi.triageTickets.merge
- [ ] `POST /v1/convai/triage-tickets/{agentqa_ticket_id}/turn-comments` — conversationalAi.triageTickets.addTurnComment
- [ ] `POST /v1/convai/twilio/outbound-call` — conversationalAi.twilio.outboundCall
- [ ] `POST /v1/convai/twilio/register-call` — conversationalAi.twilio.registerCall
- [ ] `GET /v1/convai/users` — conversationalAi.users.list
- [ ] `POST /v1/convai/whatsapp/outbound-call` — conversationalAi.whatsapp.outboundCall
- [ ] `POST /v1/convai/whatsapp/outbound-message` — conversationalAi.whatsapp.outboundMessage
- [ ] `GET /v1/convai/whatsapp-accounts/{phone_number_id}` — conversationalAi.whatsappAccounts.get
- [ ] `DELETE /v1/convai/whatsapp-accounts/{phone_number_id}` — conversationalAi.whatsappAccounts.delete
- [ ] `PATCH /v1/convai/whatsapp-accounts/{phone_number_id}` — conversationalAi.whatsappAccounts.update
- [ ] `GET /v1/convai/whatsapp-accounts` — conversationalAi.whatsappAccounts.list

## dubbing (37)
- [x] `GET /v1/dubbing` — dubbing.list
- [x] `POST /v1/dubbing` — dubbing.create
- [x] `GET /v1/dubbing/{dubbing_id}` — dubbing.get
- [x] `DELETE /v1/dubbing/{dubbing_id}` — dubbing.delete
- [x] `GET /v1/dubbing/{dubbing_id}/audio/{language_code}` — dubbing.audio.get (stream/binary)
- [x] `GET /v1/dubbing/project` — dubbing.project.list
- [x] `POST /v1/dubbing/project` — dubbing.project.create
- [x] `GET /v1/dubbing/project/{project_id}` — dubbing.project.get
- [x] `DELETE /v1/dubbing/project/{project_id}` — dubbing.project.delete
- [x] `GET /v1/dubbing/project/{project_id}/language` — dubbing.project.language.list
- [x] `POST /v1/dubbing/project/{project_id}/language` — dubbing.project.language.create
- [x] `GET /v1/dubbing/project/{project_id}/language/{language_id}` — dubbing.project.language.get
- [x] `DELETE /v1/dubbing/project/{project_id}/language/{language_id}` — dubbing.project.language.delete
- [x] `GET /v1/dubbing/project/{project_id}/language/{language_id}/transcript` — dubbing.project.language.transcript.get
- [x] `PATCH /v1/dubbing/project/{project_id}/language/{language_id}/transcript/segment/{segment_id}` — dubbing.project.language.transcript.updateSegment
- [x] `PATCH /v1/dubbing/project/{project_id}/language/{language_id}/transcript/segments` — dubbing.project.language.transcript.updateSegments
- [x] `POST /v1/dubbing/project/{project_id}/language/{language_id}/transcript/regenerate` — dubbing.project.language.transcript.regenerate
- [x] `GET /v1/dubbing/project/{project_id}/transcript` — dubbing.project.transcript.get
- [x] `DELETE /v1/dubbing/project/{project_id}/transcript/segment/{segment_id}` — dubbing.project.transcript.deleteSegment
- [x] `PATCH /v1/dubbing/project/{project_id}/transcript/segment/{segment_id}` — dubbing.project.transcript.updateSegment
- [x] `PATCH /v1/dubbing/project/{project_id}/transcript/segments` — dubbing.project.transcript.updateSegments
- [x] `POST /v1/dubbing/project/{project_id}/transcript/segment` — dubbing.project.transcript.createSegment
- [x] `GET /v1/dubbing/resource/{dubbing_id}` — dubbing.resource.get
- [x] `POST /v1/dubbing/resource/{dubbing_id}/migrate-segments` — dubbing.resource.migrateSegments
- [x] `POST /v1/dubbing/resource/{dubbing_id}/transcribe` — dubbing.resource.transcribe
- [x] `POST /v1/dubbing/resource/{dubbing_id}/translate` — dubbing.resource.translate
- [x] `POST /v1/dubbing/resource/{dubbing_id}/dub` — dubbing.resource.dub
- [x] `POST /v1/dubbing/resource/{dubbing_id}/render/{language}` — dubbing.resource.render
- [x] `POST /v1/dubbing/resource/{dubbing_id}/language` — dubbing.resource.language.add
- [x] `PATCH /v1/dubbing/resource/{dubbing_id}/segment/{segment_id}/{language}` — dubbing.resource.segment.update
- [x] `DELETE /v1/dubbing/resource/{dubbing_id}/segment/{segment_id}` — dubbing.resource.segment.delete
- [x] `PATCH /v1/dubbing/resource/{dubbing_id}/speaker/{speaker_id}` — dubbing.resource.speaker.update
- [x] `POST /v1/dubbing/resource/{dubbing_id}/speaker` — dubbing.resource.speaker.create
- [x] `GET /v1/dubbing/resource/{dubbing_id}/speaker/{speaker_id}/similar-voices` — dubbing.resource.speaker.findSimilarVoices
- [x] `POST /v1/dubbing/resource/{dubbing_id}/speaker/{speaker_id}/segment` — dubbing.resource.speaker.segment.create
- [x] `GET /v1/dubbing/{dubbing_id}/transcript/{language_code}` — dubbing.transcript.getTranscriptForDub
- [x] `GET /v1/dubbing/{dubbing_id}/transcripts/{language_code}/format/${core.url.encodePathParam(serializers.dubbing.TranscriptsGetRequestFormatType.jsonOrThrow(format_type))}` — dubbing.transcripts.get

## environmentVariables (5)
- [x] `GET /v1/convai/environment-variables` — environmentVariables.list
- [x] `POST /v1/convai/environment-variables` — environmentVariables.create
- [x] `GET /v1/convai/environment-variables/{env_var_id}` — environmentVariables.get
- [x] `DELETE /v1/convai/environment-variables/{env_var_id}` — environmentVariables.delete
- [x] `PATCH /v1/convai/environment-variables/{env_var_id}` — environmentVariables.update

## environments (3)
- [x] `GET /v1/convai/environments` — environments.list
- [x] `POST /v1/convai/environments` — environments.create
- [x] `DELETE /v1/convai/environments/{environment}` — environments.delete

## flows (14)
- [x] `GET /v1/flows/image` — flows.image.list
- [x] `POST /v1/flows/image` — flows.image.create
- [x] `GET /v1/flows/image/{generation_id}` — flows.image.get
- [x] `GET /v1/flows/templates` — flows.templates.list
- [x] `GET /v1/flows/templates/{template_id}` — flows.templates.get
- [x] `GET /v1/flows/templates/{template_id}/runs` — flows.templates.runs.list
- [x] `POST /v1/flows/templates/{template_id}/runs` — flows.templates.runs.create
- [x] `GET /v1/flows/templates/{template_id}/runs/{run_id}` — flows.templates.runs.get
- [x] `GET /v1/flows/text-to-speech` — flows.textToSpeech.list
- [x] `POST /v1/flows/text-to-speech` — flows.textToSpeech.create
- [x] `GET /v1/flows/text-to-speech/{generation_id}` — flows.textToSpeech.get
- [x] `GET /v1/flows/video` — flows.video.list
- [x] `POST /v1/flows/video` — flows.video.create
- [x] `GET /v1/flows/video/{generation_id}` — flows.video.get

## forcedAlignment (1)
- [x] `POST /v1/forced-alignment` — forcedAlignment.create

## history (5)
- [x] `GET /v1/history` — history.list
- [x] `GET /v1/history/{history_item_id}` — history.get
- [x] `DELETE /v1/history/{history_item_id}` — history.delete
- [x] `GET /v1/history/{history_item_id}/audio` — history.getAudio (stream/binary)
- [x] `POST /v1/history/download` — history.download (stream/binary)

## models (1)
- [x] `GET /v1/models` — models.list

## music (13)
- [x] `POST /v1/music/video-to-music` — music.videoToMusic (stream/binary)
- [x] `POST /v1/music` — music.compose (stream/binary)
- [x] `POST /v1/music/detailed` — music.composeDetailed (stream/binary)
- [x] `POST /v1/music/detailed/stream` — music.composeDetailedStream (stream/binary)
- [x] `POST /v1/music/stream` — music.stream (stream/binary)
- [x] `POST /v1/music/upload` — music.upload
- [x] `POST /v1/music/stem-separation` — music.separateStems (stream/binary)
- [x] `POST /v1/music/plan` — music.compositionPlan.create
- [x] `GET /v1/music/finetunes` — music.finetunes.list
- [x] `POST /v1/music/finetunes` — music.finetunes.create
- [x] `GET /v1/music/finetunes/{finetune_id}` — music.finetunes.get
- [x] `DELETE /v1/music/finetunes/{finetune_id}` — music.finetunes.delete
- [x] `PATCH /v1/music/finetunes/{finetune_id}` — music.finetunes.update

## productions (11)
- [x] `GET /v1/productions/orders` — productions.orders.list
- [x] `POST /v1/productions/orders` — productions.orders.create
- [x] `GET /v1/productions/orders/${core.url.encodePathParam(serializers.OrderId.jsonOrThrow(order_id))}` — productions.orders.get
- [x] `PATCH /v1/productions/orders/${core.url.encodePathParam(serializers.OrderId.jsonOrThrow(order_id))}` — productions.orders.update
- [x] `POST /v1/productions/orders/${core.url.encodePathParam(serializers.OrderId.jsonOrThrow(order_id))}/submit` — productions.orders.submit
- [x] `GET /v1/productions/orders/${core.url.encodePathParam(serializers.OrderId.jsonOrThrow(order_id))}/deliverables` — productions.orders.deliverables.list
- [x] `POST /v1/productions/orders/${core.url.encodePathParam(serializers.OrderId.jsonOrThrow(order_id))}/items` — productions.orders.items.upsert
- [x] `DELETE /v1/productions/orders/${core.url.encodePathParam(serializers.OrderId.jsonOrThrow(order_id))}/items/${core.url.encodePathParam(serializers.ItemId.jsonOrThrow(item_id))}` — productions.orders.items.remove
- [x] `GET /v1/productions/orders/languages/${core.url.encodePathParam(serializers.OrderItemKind.jsonOrThrow(order_item_kind))}` — productions.orders.languages.list
- [x] `POST /v1/productions/orders/${core.url.encodePathParam(serializers.OrderId.jsonOrThrow(order_id))}/media` — productions.orders.media.register
- [x] `GET /v1/productions/orders/${core.url.encodePathParam(serializers.OrderId.jsonOrThrow(order_id))}/media/${core.url.encodePathParam(serializers.MediaId.jsonOrThrow(media_id))}` — productions.orders.media.get

## pronunciationDictionaries (9)
- [x] `POST /v1/pronunciation-dictionaries/add-from-file` — pronunciationDictionaries.createFromFile
- [x] `POST /v1/pronunciation-dictionaries/add-from-rules` — pronunciationDictionaries.createFromRules
- [x] `GET /v1/pronunciation-dictionaries/{pronunciation_dictionary_id}` — pronunciationDictionaries.get
- [x] `PATCH /v1/pronunciation-dictionaries/{pronunciation_dictionary_id}` — pronunciationDictionaries.update
- [x] `GET /v1/pronunciation-dictionaries/{dictionary_id}/{version_id}/download` — pronunciationDictionaries.download (stream/binary)
- [x] `GET /v1/pronunciation-dictionaries` — pronunciationDictionaries.list
- [x] `POST /v1/pronunciation-dictionaries/{pronunciation_dictionary_id}/set-rules` — pronunciationDictionaries.rules.set
- [x] `POST /v1/pronunciation-dictionaries/{pronunciation_dictionary_id}/add-rules` — pronunciationDictionaries.rules.add
- [x] `POST /v1/pronunciation-dictionaries/{pronunciation_dictionary_id}/remove-rules` — pronunciationDictionaries.rules.remove

## samples (1)
- [x] `DELETE /v1/voices/{voice_id}/samples/{sample_id}` — samples.delete

## serviceAccounts (7)
- [x] `GET /v1/service-accounts` — serviceAccounts.list
- [x] `POST /v1/service-accounts` — serviceAccounts.create
- [x] `GET /v1/service-accounts/{service_account_user_id}/api-keys` — serviceAccounts.apiKeys.list
- [x] `POST /v1/service-accounts/{service_account_user_id}/api-keys` — serviceAccounts.apiKeys.create
- [x] `DELETE /v1/service-accounts/{service_account_user_id}/api-keys/{api_key_id}` — serviceAccounts.apiKeys.delete
- [x] `PATCH /v1/service-accounts/{service_account_user_id}/api-keys/{api_key_id}` — serviceAccounts.apiKeys.update
- [x] `POST /v1/service-accounts/{service_account_user_id}/api-keys/{api_key_id}/disable` — serviceAccounts.apiKeys.disable

## speechEngine (6)
- [x] `GET /v1/speech-engine` — speechEngine.list
- [x] `POST /v1/speech-engine` — speechEngine.create
- [x] `GET /v1/speech-engine/{speech_engine_id}` — speechEngine.get
- [x] `DELETE /v1/speech-engine/{speech_engine_id}` — speechEngine.delete
- [x] `PATCH /v1/speech-engine/{speech_engine_id}` — speechEngine.update
- [x] `POST /v1/speech-engine/{speech_engine_id}/duplicate` — speechEngine.duplicate

## speechToSpeech (2)
- [x] `POST /v1/voice-changer/{voice_id}` — speechToSpeech.convert (stream/binary)
- [x] `POST /v1/voice-changer/{voice_id}/stream` — speechToSpeech.stream (stream/binary)

## speechToText (3)
- [x] `POST /v1/speech-to-text` — speechToText.convert
- [x] `GET /v1/speech-to-text/transcripts/{transcription_id}` — speechToText.transcripts.get
- [x] `DELETE /v1/speech-to-text/transcripts/{transcription_id}` — speechToText.transcripts.delete

## studio (23)
- [x] `POST /v1/studio/podcasts` — studio.createPodcast
- [x] `GET /v1/studio/projects` — studio.projects.list
- [x] `POST /v1/studio/projects` — studio.projects.create
- [x] `GET /v1/studio/projects/{project_id}` — studio.projects.get
- [x] `POST /v1/studio/projects/{project_id}` — studio.projects.update
- [x] `DELETE /v1/studio/projects/{project_id}` — studio.projects.delete
- [x] `POST /v1/studio/projects/{project_id}/convert` — studio.projects.convert
- [x] `GET /v1/studio/projects/{project_id}/muted-tracks` — studio.projects.getMutedTracks
- [x] `GET /v1/studio/projects/{project_id}/chapters` — studio.projects.chapters.list
- [x] `POST /v1/studio/projects/{project_id}/chapters` — studio.projects.chapters.create
- [x] `GET /v1/studio/projects/{project_id}/chapters/{chapter_id}` — studio.projects.chapters.get
- [x] `POST /v1/studio/projects/{project_id}/chapters/{chapter_id}` — studio.projects.chapters.update
- [x] `DELETE /v1/studio/projects/{project_id}/chapters/{chapter_id}` — studio.projects.chapters.delete
- [x] `POST /v1/studio/projects/{project_id}/chapters/{chapter_id}/convert` — studio.projects.chapters.convert
- [x] `GET /v1/studio/projects/{project_id}/chapters/{chapter_id}/snapshots` — studio.projects.chapters.snapshots.list
- [x] `GET /v1/studio/projects/{project_id}/chapters/{chapter_id}/snapshots/{chapter_snapshot_id}` — studio.projects.chapters.snapshots.get
- [x] `POST /v1/studio/projects/{project_id}/chapters/{chapter_id}/snapshots/{chapter_snapshot_id}/stream` — studio.projects.chapters.snapshots.stream (stream/binary)
- [x] `POST /v1/studio/projects/{project_id}/content` — studio.projects.content.update
- [x] `POST /v1/studio/projects/{project_id}/pronunciation-dictionaries` — studio.projects.pronunciationDictionaries.create
- [x] `GET /v1/studio/projects/{project_id}/snapshots` — studio.projects.snapshots.list
- [x] `GET /v1/studio/projects/{project_id}/snapshots/{project_snapshot_id}` — studio.projects.snapshots.get
- [x] `POST /v1/studio/projects/{project_id}/snapshots/{project_snapshot_id}/stream` — studio.projects.snapshots.stream (stream/binary)
- [x] `POST /v1/studio/projects/{project_id}/snapshots/{project_snapshot_id}/archive` — studio.projects.snapshots.streamArchive (stream/binary)

## textToDialogue (4)
- [x] `POST /v1/text-to-dialogue` — textToDialogue.convert (stream/binary)
- [x] `POST /v1/text-to-dialogue/stream` — textToDialogue.stream (stream/binary)
- [x] `POST /v1/text-to-dialogue/stream/with-timestamps` — textToDialogue.streamWithTimestamps (stream/binary)
- [x] `POST /v1/text-to-dialogue/with-timestamps` — textToDialogue.convertWithTimestamps

## textToSoundEffects (1)
- [x] `POST /v1/sound-generation` — textToSoundEffects.convert (stream/binary)

## textToSpeech (4)
- [x] `POST /v1/text-to-speech/{voice_id}` — textToSpeech.convert (stream/binary)
- [x] `POST /v1/text-to-speech/{voice_id}/with-timestamps` — textToSpeech.convertWithTimestamps
- [x] `POST /v1/text-to-speech/{voice_id}/stream` — textToSpeech.stream (stream/binary)
- [x] `POST /v1/text-to-speech/{voice_id}/stream/with-timestamps` — textToSpeech.streamWithTimestamps (stream/binary)

## textToVoice (5)
- [x] `POST /v1/text-to-voice/create-previews` — textToVoice.createPreviews
- [x] `POST /v1/text-to-voice` — textToVoice.create
- [x] `POST /v1/text-to-voice/design` — textToVoice.design
- [x] `POST /v1/text-to-voice/{voice_id}/remix` — textToVoice.remix
- [x] `GET /v1/text-to-voice/{generated_voice_id}/stream` — textToVoice.preview.stream (stream/binary)

## tokens (1)
- [x] `POST /v1/single-use-token/${core.url.encodePathParam(serializers.SingleUseTokenType.jsonOrThrow(token_type))}` — tokens.singleUse.create

## usage (1)
- [x] `GET /v1/usage/character-stats` — usage.get

## user (2)
- [x] `GET /v1/user` — user.get
- [x] `GET /v1/user/subscription` — user.subscription.get

## voices (36)
- [x] `GET /v1/voices` — voices.getAll
- [x] `GET /v1/voices/{voice_id}` — voices.get
- [x] `DELETE /v1/voices/{voice_id}` — voices.delete
- [x] `POST /v1/voices/{voice_id}/edit` — voices.update
- [x] `GET /v2/voices` — voices.search
- [x] `POST /v1/voices/{voice_id}/replicate-to-isolated-environment` — voices.replicateToIsolatedEnvironment
- [x] `POST /v1/voices/add/{public_user_id}/{voice_id}` — voices.share
- [x] `POST /v1/similar-voices` — voices.findSimilarVoices
- [x] `GET /v1/shared-voices` — voices.getShared
- [x] `GET /v1/voices/accents` — voices.accents.get
- [x] `GET /v1/voices/collections` — voices.collections.list
- [x] `POST /v1/voices/collections` — voices.collections.create
- [x] `GET /v1/voices/collections/{collection_id}` — voices.collections.get
- [x] `DELETE /v1/voices/collections/{collection_id}` — voices.collections.delete
- [x] `PATCH /v1/voices/collections/{collection_id}` — voices.collections.update
- [x] `POST /v1/voices/collections/{collection_id}/voice/{voice_id}` — voices.collections.addVoice
- [x] `DELETE /v1/voices/collections/{collection_id}/voice/{voice_id}` — voices.collections.removeVoice
- [x] `POST /v1/voices/add` — voices.ivc.create
- [x] `POST /v1/voices/pvc` — voices.pvc.create
- [x] `POST /v1/voices/pvc/{voice_id}` — voices.pvc.update
- [x] `POST /v1/voices/pvc/{voice_id}/train` — voices.pvc.train
- [x] `POST /v1/voices/pvc/{voice_id}/samples` — voices.pvc.samples.create
- [x] `POST /v1/voices/pvc/{voice_id}/samples/{sample_id}` — voices.pvc.samples.update
- [x] `DELETE /v1/voices/pvc/{voice_id}/samples/{sample_id}` — voices.pvc.samples.delete
- [x] `GET /v1/voices/pvc/{voice_id}/samples/{sample_id}/audio` — voices.pvc.samples.audio.get
- [x] `GET /v1/voices/pvc/{voice_id}/samples/{sample_id}/speakers` — voices.pvc.samples.speakers.get
- [x] `POST /v1/voices/pvc/{voice_id}/samples/{sample_id}/separate-speakers` — voices.pvc.samples.speakers.separate
- [x] `GET /v1/voices/pvc/{voice_id}/samples/{sample_id}/speakers/{speaker_id}/audio` — voices.pvc.samples.speakers.audio.get
- [x] `GET /v1/voices/pvc/{voice_id}/samples/{sample_id}/waveform` — voices.pvc.samples.waveform.get
- [x] `POST /v1/voices/pvc/{voice_id}/verification` — voices.pvc.verification.request
- [x] `GET /v1/voices/pvc/{voice_id}/captcha` — voices.pvc.verification.captcha.get
- [x] `POST /v1/voices/pvc/{voice_id}/captcha` — voices.pvc.verification.captcha.verify
- [x] `GET /v1/voices/{voice_id}/samples/{sample_id}/audio` — voices.samples.audio.get (stream/binary)
- [x] `GET /v1/voices/settings/default` — voices.settings.getDefault
- [x] `GET /v1/voices/{voice_id}/settings` — voices.settings.get
- [x] `POST /v1/voices/{voice_id}/settings/edit` — voices.settings.update

## webhooks (4)
- [x] `GET /v1/workspace/webhooks` — webhooks.list
- [x] `POST /v1/workspace/webhooks` — webhooks.create
- [x] `DELETE /v1/workspace/webhooks/{webhook_id}` — webhooks.delete
- [x] `PATCH /v1/workspace/webhooks/{webhook_id}` — webhooks.update

## workspace (20)
- [x] `POST /v1/workspaces/api-keys/third-party-disabling` — workspace.setThirdPartyDisablingPolicy
- [x] `POST /v1/workspace/analytics/requests` — workspace.analytics.requests.get
- [x] `GET /v1/workspace/audit-logs` — workspace.auditLogs.list
- [x] `GET /v1/workspace/auth-connections` — workspace.authConnections.list
- [x] `POST /v1/workspace/auth-connections` — workspace.authConnections.create
- [x] `DELETE /v1/workspace/auth-connections/{auth_connection_id}` — workspace.authConnections.delete
- [x] `PATCH /v1/workspace/auth-connections/{auth_connection_id}` — workspace.authConnections.update
- [x] `GET /v1/workspace/groups` — workspace.groups.list
- [x] `GET /v1/workspace/groups/search` — workspace.groups.search
- [x] `POST /v1/workspace/groups/{group_id}/members/remove` — workspace.groups.members.remove
- [x] `POST /v1/workspace/groups/{group_id}/members` — workspace.groups.members.add
- [x] `POST /v1/workspace/invites/add` — workspace.invites.create
- [x] `POST /v1/workspace/invites/add-bulk` — workspace.invites.createBatch
- [x] `DELETE /v1/workspace/invites` — workspace.invites.delete
- [x] `GET /v1/workspace/members` — workspace.members.list
- [x] `POST /v1/workspace/members` — workspace.members.update
- [x] `GET /v1/workspace/resources/{resource_id}` — workspace.resources.get
- [x] `POST /v1/workspace/resources/{resource_id}/share` — workspace.resources.share
- [x] `POST /v1/workspace/resources/{resource_id}/unshare` — workspace.resources.unshare
- [x] `POST /v1/workspace/analytics/query/usage-by-product-over-time` — workspace.usage.getUsageByProductOverTime

## workspaces (1)
- [x] `POST /v1/workspaces/api-keys/disable` — workspaces.apiKeys.disable
