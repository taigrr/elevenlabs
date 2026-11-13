package ws

import (
	"errors"
	"io"
	"testing"
)

func TestAudioReaderEmptyReadDoesNotFetch(t *testing.T) {
	reader := &AudioReader{Next: func() ([]byte, error) {
		t.Fatal("empty read fetched a chunk")
		return nil, io.EOF
	}}
	for _, buffer := range [][]byte{nil, make([]byte, 0)} {
		count, err := reader.Read(buffer)
		if count != 0 || err != nil {
			t.Fatalf("empty read = %d, %v; want 0, nil", count, err)
		}
	}
}

func TestAudioReaderPreservesChunkAndTerminalError(t *testing.T) {
	terminal := errors.New("stream failed")
	calls := 0
	reader := &AudioReader{Next: func() ([]byte, error) {
		calls++
		return []byte("audio"), terminal
	}}
	buffer := make([]byte, 2)
	count, err := reader.Read(buffer)
	if count != 2 || err != nil || string(buffer) != "au" {
		t.Fatalf("first read = %q, %d, %v", buffer, count, err)
	}
	if count, err := reader.Read(nil); count != 0 || err != nil {
		t.Fatalf("empty buffered read = %d, %v", count, err)
	}
	rest, err := io.ReadAll(reader)
	if string(rest) != "dio" || !errors.Is(err, terminal) || calls != 1 {
		t.Fatalf("remaining read = %q, %v; calls = %d", rest, err, calls)
	}
}
