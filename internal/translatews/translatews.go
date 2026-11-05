package translatews

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/taigrr/elevenlabs/v2/internal/ws"
)

type audioChunk struct {
	MessageType string `json:"message_type"`
	Audio       []byte `json:"audio_base_64"`
}

var endOfStreamMessage = json.RawMessage(`{"message_type":"end_of_stream"}`)

// Message is the constraint satisfied by a pointer to a server message type.
type Message[M any] interface {
	*M
	Err() error
	// Audio returns the decoded audio of audio messages.
	Audio() ([]byte, bool)
}

// Stream implements the shared protocol of the realtime dubbing and
// translation sockets for server messages of type M.
type Stream[M any, PM Message[M]] struct{ Socket *ws.Socket }

// SendAudio sends an input_audio_chunk.
func (t Stream[M, PM]) SendAudio(ctx context.Context, audio []byte) error {
	return t.Socket.WriteJSON(ctx, audioChunk{"input_audio_chunk", ws.NonNilBytes(audio)})
}

// EndOfStream sends end_of_stream.
func (t Stream[M, PM]) EndOfStream(ctx context.Context) error {
	return t.Socket.WriteJSON(ctx, endOfStreamMessage)
}

// AudioWriter returns an io.Writer that sends every Write as an audio chunk.
func (t Stream[M, PM]) AudioWriter(ctx context.Context) io.Writer {
	return ws.WriterFunc(func(p []byte) (int, error) {
		if err := t.SendAudio(ctx, bytes.Clone(p)); err != nil {
			return 0, err
		}
		return len(p), nil
	})
}

// Recv reads the next message; error messages are returned with their error.
func (t Stream[M, PM]) Recv(ctx context.Context) (*M, error) {
	var msg M
	if err := t.Socket.ReadJSON(ctx, &msg); err != nil {
		return nil, err
	}
	return &msg, PM(&msg).Err()
}

// AudioReader returns a reader over audio messages, skipping other types.
func (t Stream[M, PM]) AudioReader(ctx context.Context) io.Reader {
	return &ws.AudioReader{Next: func() ([]byte, error) {
		for {
			msg, err := t.Recv(ctx)
			if err != nil {
				return nil, err
			}
			if data, ok := PM(msg).Audio(); ok {
				return data, nil
			}
		}
	}}
}

// Close performs a normal WebSocket closure.
func (t Stream[M, PM]) Close() error { return t.Socket.Close() }
