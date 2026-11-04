# ElevenLabs Go parity checklist

Generated from elevenlabs-js fe015f4 (Mon Nov 2 17:14:23 2026 +0100) v2.75.0, `src/api/resources/**/client/Client.ts`. `[x]` = implemented on the `parity` branch (package `github.com/taigrr/elevenlabs`).

**Totals:** 24/419 endpoints covered.  assets: 0/4; audioEffects: 0/1; audioIsolation: 0/4; audioNative: 0/4; conversationalAi: 0/186; dubbing: 0/37; environmentVariables: 0/5; environments: 0/3; flows: 0/14; forcedAlignment: 0/1; history: 5/5; models: 0/1; music: 0/13; productions: 0/11; pronunciationDictionaries: 0/9; samples: 1/1; serviceAccounts: 0/7; speechEngine: 0/6; speechToSpeech: 0/2; speechToText: 1/3; studio: 0/23; textToDialogue: 0/4; textToSoundEffects: 1/1; textToSpeech: 4/4; textToVoice: 0/5; tokens: 0/1; usage: 0/1; user: 2/2; voices: 10/36; webhooks: 0/4; workspace: 0/20; workspaces: 0/1

## core (beyond REST, not counted in totals)
- [x] Typed `*APIError` + sentinels (400/401/403/404/409/422/425/429/5xx) and `ErrTimeout` (ElevenLabsError/ElevenLabsTimeoutError)
- [x] Retries with exponential backoff + jitter (maxRetries, Retry-After, X-RateLimit-Reset)
- [x] Per-request options (headers, query params, body params, timeout, retries, API key)
- [x] Base URL environments (production, US, EU, India, Singapore residency)
- [x] Pagination iterators (`iter.Seq2`)
- [x] Streaming multipart uploads and streaming binary / NDJSON responses
- [ ] WebSocket realtime: TTS stream-input, multi-stream-input, dialogue stream-input, STT realtime, dubbing/translation realtime
- [ ] Conversational AI websocket
- [ ] Webhooks signature verification (src/wrapper/webhooks.ts)
- [ ] Music multipart helper (src/wrapper/music.ts)
- [ ] speechToText wrapper helpers (src/wrapper/speechToText.ts)

## assets (4)
- [ ] `GET /v1/assets` — assets.list
- [ ] `POST /v1/assets` — assets.create
- [ ] `GET /v1/assets/{asset_id}` — assets.get
- [ ] `DELETE /v1/assets/{asset_id}` — assets.delete

## audioEffects (1)
- [ ] `GET /v1/audio-effects/catalog` — audioEffects.getCatalog

## audioIsolation (4)
- [ ] `POST /v1/audio-isolation` — audioIsolation.convert (stream/binary)
- [ ] `GET /v1/audio-isolation/history` — audioIsolation.list
- [ ] `DELETE /v1/audio-isolation/history/{history_item_id}` — audioIsolation.delete
- [ ] `POST /v1/audio-isolation/stream` — audioIsolation.stream (stream/binary)

## audioNative (4)
- [ ] `POST /v1/audio-native` — audioNative.create
- [ ] `GET /v1/audio-native/{project_id}/settings` — audioNative.getSettings
- [ ] `POST /v1/audio-native/{project_id}/content` — audioNative.update
- [ ] `POST /v1/audio-native/content` — audioNative.updateContentFromUrl

