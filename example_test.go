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

func ExampleTextToSpeechService_Convert() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	audio, err := client.TextToSpeech.Convert(ctx, "JBFqnCBsd6RMkjVDRZzb", &elevenlabs.TextToSpeechRequest{
		Text:         "The first move is what sets everything in motion.",
		ModelID:      "eleven_multilingual_v2",
		OutputFormat: elevenlabs.OutputFormatMP3_44100_128,
		VoiceSettings: &elevenlabs.VoiceSettings{
			Stability: elevenlabs.Ptr(0.5),
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer audio.Close()

	f, err := os.Create("speech.mp3")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if _, err := io.Copy(f, audio); err != nil {
		log.Fatal(err)
	}
}

func ExampleSpeechToTextService_Convert() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	f, err := os.Open("meeting.mp3")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	resp, err := client.SpeechToText.Convert(ctx, &elevenlabs.SpeechToTextConvertRequest{
		ModelID: elevenlabs.SpeechToTextModelScribeV1,
		File:    &elevenlabs.File{Reader: f, Name: "meeting.mp3"},
		Diarize: elevenlabs.Ptr(true),
	})
	if err != nil {
		log.Fatal(err)
	}
	if resp.Chunk != nil {
		fmt.Println(resp.Chunk.Text)
	}
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
