package elevenlabs_test

import (
	"context"
	"fmt"
	"log"

	"github.com/taigrr/elevenlabs/v2"
)

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