## conversationalAi (186)
- [ ] `POST /v1/convai/knowledge-base` — conversationalAi.addToKnowledgeBase
- [ ] `GET /v1/convai/knowledge-base/rag-index` — conversationalAi.ragIndexOverview
- [ ] `GET /v1/convai/knowledge-base/{documentation_id}/rag-index` — conversationalAi.getDocumentRagIndexes
- [ ] `DELETE /v1/convai/knowledge-base/{documentation_id}/rag-index/{rag_index_id}` — conversationalAi.deleteDocumentRagIndex
- [ ] `POST /v1/convai/agents/create` — conversationalAi.agents.create
- [ ] `GET /v1/convai/agents/{agent_id}` — conversationalAi.agents.get
- [ ] `DELETE /v1/convai/agents/{agent_id}` — conversationalAi.agents.delete
- [ ] `PATCH /v1/convai/agents/{agent_id}` — conversationalAi.agents.update
- [ ] `GET /v1/convai/agents` — conversationalAi.agents.list
- [ ] `POST /v1/convai/agents/{agent_id}/duplicate` — conversationalAi.agents.duplicate
- [ ] `POST /v1/convai/agents/{agent_id}/simulate-conversation` — conversationalAi.agents.simulateConversation
- [ ] `POST /v1/convai/agents/{agent_id}/simulate-conversation/stream` — conversationalAi.agents.simulateConversationStream (stream/binary)
- [ ] `POST /v1/convai/agents/{agent_id}/run-tests` — conversationalAi.agents.runTests
- [ ] `GET /v1/convai/agents/{agent_id}/branches` — conversationalAi.agents.branches.list
- [ ] `POST /v1/convai/agents/{agent_id}/branches` — conversationalAi.agents.branches.create
- [ ] `GET /v1/convai/agents/{agent_id}/branches/{branch_id}` — conversationalAi.agents.branches.get
- [ ] `PATCH /v1/convai/agents/{agent_id}/branches/{branch_id}` — conversationalAi.agents.branches.update
- [ ] `GET /v1/convai/agents/{agent_id}/branches/{source_branch_id}/merge-preview` — conversationalAi.agents.branches.previewMerge
- [ ] `POST /v1/convai/agents/{agent_id}/branches/{source_branch_id}/merge` — conversationalAi.agents.branches.merge
- [ ] `GET /v1/convai/agents/{agent_id}/branches/{branch_id}/rebase-preview` — conversationalAi.agents.branches.previewRebase
- [ ] `POST /v1/convai/agents/{agent_id}/branches/{branch_id}/rebase` — conversationalAi.agents.branches.rebase
- [ ] `GET /v1/convai/agents/{agent_id}/deployments` — conversationalAi.agents.deployments.list
- [ ] `POST /v1/convai/agents/{agent_id}/deployments` — conversationalAi.agents.deployments.create
- [ ] `POST /v1/convai/agents/{agent_id}/drafts` — conversationalAi.agents.drafts.create
- [ ] `DELETE /v1/convai/agents/{agent_id}/drafts` — conversationalAi.agents.drafts.delete
- [ ] `POST /v1/convai/agents/{agent_id}/hold-audio` — conversationalAi.agents.holdAudio.create
- [ ] `DELETE /v1/convai/agents/{agent_id}/hold-audio` — conversationalAi.agents.holdAudio.delete
- [ ] `GET /v1/convai/agent/{agent_id}/knowledge-base/size` — conversationalAi.agents.knowledgeBase.size
- [ ] `GET /v1/convai/agents/{agent_id}/link` — conversationalAi.agents.link.get
- [ ] `POST /v1/convai/agent/{agent_id}/llm-usage/calculate` — conversationalAi.agents.llmUsage.calculate
- [ ] `GET /v1/convai/agents/{agent_id}/merge-proposals` — conversationalAi.agents.mergeProposals.list
- [ ] `POST /v1/convai/agents/{agent_id}/merge-proposals` — conversationalAi.agents.mergeProposals.create
- [ ] `GET /v1/convai/agents/{agent_id}/merge-proposals/{merge_proposal_id}` — conversationalAi.agents.mergeProposals.get
- [ ] `PATCH /v1/convai/agents/{agent_id}/merge-proposals/{merge_proposal_id}` — conversationalAi.agents.mergeProposals.update
- [ ] `POST /v1/convai/agents/{agent_id}/merge-proposals/{merge_proposal_id}/reviews` — conversationalAi.agents.mergeProposals.submitReview
- [ ] `POST /v1/convai/agents/{agent_id}/merge-proposals/{merge_proposal_id}/comments` — conversationalAi.agents.mergeProposals.addComment
- [ ] `POST /v1/convai/agents/{agent_id}/merge-proposals/{merge_proposal_id}/merge` — conversationalAi.agents.mergeProposals.merge
- [ ] `GET /v1/convai/agents/{agent_id}/branches/{branch_id}/procedures` — conversationalAi.agents.procedures.list
- [ ] `POST /v1/convai/agents/{agent_id}/branches/{branch_id}/procedures` — conversationalAi.agents.procedures.create
- [ ] `POST /v1/convai/agents/{agent_id}/branches/{branch_id}/procedures/compile` — conversationalAi.agents.procedures.compile
- [ ] `GET /v1/convai/agents/{agent_id}/branches/{branch_id}/procedures/{procedure_id}` — conversationalAi.agents.procedures.get
- [ ] `DELETE /v1/convai/agents/{agent_id}/branches/{branch_id}/procedures/{procedure_id}` — conversationalAi.agents.procedures.remove
- [ ] `GET /v1/convai/agents/{agent_id}/branches/{branch_id}/procedures/{procedure_id}/draft` — conversationalAi.agents.procedures.drafts.get
- [ ] `DELETE /v1/convai/agents/{agent_id}/branches/{branch_id}/procedures/{procedure_id}/draft` — conversationalAi.agents.procedures.drafts.delete
- [ ] `PATCH /v1/convai/agents/{agent_id}/branches/{branch_id}/procedures/{procedure_id}/draft` — conversationalAi.agents.procedures.drafts.update
- [ ] `GET /v1/convai/agents/summaries` — conversationalAi.agents.summaries.get
- [ ] `GET /v1/convai/agents/{agent_id}/versions/{version_id}` — conversationalAi.agents.versions.get
- [ ] `GET /v1/convai/agents/{agent_id}/widget` — conversationalAi.agents.widget.get
- [ ] `POST /v1/convai/agents/{agent_id}/avatar` — conversationalAi.agents.widget.avatar.create
- [ ] `GET /v1/convai/analytics/live-count` — conversationalAi.analytics.liveCount.get
- [ ] `POST /v1/convai/batch-calling/submit` — conversationalAi.batchCalls.create
- [ ] `GET /v1/convai/batch-calling/workspace` — conversationalAi.batchCalls.list
- [ ] `GET /v1/convai/batch-calling/{batch_id}` — conversationalAi.batchCalls.get
- [ ] `DELETE /v1/convai/batch-calling/{batch_id}` — conversationalAi.batchCalls.delete
- [ ] `POST /v1/convai/batch-calling/{batch_id}/cancel` — conversationalAi.batchCalls.cancel
- [ ] `POST /v1/convai/batch-calling/{batch_id}/retry` — conversationalAi.batchCalls.retry
- [ ] `GET /v1/convai/batch-calling/{batch_id}/export` — conversationalAi.batchCalls.export (stream/binary)
- [ ] `GET /v1/convai/conversation/get-signed-url` — conversationalAi.conversations.getSignedUrl
- [ ] `GET /v1/convai/conversation/token` — conversationalAi.conversations.getWebrtcToken
- [ ] `GET /v1/convai/conversations` — conversationalAi.conversations.list
- [ ] `GET /v1/convai/conversations/resolve` — conversationalAi.conversations.resolve
- [ ] `GET /v1/convai/conversations/{conversation_id}` — conversationalAi.conversations.get
- [ ] `DELETE /v1/convai/conversations/{conversation_id}` — conversationalAi.conversations.delete
- [ ] `GET /v1/convai/conversations/{conversation_id}/summary` — conversationalAi.conversations.getSummary
- [ ] `GET /v1/convai/conversations/{conversation_id}/sip-messages` — conversationalAi.conversations.getSipMessages
- [ ] `POST /v1/convai/conversations/{conversation_id}/analysis/run` — conversationalAi.conversations.analysis.run
- [ ] `POST /v1/convai/conversations/{conversation_id}/analysis/evaluations/run` — conversationalAi.conversations.analysis.runEvaluation
- [ ] `GET /v1/convai/conversations/{conversation_id}/audio` — conversationalAi.conversations.audio.get (stream/binary)
- [ ] `POST /v1/convai/conversations/{conversation_id}/feedback` — conversationalAi.conversations.feedback.create
- [ ] `POST /v1/convai/conversations/{conversation_id}/files` — conversationalAi.conversations.files.create
- [ ] `DELETE /v1/convai/conversations/{conversation_id}/files/{file_id}` — conversationalAi.conversations.files.delete
- [ ] `GET /v1/convai/conversations/messages/text-search` — conversationalAi.conversations.messages.textSearch
- [ ] `GET /v1/convai/conversations/messages/smart-search` — conversationalAi.conversations.messages.search
- [ ] `POST /v1/convai/conversations/{conversation_id}/tags` — conversationalAi.conversations.tags.assign
- [ ] `DELETE /v1/convai/conversations/{conversation_id}/tags/{tag_id}` — conversationalAi.conversations.tags.unassign
- [ ] `GET /v1/convai/tags` — conversationalAi.conversations.tags.list
- [ ] `POST /v1/convai/tags` — conversationalAi.conversations.tags.create
- [ ] `GET /v1/convai/tags/{tag_id}` — conversationalAi.conversations.tags.get
- [ ] `DELETE /v1/convai/tags/{tag_id}` — conversationalAi.conversations.tags.delete
- [ ] `PATCH /v1/convai/tags/{tag_id}` — conversationalAi.conversations.tags.update
- [ ] `GET /v1/convai/agents/{agent_id}/topics` — conversationalAi.conversations.topics.get
- [ ] `GET /v1/convai/settings/dashboard` — conversationalAi.dashboard.settings.get
- [ ] `PATCH /v1/convai/settings/dashboard` — conversationalAi.dashboard.settings.update
- [ ] `POST /v1/convai/exotel/outbound-call` — conversationalAi.exotel.outboundCall
- [ ] `GET /v1/convai/knowledge-base` — conversationalAi.knowledgeBase.list
- [ ] `POST /v1/convai/knowledge-base/rag-index` — conversationalAi.knowledgeBase.getOrCreateRagIndexes
- [ ] `GET /v1/convai/knowledge-base/search` — conversationalAi.knowledgeBase.search
- [ ] `GET /v1/convai/knowledge-base/crawl` — conversationalAi.knowledgeBase.crawlJobs.list
- [ ] `POST /v1/convai/knowledge-base/crawl` — conversationalAi.knowledgeBase.crawlJobs.create
- [ ] `GET /v1/convai/knowledge-base/crawl/{crawl_job_id}` — conversationalAi.knowledgeBase.crawlJobs.get
- [ ] `POST /v1/convai/knowledge-base/crawl/{crawl_job_id}/cancel` — conversationalAi.knowledgeBase.crawlJobs.cancel
- [ ] `PATCH /v1/convai/knowledge-base/{documentation_id}/update-file` — conversationalAi.knowledgeBase.document.updateFile
- [ ] `POST /v1/convai/knowledge-base/{documentation_id}/refresh` — conversationalAi.knowledgeBase.document.refresh
- [ ] `POST /v1/convai/knowledge-base/{documentation_id}/rag-index` — conversationalAi.knowledgeBase.document.computeRagIndex
- [ ] `POST /v1/convai/knowledge-base/url` — conversationalAi.knowledgeBase.documents.createFromUrl
- [ ] `POST /v1/convai/knowledge-base/file` — conversationalAi.knowledgeBase.documents.createFromFile
- [ ] `POST /v1/convai/knowledge-base/text` — conversationalAi.knowledgeBase.documents.createFromText
- [ ] `POST /v1/convai/knowledge-base/folder` — conversationalAi.knowledgeBase.documents.createFolder
- [ ] `GET /v1/convai/knowledge-base/{documentation_id}` — conversationalAi.knowledgeBase.documents.get
- [ ] `DELETE /v1/convai/knowledge-base/{documentation_id}` — conversationalAi.knowledgeBase.documents.delete
- [ ] `PATCH /v1/convai/knowledge-base/{documentation_id}` — conversationalAi.knowledgeBase.documents.update
- [ ] `GET /v1/convai/knowledge-base/{documentation_id}/dependent-agents` — conversationalAi.knowledgeBase.documents.getAgents
- [ ] `POST /v1/convai/knowledge-base/dependent-agents` — conversationalAi.knowledgeBase.documents.getBulkAgents
- [ ] `GET /v1/convai/knowledge-base/{documentation_id}/content` — conversationalAi.knowledgeBase.documents.getContent
- [ ] `GET /v1/convai/knowledge-base/{documentation_id}/source-file-url` — conversationalAi.knowledgeBase.documents.getSourceFileUrl
- [ ] `POST /v1/convai/knowledge-base/{document_id}/move` — conversationalAi.knowledgeBase.documents.move
- [ ] `POST /v1/convai/knowledge-base/bulk-move` — conversationalAi.knowledgeBase.documents.bulkMove
- [ ] `POST /v1/convai/knowledge-base/bulk-delete` — conversationalAi.knowledgeBase.documents.bulkDelete
- [ ] `GET /v1/convai/knowledge-base/{documentation_id}/chunk/{chunk_id}` — conversationalAi.knowledgeBase.documents.chunk.get
- [ ] `GET /v1/convai/knowledge-base/{documentation_id}/chunks` — conversationalAi.knowledgeBase.documents.chunks.list
- [ ] `GET /v1/convai/knowledge-base/summaries` — conversationalAi.knowledgeBase.documents.summaries.get
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
- [ ] `GET /v1/convai/secrets` — conversationalAi.secrets.list
- [ ] `POST /v1/convai/secrets` — conversationalAi.secrets.create
- [ ] `GET /v1/convai/secrets/{secret_id}` — conversationalAi.secrets.get
- [ ] `DELETE /v1/convai/secrets/{secret_id}` — conversationalAi.secrets.delete
- [ ] `PATCH /v1/convai/secrets/{secret_id}` — conversationalAi.secrets.update
- [ ] `GET /v1/convai/secrets/{secret_id}/dependencies/${core.url.encodePathParam(serializers.SecretDependencyResourceType.jsonOrThrow(resource_type))}` — conversationalAi.secrets.getDependencies
- [ ] `GET /v1/convai/settings` — conversationalAi.settings.get
- [ ] `PATCH /v1/convai/settings` — conversationalAi.settings.update
- [ ] `POST /v1/convai/sip-trunk/outbound-call` — conversationalAi.sipTrunk.outboundCall
- [ ] `POST /v1/convai/agent-testing/create` — conversationalAi.tests.create
- [ ] `POST /v1/convai/agent-testing/bulk-move` — conversationalAi.tests.move
- [ ] `GET /v1/convai/agent-testing/{test_id}` — conversationalAi.tests.get
- [ ] `PUT /v1/convai/agent-testing/{test_id}` — conversationalAi.tests.update
- [ ] `DELETE /v1/convai/agent-testing/{test_id}` — conversationalAi.tests.delete
- [ ] `POST /v1/convai/agent-testing/summaries` — conversationalAi.tests.summaries
- [ ] `GET /v1/convai/agent-testing` — conversationalAi.tests.list
- [ ] `POST /v1/convai/agent-testing/folders` — conversationalAi.tests.folders.create
- [ ] `GET /v1/convai/agent-testing/folders/{folder_id}` — conversationalAi.tests.folders.get
- [ ] `DELETE /v1/convai/agent-testing/folders/{folder_id}` — conversationalAi.tests.folders.delete
- [ ] `PATCH /v1/convai/agent-testing/folders/{folder_id}` — conversationalAi.tests.folders.update
- [ ] `GET /v1/convai/test-invocations` — conversationalAi.tests.invocations.list
- [ ] `GET /v1/convai/test-invocations/{test_invocation_id}` — conversationalAi.tests.invocations.get
- [ ] `PATCH /v1/convai/test-invocations/{test_invocation_id}` — conversationalAi.tests.invocations.update
- [ ] `POST /v1/convai/test-invocations/{test_invocation_id}/cancel` — conversationalAi.tests.invocations.cancel
- [ ] `POST /v1/convai/test-invocations/{test_invocation_id}/resubmit` — conversationalAi.tests.invocations.resubmit
- [ ] `GET /v1/convai/tools` — conversationalAi.tools.list
- [ ] `POST /v1/convai/tools` — conversationalAi.tools.create
- [ ] `GET /v1/convai/tools/{tool_id}` — conversationalAi.tools.get
- [ ] `DELETE /v1/convai/tools/{tool_id}` — conversationalAi.tools.delete
- [ ] `PATCH /v1/convai/tools/{tool_id}` — conversationalAi.tools.update
- [ ] `GET /v1/convai/tools/{tool_id}/dependent-agents` — conversationalAi.tools.getDependentAgents
- [ ] `GET /v1/convai/tools/{tool_id}/executions` — conversationalAi.tools.executions.get
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
- [ ] `GET /v1/dubbing` — dubbing.list
- [ ] `POST /v1/dubbing` — dubbing.create
- [ ] `GET /v1/dubbing/{dubbing_id}` — dubbing.get
- [ ] `DELETE /v1/dubbing/{dubbing_id}` — dubbing.delete
- [ ] `GET /v1/dubbing/{dubbing_id}/audio/{language_code}` — dubbing.audio.get (stream/binary)
- [ ] `GET /v1/dubbing/project` — dubbing.project.list
- [ ] `POST /v1/dubbing/project` — dubbing.project.create
- [ ] `GET /v1/dubbing/project/{project_id}` — dubbing.project.get
- [ ] `DELETE /v1/dubbing/project/{project_id}` — dubbing.project.delete
- [ ] `GET /v1/dubbing/project/{project_id}/language` — dubbing.project.language.list
- [ ] `POST /v1/dubbing/project/{project_id}/language` — dubbing.project.language.create
- [ ] `GET /v1/dubbing/project/{project_id}/language/{language_id}` — dubbing.project.language.get
- [ ] `DELETE /v1/dubbing/project/{project_id}/language/{language_id}` — dubbing.project.language.delete
- [ ] `GET /v1/dubbing/project/{project_id}/language/{language_id}/transcript` — dubbing.project.language.transcript.get
- [ ] `PATCH /v1/dubbing/project/{project_id}/language/{language_id}/transcript/segment/{segment_id}` — dubbing.project.language.transcript.updateSegment
- [ ] `PATCH /v1/dubbing/project/{project_id}/language/{language_id}/transcript/segments` — dubbing.project.language.transcript.updateSegments
- [ ] `POST /v1/dubbing/project/{project_id}/language/{language_id}/transcript/regenerate` — dubbing.project.language.transcript.regenerate
- [ ] `GET /v1/dubbing/project/{project_id}/transcript` — dubbing.project.transcript.get
- [ ] `DELETE /v1/dubbing/project/{project_id}/transcript/segment/{segment_id}` — dubbing.project.transcript.deleteSegment
- [ ] `PATCH /v1/dubbing/project/{project_id}/transcript/segment/{segment_id}` — dubbing.project.transcript.updateSegment
- [ ] `PATCH /v1/dubbing/project/{project_id}/transcript/segments` — dubbing.project.transcript.updateSegments
- [ ] `POST /v1/dubbing/project/{project_id}/transcript/segment` — dubbing.project.transcript.createSegment
- [ ] `GET /v1/dubbing/resource/{dubbing_id}` — dubbing.resource.get
- [ ] `POST /v1/dubbing/resource/{dubbing_id}/migrate-segments` — dubbing.resource.migrateSegments
- [ ] `POST /v1/dubbing/resource/{dubbing_id}/transcribe` — dubbing.resource.transcribe
- [ ] `POST /v1/dubbing/resource/{dubbing_id}/translate` — dubbing.resource.translate
- [ ] `POST /v1/dubbing/resource/{dubbing_id}/dub` — dubbing.resource.dub
- [ ] `POST /v1/dubbing/resource/{dubbing_id}/render/{language}` — dubbing.resource.render
- [ ] `POST /v1/dubbing/resource/{dubbing_id}/language` — dubbing.resource.language.add
- [ ] `PATCH /v1/dubbing/resource/{dubbing_id}/segment/{segment_id}/{language}` — dubbing.resource.segment.update
- [ ] `DELETE /v1/dubbing/resource/{dubbing_id}/segment/{segment_id}` — dubbing.resource.segment.delete
- [ ] `PATCH /v1/dubbing/resource/{dubbing_id}/speaker/{speaker_id}` — dubbing.resource.speaker.update
- [ ] `POST /v1/dubbing/resource/{dubbing_id}/speaker` — dubbing.resource.speaker.create
- [ ] `GET /v1/dubbing/resource/{dubbing_id}/speaker/{speaker_id}/similar-voices` — dubbing.resource.speaker.findSimilarVoices
- [ ] `POST /v1/dubbing/resource/{dubbing_id}/speaker/{speaker_id}/segment` — dubbing.resource.speaker.segment.create
- [ ] `GET /v1/dubbing/{dubbing_id}/transcript/{language_code}` — dubbing.transcript.getTranscriptForDub
- [ ] `GET /v1/dubbing/{dubbing_id}/transcripts/{language_code}/format/${core.url.encodePathParam(serializers.dubbing.TranscriptsGetRequestFormatType.jsonOrThrow(format_type))}` — dubbing.transcripts.get

