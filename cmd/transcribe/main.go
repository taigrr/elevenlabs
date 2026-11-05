// Command transcribe prints the JSON transcript of an audio file using the
// ElevenLabs speech-to-text endpoint.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/taigrr/elevenlabs/v2"
	"github.com/taigrr/elevenlabs/v2/speechtotext"
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	if len(os.Args) < 2 {
		return errors.New("usage: transcribe <audio-file>")
	}
	opts := []elevenlabs.Option{}
	if key := os.Getenv("XI_API_KEY"); key != "" {
		opts = append(opts, elevenlabs.WithAPIKey(key))
	}
	client := elevenlabs.NewClient(opts...)

	f, err := os.Open(os.Args[1])
	if err != nil {
		return err
	}
	defer f.Close()

	resp, err := client.SpeechToText.Convert(ctx, &speechtotext.ConvertRequest{
		ModelID:               speechtotext.ModelScribeV1,
		File:                  &elevenlabs.File{Reader: f, Name: filepath.Base(f.Name())},
		TimestampsGranularity: speechtotext.TimestampsGranularityWord,
		Diarize:               elevenlabs.Ptr(true),
	})
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(resp)
}
