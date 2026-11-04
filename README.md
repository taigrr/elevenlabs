# elevenlabs

[![License 0BSD](https://img.shields.io/badge/License-0BSD-pink.svg)](https://opensource.org/licenses/0BSD)
[![GoDoc](https://pkg.go.dev/badge/github.com/taigrr/elevenlabs/v2.svg)](https://pkg.go.dev/github.com/taigrr/elevenlabs/v2)
[![CI](https://github.com/taigrr/elevenlabs/actions/workflows/ci.yml/badge.svg)](https://github.com/taigrr/elevenlabs/actions/workflows/ci.yml)
[![Go Mod](https://img.shields.io/badge/go.mod-v1.27.2-blue)](go.mod)
[![Go Report Card](https://goreportcard.com/badge/github.com/taigrr/elevenlabs?branch=master)](https://goreportcard.com/report/github.com/taigrr/elevenlabs)

Unofficial [elevenlabs.io](https://beta.elevenlabs.io/) ([11.ai](http://11.ai)) voice synthesis client

This library is not affiliated with, nor associated with ElevenLabs in any way.

ElevenLabs' official api documentation, upon which this client has been
derived, [can be found here](https://api.elevenlabs.io/docs).

## Purpose

This Go client provides an idiomatic interface to the elevenlabs.io API.
It is modeled on the official TypeScript SDK: endpoints are grouped into services such as `client.TextToSpeech`, `client.Voices` and `client.History`, and every call takes a `context.Context`.
See [PARITY.md](PARITY.md) for endpoint coverage and [MIGRATION.md](MIGRATION.md) for upgrading from the previous `client` package.

As a prerequisite, you must already have an account with elevenlabs.io.
After creating your account, you can get your API key [from here](https://help.elevenlabs.io/hc/en-us/articles/14599447207697-How-to-authorize-yourself-using-your-xi-api-key-).

## Test Program

To test out an example `say` program, run:

`go install github.com/taigrr/elevenlabs/v2/cmd/say@latest`

Set the `XI_API_KEY` environment variable, and pipe it some text to give it a whirl!

## Example Code

### Creating a client

```go
client := elevenlabs.NewClient(
        elevenlabs.WithAPIKey(os.Getenv("XI_API_KEY")), // defaults to $ELEVENLABS_API_KEY
        elevenlabs.WithEnvironment(elevenlabs.EnvironmentProductionEU), // optional data residency
        elevenlabs.WithMaxRetries(3),
        elevenlabs.WithTimeout(60*time.Second),
)
```

Per-call behavior can be adjusted with request options such as `elevenlabs.WithRequestTimeout`, `elevenlabs.WithRequestHeader` and `elevenlabs.WithQueryParam`.

### Text-to-Speech Example

The following program replicates the say/espeak command using the streaming endpoint and gopxl's beep package.
Audio responses are `io.ReadCloser`s that stream from the network, so you can just as easily copy them to a file.

```go
package main

import (
        "context"
        "io"
        "log"
        "os"
        "time"

        "github.com/gopxl/beep/v2"
        "github.com/gopxl/beep/v2/mp3"
        "github.com/gopxl/beep/v2/speaker"

        "github.com/taigrr/elevenlabs/v2"
)

func main() {
        ctx := context.Background()
        client := elevenlabs.NewClient(elevenlabs.WithAPIKey(os.Getenv("XI_API_KEY")))

        voices, err := client.Voices.GetAll(ctx, nil)
        if err != nil {
                log.Fatal(err)
        }
        text, _ := io.ReadAll(os.Stdin)

        audio, err := client.TextToSpeech.Stream(ctx, voices.Voices[0].VoiceID, &elevenlabs.TextToSpeechRequest{
                Text:    string(text),
                ModelID: "eleven_multilingual_v2",
                VoiceSettings: &elevenlabs.VoiceSettings{
                        Stability:       elevenlabs.Ptr(0.75),
                        SimilarityBoost: elevenlabs.Ptr(0.75),
                },
        })
        if err != nil {
                log.Fatal(err)
        }
        defer audio.Close()

        streamer, format, err := mp3.Decode(audio)
        if err != nil {
                log.Fatal(err)
        }
        defer streamer.Close()
        speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
        done := make(chan struct{})
        speaker.Play(beep.Seq(streamer, beep.Callback(func() { close(done) })))
        <-done
}
```

### Sound Generation Example

```go
audio, err := client.TextToSoundEffects.Convert(ctx, &elevenlabs.CreateSoundEffectRequest{
        Text:            "heavy rain on a tin roof",
        DurationSeconds: elevenlabs.Ptr(5.0),
        PromptInfluence: elevenlabs.Ptr(0.5),
})
if err != nil {
        log.Fatal(err)
}
defer audio.Close()

f, err := os.Create("rain.mp3")
if err != nil {
        log.Fatal(err)
}
defer f.Close()
io.Copy(f, audio)
```

### Speech-to-Text Example

Uploads accept any `io.Reader` and are streamed rather than buffered.

```go
f, err := os.Open("meeting.mp3")
if err != nil {
        log.Fatal(err)
}
defer f.Close()

resp, err := client.SpeechToText.Convert(ctx, &elevenlabs.SpeechToTextConvertRequest{
        ModelID:               elevenlabs.SpeechToTextModelScribeV1,
        File:                  &elevenlabs.File{Reader: f, Name: "meeting.mp3"},
        TimestampsGranularity: elevenlabs.TimestampsGranularityWord,
        Diarize:               elevenlabs.Ptr(true),
})
if err != nil {
        log.Fatal(err)
}
fmt.Println(resp.Chunk.Text)
```

### Music Example

`ComposeDetailed` decodes the metadata part and streams the audio part.

```go
song, err := client.Music.ComposeDetailed(ctx, &elevenlabs.MusicComposeDetailedRequest{
        MusicGenerationParams: elevenlabs.MusicGenerationParams{Prompt: "upbeat synthwave", MusicLengthMs: elevenlabs.Ptr(int64(30000))},
})
if err != nil {
        log.Fatal(err)
}
defer song.Close()
fmt.Println(song.Metadata.SongMetadata.Title)
out, _ := os.Create(song.Filename)
defer out.Close()
io.Copy(out, song)
```

### Pronunciation Dictionary Example

```go
dict, err := client.PronunciationDictionaries.CreateFromRules(ctx, &elevenlabs.PronunciationDictionaryCreateFromRulesRequest{
        Name: "My Dictionary",
        Rules: []elevenlabs.PronunciationRule{
                elevenlabs.AliasRule("Thailand", "tie-land"),
                elevenlabs.PhonemeRule("tomato", "təˈmeɪtoʊ", "ipa"),
        },
})
if err != nil {
        log.Fatal(err)
}
_, err = client.PronunciationDictionaries.Rules.Add(ctx, dict.ID, []elevenlabs.PronunciationRule{elevenlabs.AliasRule("SQL", "sequel")})
```

### Dubbing Example

```go
f, _ := os.Open("talk.mp4")
defer f.Close()
dub, err := client.Dubbing.Create(ctx, &elevenlabs.DubbingCreateRequest{
        File:       &elevenlabs.File{Reader: f, Name: "talk.mp4", ContentType: "video/mp4"},
        TargetLang: "es",
})
if err != nil {
        log.Fatal(err)
}
// once client.Dubbing.Get reports status "dubbed":
audio, err := client.Dubbing.Audio.Get(ctx, dub.DubbingID, "es")
```

### Flows Example

```go
generation, err := client.Flows.Image.CreateAndWait(ctx, &elevenlabs.ImageGenerationRequest{
        ModelID: "bytedance-seedream-5-lite",
        Prompt:  "A corgi on a surfboard",
}, &elevenlabs.FlowsWaitOptions{WaitTimeout: 5 * time.Minute})
if err != nil {
        log.Fatal(err)
}
fmt.Println(generation.Status, generation.ContentURL)
```

### Conversational AI Example

```go
agent, err := client.ConversationalAI.Agents.Create(ctx, &elevenlabs.AgentCreateRequest{
        Name: "Support",
        ConversationConfig: elevenlabs.ConversationalConfig{
                Agent: &elevenlabs.AgentConfig{
                        FirstMessage: "Hi, how can I help?",
                        Prompt:       &elevenlabs.AgentPrompt{Prompt: "You are a helpful agent.", LLM: elevenlabs.LLMGPT4oMini},
                },
                TTS: &elevenlabs.AgentTTSConfig{VoiceID: "21m00Tcm4TlvDq8ikWAM"},
        },
})
if err != nil {
        log.Fatal(err)
}

f, _ := os.Open("faq.pdf")
defer f.Close()
doc, err := client.ConversationalAI.KnowledgeBase.Documents.CreateFromFile(ctx, &elevenlabs.DocumentFromFileRequest{
        File: elevenlabs.NewFile(f, "faq.pdf"),
})
if err != nil {
        log.Fatal(err)
}
fmt.Println("document", doc.ID)

for conv, err := range client.ConversationalAI.Conversations.ListAll(ctx, &elevenlabs.ConversationListRequest{
        ConversationFilter: elevenlabs.ConversationFilter{AgentID: agent.AgentID},
}) {
        if err != nil {
                log.Fatal(err)
        }
        audio, err := client.ConversationalAI.Conversations.Audio.Get(ctx, conv.ConversationID)
        // ... copy audio somewhere, then audio.Close()
}
```

Telephony, batch calls and MCP servers hang off the same service:

```go
call, err := client.ConversationalAI.Twilio.OutboundCall(ctx, &elevenlabs.TwilioOutboundCallRequest{
        OutboundCallRequest: elevenlabs.OutboundCallRequest{
                AgentID:            agent.AgentID,
                AgentPhoneNumberID: "phnum_123",
                ToNumber:           "+15551234567",
        },
})
if err != nil {
        log.Fatal(err)
}
fmt.Println(call.Success, call.Message)

batch, err := client.ConversationalAI.BatchCalls.Create(ctx, &elevenlabs.CreateBatchCallRequest{
        CallName:           "reminders",
        AgentID:            agent.AgentID,
        AgentPhoneNumberID: "phnum_123",
        Recipients:         []elevenlabs.BatchCallRecipient{{PhoneNumber: "+15551234567"}},
})
if err != nil {
        log.Fatal(err)
}
fmt.Println(batch.ID, batch.Status)

server, err := client.ConversationalAI.MCPServers.Create(ctx, &elevenlabs.MCPServerConfig{
        Name: "docs",
        URL:  "https://mcp.example.com/sse",
})
if err != nil {
        log.Fatal(err)
}
_, err = client.ConversationalAI.MCPServers.ApprovalPolicy.Update(ctx, server.ID, elevenlabs.MCPApprovalAutoApproveAll)
```

Large agent configuration sections (platform settings, workflows, tool
schemas) are exposed as `json.RawMessage` so they round-trip unchanged.

### Realtime (WebSocket) Example

Text-to-speech from text that arrives in pieces (for example from an LLM):

```go
stream, err := client.TextToSpeech.StreamInput(ctx, "JBFqnCBsd6RMkjVDRZzb", &elevenlabs.TextToSpeechStreamInputRequest{
        ModelID: "eleven_flash_v2_5",
})
if err != nil {
        log.Fatal(err)
}
defer stream.Close()

go func() {
        _ = stream.Send(ctx, &elevenlabs.InitializeConnection{})
        for _, chunk := range []string{"Hello ", "world. "} {
                _ = stream.SendText(ctx, chunk)
        }
        _ = stream.CloseSend(ctx)
}()
_, err = io.Copy(out, stream.AudioReader(ctx))
```

Realtime transcription (Scribe), agent conversations, dialogue, dubbing and
translation follow the same Send/Recv pattern:

```go
conn, err := client.SpeechToText.Realtime(ctx, &elevenlabs.SpeechToTextRealtimeRequest{
        ModelID:     "scribe_v2_realtime",
        AudioFormat: elevenlabs.RealtimeAudioPCM16000,
})
if err != nil {
        log.Fatal(err)
}
defer conn.Close()
go conn.SendFrom(ctx, pcm) // or io.Copy(conn.AudioWriter(ctx), pcm); conn.Commit(ctx)
for {
        msg, err := conn.Recv(ctx)
        if err != nil {
                break // io.EOF, *elevenlabs.CloseError or *elevenlabs.RealtimeError
        }
        if msg.MessageType == elevenlabs.RealtimeCommittedTranscript {
                fmt.Println(msg.Text)
        }
}

conv, err := client.ConversationalAI.Connect(ctx, &elevenlabs.ConversationRequest{AgentID: "agent_123"})
if err != nil {
        log.Fatal(err)
}
defer conv.Close()
_ = conv.SendUserMessage(ctx, "Hi!")
ev, err := conv.Recv(ctx) // pings are answered automatically
```

### Webhook Verification Example

```go
http.HandleFunc("/webhooks/elevenlabs", func(w http.ResponseWriter, r *http.Request) {
        payload, _ := io.ReadAll(r.Body)
        event, err := elevenlabs.ConstructWebhookEvent(payload, r.Header.Get(elevenlabs.WebhookSignatureHeader), secret)
        if err != nil {
                http.Error(w, "invalid signature", http.StatusUnauthorized)
                return
        }
        log.Println(event.Type, string(event.Data))
})
```

### Pagination

```go
for item, err := range client.History.ListAll(ctx, &elevenlabs.HistoryListRequest{PageSize: elevenlabs.Ptr(100)}) {
        if err != nil {
                log.Fatal(err)
        }
        fmt.Println(item.HistoryItemID)
}
```

### Errors

Non-2xx responses are returned as `*elevenlabs.APIError`, which matches status sentinels with `errors.Is`:

```go
_, err := client.Voices.Get(ctx, "missing", nil)
if errors.Is(err, elevenlabs.ErrNotFound) {
        // handle missing voice
}
var apiErr *elevenlabs.APIError
if errors.As(err, &apiErr) {
        log.Println(apiErr.StatusCode, apiErr.Message)
}
```

Requests with status 408, 409, 429 or 5xx are retried automatically with exponential backoff, honoring `Retry-After`.
