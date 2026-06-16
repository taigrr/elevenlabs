package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/taigrr/elevenlabs/client/types"
)

func TestTTSWriterReturnsCopyError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "audio-stream")
	}))
	defer ts.Close()

	c := newTestClient(ts)
	err := c.TTSWriter(context.Background(), failingWriter{}, "hello", "model1", "voice1", types.SynthesisOptions{})
	if err == nil || err.Error() != "write failed" {
		t.Fatalf("err = %v, want write failed", err)
	}
}

func TestTTSStreamReturnsCopyError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "audio-stream")
	}))
	defer ts.Close()

	c := newTestClient(ts)
	err := c.TTSStream(context.Background(), failingWriter{}, "hello", "voice1", types.SynthesisOptions{})
	if err == nil || err.Error() != "write failed" {
		t.Fatalf("err = %v, want write failed", err)
	}
}

func TestDownloadVoiceSampleWriterReturnsCopyError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "sample-stream")
	}))
	defer ts.Close()

	c := newTestClient(ts)
	err := c.DownloadVoiceSampleWriter(context.Background(), failingWriter{}, "voice1", "sample1")
	if err == nil || err.Error() != "write failed" {
		t.Fatalf("err = %v, want write failed", err)
	}
}
