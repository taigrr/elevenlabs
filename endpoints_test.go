package elevenlabs

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func assertJSON(t *testing.T, got []byte, want string) {
	t.Helper()
	var gotV, wantV any
	if err := json.Unmarshal(got, &gotV); err != nil {
		t.Fatalf("invalid JSON body %q: %v", got, err)
	}
	if err := json.Unmarshal([]byte(want), &wantV); err != nil {
		t.Fatalf("invalid want JSON: %v", err)
	}
	if !reflect.DeepEqual(gotV, wantV) {
		t.Errorf("body = %s\nwant   %s", got, want)
	}
}

// endpointCase describes one API call and what the server should observe.
type endpointCase struct {
	name     string
	call     func(ctx context.Context, c *Client) (any, error)
	method   string
	path     string
	query    string
	body     string              // expected JSON body, if any
	form     map[string][]string // expected multipart fields, if any
	accept   string
	respType string
	respBody string
	check    func(t *testing.T, result any)
}

func readAll(rc io.ReadCloser, err error) (any, error) {
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	b, err := io.ReadAll(rc)
	return string(b), err
}

func checkString(want string) func(t *testing.T, r any) {
	return func(t *testing.T, r any) {
		t.Helper()
		if r.(string) != want {
			t.Errorf("result = %q, want %q", r, want)
		}
	}
}

func TestEndpoints(t *testing.T) {
	for _, tt := range endpointCases() {
		t.Run(tt.name, func(t *testing.T) {
			respType := tt.respType
			if respType == "" {
				respType = contentTypeJSON
			}
			client, ts := newTestServer(t, respond(http.StatusOK, respType, tt.respBody))
			result, err := tt.call(context.Background(), client)
			if err != nil {
				t.Fatalf("call failed: %v", err)
			}
			got := ts.last(t)
			if got.Method != tt.method || got.Path != tt.path {
				t.Errorf("request = %s %s, want %s %s", got.Method, got.Path, tt.method, tt.path)
			}
			if got.RawQuery != tt.query {
				t.Errorf("query = %q\nwant    %q", got.RawQuery, tt.query)
			}
			wantAccept := tt.accept
			if wantAccept == "" {
				wantAccept = contentTypeJSON
			}
			if a := got.Header.Get(headerAccept); a != wantAccept {
				t.Errorf("accept = %q, want %q", a, wantAccept)
			}
			if k := got.Header.Get(headerAPIKey); k != "test-key" {
				t.Errorf("api key = %q", k)
			}
			switch {
			case tt.body != "":
				if got.ContentType != contentTypeJSON {
					t.Errorf("content type = %q", got.ContentType)
				}
				assertJSON(t, got.Body, tt.body)
			case tt.form != nil:
				if fields := formFields(t, got); !reflect.DeepEqual(fields, tt.form) {
					t.Errorf("form = %v\nwant   %v", fields, tt.form)
				}
			default:
				if len(got.Body) != 0 {
					t.Errorf("unexpected body %q", got.Body)
				}
			}
			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestEndpointsMapErrors(t *testing.T) {
	statuses := []struct {
		code     int
		sentinel error
	}{
		{http.StatusUnauthorized, ErrUnauthorized},
		{http.StatusNotFound, ErrNotFound},
		{http.StatusUnprocessableEntity, ErrUnprocessableEntity},
	}
	for _, tt := range endpointCases() {
		for _, st := range statuses {
			t.Run(tt.name+"/"+http.StatusText(st.code), func(t *testing.T) {
				client, _ := newTestServer(t, respond(st.code, contentTypeJSON, `{"detail":{"status":"x","message":"nope"}}`))
				_, err := tt.call(context.Background(), client)
				if !errors.Is(err, st.sentinel) {
					t.Fatalf("error = %v, want %v", err, st.sentinel)
				}
				var apiErr *APIError
				if !errors.As(err, &apiErr) || apiErr.Message != "nope" {
					t.Errorf("api error = %+v", apiErr)
				}
			})
		}
	}
}

func endpointCases() []endpointCase {
	cases := platformEndpointCases()
	cases = append(cases, convaiEndpointCases()...)
	return append(cases, productionEndpointCases()...)
}

func testFile(name, data string) File {
	return File{Reader: strings.NewReader(data), Name: name}
}

func voidCall(err error) (any, error) { return nil, err }
