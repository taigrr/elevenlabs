package testutil

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/taigrr/elevenlabs/v2/apierror"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

// Endpoint describes one API call made through a service of type S and
// what the server should observe.
type Endpoint[S any] struct {
	Name     string
	Call     func(ctx context.Context, s S) (any, error)
	Method   string
	Path     string
	Query    string
	Body     string              // expected JSON body, if any
	Form     map[string][]string // expected multipart fields, if any
	Accept   string              // expected Accept header, default application/json
	RespType string              // response content type, default application/json
	RespBody string
	Check    func(t *testing.T, result any)
}

// RunEndpoints checks every case against a recording server, then checks
// that 401, 404 and 422 responses map to the matching apierror sentinels.
func RunEndpoints[S any](t *testing.T, newService func(...option.RequestOption) S, cases []Endpoint[S]) {
	t.Helper()
	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			respType := tt.RespType
			if respType == "" {
				respType = "application/json"
			}
			ts, opts := NewServer(t, Respond(http.StatusOK, respType, tt.RespBody))
			result, err := tt.Call(context.Background(), newService(opts...))
			if err != nil {
				t.Fatalf("call failed: %v", err)
			}
			got := ts.Last(t)
			if got.Method != tt.Method || got.Path != tt.Path {
				t.Errorf("request = %s %s, want %s %s", got.Method, got.Path, tt.Method, tt.Path)
			}
			if got.RawQuery != tt.Query {
				t.Errorf("query = %q\nwant    %q", got.RawQuery, tt.Query)
			}
			wantAccept := tt.Accept
			if wantAccept == "" {
				wantAccept = "application/json"
			}
			if a := got.Header.Get("Accept"); a != wantAccept {
				t.Errorf("accept = %q, want %q", a, wantAccept)
			}
			if k := got.Header.Get("xi-api-key"); k != APIKey {
				t.Errorf("api key = %q", k)
			}
			switch {
			case tt.Body != "":
				if got.ContentType != "application/json" {
					t.Errorf("content type = %q", got.ContentType)
				}
				AssertJSON(t, got.Body, tt.Body)
			case tt.Form != nil:
				if fields := FormFields(t, got); !reflect.DeepEqual(fields, tt.Form) {
					t.Errorf("form = %v\nwant   %v", fields, tt.Form)
				}
			default:
				if len(got.Body) != 0 {
					t.Errorf("unexpected body %q", got.Body)
				}
			}
			if tt.Check != nil {
				tt.Check(t, result)
			}
		})
	}
	statuses := []struct {
		code     int
		sentinel error
	}{
		{http.StatusUnauthorized, apierror.ErrUnauthorized},
		{http.StatusNotFound, apierror.ErrNotFound},
		{http.StatusUnprocessableEntity, apierror.ErrUnprocessableEntity},
	}
	for _, tt := range cases {
		for _, st := range statuses {
			t.Run(tt.Name+"/"+http.StatusText(st.code), func(t *testing.T) {
				_, opts := NewServer(t, Respond(st.code, "application/json", `{"detail":{"status":"x","message":"nope"}}`))
				_, err := tt.Call(context.Background(), newService(opts...))
				if !errors.Is(err, st.sentinel) {
					t.Fatalf("error = %v, want %v", err, st.sentinel)
				}
				var apiErr *apierror.APIError
				if !errors.As(err, &apiErr) || apiErr.Message != "nope" {
					t.Errorf("api error = %+v", apiErr)
				}
			})
		}
	}
}

// AssertJSON reports whether got and want encode the same JSON value.
func AssertJSON(t testing.TB, got []byte, want string) {
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

// AssertJSONSubset checks every key of want is present with the same value in got.
func AssertJSONSubset(t testing.TB, got []byte, want string) {
	t.Helper()
	var gotM, wantM map[string]any
	_ = json.Unmarshal(got, &gotM)
	_ = json.Unmarshal([]byte(want), &wantM)
	for k, v := range wantM {
		if !reflect.DeepEqual(gotM[k], v) {
			t.Errorf("key %s = %v, want %v", k, gotM[k], v)
		}
	}
}

// ReadAll reads and closes rc, returning its contents as a string.
func ReadAll(rc io.ReadCloser, err error) (any, error) {
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	b, err := io.ReadAll(rc)
	return string(b), err
}

// DrainStream collects every item of s.
func DrainStream[T any](s *types.Stream[T], err error) (any, error) {
	if err != nil {
		return nil, err
	}
	defer s.Close()
	var out []T
	for v, err := range s.All() {
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

// CheckString returns a check asserting the result is the string want.
func CheckString(want string) func(t *testing.T, r any) {
	return func(t *testing.T, r any) {
		t.Helper()
		if r.(string) != want {
			t.Errorf("result = %q, want %q", r, want)
		}
	}
}

// File returns an in-memory file part.
func File(name, data string) types.File {
	return types.File{Reader: strings.NewReader(data), Name: name}
}

// VoidCall adapts an error-only call to the Endpoint.Call signature.
func VoidCall(err error) (any, error) { return nil, err }
