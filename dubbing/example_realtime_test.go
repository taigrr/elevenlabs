package dubbing_test

import (
	"context"
	"errors"
	"io"
	"log"
	"os"

	"github.com/taigrr/elevenlabs/v2"
	"github.com/taigrr/elevenlabs/v2/dubbing"
	"github.com/taigrr/elevenlabs/v2/speechtotext"
)

func ExampleService_Realtime() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	conn, err := client.Dubbing.Realtime(ctx, &dubbing.RealtimeRequest{
		SourceLanguage: "en",
		TargetLanguage: "es",
		InputFormat:    speechtotext.RealtimeAudioPCM16000,
		OutputFormat:   speechtotext.RealtimeAudioPCM16000,
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

	out, err := os.Create("dubbed.pcm")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	if _, err := io.Copy(out, conn.AudioReader(ctx)); err != nil && !errors.Is(err, io.EOF) {
		log.Fatal(err)
	}
}
