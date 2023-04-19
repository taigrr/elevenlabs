# elevenlabs
[![License 0BSD](https://img.shields.io/badge/License-0BSD-pink.svg)](https://opensource.org/licenses/0BSD)
[![GoDoc](https://godoc.org/github.com/taigrr/elevenlabs?status.svg)](https://godoc.org/github.com/taigrr/elevenlabs)
[![Go Mod](https://img.shields.io/badge/go.mod-v1.20-blue)](go.mod)
[![Go Report Card](https://goreportcard.com/badge/github.com/taigrr/elevenlabs?branch=master)](https://goreportcard.com/report/github.com/taigrr/elevenlabs)
Unofficial [elevenlabs.io](https://beta.elevenlabs.io/) ([11.ai](http://11.ai)) voice synthesis client

This library is not affiliated with, nor associated with ElevenLabs in any way.

ElevenLabs' official api documentation, upon which this client has been
derived, [can be found here](https://api.elevenlabs.io/docs).

## Purpose
This go client provides an easy interface to create synthesized voices and
make TTS (text-to-speech) requests to elevenlabs.io


As a prerequisite, you must already have an account with elevenlabs.io.
After creating your account, you can get you API key [from here](https://help.elevenlabs.io/hc/en-us/articles/14599447207697-How-to-authorize-yourself-using-your-xi-api-key-).


To use this library, create a new client and send a TTS request to a voice.
The following code block illustrates how one might replicate the say/espeak
command, using the streaming endpoint.
I've opted to go with faiface's beep package, but you can also save the file
to an mp3 on-disk.
```go
package main

import (
        "bufio"
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
        pipeReader, pipeWriter := io.Pipe()
        reader := bufio.NewReader(os.Stdin)
        text, _ := reader.ReadString('\n')
        go func() {
                err = client.TTSStream(ctx, pipeWriter, text, ids[0], types.SynthesisOptions{Stability: 0.75, SimilarityBoost: 0.75})
                if err != nil {
                        panic(err)
                }
                pipeWriter.Close()
        }()
        streamer, format, err := mp3.Decode(pipeReader)
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
```
