package translation_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/taigrr/elevenlabs/v2"
	"github.com/taigrr/elevenlabs/v2/dubbing"
	"github.com/taigrr/elevenlabs/v2/speechtotext"
	"github.com/taigrr/elevenlabs/v2/translation"
)

func ExampleService_Realtime() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	conn, err := client.Translation.Realtime(ctx, &translation.RealtimeRequest{
		SourceLanguage: "en",
		TargetLanguage: "es",
		VoiceID:        "JBFqnCBsd6RMkjVDRZzb",
		InputFormat:    speechtotext.RealtimeAudioPCM16000,
		OutputFormat:   "pcm_16000",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	go func() {
		pcm, err := os.Open("speech.pcm")
		if err != nil {
			log.Print(err)
			return
		}
		defer pcm.Close()
		if _, err := io.Copy(conn.AudioWriter(ctx), pcm); err != nil {
			log.Print(err)
		}
		_ = conn.EndOfStream(ctx)
	}()

	for {
		msg, err := conn.Recv(ctx)
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			log.Fatal(err)
		}
		if msg.MessageType == dubbing.RealtimeTranslation {
			fmt.Println(msg.Text)
		}
	}
}
