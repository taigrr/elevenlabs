# Migration guide

This document lists every breaking change made on the `parity` branch, which rebuilds the client to match the official ElevenLabs TypeScript SDK.
The module path is now `github.com/taigrr/elevenlabs/v2`.

## Module path (v2)

The module path is now `github.com/taigrr/elevenlabs/v2`.
Update your imports from `github.com/taigrr/elevenlabs` to `github.com/taigrr/elevenlabs/v2` and fetch the new version with:

```sh
go get github.com/taigrr/elevenlabs/v2@latest
```

## Packages

| Old | New |
| --- | --- |
| `github.com/taigrr/elevenlabs/client` | `github.com/taigrr/elevenlabs/v2` (package `elevenlabs`) |
| `github.com/taigrr/elevenlabs/client/types` | merged into `github.com/taigrr/elevenlabs/v2` |

## Client construction

| Old | New |
| --- | --- |
| `client.New(apiKey)` (returns `Client` value) | `elevenlabs.NewClient(elevenlabs.WithAPIKey(apiKey))` (returns `*Client`; defaults to `$ELEVENLABS_API_KEY`) |
| `c.WithEndpoint(url)` | `elevenlabs.WithBaseURL(url)` or `elevenlabs.WithEnvironment(env)` |
| `c.WithAPIKey(key)` | `elevenlabs.WithAPIKey(key)` option, passed to NewClient or to a single call |
| `c.WithHTTPClient(hc)` | `elevenlabs.WithHTTPClient(hc)` |
| (none) | `WithTimeout`, `WithMaxRetries`, `WithHeader`, `WithUserAgent`, the same options per call |

Requests are now retried by default (2 retries for 408/409/429/5xx) and time out after 240 seconds per attempt; use `WithMaxRetries(0)` / `WithTimeout(0)` to restore the old behavior.

## Endpoints

| Old | New |
| --- | --- |
| `c.GetUserInfo(ctx)` | `client.User.Get(ctx)` returning `*User` |
| `c.GetSubscriptionInfo(ctx)` (read from `/v1/user`) | `client.User.Subscription.Get(ctx)` (calls `/v1/user/subscription`) returning `*Subscription` |
| `c.GetVoices(ctx)` | `client.Voices.GetAll(ctx, nil)` returning `*GetVoicesResponse` (use `.Voices`) |
| `c.GetVoiceIDs(ctx)` | removed; iterate `client.Voices.GetAll` or `client.Voices.SearchAll` |
| `c.GetVoice(ctx, id)` | `client.Voices.Get(ctx, id, nil)` returning `*Voice` |
| `c.DeleteVoice(ctx, id)` | `client.Voices.Delete(ctx, id)` returning `*StatusResponse` |
| `c.CreateVoice(ctx, name, desc, labels, files)` | `client.Voices.IVC.Create(ctx, &IVCCreateRequest{...})` returning `*AddVoiceIVCResponse` |
| `c.EditVoice(ctx, id, name, desc, labels, files)` | `client.Voices.Update(ctx, id, &VoicesUpdateRequest{...})` |
| `c.GetVoiceSettings(ctx, id)` | `client.Voices.Settings.Get(ctx, id)` returning `*VoiceSettings` |
| `c.EditVoiceSettings(ctx, id, opts)` | `client.Voices.Settings.Update(ctx, id, &VoiceSettings{...})` |
| `c.DownloadVoiceSample(ctx, voiceID, sampleID)` | `client.Voices.Samples.Audio.Get(ctx, voiceID, sampleID)` |
| `c.DeleteVoiceSample(ctx, voiceID, sampleID)` (returned `bool`) | `client.Samples.Delete(ctx, voiceID, sampleID)` returning `*StatusResponse` |
| `c.GetHistoryItemList(ctx)` | `client.History.List(ctx, &HistoryListRequest{...})` or `client.History.ListAll(...)` iterator |
| `c.GetHistoryIDs(ctx, voiceIDs...)` | removed; use `client.History.ListAll(ctx, &HistoryListRequest{VoiceID: id})` |
| `c.HistoryDelete(ctx, id)` (returned `bool`) | `client.History.Delete(ctx, id)` returning `*StatusResponse` |
| `c.HistoryDownloadAudio(ctx, id)` | `client.History.GetAudio(ctx, id)` |
| `c.HistoryDownloadZip(ctx, id1, id2, ids...)` | `client.History.Download(ctx, &DownloadHistoryRequest{HistoryItemIDs: ids})` |
| `c.TTS(ctx, text, voiceID, modelID, opts, params...)` | `client.TextToSpeech.Convert(ctx, voiceID, &TextToSpeechRequest{Text: text, ModelID: modelID, ...})` |
| `c.TTSStream(ctx, text, voiceID, opts, params...)` | `client.TextToSpeech.Stream(ctx, voiceID, &TextToSpeechRequest{...})` |
| `client.WithPreviousText` / `client.WithNextText` | `TextToSpeechRequest.PreviousText` / `.NextText` fields |
| `c.SoundGeneration(ctx, text, duration, influence)` | `client.TextToSoundEffects.Convert(ctx, &CreateSoundEffectRequest{...})` |
| `c.ConvertSpeechToText(ctx, path, req)` | `client.SpeechToText.Convert(ctx, &SpeechToTextConvertRequest{File: &File{Reader: f, Name: name}, ...})` |
| `c.ConvertSpeechToTextFromReader(ctx, r, name, req)` | same as above |

