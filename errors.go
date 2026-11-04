package elevenlabs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Sentinel errors. Every *APIError unwraps to the sentinel matching its
// status code, so callers can write errors.Is(err, elevenlabs.ErrNotFound).
var (
	ErrBadRequest          = errors.New("elevenlabs: bad request")
	ErrUnauthorized        = errors.New("elevenlabs: unauthorized")
	ErrForbidden           = errors.New("elevenlabs: forbidden")
	ErrNotFound            = errors.New("elevenlabs: not found")
	ErrConflict            = errors.New("elevenlabs: conflict")
	ErrTooEarly            = errors.New("elevenlabs: too early")
	ErrUnprocessableEntity = errors.New("elevenlabs: unprocessable entity")
	ErrTooManyRequests     = errors.New("elevenlabs: too many requests")
	ErrServer              = errors.New("elevenlabs: server error")
	// ErrUnexpectedStatus is wrapped by API errors whose status has no dedicated sentinel.
	ErrUnexpectedStatus = errors.New("elevenlabs: unexpected status")
	// ErrTimeout is wrapped when a request exceeds the client or per-request
	// timeout. The returned error also matches context.DeadlineExceeded.
	ErrTimeout = errors.New("elevenlabs: timeout exceeded")
)

// maxErrorBody bounds how much of an error response body is retained.
const maxErrorBody = 1 << 20

// APIError is returned for any non-2xx response. Use errors.As to inspect it.
type APIError struct {
	// StatusCode is the HTTP status code of the response.
	StatusCode int
	// Header holds the response headers.
	Header http.Header
	// Body is the raw response body (truncated to 1 MiB).
	Body []byte
	// Status is the machine-readable detail.status value, when present.
	Status string
	// Message is the human-readable detail message, when present.
	Message string
	// ValidationErrors holds the parsed detail of 422 validation responses.
	ValidationErrors []ValidationError
}

// ValidationError is one entry of an HTTP validation error response.
type ValidationError struct {
	// Loc is the location of the offending field; numeric indices are rendered as strings.
	Loc  []string `json:"loc"`
	Msg  string   `json:"msg"`
	Type string   `json:"type"`
}

// UnmarshalJSON accepts loc items that are either strings or numbers.
func (v *ValidationError) UnmarshalJSON(data []byte) error {
	var raw struct {
		Loc  []json.RawMessage `json:"loc"`
		Msg  string            `json:"msg"`
		Type string            `json:"type"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	v.Msg, v.Type, v.Loc = raw.Msg, raw.Type, nil
	for _, item := range raw.Loc {
		var s string
		if err := json.Unmarshal(item, &s); err != nil {
			s = string(bytes.TrimSpace(item))
		}
		v.Loc = append(v.Loc, s)
	}
	return nil
}

// String formats the validation error as "loc: msg".
func (v ValidationError) String() string {
	if len(v.Loc) == 0 {
		return v.Msg
	}
	return strings.Join(v.Loc, ".") + ": " + v.Msg
}

func newAPIError(resp *http.Response, body []byte) *APIError {
	apiErr := &APIError{StatusCode: resp.StatusCode, Header: resp.Header, Body: body}
	var envelope struct {
		Detail json.RawMessage `json:"detail"`
	}
	if json.Unmarshal(body, &envelope) != nil || len(envelope.Detail) == 0 {
		return apiErr
	}
	switch envelope.Detail[0] {
	case '[':
		_ = json.Unmarshal(envelope.Detail, &apiErr.ValidationErrors)
	case '"':
		_ = json.Unmarshal(envelope.Detail, &apiErr.Message)
	case '{':
		var detail struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}
		if json.Unmarshal(envelope.Detail, &detail) == nil {
			apiErr.Status, apiErr.Message = detail.Status, detail.Message
		}
	}
	return apiErr
}

// Error implements error.
func (e *APIError) Error() string {
	var sb strings.Builder
	sb.WriteString("elevenlabs: status ")
	sb.WriteString(strconv.Itoa(e.StatusCode))
	if text := http.StatusText(e.StatusCode); text != "" {
		sb.WriteString(" ")
		sb.WriteString(strings.ToLower(text))
	}
	switch {
	case e.Message != "":
		sb.WriteString(": ")
		if e.Status != "" {
			sb.WriteString(e.Status + ": ")
		}
		sb.WriteString(e.Message)
	case len(e.ValidationErrors) > 0:
		parts := make([]string, len(e.ValidationErrors))
		for i, v := range e.ValidationErrors {
			parts[i] = v.String()
		}
		sb.WriteString(": ")
		sb.WriteString(strings.Join(parts, "; "))
	case len(e.Body) > 0 && len(e.Body) <= 512:
		sb.WriteString(": ")
		sb.WriteString(strings.TrimSpace(string(e.Body)))
	}
	return sb.String()
}

// Unwrap returns the sentinel error matching the status code.
func (e *APIError) Unwrap() error {
	switch code := e.StatusCode; {
	case code == http.StatusBadRequest:
		return ErrBadRequest
	case code == http.StatusUnauthorized:
		return ErrUnauthorized
	case code == http.StatusForbidden:
		return ErrForbidden
	case code == http.StatusNotFound:
		return ErrNotFound
	case code == http.StatusConflict:
		return ErrConflict
	case code == http.StatusTooEarly:
		return ErrTooEarly
	case code == http.StatusUnprocessableEntity:
		return ErrUnprocessableEntity
	case code == http.StatusTooManyRequests:
		return ErrTooManyRequests
	case code >= 500:
		return ErrServer
	default:
		return ErrUnexpectedStatus
	}
}

func timeoutError(method, path string, err error) error {
	return fmt.Errorf("%w: %s %s: %w", ErrTimeout, method, path, err)
}
