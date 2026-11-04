package elevenlabs_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/taigrr/elevenlabs/v2"
)

func ExampleTextToSpeechService_StreamInput() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	stream, err := client.TextToSpeech.StreamInput(ctx, "JBFqnCBsd6RMkjVDRZzb", &elevenlabs.TextToSpeechStreamInputRequest{
		ModelID:      "eleven_flash_v2_5",
		OutputFormat: "mp3_44100_128",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	go func() {
		_ = stream.Send(ctx, &elevenlabs.InitializeConnection{})
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

func ExampleSpeechToTextService_Realtime() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	conn, err := client.SpeechToText.Realtime(ctx, &elevenlabs.SpeechToTextRealtimeRequest{
		ModelID:     "scribe_v2_realtime",
		AudioFormat: elevenlabs.RealtimeAudioPCM16000,
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
		if msg.MessageType == elevenlabs.RealtimeCommittedTranscript {
			fmt.Println(msg.Text)
			return
		}
	}
}

func ExampleConversationalAIService_Connect() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	tools := &elevenlabs.ClientTools{}
	_ = tools.Register("get_time", func(context.Context, map[string]any) (any, error) {
		return "12:00", nil
	})
	conv, err := client.ConversationalAI.Connect(ctx, &elevenlabs.ConversationRequest{
		AgentID:     "agent_123",
		ClientTools: tools,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer conv.Close()

	_ = conv.SendUserMessage(ctx, "What time is it?")
	for {
		ev, err := conv.Recv(ctx)
		if err != nil {
			log.Fatal(err)
		}
		if ev.Type == elevenlabs.ConversationEventAgentResponse {
			fmt.Println(ev.AgentResponse.AgentResponse)
			return
		}
	}
}