## Types

| Old | New |
| --- | --- |
| `types.SynthesisOptions` (values, silently clamped) | `VoiceSettings` (optional pointer fields, sent as given) |
| `types.TTS`, `types.TTSParam` | `TextToSpeechRequest` |
| `types.SoundGeneration` | `CreateSoundEffectRequest` (prompt influence no longer defaults to 0.3 client-side) |
| `types.SpeechToTextRequest` | `SpeechToTextConvertRequest`; `ModelID` is now a `string` (constants `SpeechToTextModelScribeV1`, ...) and bool fields are `*bool` |
| `types.SpeechToTextResponse` | `SpeechToTextConvertResponse` union with `Chunk`, `Multichannel` and `Webhook` variants |
| `types.TranscriptionWord` / `types.TranscriptionCharacter` | `SpeechToTextWord` / `SpeechToTextCharacter` (start/end are `*float64`) |
| `types.UserResponseModel` | `User` (the `xi_api_key` field no longer exists upstream; see `XiAPIKeyPreview`) |
| `types.Subscription` / `types.ExtendedSubscriptionResponseModel` | `SubscriptionResponse` / `Subscription` |
| `types.VoiceResponseModel`, `types.GetVoicesResponseModel` | `Voice`, `GetVoicesResponse` |
| `types.HistoryItemList`, `types.GetHistoryResponse` | `SpeechHistoryItem`, `GetSpeechHistoryResponse` |
| `types.Sample`, `types.FineTuningResponseModel`, `types.SharingOptions` | `VoiceSample`, `FineTuning`, `VoiceSharing` |
| `types.TimestampsGranularity*` | `TimestampsGranularity*` (same values) |

## Uploads

Uploads take `elevenlabs.File{Reader, Name, ContentType}` instead of `[]*os.File` or file paths and are streamed instead of buffered in memory.
Uploads are retried only when every reader implements `io.Seeker` (for example `*os.File` or `*bytes.Reader`).
The multipart field for voice samples is now `files` (matching the official SDK) instead of `files[]`, and labels are sent as a JSON object instead of a comma-separated string.

## Errors

| Old | New |
| --- | --- |
| `client.ErrUnauthorized` | `elevenlabs.ErrUnauthorized` (matched with `errors.Is`) |
| `client.ErrUnspecified` | removed; every non-2xx response is an `*APIError` unwrapping to a status sentinel (`ErrUnexpectedStatus` when there is none) |
| `types.ValidationError` joined into the error | `APIError.ValidationErrors` (`[]ValidationError`) |
| `types.ParamError` joined into the error | `APIError.Status` / `APIError.Message` |

## Package layout (pre-release root names)

Services moved out of the root package. Each package exports `Service` and
`NewService(opts ...option.RequestOption)`; the root `Client` fields keep
their names. Shared models moved to `types`.

