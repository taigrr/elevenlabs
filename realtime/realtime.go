package realtime

import (
	"errors"
	"fmt"
	"strings"

	"github.com/taigrr/elevenlabs/v2/apierror"
)

// Sentinel errors for realtime (WebSocket) connections.
var (
	// ErrSocketClosed is wrapped by errors returned once a WebSocket
	// connection has been closed, either locally or abnormally by the server.
	// A normal closure by the server is reported as io.EOF instead.
	ErrSocketClosed = errors.New("elevenlabs: websocket closed")
	// ErrRealtime is wrapped by every *Error, i.e. by error messages
	// the server sends over a WebSocket connection.
	ErrRealtime = errors.New("elevenlabs: realtime error")
	// ErrInvalidParameter is returned when realtime connection parameters fail
	// client-side validation.
	ErrInvalidParameter = errors.New("elevenlabs: invalid parameter")
	// ErrUnexpectedMessage is wrapped when the server sends a message that
	// cannot be decoded.
	ErrUnexpectedMessage = errors.New("elevenlabs: unexpected websocket message")
)

// CloseError reports that the server closed a WebSocket connection with a
// status other than a normal closure. It matches ErrSocketClosed with
// errors.Is.
type CloseError struct {
	// Code is the WebSocket close status code, e.g. 1008 for policy violations.
	Code int
	// Reason is the close reason sent by the server.
	Reason string
}

// Error implements error.
func (e *CloseError) Error() string {
	if e.Reason == "" {
		return fmt.Sprintf("elevenlabs: websocket closed with status %d", e.Code)
	}
	return fmt.Sprintf("elevenlabs: websocket closed with status %d: %s", e.Code, e.Reason)
}

// Is reports whether target is ErrSocketClosed.
func (e *CloseError) Is(target error) bool { return target == ErrSocketClosed }

// Error is an error message received over a WebSocket connection,
// such as a Scribe "auth_error" or a dialogue stream error. It always matches
// ErrRealtime, and additionally the closest HTTP sentinel: authentication
// failures match ErrUnauthorized, rate limits and quotas match
// ErrTooManyRequests, and malformed input matches ErrBadRequest.
type Error struct {
	// MessageType is the server's message type, e.g. "auth_error" or
	// "rate_limited". It is empty for socket types without message types.
	MessageType string
	// Message is the human-readable error description.
	Message string
	// Code is the numeric error code, when the server provides one.
	Code int
	// Param names the offending parameter, when the server provides one.
	Param string
}

// Error implements error.
func (e *Error) Error() string {
	var sb strings.Builder
	sb.WriteString("elevenlabs: realtime")
	if e.MessageType != "" {
		sb.WriteString(" " + e.MessageType)
	} else {
		sb.WriteString(" error")
	}
	if e.Code != 0 {
		fmt.Fprintf(&sb, " (%d)", e.Code)
	}
	if e.Message != "" {
		sb.WriteString(": " + e.Message)
	}
	if e.Param != "" {
		sb.WriteString(" [param " + e.Param + "]")
	}
	return sb.String()
}

// Is matches ErrRealtime and the HTTP sentinel closest to the error type.
func (e *Error) Is(target error) bool {
	if target == ErrRealtime {
		return true
	}
	switch e.MessageType {
	case "auth_error":
		return target == apierror.ErrUnauthorized
	case "rate_limited", "quota_exceeded", "commit_throttled", "queue_overflow", "resource_exhausted":
		return target == apierror.ErrTooManyRequests
	case "invalid_request", "input_error", "chunk_size_exceeded":
		return target == apierror.ErrBadRequest
	case "unaccepted_terms", "payment_issue":
		return target == apierror.ErrForbidden
	}
	if e.Code >= 400 && e.Code < 600 {
		return apierror.ForStatus(e.Code) == target
	}
	return false
}
