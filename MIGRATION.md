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
