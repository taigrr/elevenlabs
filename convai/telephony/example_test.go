package telephony_test

import (
	"context"
	"fmt"
	"log"

	"github.com/taigrr/elevenlabs/v2"
	"github.com/taigrr/elevenlabs/v2/convai/telephony"
)

func ExampleBatchCallsService_ListAll() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	for batch, err := range client.ConvAI.BatchCalls.ListAll(ctx, &telephony.BatchCallListRequest{AgentID: "agent-id"}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(batch.Name, batch.Status, batch.TotalCallsFinished)
	}
}
