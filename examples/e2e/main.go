//go:build integration

// Command e2e runs the live end-to-end harness against the ElevenLabs API.
//
//	ELEVENLABS_API_KEY=... go run -tags integration ./examples/e2e [-out dir]
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/taigrr/elevenlabs/v2"
	"github.com/taigrr/elevenlabs/v2/integration/e2e"
)

func main() {
	outDir := flag.String("out", e2e.DefaultOutDir, "output directory")
	flag.Parse()
	key := os.Getenv(elevenlabs.APIKeyEnvVar)
	if key == "" {
		key = os.Getenv("XI_API_KEY")
	}
	if key == "" {
		log.Fatal("ELEVENLABS_API_KEY or XI_API_KEY must be set")
	}
	client := elevenlabs.NewClient(elevenlabs.WithAPIKey(key), elevenlabs.WithTimeout(90*time.Second))
	h, err := e2e.New(e2e.Config{Client: client, OutDir: *outDir, Logf: log.Printf})
	if err != nil {
		log.Fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	failed := 0
	for _, r := range h.RunAll(ctx) {
		fmt.Printf("%-28s %-7s %s\n", r.Area, r.Status, r.Detail)
		if r.Status == e2e.Fail {
			failed++
		}
	}
	if failed > 0 {
		os.Exit(1)
	}
}
