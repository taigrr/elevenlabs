package realtime

import (
	"errors"
	"testing"

	"github.com/taigrr/elevenlabs/v2/apierror"
)

func TestErrorIs(t *testing.T) {
	tests := []struct {
		err  *Error
		want error
	}{
		{&Error{MessageType: "auth_error"}, apierror.ErrUnauthorized},
		{&Error{MessageType: "quota_exceeded"}, apierror.ErrTooManyRequests},
		{&Error{MessageType: "commit_throttled"}, apierror.ErrTooManyRequests},
		{&Error{MessageType: "input_error"}, apierror.ErrBadRequest},
		{&Error{MessageType: "unaccepted_terms"}, apierror.ErrForbidden},
		{&Error{MessageType: "x", Code: 404}, apierror.ErrNotFound},
		{&Error{MessageType: "transcriber_error"}, ErrRealtime},
	}
	for _, tt := range tests {
		if !errors.Is(tt.err, tt.want) || !errors.Is(tt.err, ErrRealtime) {
			t.Errorf("%v: Is(%v) = false", tt.err, tt.want)
		}
	}
	if errors.Is(&Error{MessageType: "transcriber_error"}, apierror.ErrUnauthorized) {
		t.Error("transcriber_error matched apierror.ErrUnauthorized")
	}
	if got := (&Error{MessageType: "e", Code: 3, Message: "m", Param: "p"}).Error(); got != "elevenlabs: realtime e (3): m [param p]" {
		t.Errorf("Error() = %q", got)
	}
	if got := (&CloseError{Code: 1011, Reason: "boom"}).Error(); got != "elevenlabs: websocket closed with status 1011: boom" {
		t.Errorf("CloseError.Error() = %q", got)
	}
}
