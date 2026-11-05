package speechtotext_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/taigrr/elevenlabs/v2"
	"github.com/taigrr/elevenlabs/v2/speechtotext"
)

func ExampleService_Convert() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	f, err := os.Open("meeting.mp3")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	resp, err := client.SpeechToText.Convert(ctx, &speechtotext.ConvertRequest{
		ModelID: speechtotext.ModelScribeV1,
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

func ExampleService_Realtime() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	conn, err := client.SpeechToText.Realtime(ctx, &speechtotext.RealtimeRequest{
		ModelID:     "scribe_v2_realtime",
		AudioFormat: speechtotext.RealtimeAudioPCM16000,
		SampleRate:  16000,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	go func() {
		// Raw 16 kHz mono 16-bit PCM; SendFrom commits at EOF.
		pcm, err := os.Open("speech.pcm")
		if err != nil {
			log.Print(err)
			return
		}
		defer pcm.Close()
		_ = conn.SendFrom(ctx, pcm)
	}()

	for {
		msg, err := conn.Recv(ctx)
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			log.Fatal(err)
		}
		if msg.MessageType == speechtotext.RealtimeCommittedTranscript {
			fmt.Println(msg.Text)
			return
		}
	}
}
