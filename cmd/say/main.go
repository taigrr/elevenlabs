package main

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"

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
	buf := bytes.Buffer{}
	w := bufio.NewWriter(&buf)

	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')

	err = client.TTSWriter(ctx, w, text, ids[0], types.SynthesisOptions{Stability: 0.75, SimilarityBoost: 0.75})
	if err != nil {
		panic(err)
	}
	r := io.NopCloser(bytes.NewReader(buf.Bytes()))
	streamer, format, err := mp3.Decode(r)
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
