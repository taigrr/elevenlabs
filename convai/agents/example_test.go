package agents_test

import (
	"context"
	"fmt"
	"log"

	"github.com/taigrr/elevenlabs/v2"
	"github.com/taigrr/elevenlabs/v2/convai/agents"
)

func ExampleService_Create() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	agent, err := client.ConvAI.Agents.Create(ctx, &agents.CreateRequest{
		Name: "Support",
		ConversationConfig: agents.ConversationalConfig{
			Agent: &agents.AgentConfig{
				FirstMessage: "Hi, how can I help?",
				Prompt:       &agents.AgentPrompt{Prompt: "You are a helpful support agent.", LLM: agents.LLMGPT4oMini},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(agent.AgentID)
}