## environmentVariables (5)
- [ ] `GET /v1/convai/environment-variables` — environmentVariables.list
- [ ] `POST /v1/convai/environment-variables` — environmentVariables.create
- [ ] `GET /v1/convai/environment-variables/{env_var_id}` — environmentVariables.get
- [ ] `DELETE /v1/convai/environment-variables/{env_var_id}` — environmentVariables.delete
- [ ] `PATCH /v1/convai/environment-variables/{env_var_id}` — environmentVariables.update

## environments (3)
- [ ] `GET /v1/convai/environments` — environments.list
- [ ] `POST /v1/convai/environments` — environments.create
- [ ] `DELETE /v1/convai/environments/{environment}` — environments.delete

## flows (14)
- [ ] `GET /v1/flows/image` — flows.image.list
- [ ] `POST /v1/flows/image` — flows.image.create
- [ ] `GET /v1/flows/image/{generation_id}` — flows.image.get
- [ ] `GET /v1/flows/templates` — flows.templates.list
- [ ] `GET /v1/flows/templates/{template_id}` — flows.templates.get
- [ ] `GET /v1/flows/templates/{template_id}/runs` — flows.templates.runs.list
- [ ] `POST /v1/flows/templates/{template_id}/runs` — flows.templates.runs.create
- [ ] `GET /v1/flows/templates/{template_id}/runs/{run_id}` — flows.templates.runs.get
- [ ] `GET /v1/flows/text-to-speech` — flows.textToSpeech.list
- [ ] `POST /v1/flows/text-to-speech` — flows.textToSpeech.create
- [ ] `GET /v1/flows/text-to-speech/{generation_id}` — flows.textToSpeech.get
- [ ] `GET /v1/flows/video` — flows.video.list
- [ ] `POST /v1/flows/video` — flows.video.create
- [ ] `GET /v1/flows/video/{generation_id}` — flows.video.get

