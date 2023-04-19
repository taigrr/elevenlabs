package main

import (
	"context"
	"os"

	"github.com/taigrr/elevenlabs/client"
	"github.com/taigrr/elevenlabs/client/types"
)

func main() {
	ctx := context.Background()
	client := client.New(os.Getenv("XI_API_KEY"))
	ids, err := client.GetVoiceIDs(ctx)
	if err != nil {
		panic(err)
	}
	saveFile, err := os.Create("sample.mp3")
	if err != nil {
		panic(err)
	}
	defer saveFile.Close()
	err = client.TTSWriter(ctx, saveFile, "hello, golang", ids[0], types.SynthesisOptions{Stability: 0.75, SimilarityBoost: 0.75})
	if err != nil {
		panic(err)
	}
}
