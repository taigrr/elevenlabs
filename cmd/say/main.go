package main

import (
	"bufio"
	"context"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/mp3"
	"github.com/gopxl/beep/v2/speaker"

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

	start := time.Now()
	defer func() {
		log.Println(time.Since(start))
	}()

	var text string
	if len(os.Args) > 1 {
		text = strings.Join(os.Args[1:], " ")
	} else {
		reader := bufio.NewReader(os.Stdin)
		b, _ := io.ReadAll(reader)
		text = string(b)
	}

	rc, err := client.TTSStream(ctx, text, ids[0], types.SynthesisOptions{Stability: 0.75, SimilarityBoost: 0.75, Style: 0.0, UseSpeakerBoost: false})
	if err != nil {
		panic(err)
	}
	defer rc.Close()

	streamer, format, err := mp3.Decode(rc)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}