## forcedAlignment (1)
- [ ] `POST /v1/forced-alignment` — forcedAlignment.create

## history (5)
- [x] `GET /v1/history` — history.list
- [x] `GET /v1/history/{history_item_id}` — history.get
- [x] `DELETE /v1/history/{history_item_id}` — history.delete
- [x] `GET /v1/history/{history_item_id}/audio` — history.getAudio (stream/binary)
- [x] `POST /v1/history/download` — history.download (stream/binary)

## models (1)
- [ ] `GET /v1/models` — models.list

## music (13)
- [ ] `POST /v1/music/video-to-music` — music.videoToMusic (stream/binary)
- [ ] `POST /v1/music` — music.compose (stream/binary)
- [ ] `POST /v1/music/detailed` — music.composeDetailed (stream/binary)
- [ ] `POST /v1/music/detailed/stream` — music.composeDetailedStream (stream/binary)
- [ ] `POST /v1/music/stream` — music.stream (stream/binary)
- [ ] `POST /v1/music/upload` — music.upload
- [ ] `POST /v1/music/stem-separation` — music.separateStems (stream/binary)
- [ ] `POST /v1/music/plan` — music.compositionPlan.create
- [ ] `GET /v1/music/finetunes` — music.finetunes.list
- [ ] `POST /v1/music/finetunes` — music.finetunes.create
- [ ] `GET /v1/music/finetunes/{finetune_id}` — music.finetunes.get
- [ ] `DELETE /v1/music/finetunes/{finetune_id}` — music.finetunes.delete
- [ ] `PATCH /v1/music/finetunes/{finetune_id}` — music.finetunes.update

