// Command say reads text from its arguments or stdin and speaks it using the
// ElevenLabs streaming text-to-speech endpoint.
package main

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"

	"github.com/taigrr/elevenlabs"
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	opts := []elevenlabs.Option{}
	if key := os.Getenv("XI_API_KEY"); key != "" {
		opts = append(opts, elevenlabs.WithAPIKey(key))
	}
	client := elevenlabs.NewClient(opts...)

	voices, err := client.Voices.GetAll(ctx, nil)
	if err != nil {
		return err
	}
	if len(voices.Voices) == 0 {
		return errors.New("no voices available")
	}

	start := time.Now()
	defer func() { log.Println(time.Since(start)) }()

	var text string
	if len(os.Args) > 1 {
		text = strings.Join(os.Args[1:], " ")
	} else {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		text = string(b)
	}

	audio, err := client.TextToSpeech.Stream(ctx, voices.Voices[0].VoiceID, &elevenlabs.TextToSpeechRequest{
		Text: text,
		VoiceSettings: &elevenlabs.VoiceSettings{
			Stability:       elevenlabs.Ptr(0.75),
			SimilarityBoost: elevenlabs.Ptr(0.75),
		},
	})
	if err != nil {
		return err
	}
	defer audio.Close()

	streamer, format, err := mp3.Decode(audio)
	if err != nil {
		return err
	}
	defer streamer.Close()

	if err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)); err != nil {
		return err
	}
	done := make(chan struct{})
	speaker.Play(beep.Seq(streamer, beep.Callback(func() { close(done) })))
	<-done
	return nil
}
