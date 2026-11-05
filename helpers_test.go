package elevenlabs

import (
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/internal/testutil"
)

// newTestServer starts a recording server and a client pointed at it that
// never really sleeps between retries and has no jitter.
func newTestServer(t *testing.T, handler http.HandlerFunc, opts ...Option) (*Client, *testutil.Server) {
	t.Helper()
	ts, clientOpts := testutil.NewServer(t, handler, opts...)
	return NewClient(clientOpts...), ts
}

func respond(status int, contentType, body string) http.HandlerFunc {
	return testutil.Respond(status, contentType, body)
}