## productions (11)
- [ ] `GET /v1/productions/orders` — productions.orders.list
- [ ] `POST /v1/productions/orders` — productions.orders.create
- [ ] `GET /v1/productions/orders/${core.url.encodePathParam(serializers.OrderId.jsonOrThrow(order_id))}` — productions.orders.get
- [ ] `PATCH /v1/productions/orders/${core.url.encodePathParam(serializers.OrderId.jsonOrThrow(order_id))}` — productions.orders.update
- [ ] `POST /v1/productions/orders/${core.url.encodePathParam(serializers.OrderId.jsonOrThrow(order_id))}/submit` — productions.orders.submit
- [ ] `GET /v1/productions/orders/${core.url.encodePathParam(serializers.OrderId.jsonOrThrow(order_id))}/deliverables` — productions.orders.deliverables.list
- [ ] `POST /v1/productions/orders/${core.url.encodePathParam(serializers.OrderId.jsonOrThrow(order_id))}/items` — productions.orders.items.upsert
- [ ] `DELETE /v1/productions/orders/${core.url.encodePathParam(serializers.OrderId.jsonOrThrow(order_id))}/items/${core.url.encodePathParam(serializers.ItemId.jsonOrThrow(item_id))}` — productions.orders.items.remove
- [ ] `GET /v1/productions/orders/languages/${core.url.encodePathParam(serializers.OrderItemKind.jsonOrThrow(order_item_kind))}` — productions.orders.languages.list
- [ ] `POST /v1/productions/orders/${core.url.encodePathParam(serializers.OrderId.jsonOrThrow(order_id))}/media` — productions.orders.media.register
- [ ] `GET /v1/productions/orders/${core.url.encodePathParam(serializers.OrderId.jsonOrThrow(order_id))}/media/${core.url.encodePathParam(serializers.MediaId.jsonOrThrow(media_id))}` — productions.orders.media.get

