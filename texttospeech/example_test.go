package texttospeech_test

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/taigrr/elevenlabs/v2"
	"github.com/taigrr/elevenlabs/v2/texttospeech"
	"github.com/taigrr/elevenlabs/v2/types"
)

func ExampleService_Convert() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	audio, err := client.TextToSpeech.Convert(ctx, "JBFqnCBsd6RMkjVDRZzb", &texttospeech.ConvertRequest{
		Text:         "The first move is what sets everything in motion.",
		ModelID:      "eleven_multilingual_v2",
		OutputFormat: types.OutputFormatMP3_44100_128,
		VoiceSettings: &types.VoiceSettings{
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

func ExampleService_StreamInput() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	stream, err := client.TextToSpeech.StreamInput(ctx, "JBFqnCBsd6RMkjVDRZzb", &texttospeech.StreamInputRequest{
		ModelID:      "eleven_flash_v2_5",
		OutputFormat: "mp3_44100_128",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	go func() {
		_ = stream.Send(ctx, &texttospeech.InitializeConnection{})
		for _, chunk := range []string{"Hello ", "from ", "a streaming ", "socket. "} {
			_ = stream.SendText(ctx, chunk)
		}
		_ = stream.CloseSend(ctx)
	}()

	f, err := os.Create("streamed.mp3")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if _, err := io.Copy(f, stream.AudioReader(ctx)); err != nil {
		log.Fatal(err)
	}
}
