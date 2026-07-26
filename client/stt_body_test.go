package client

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/taigrr/elevenlabs/client/types"
)

type closeTrackingBody struct {
	*strings.Reader
	closed bool
}

func (b *closeTrackingBody) Close() error {
	b.closed = true
	return nil
}

func TestConvertSpeechToTextFromReaderClosesResponseBody(t *testing.T) {
	body := &closeTrackingBody{
		Reader: strings.NewReader(`{"language_code":"en","language_probability":0.99,"text":"hello"}`),
	}

	c := New("test-api-key").WithHTTPClient(&http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/v1/speech-to-text" {
				t.Fatalf("path = %q, want /v1/speech-to-text", req.URL.Path)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       body,
				Header:     make(http.Header),
				Request:    req,
			}, nil
		}),
	})

	response, err := c.ConvertSpeechToTextFromReader(
		context.Background(),
		strings.NewReader("audio"),
		"sample.mp3",
		types.SpeechToTextRequest{ModelID: types.SpeechToTextModelScribeV1},
	)
	if err != nil {
		t.Fatal(err)
	}
	if response.Text != "hello" {
		t.Fatalf("text = %q, want hello", response.Text)
	}
	if !body.closed {
		t.Fatal("response body was not closed")
	}
}