## pronunciationDictionaries (9)
- [ ] `POST /v1/pronunciation-dictionaries/add-from-file` — pronunciationDictionaries.createFromFile
- [ ] `POST /v1/pronunciation-dictionaries/add-from-rules` — pronunciationDictionaries.createFromRules
- [ ] `GET /v1/pronunciation-dictionaries/{pronunciation_dictionary_id}` — pronunciationDictionaries.get
- [ ] `PATCH /v1/pronunciation-dictionaries/{pronunciation_dictionary_id}` — pronunciationDictionaries.update
- [ ] `GET /v1/pronunciation-dictionaries/{dictionary_id}/{version_id}/download` — pronunciationDictionaries.download (stream/binary)
- [ ] `GET /v1/pronunciation-dictionaries` — pronunciationDictionaries.list
- [ ] `POST /v1/pronunciation-dictionaries/{pronunciation_dictionary_id}/set-rules` — pronunciationDictionaries.rules.set
- [ ] `POST /v1/pronunciation-dictionaries/{pronunciation_dictionary_id}/add-rules` — pronunciationDictionaries.rules.add
- [ ] `POST /v1/pronunciation-dictionaries/{pronunciation_dictionary_id}/remove-rules` — pronunciationDictionaries.rules.remove

## samples (1)
- [x] `DELETE /v1/voices/{voice_id}/samples/{sample_id}` — samples.delete

## serviceAccounts (7)
- [ ] `GET /v1/service-accounts` — serviceAccounts.list
- [ ] `POST /v1/service-accounts` — serviceAccounts.create
- [ ] `GET /v1/service-accounts/{service_account_user_id}/api-keys` — serviceAccounts.apiKeys.list
- [ ] `POST /v1/service-accounts/{service_account_user_id}/api-keys` — serviceAccounts.apiKeys.create
- [ ] `DELETE /v1/service-accounts/{service_account_user_id}/api-keys/{api_key_id}` — serviceAccounts.apiKeys.delete
- [ ] `PATCH /v1/service-accounts/{service_account_user_id}/api-keys/{api_key_id}` — serviceAccounts.apiKeys.update
- [ ] `POST /v1/service-accounts/{service_account_user_id}/api-keys/{api_key_id}/disable` — serviceAccounts.apiKeys.disable

## speechEngine (6)
- [ ] `GET /v1/speech-engine` — speechEngine.list
- [ ] `POST /v1/speech-engine` — speechEngine.create
- [ ] `GET /v1/speech-engine/{speech_engine_id}` — speechEngine.get
- [ ] `DELETE /v1/speech-engine/{speech_engine_id}` — speechEngine.delete
- [ ] `PATCH /v1/speech-engine/{speech_engine_id}` — speechEngine.update
- [ ] `POST /v1/speech-engine/{speech_engine_id}/duplicate` — speechEngine.duplicate

