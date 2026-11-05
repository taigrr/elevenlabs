package apierror_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/taigrr/elevenlabs/v2/apierror"
)

func newError(status int, body string) error {
	resp := &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"application/json"}}}
	return apierror.New(resp, []byte(body))
}

func TestAPIErrorMapping(t *testing.T) {
	tests := []struct {
		status      int
		body        string
		sentinel    error
		wantMessage string
		wantStatus  string
		wantInError string
	}{
		{400, `{"detail":{"status":"invalid_input","message":"bad text"}}`, apierror.ErrBadRequest, "bad text", "invalid_input", "invalid_input: bad text"},
		{401, `{"detail":{"status":"invalid_api_key","message":"Invalid API key"}}`, apierror.ErrUnauthorized, "Invalid API key", "invalid_api_key", "401 unauthorized"},
		{403, `{"detail":"forbidden thing"}`, apierror.ErrForbidden, "forbidden thing", "", "forbidden thing"},
		{404, `{"detail":{"message":"voice not found"}}`, apierror.ErrNotFound, "voice not found", "", "voice not found"},
		{409, `{}`, apierror.ErrConflict, "", "", "409 conflict"},
		{425, `{}`, apierror.ErrTooEarly, "", "", "425"},
		{422, `{"detail":[{"loc":["body","text"],"msg":"field required","type":"missing"},{"loc":["query",0],"msg":"bad","type":"x"}]}`, apierror.ErrUnprocessableEntity, "", "", "body.text: field required; query.0: bad"},
		{429, `not json`, apierror.ErrTooManyRequests, "", "", "not json"},
		{500, ``, apierror.ErrServer, "", "", "500 internal server error"},
		{503, `{}`, apierror.ErrServer, "", "", "503"},
		{402, `{}`, apierror.ErrUnexpectedStatus, "", "", "402 payment required"},
	}
	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			err := newError(tt.status, tt.body)
			if !errors.Is(err, tt.sentinel) {
				t.Fatalf("errors.Is(%v, %v) = false", err, tt.sentinel)
			}
			var apiErr *apierror.APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("errors.As failed for %T", err)
			}
			if apiErr.StatusCode != tt.status || string(apiErr.Body) != tt.body || apiErr.Header == nil {
				t.Errorf("api error fields: %+v", apiErr)
			}
			if apiErr.Message != tt.wantMessage || apiErr.Status != tt.wantStatus {
				t.Errorf("message/status = %q/%q", apiErr.Message, apiErr.Status)
			}
			if !strings.Contains(err.Error(), tt.wantInError) {
				t.Errorf("Error() = %q, want substring %q", err.Error(), tt.wantInError)
			}
		})
	}
}

func TestValidationErrorsParsed(t *testing.T) {
	body := `{"detail":[{"loc":["body",3,"name"],"msg":"too long","type":"value_error"}]}`
	var apiErr *apierror.APIError
	if !errors.As(newError(422, body), &apiErr) {
		t.Fatal("not an APIError")
	}
	want := apierror.ValidationError{Loc: []string{"body", "3", "name"}, Msg: "too long", Type: "value_error"}
	if len(apiErr.ValidationErrors) != 1 || apiErr.ValidationErrors[0].Msg != want.Msg ||
		strings.Join(apiErr.ValidationErrors[0].Loc, ",") != "body,3,name" || apiErr.ValidationErrors[0].Type != want.Type {
		t.Errorf("validation errors = %+v", apiErr.ValidationErrors)
	}
	if got := apiErr.ValidationErrors[0].String(); got != "body.3.name: too long" {
		t.Errorf("String() = %q", got)
	}
}

func TestForStatus(t *testing.T) {
	if apierror.ForStatus(404) != apierror.ErrNotFound || apierror.ForStatus(599) != apierror.ErrServer || apierror.ForStatus(418) != apierror.ErrUnexpectedStatus {
		t.Error("ForStatus mapping mismatch")
	}
}
