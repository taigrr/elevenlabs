package elevenlabs_test

import (
	"context"
	"log"
	"net/http"

	"github.com/taigrr/elevenlabs/v2"
)

func ExampleSpeechEngineService_Handler() {
	client := elevenlabs.NewClient()
	handler := client.SpeechEngine.Handler(elevenlabs.SpeechEngineCallbacks{
		OnTranscript: func(ctx context.Context, s *elevenlabs.SpeechEngineSession, transcript []elevenlabs.SpeechEngineTranscriptMessage) error {
			last := transcript[len(transcript)-1].Content
			return s.SendResponse(ctx, "You said: "+last)
		},
		OnError: func(_ *elevenlabs.SpeechEngineSession, err error) { log.Println(err) },
	})
	http.Handle("/api/speech-engine/ws", handler)
	// log.Fatal(http.ListenAndServe(":3001", nil))
}

func ExampleAudioEffectsService_GetCatalog() {
	client := elevenlabs.NewClient()
	catalog, err := client.AudioEffects.GetCatalog(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, preset := range catalog.FilterPresets {
		log.Println(preset.FilterPresetID, preset.Name)
	}
}
