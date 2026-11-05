package conversations_test

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/taigrr/elevenlabs/v2"
	"github.com/taigrr/elevenlabs/v2/convai/conversations"
)

func ExampleAudioService_Get() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	var svc *conversations.AudioService = client.ConvAI.Conversations.Audio
	audio, err := svc.Get(ctx, "conversation-id")
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
