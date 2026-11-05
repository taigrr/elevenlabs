package dubbing_test

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/taigrr/elevenlabs/v2"
	"github.com/taigrr/elevenlabs/v2/dubbing"
	"github.com/taigrr/elevenlabs/v2/types"
)

func ExampleService_Create() {
	ctx := context.Background()
	client := elevenlabs.NewClient()

	f, err := os.Open("talk.mp4")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	dub, err := client.Dubbing.Create(ctx, &dubbing.CreateRequest{
		File:       &types.File{Reader: f, Name: "talk.mp4", ContentType: "video/mp4"},
		TargetLang: "es",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dub.DubbingID, dub.ExpectedDurationSec)
}
