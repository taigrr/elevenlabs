package elevenlabs_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/taigrr/elevenlabs/v2"
)

func ExampleNewClient() {
	client := elevenlabs.NewClient(
		elevenlabs.WithAPIKey("my-api-key"),
		elevenlabs.WithEnvironment(elevenlabs.EnvironmentProductionEU),
		elevenlabs.WithMaxRetries(3),
	)
	fmt.Println(client.BaseURL())
	// Output: https://api.eu.residency.elevenlabs.io
}

func ExampleAPIError() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(w, `{"detail":{"status":"voice_not_found","message":"A voice with that ID does not exist."}}`)
	}))
	defer server.Close()

	client := elevenlabs.NewClient(elevenlabs.WithBaseURL(server.URL), elevenlabs.WithAPIKey("key"))
	_, err := client.Voices.Get(context.Background(), "missing", nil)

	fmt.Println(errors.Is(err, elevenlabs.ErrNotFound))
	var apiErr *elevenlabs.APIError
	if errors.As(err, &apiErr) {
		fmt.Println(apiErr.StatusCode, apiErr.Status)
	}
	// Output:
	// true
	// 404 voice_not_found
}