| Old root name | New name |
| --- | --- |
| `OutputFormat*`, `TextNormalization*`, `PronunciationDictionaryLocator`, `AudioEffects`, `VoiceSettings`, `SortDirection*`, `CharacterAlignment`, `AudioFileFormat*` | `types.` + same name |
| `Voice`, `VoiceCategory*`, `VoiceSample`, `VoiceSharing`, `VoiceVerification`, `VerifiedVoiceLanguage`, `VoiceSafetyControl*`, `VoiceSharingModerationCheck` | `voices.Voice`, `voices.Category*`, `voices.Sample`, `voices.Sharing`, `voices.Verification`, `voices.VerifiedLanguage`, `voices.SafetyControl*`, `voices.SharingModerationCheck` |
| `TextToSpeechService`, `TextToSpeechRequest` | `texttospeech.Service`, `texttospeech.ConvertRequest` |
| `TextToSpeechStreamInputRequest`, `TextToSpeechStream`, `TextToSpeechMultiStream` | `texttospeech.StreamInputRequest`, `InputStream`, `MultiInputStream` |
| `TextToSpeechStreamInputMessage`, `TextToSpeechStreamOutput`, `TextToSpeechMultiStream{InputMessage,Output}` | `texttospeech.StreamInputMessage`, `StreamOutput`, `MultiStream{InputMessage,Output}` |
| `SpeechToTextService`, `SpeechToTextTranscriptsService` | `speechtotext.Service`, `speechtotext.TranscriptsService` |
| `SpeechToTextRealtime`, `SpeechToTextRealtimeRequest` | `speechtotext.RealtimeSession`, `speechtotext.RealtimeRequest` |
| `SpeechToText*` (e.g. `SpeechToTextConvertRequest`, `SpeechToTextModelScribeV1`, `SpeechToTextWord`), `MultichannelSpeechToText` | `speechtotext.*` without the prefix (`ConvertRequest`, `ModelScribeV1`, `Word`), `speechtotext.Multichannel` |
| `SpeechToSpeechService`, `SpeechToSpeechRequest` | `speechtospeech.Service`, `speechtospeech.ConvertRequest` |
| `TextToSoundEffectsService`, `CreateSoundEffectRequest`, `SoundEffectsModel*` | `soundeffects.Service`, `soundeffects.ConvertRequest`, `soundeffects.Model*` |
| `AudioIsolationService`, `AudioIsolationRequest`, `AudioIsolation*` | `audioisolation.Service`, `audioisolation.ConvertRequest`, `audioisolation.*` |
| `AudioEffectsService`, `AudioEffects{FilterPreset,Environment,BackgroundNoise,Catalog}` | `audioeffects.Service`, `audioeffects.{FilterPreset,Environment,BackgroundNoise,Catalog}` |
| `TextToDialogueService`, `TextToDialogueRequest`, `TextToDialogueInput` | `texttodialogue.Service`, `ConvertRequest`, `Input` |
| `TextToDialogueStreamInputRequest`, `TextToDialogueStream`, `TextToDialogueMultiStream`, `Dialogue*` | `texttodialogue.StreamInputRequest`, `InputStream`, `MultiInputStream`, `texttodialogue.*` without the prefix |
| `TextToVoiceService`, `TextToVoicePreviewService`, `VoiceDesignModel*` | `texttovoice.Service`, `PreviewService`, `Model*` |
| `VoicePreviewsRequest`, `VoiceDesignRequest`, `VoiceRemixRequest`, `CreateVoiceFromPreviewRequest`, `VoicePreview`, `VoiceDesignPreviews` | `texttovoice.PreviewsRequest`, `DesignRequest`, `RemixRequest`, `CreateRequest`, `Preview`, `DesignPreviews` |
| `ForcedAlignmentService`, `ForcedAlignmentRequest`, `ForcedAlignment{Character,Word}`, `ForcedAlignment` | `forcedalignment.Service`, `CreateRequest`, `Alignment{Character,Word}`, `Alignment` |
| `MusicService`, `MusicCompositionPlanService`, `MusicFinetunesService`, `Music*` | `music.Service`, `CompositionPlanService`, `FinetunesService`, `music.*` without the prefix (`MusicCompositionPlan` keeps its name) |
| `ModelsService`, `ModelRates` | `models.Service`, `models.Rates` |
| `UsageService`, `UsageGetRequest`, `UsageCharacters`, `UsageAggregationInterval*` | `usage.Service`, `GetRequest`, `Characters`, `AggregationInterval*` |
| `UserService`, `UserSubscriptionService` | `user.Service`, `user.SubscriptionService` |
| `TokensService`, `TokensSingleUseService` | `tokens.Service`, `tokens.SingleUseService` |
| `StatusResponse`, `ResourcePermission*` | `types.StatusResponse`, `types.ResourcePermission*` |
| `VoicesService`, `Voices{IVC,Settings,Samples,SamplesAudio,PVC,Accents,Collections}Service` | `voices.Service`, `voices.{IVC,Settings,Samples,SamplesAudio,PVC,Accents,Collections}Service` (`PVC*Service` unchanged) |
| `Voices{GetAll,Search,Get,Update,GetShared}Request`, `GetVoicesResponse`, `GetVoicesV2Response` | `voices.{GetAll,Search,Get,Update,GetShared}Request`, `GetAllResponse`, `SearchResponse` |
| `VoiceCollection`, `VoiceCollectionRequest`, `GetVoiceCollectionsResponse`, `DeleteVoiceCollectionResponse` | `voices.Collection`, `CollectionRequest`, `ListCollectionsResponse`, `DeleteCollectionResponse` |
| `VoiceAccent`, `GetVoiceAccentsResponse`, `GetLibraryVoicesResponse`, `VoiceSamplePreview`, `VoiceSampleWaveform` | `voices.Accent`, `AccentsResponse`, `LibraryVoicesResponse`, `SamplePreview`, `SampleWaveform` |
| other voice/IVC/PVC/library types (`IVCCreateRequest`, `PVC*Request`, `LibraryVoice`, `SharedVoice*`, `VoiceIDResponse`, ...) | `voices.` + same name |
| `SamplesService` | `samples.Service` |
| `HistoryService`, `HistorySource*`, `HistoryListRequest`, `GetSpeechHistoryResponse`, `SpeechHistoryItem`, `HistoryAlignments`, `DownloadHistoryRequest` | `history.Service`, `Source*`, `ListRequest`, `ListResponse`, `Item`, `Alignments`, `DownloadRequest` (`DialogueInput` unchanged) |
| `PronunciationDictionariesService`, `PronunciationDictionaryRulesService`, `PronunciationRule`, `PronunciationRuleType*` | `pronunciation.Service`, `RulesService`, `Rule`, `RuleType*` (`AliasRule`, `PhonemeRule` unchanged) |
| `PronunciationDictionaryCreateFrom{File,Rules}Request`, `AddPronunciationDictionaryResponse`, `PronunciationDictionaryMetadata`, `PronunciationDictionaryWithRules`, `PronunciationDictionaryUpdateRequest`, `PronunciationDictionarySort*`, `PronunciationDictionariesListRequest`, `GetPronunciationDictionariesResponse`, `PronunciationDictionaryRulesResponse` | `pronunciation.CreateFrom{File,Rules}Request`, `AddResponse`, `Metadata`, `DictionaryWithRules`, `UpdateRequest`, `Sort*`, `ListRequest`, `ListResponse`, `RulesResponse` |
| `AudioNativeService`, `AudioNative*` | `audionative.Service`, `audionative.*` without the prefix |
| `AssetsService`, `AssetsListRequest`, `AssetListResponse`, `AssetCreateRequest` | `assets.Service`, `ListRequest`, `ListResponse`, `CreateRequest` (`Asset` unchanged) |
| `DubbingService`, `Dubbing*Service`, `Dubbing*` (e.g. `DubbingCreateRequest`, `DubbingMetadata`, `DubbingStatusDubbed`) | `dubbing.Service`, `dubbing.*Service`, `dubbing.*` without the prefix (`CreateRequest`, `Metadata`, `StatusDubbed`); struct fields keep their names (`DubbingID`, `DubbingStatuses`, ...) |
| `DubbingRealtime`, `DubbingRealtimeRequest`, `RealtimeTranslateMessage{,Type}`, `Translate*` message types | `dubbing.RealtimeSession`, `dubbing.RealtimeRequest`, `dubbing.RealtimeMessage{,Type}`, `dubbing.Realtime*` (e.g. `RealtimeAudio`, `RealtimeErrorAuth`) |
| `TranslationService`, `TranslationRealtime`, `TranslationRealtimeRequest`, `TranslationEagerness*`, `TranslationTTSModel*` | `translation.Service`, `translation.Session`, `RealtimeRequest`, `Eagerness*`, `TTSModel*` (messages are `*dubbing.RealtimeMessage`) |
| `StudioService`, `Studio*Service` | `studio.Service`, `studio.*Service` (model types unchanged) |
| `ProductionsService`, `Production{Orders,Deliverables,Items,Languages,Media}Service`, `ProductionLanguages` | `productions.Service`, `{Orders,Deliverables,Items,Languages,Media}Service`, `Languages` |
| `WorkspaceService`, `Workspace*Service`, `WorkspaceGroup`, `WorkspaceMember`, `WorkspaceResourceType*`, ... | `workspace.Service`, `workspace.*Service`, `Group`, `Member`, `ResourceType*`, ... (auth connection types: `workspace.` + same name) |
| `client.Workspaces.APIKeys` (`WorkspacesService`, `WorkspacesAPIKeysService`) | `client.Workspace.APIKeys` (`workspace.APIKeysService`) |
| `ResourceRole*` | `types.ResourcePermission*` (identical values) |
| `ResourceAccessInfo` | `types.ResourceAccessInfo` |
| `ServiceAccountsService`, `ServiceAccountAPIKeysService`, `ServiceAccount`, `ServiceAccountList`, `ServiceAccountCreate{Request,Response}`, `WorkspaceAPIKey{,List}` | `serviceaccounts.Service`, `APIKeysService`, `Account`, `ListResponse`, `Create{Request,Response}`, `APIKey{,List}` (permission types unchanged) |
| `EnvironmentVariablesService`, `EnvironmentVariable`, `EnvironmentVariable{Type,Value,CreateRequest,UpdateRequest}`, `EnvironmentVariables{ListRequest,Page}` | `envvars.Service`, `Variable`, `{Type,Value,CreateRequest,UpdateRequest}`, `{ListRequest,Page}` (`StringValue`, `SecretValue`, `AuthConnectionValue` unchanged) |
| `EnvironmentsService`, `AgentEnvironment`, `AgentEnvironmentList` | `environments.Service`, `Environment`, `ListResponse` |
| `WebhooksService`, `Webhook*` types, `WorkspaceWebhook{,List}`, `ConstructWebhookEvent`, `VerifyWebhookSignature`, `WithWebhook{Clock,Tolerance}`, `WebhookSignatureHeader`, `ErrWebhook*` | `webhooks.Service`, `webhooks.*` without the prefix, `Webhook`, `ListResponse`, `ConstructEvent`, `VerifySignature`, `With{Clock,Tolerance}`, `SignatureHeader`, `Err*` |
| `FlowsService`, `Flows*Service`, `FlowsWaitOptions`, `FlowsWaitTimeoutError`, `ErrFlowsWaitTimeout`, `FlowsVoiceSettings` | `flows.Service`, `flows.*Service`, `WaitOptions`, `WaitTimeoutError`, `ErrWaitTimeout`, `VoiceSettings` (generation/template types unchanged) |
| `SpeechEngineService`, `SpeechEngine`, `SpeechEngine*` types, `SpeechEngineSession`, `NewSpeechEngineSession`, `VerifySpeechEngineToken`, `ErrSpeechEngine*` | `speechengine.Service`, `Engine`, `speechengine.*` without the prefix (`Config`, `Callbacks`, `Claims`, ...), `Session`, `NewSession`, `VerifyToken`, `Err*` |
| `client.ConversationalAI` (`ConversationalAIService`) | `client.ConvAI` (`convai.Service`); sub-service field names unchanged |
| `ConvAI{Settings,Secrets,Dashboard,DashboardSettings,Analytics,LiveCount,LLM,LLMUsage,Realtime,Users}Service`, `ConvAISettings`, `ConvAIWebhooks` | `convai.{Settings,Secrets,...}Service`, `convai.Settings`, `convai.Webhooks` (other settings/secret/misc types: `convai.` + same name) |
| `ConversationSession`, `ConversationRequest`, `ConversationEvent*`, `ClientTools`, `ClientToolHandler`, `ErrTool*` | `convai.Session`, `convai.ConnectRequest`, `convai.ConversationEvent*`, `convai.ClientTools`, ... (same names) |
| `AddToKnowledgeBase`, `RAGIndexOverview`, `GetDocumentRAGIndexes`, `DeleteDocumentRAGIndex` | methods on `convai.Service`; their types live in `knowledgebase` (`KnowledgeBaseAddRequest` -> `knowledgebase.AddRequest`) |
| `AgentsService`, `Agent*Service`, `Agent*` / `Agents*` request and response types (e.g. `AgentsListRequest`, `AgentCreateRequest`, `AgentBranch`, `AgentSortBy*`) | `agents.Service`, `agents.*Service`, `agents.*` without the prefix (`ListRequest`, `CreateRequest`, `Branch`, `SortBy*`) |
| `AgentConfig`, `AgentPrompt`, `AgentTTSConfig`, `AgentPlatformSettings`, `AgentWorkflow`, `AgentConfigOverride`, `Agent`, `ConversationalConfig`, `LLM*`, `KnowledgeBaseLocator`, `KnowledgeBaseDocumentType*`, `ListMeta`, `PageRequest`, `ForceRequest`, `BatchResult`, `SignedURL`, `SIPMessage{,sPage}`, `DependentAgent{,sPage}`, `DependentBranch`, `ConversationInitiationClientData`, `ToolResponseMock` | `agents.` + same name (shared by the other convai packages) |
| `ConversationsService`, `Conversation*Service`, `Conversation*` types (`ConversationStatus*`, `ConversationFilter`, `ConversationListRequest`, `ConversationsPage`, `ConversationTag`, ...) | `conversations.Service`, `conversations.*Service`, prefix dropped (`Status*`, `Filter`, `ListRequest`, `Page`, `Tag`, ...); `Conversation` unchanged; the embedded `ConversationFilter` field is now `Filter` |
| `KnowledgeBase*Service`, `KnowledgeBase*` types (`KnowledgeBaseDocument`, `KnowledgeBaseListRequest`, `KnowledgeBaseSortBy*`, `KnowledgeBaseChunk`, ...) | `knowledgebase.Service`, `knowledgebase.*Service`, prefix dropped (`Document`, `ListRequest`, `SortBy*`, `Chunk`, ...) |
| `ToolsService`, `ToolExecutionsService`, `ToolType*`, `ToolConfig`, `ToolRequest`, `ToolsPage`, `ToolListRequest`, `ToolExecution*` | `tools.Service`, `ExecutionsService`, `Type*`, `Config`, `Request`, `Page`, `ListRequest`, `Execution*` (`Tool` unchanged) |
| `AgentTestsService`, `AgentTestFoldersService`, `TestInvocationsService`, `TestType*`, `AgentTest{Summary,Folder,ListRequest}`, `AgentTestsPage`, `TestInvocation*` | `agenttests.Service`, `FoldersService`, `InvocationsService`, `Type*`, `{Summary,Folder,ListRequest}`, `Page`, `Invocation*` (`AgentTest` unchanged) |
| `PhoneNumbersService`, `TwilioService`, `SIPTrunkService`, `ExotelService`, `WhatsAppService`, `WhatsAppAccountsService`, `BatchCallsService`, `TelephonyProvider*`, `TelephonyDirection*`, `TelephonyCallConfig` | `telephony.` + same service names (also grouped by `telephony.Service`), `Provider*`, `Direction*`, `CallConfig` (other telephony types: `telephony.` + same name) |
| `MCPServersService`, `MCP*Service`, `MCP*` types, `UpdateMCPServerRequest` | `mcp.Service`, `mcp.*Service`, prefix dropped (`Server`, `ServerConfig`, `ApprovalPolicy`, `ToolConfig`, ...), `UpdateServerRequest` |
| `TriageTicketsService`, `TriageTicket`, `TriageTicket*` types, `TriageTicketOpen` etc., `TriageIssue*`, `Create/UpdateTriageTicketRequest` | `triage.Service`, `Ticket`, prefix dropped (`Priority*`, `Source*`, `SortBy*`, `ListRequest`, `Page`), `StatusOpen` etc., `Issue*`, `CreateRequest`/`UpdateRequest` |
| `elevenlabs.File`, `NewFile`, `SetTo`, `ClearField`, `RealtimeError`, `CloseError`, `ErrSocketClosed`, ... (temporary root aliases) | `types.File`, `types.NewFile`, `types.SetTo`, `types.ClearField`, `realtime.Error`, `realtime.CloseError`, `realtime.ErrSocketClosed`, ... |