## speechToSpeech (2)
- [ ] `POST /v1/voice-changer/{voice_id}` — speechToSpeech.convert (stream/binary)
- [ ] `POST /v1/voice-changer/{voice_id}/stream` — speechToSpeech.stream (stream/binary)

## speechToText (3)
- [x] `POST /v1/speech-to-text` — speechToText.convert
- [ ] `GET /v1/speech-to-text/transcripts/{transcription_id}` — speechToText.transcripts.get
- [ ] `DELETE /v1/speech-to-text/transcripts/{transcription_id}` — speechToText.transcripts.delete

## studio (23)
- [ ] `POST /v1/studio/podcasts` — studio.createPodcast
- [ ] `GET /v1/studio/projects` — studio.projects.list
- [ ] `POST /v1/studio/projects` — studio.projects.create
- [ ] `GET /v1/studio/projects/{project_id}` — studio.projects.get
- [ ] `POST /v1/studio/projects/{project_id}` — studio.projects.update
- [ ] `DELETE /v1/studio/projects/{project_id}` — studio.projects.delete
- [ ] `POST /v1/studio/projects/{project_id}/convert` — studio.projects.convert
- [ ] `GET /v1/studio/projects/{project_id}/muted-tracks` — studio.projects.getMutedTracks
- [ ] `GET /v1/studio/projects/{project_id}/chapters` — studio.projects.chapters.list
- [ ] `POST /v1/studio/projects/{project_id}/chapters` — studio.projects.chapters.create
- [ ] `GET /v1/studio/projects/{project_id}/chapters/{chapter_id}` — studio.projects.chapters.get
- [ ] `POST /v1/studio/projects/{project_id}/chapters/{chapter_id}` — studio.projects.chapters.update
- [ ] `DELETE /v1/studio/projects/{project_id}/chapters/{chapter_id}` — studio.projects.chapters.delete
- [ ] `POST /v1/studio/projects/{project_id}/chapters/{chapter_id}/convert` — studio.projects.chapters.convert
- [ ] `GET /v1/studio/projects/{project_id}/chapters/{chapter_id}/snapshots` — studio.projects.chapters.snapshots.list
- [ ] `GET /v1/studio/projects/{project_id}/chapters/{chapter_id}/snapshots/{chapter_snapshot_id}` — studio.projects.chapters.snapshots.get
- [ ] `POST /v1/studio/projects/{project_id}/chapters/{chapter_id}/snapshots/{chapter_snapshot_id}/stream` — studio.projects.chapters.snapshots.stream (stream/binary)
- [ ] `POST /v1/studio/projects/{project_id}/content` — studio.projects.content.update
- [ ] `POST /v1/studio/projects/{project_id}/pronunciation-dictionaries` — studio.projects.pronunciationDictionaries.create
- [ ] `GET /v1/studio/projects/{project_id}/snapshots` — studio.projects.snapshots.list
- [ ] `GET /v1/studio/projects/{project_id}/snapshots/{project_snapshot_id}` — studio.projects.snapshots.get
- [ ] `POST /v1/studio/projects/{project_id}/snapshots/{project_snapshot_id}/stream` — studio.projects.snapshots.stream (stream/binary)
- [ ] `POST /v1/studio/projects/{project_id}/snapshots/{project_snapshot_id}/archive` — studio.projects.snapshots.streamArchive (stream/binary)

## textToDialogue (4)
- [ ] `POST /v1/text-to-dialogue` — textToDialogue.convert (stream/binary)
- [ ] `POST /v1/text-to-dialogue/stream` — textToDialogue.stream (stream/binary)
- [ ] `POST /v1/text-to-dialogue/stream/with-timestamps` — textToDialogue.streamWithTimestamps (stream/binary)
- [ ] `POST /v1/text-to-dialogue/with-timestamps` — textToDialogue.convertWithTimestamps

## textToSoundEffects (1)
- [x] `POST /v1/sound-generation` — textToSoundEffects.convert (stream/binary)

## textToSpeech (4)
- [x] `POST /v1/text-to-speech/{voice_id}` — textToSpeech.convert (stream/binary)
- [x] `POST /v1/text-to-speech/{voice_id}/with-timestamps` — textToSpeech.convertWithTimestamps
- [x] `POST /v1/text-to-speech/{voice_id}/stream` — textToSpeech.stream (stream/binary)
- [x] `POST /v1/text-to-speech/{voice_id}/stream/with-timestamps` — textToSpeech.streamWithTimestamps (stream/binary)

## textToVoice (5)
- [ ] `POST /v1/text-to-voice/create-previews` — textToVoice.createPreviews
- [ ] `POST /v1/text-to-voice` — textToVoice.create
- [ ] `POST /v1/text-to-voice/design` — textToVoice.design
- [ ] `POST /v1/text-to-voice/{voice_id}/remix` — textToVoice.remix
- [ ] `GET /v1/text-to-voice/{generated_voice_id}/stream` — textToVoice.preview.stream (stream/binary)

## tokens (1)
- [ ] `POST /v1/single-use-token/${core.url.encodePathParam(serializers.SingleUseTokenType.jsonOrThrow(token_type))}` — tokens.singleUse.create

## usage (1)
- [ ] `GET /v1/usage/character-stats` — usage.get

## user (2)
- [x] `GET /v1/user` — user.get
- [x] `GET /v1/user/subscription` — user.subscription.get

