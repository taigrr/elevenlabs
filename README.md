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
After creating your account, you can get your API key [from here](https://help.elevenlabs.io/hc/en-us/articles/14599447207697-How-to-authorize-yourself-using-your-xi-api-key-).

## Test Program

To test out an example `say` program, run:

`go install github.com/taigrr/elevenlabs/cmd/say@latest`

Set the `XI_API_KEY` environment variable, and pipe it some text to give it a whirl!

## Example Code

### Text-to-Speech Example

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
        // load in an API key to create a client
        client := client.New(os.Getenv("XI_API_KEY"))
        // fetch a list of voice IDs from elevenlabs
        ids, err := client.GetVoiceIDs(ctx)
        if err != nil {
                panic(err)
        }
        // prepare a pipe for streaming audio directly to beep
        pipeReader, pipeWriter := io.Pipe()
        reader := bufio.NewReader(os.Stdin)
        text, _ := reader.ReadString('\n')
        go func() {
                // stream audio from elevenlabs using the first voice we found
                err = client.TTSStream(ctx, pipeWriter, text, ids[0], types.SynthesisOptions{Stability: 0.75, SimilarityBoost: 0.75, Style: 0.0, UseSpeakerBoost: true})
                if err != nil {
                        panic(err)
                }
                pipeWriter.Close()
        }()
        // decode and prepare the streaming mp3 as it comes through
        streamer, format, err := mp3.Decode(pipeReader)
        if err != nil {
                log.Fatal(err)
        }
        defer streamer.Close()
        speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
        done := make(chan bool)
        // play the audio
        speaker.Play(beep.Seq(streamer, beep.Callback(func() {
                done <- true
        })))
        <-done
}
```

### Sound Generation Example

The following example demonstrates how to generate sound effects using the Sound Generation API:

```go
package main

import (
        "context"
        "os"

        "github.com/taigrr/elevenlabs/client"
)

func main() {
        ctx := context.Background()
        // Create a new client with your API key
        client := client.New(os.Getenv("XI_API_KEY"))

        // Generate a sound effect and save it to a file
        f, err := os.Create("footsteps.mp3")
        if err != nil {
                panic(err)
        }
        defer f.Close()

        // Basic usage (using default duration and prompt influence)
        err = client.SoundGenerationWriter(ctx, f, "footsteps on wooden floor", 0, 0)
        if err != nil {
                panic(err)
        }

        // Advanced usage with custom duration and prompt influence
        audio, err := client.SoundGeneration(
                ctx,
                "heavy rain on a tin roof",
                5.0,    // Set duration to 5 seconds
                0.5,    // Set prompt influence to 0.5
        )
        if err != nil {
                panic(err)
        }
        os.WriteFile("rain.mp3", audio, 0644)
}
```
