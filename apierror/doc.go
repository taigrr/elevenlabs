// Package apierror defines the errors returned by every ElevenLabs service.
//
// Non-2xx HTTP responses are returned as *APIError, which unwraps to one of
// the sentinel errors (ErrNotFound, ErrUnauthorized, ...) so callers can use
// errors.Is and errors.As. Requests exceeding their timeout wrap ErrTimeout.
// The root elevenlabs package re-exports all of these names.
package apierror
