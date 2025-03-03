package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/taigrr/elevenlabs/client"
	"github.com/taigrr/elevenlabs/client/types"
)

func main() {
	ctx := context.Background()
	client := client.New(os.Getenv("XI_API_KEY"))

	filePath := os.Args[1]

	resp, err := client.ConvertSpeechToText(ctx, filePath, types.SpeechToTextRequest{
		ModelID:               types.SpeehToTextModelScribeV1,
		TimestampsGranularity: types.TimestampsGranularityWord,
		Diarize:               true,
	})

	if err != nil {
		panic(err)
	}

	bytes, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bytes))
}
