package elevenlabs_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/taigrr/elevenlabs/v2"
)

func ExampleNewClient() {
	client := elevenlabs.NewClient(
		elevenlabs.WithAPIKey("my-api-key"),
		elevenlabs.WithEnvironment(elevenlabs.EnvironmentProductionEU),
		elevenlabs.WithMaxRetries(3),
	)
	fmt.Println(client.BaseURL())
	// Output: https://api.eu.residency.elevenlabs.io
}

func ExampleHistoryService_ListAll() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	for item, err := range client.History.ListAll(ctx, &elevenlabs.HistoryListRequest{PageSize: elevenlabs.Ptr(100)}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(item.HistoryItemID)
	}
}

func ExamplePronunciationDictionariesService_CreateFromRules() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

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
	fmt.Println(dict.ID, dict.VersionID)
}

func ExampleAPIError() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"detail":{"status":"voice_not_found","message":"A voice with that ID does not exist."}}`)
	}))
	defer server.Close()

	client := elevenlabs.NewClient(elevenlabs.WithBaseURL(server.URL), elevenlabs.WithAPIKey("key"))
	_, err := client.Voices.Get(context.Background(), "missing", nil)

	fmt.Println(errors.Is(err, elevenlabs.ErrNotFound))
	var apiErr *elevenlabs.APIError
	if errors.As(err, &apiErr) {
		fmt.Println(apiErr.StatusCode, apiErr.Status)
	}
	// Output:
	// true
	// 404 voice_not_found
}

func ExampleConstructWebhookEvent() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}
		event, err := elevenlabs.ConstructWebhookEvent(payload, r.Header.Get(elevenlabs.WebhookSignatureHeader), os.Getenv("WEBHOOK_SECRET"))
		if err != nil {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}
		fmt.Println(event.Type)
	})
	mux := http.NewServeMux()
	mux.Handle("/webhooks/elevenlabs", handler)
}

func ExampleFlowsImageService_CreateAndWait() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	generation, err := client.Flows.Image.CreateAndWait(ctx, &elevenlabs.ImageGenerationRequest{
		ModelID: "bytedance-seedream-5-lite",
		Prompt:  "A corgi in a tiny lifeguard chair on a sunlit beach at golden hour, photorealistic",
	}, &elevenlabs.FlowsWaitOptions{WaitTimeout: 5 * time.Minute})
	if errors.Is(err, elevenlabs.ErrFlowsWaitTimeout) {
		log.Fatal("still generating; resume later with client.Flows.Image.Get")
	}
	if err != nil {
		log.Fatal(err)
	}
	if generation.Status == elevenlabs.GenerationStatusFailed {
		log.Fatal(generation.ErrorMessage)
	}
	fmt.Println(generation.ContentURL)
}

func ExampleDubbingService_Create() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	f, err := os.Open("talk.mp4")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	dub, err := client.Dubbing.Create(ctx, &elevenlabs.DubbingCreateRequest{
		File:       &elevenlabs.File{Reader: f, Name: "talk.mp4", ContentType: "video/mp4"},
		TargetLang: "es",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dub.DubbingID, dub.ExpectedDurationSec)
}

func ExampleAgentsService_Create() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	agent, err := client.ConversationalAI.Agents.Create(ctx, &elevenlabs.AgentCreateRequest{
		Name: "Support",
		ConversationConfig: elevenlabs.ConversationalConfig{
			Agent: &elevenlabs.AgentConfig{
				FirstMessage: "Hi, how can I help?",
				Prompt:       &elevenlabs.AgentPrompt{Prompt: "You are a helpful support agent.", LLM: elevenlabs.LLMGPT4oMini},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(agent.AgentID)
}

func ExampleConversationAudioService_Get() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	audio, err := client.ConversationalAI.Conversations.Audio.Get(ctx, "conversation-id")
	if err != nil {
		log.Fatal(err)
	}
	defer audio.Close()

	out, err := os.Create("conversation.mp3")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if _, err := io.Copy(out, audio); err != nil {
		log.Fatal(err)
	}
}

func ExampleBatchCallsService_ListAll() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	for batch, err := range client.ConversationalAI.BatchCalls.ListAll(ctx, &elevenlabs.BatchCallListRequest{AgentID: "agent-id"}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(batch.Name, batch.Status, batch.TotalCallsFinished)
	}
}
