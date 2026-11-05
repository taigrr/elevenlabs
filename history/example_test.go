package history_test

import (
	"context"
	"fmt"
	"log"

	"github.com/taigrr/elevenlabs/v2"
	"github.com/taigrr/elevenlabs/v2/history"
)

func ExampleService_ListAll() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	for item, err := range client.History.ListAll(ctx, &history.ListRequest{PageSize: elevenlabs.Ptr(100)}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(item.HistoryItemID)
	}
}
