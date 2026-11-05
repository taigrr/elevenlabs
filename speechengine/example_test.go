package speechengine_test

import (
	"context"
	"log"
	"net/http"

	"github.com/taigrr/elevenlabs/v2"
	"github.com/taigrr/elevenlabs/v2/speechengine"
)

func ExampleService_Handler() {
	client := elevenlabs.NewClient()
	handler := client.SpeechEngine.Handler(speechengine.Callbacks{
		OnTranscript: func(ctx context.Context, s *speechengine.Session, transcript []speechengine.TranscriptMessage) error {
			last := transcript[len(transcript)-1].Content
			return s.SendResponse(ctx, "You said: "+last)
		},
		OnError: func(_ *speechengine.Session, err error) { log.Println(err) },
	})
	http.Handle("/api/speech-engine/ws", handler)
	// log.Fatal(http.ListenAndServe(":3001", nil))
}
