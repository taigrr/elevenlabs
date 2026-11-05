package elevenlabs

import (
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/internal/testutil"
)

type recordedRequest = testutil.Recorded

type testServer struct{ *testutil.Server }

func (ts *testServer) all() []recordedRequest { return ts.All() }

func (ts *testServer) last(t *testing.T) recordedRequest {
	t.Helper()
	return ts.Last(t)
}

// newTestServer starts a recording server and a client pointed at it that
// never really sleeps between retries and has no jitter.
func newTestServer(t *testing.T, handler http.HandlerFunc, opts ...Option) (*Client, *testServer) {
	t.Helper()
	ts, clientOpts := testutil.NewServer(t, handler, opts...)
	return NewClient(clientOpts...), &testServer{ts}
}

func respond(status int, contentType, body string) http.HandlerFunc {
	return testutil.Respond(status, contentType, body)
}

func sequence(handlers ...http.HandlerFunc) http.HandlerFunc {
	return testutil.Sequence(handlers...)
}

func formFields(t *testing.T, r recordedRequest) map[string][]string {
	t.Helper()
	return testutil.FormFields(t, r)
}
