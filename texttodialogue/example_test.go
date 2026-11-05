package texttodialogue_test

import (
	"context"
	"errors"
	"io"
	"log"
	"os"

	"github.com/taigrr/elevenlabs/v2"
	"github.com/taigrr/elevenlabs/v2/texttodialogue"
)

func ExampleService_StreamInput() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	const alice, bob = "JBFqnCBsd6RMkjVDRZzb", "21m00Tcm4TlvDq8ikWAM"
	stream, err := client.TextToDialogue.StreamInput(ctx, &texttodialogue.StreamInputRequest{
		ModelID:      "eleven_v3",
		OutputFormat: "mp3_44100_128",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	if err := errors.Join(
		stream.Send(ctx, &texttodialogue.StreamMessage{Voices: []string{alice, bob}}),
		stream.SendInputs(ctx,
			texttodialogue.StreamInput{Text: "Is the Go client ready?", VoiceID: alice},
			texttodialogue.StreamInput{Text: "It is!", VoiceID: bob, NewTurn: elevenlabs.Ptr(true)}),
		stream.Flush(ctx),
		stream.CloseSend(ctx),
	); err != nil {
		log.Fatal(err)
	}

	f, err := os.Create("dialogue.mp3")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if _, err := io.Copy(f, stream.AudioReader(ctx)); err != nil {
		log.Fatal(err)
	}
}

func ExampleService_MultiStreamInput() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	stream, err := client.TextToDialogue.MultiStreamInput(ctx, &texttodialogue.StreamInputRequest{ModelID: "eleven_v3"})
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	const voice = "JBFqnCBsd6RMkjVDRZzb"
	if err := errors.Join(
		stream.Send(ctx, &texttodialogue.MultiStreamMessage{ContextID: "intro", Voices: []string{voice}}),
		stream.SendInputs(ctx, "intro", texttodialogue.StreamInput{Text: "Welcome back.", VoiceID: voice}),
		stream.FlushContext(ctx, "intro"),
		stream.CloseContext(ctx, "intro"),
	); err != nil {
		log.Fatal(err)
	}

	var audio []byte
	for {
		out, err := stream.Recv(ctx)
		if err != nil {
			log.Fatal(err)
		}
		audio = append(audio, out.Audio...)
		if out.IsFinal {
			break
		}
	}
	_ = stream.CloseSend(ctx)
	_ = os.WriteFile("intro.mp3", audio, 0o644)
}
