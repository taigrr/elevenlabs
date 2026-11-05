package flows_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/taigrr/elevenlabs/v2"
	"github.com/taigrr/elevenlabs/v2/flows"
)

func ExampleImageService_CreateAndWait() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	generation, err := client.Flows.Image.CreateAndWait(ctx, &flows.ImageGenerationRequest{
		ModelID: "bytedance-seedream-5-lite",
		Prompt:  "A corgi in a tiny lifeguard chair on a sunlit beach at golden hour, photorealistic",
	}, &flows.WaitOptions{WaitTimeout: 5 * time.Minute})
	if errors.Is(err, flows.ErrWaitTimeout) {
		log.Fatal("still generating; resume later with client.Flows.Image.Get")
	}
	if err != nil {
		log.Fatal(err)
	}
	if generation.Status == flows.GenerationStatusFailed {
		log.Fatal(generation.ErrorMessage)
	}
	fmt.Println(generation.ContentURL)
}