## voices (36)
- [x] `GET /v1/voices` — voices.getAll
- [x] `GET /v1/voices/{voice_id}` — voices.get
- [x] `DELETE /v1/voices/{voice_id}` — voices.delete
- [x] `POST /v1/voices/{voice_id}/edit` — voices.update
- [x] `GET /v2/voices` — voices.search
- [ ] `POST /v1/voices/{voice_id}/replicate-to-isolated-environment` — voices.replicateToIsolatedEnvironment
- [ ] `POST /v1/voices/add/{public_user_id}/{voice_id}` — voices.share
- [ ] `POST /v1/similar-voices` — voices.findSimilarVoices
- [ ] `GET /v1/shared-voices` — voices.getShared
- [ ] `GET /v1/voices/accents` — voices.accents.get
- [ ] `GET /v1/voices/collections` — voices.collections.list
- [ ] `POST /v1/voices/collections` — voices.collections.create
- [ ] `GET /v1/voices/collections/{collection_id}` — voices.collections.get
- [ ] `DELETE /v1/voices/collections/{collection_id}` — voices.collections.delete
- [ ] `PATCH /v1/voices/collections/{collection_id}` — voices.collections.update
- [ ] `POST /v1/voices/collections/{collection_id}/voice/{voice_id}` — voices.collections.addVoice
- [ ] `DELETE /v1/voices/collections/{collection_id}/voice/{voice_id}` — voices.collections.removeVoice
- [x] `POST /v1/voices/add` — voices.ivc.create
- [ ] `POST /v1/voices/pvc` — voices.pvc.create
- [ ] `POST /v1/voices/pvc/{voice_id}` — voices.pvc.update
- [ ] `POST /v1/voices/pvc/{voice_id}/train` — voices.pvc.train
- [ ] `POST /v1/voices/pvc/{voice_id}/samples` — voices.pvc.samples.create
- [ ] `POST /v1/voices/pvc/{voice_id}/samples/{sample_id}` — voices.pvc.samples.update
- [ ] `DELETE /v1/voices/pvc/{voice_id}/samples/{sample_id}` — voices.pvc.samples.delete
- [ ] `GET /v1/voices/pvc/{voice_id}/samples/{sample_id}/audio` — voices.pvc.samples.audio.get
- [ ] `GET /v1/voices/pvc/{voice_id}/samples/{sample_id}/speakers` — voices.pvc.samples.speakers.get
- [ ] `POST /v1/voices/pvc/{voice_id}/samples/{sample_id}/separate-speakers` — voices.pvc.samples.speakers.separate
- [ ] `GET /v1/voices/pvc/{voice_id}/samples/{sample_id}/speakers/{speaker_id}/audio` — voices.pvc.samples.speakers.audio.get
- [ ] `GET /v1/voices/pvc/{voice_id}/samples/{sample_id}/waveform` — voices.pvc.samples.waveform.get
- [ ] `POST /v1/voices/pvc/{voice_id}/verification` — voices.pvc.verification.request
- [ ] `GET /v1/voices/pvc/{voice_id}/captcha` — voices.pvc.verification.captcha.get
- [ ] `POST /v1/voices/pvc/{voice_id}/captcha` — voices.pvc.verification.captcha.verify
- [x] `GET /v1/voices/{voice_id}/samples/{sample_id}/audio` — voices.samples.audio.get (stream/binary)
- [x] `GET /v1/voices/settings/default` — voices.settings.getDefault
- [x] `GET /v1/voices/{voice_id}/settings` — voices.settings.get
- [x] `POST /v1/voices/{voice_id}/settings/edit` — voices.settings.update

## webhooks (4)
- [ ] `GET /v1/workspace/webhooks` — webhooks.list
- [ ] `POST /v1/workspace/webhooks` — webhooks.create
- [ ] `DELETE /v1/workspace/webhooks/{webhook_id}` — webhooks.delete
- [ ] `PATCH /v1/workspace/webhooks/{webhook_id}` — webhooks.update

## workspace (20)
- [ ] `POST /v1/workspaces/api-keys/third-party-disabling` — workspace.setThirdPartyDisablingPolicy
- [ ] `POST /v1/workspace/analytics/requests` — workspace.analytics.requests.get
- [ ] `GET /v1/workspace/audit-logs` — workspace.auditLogs.list
- [ ] `GET /v1/workspace/auth-connections` — workspace.authConnections.list
- [ ] `POST /v1/workspace/auth-connections` — workspace.authConnections.create
- [ ] `DELETE /v1/workspace/auth-connections/{auth_connection_id}` — workspace.authConnections.delete
- [ ] `PATCH /v1/workspace/auth-connections/{auth_connection_id}` — workspace.authConnections.update
- [ ] `GET /v1/workspace/groups` — workspace.groups.list
- [ ] `GET /v1/workspace/groups/search` — workspace.groups.search
- [ ] `POST /v1/workspace/groups/{group_id}/members/remove` — workspace.groups.members.remove
- [ ] `POST /v1/workspace/groups/{group_id}/members` — workspace.groups.members.add
- [ ] `POST /v1/workspace/invites/add` — workspace.invites.create
- [ ] `POST /v1/workspace/invites/add-bulk` — workspace.invites.createBatch
- [ ] `DELETE /v1/workspace/invites` — workspace.invites.delete
- [ ] `GET /v1/workspace/members` — workspace.members.list
- [ ] `POST /v1/workspace/members` — workspace.members.update
- [ ] `GET /v1/workspace/resources/{resource_id}` — workspace.resources.get
- [ ] `POST /v1/workspace/resources/{resource_id}/share` — workspace.resources.share
- [ ] `POST /v1/workspace/resources/{resource_id}/unshare` — workspace.resources.unshare
- [ ] `POST /v1/workspace/analytics/query/usage-by-product-over-time` — workspace.usage.getUsageByProductOverTime

## workspaces (1)
- [ ] `POST /v1/workspaces/api-keys/disable` — workspaces.apiKeys.disable
