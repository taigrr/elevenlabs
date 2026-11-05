// Package option provides the functional options accepted by
// elevenlabs.NewClient, by every service constructor and by every API call.
//
// A single RequestOption type is used everywhere: options passed to
// NewClient apply to every request, and options passed to a method apply on
// top of them for that call only.
//
//	client := elevenlabs.NewClient(option.WithMaxRetries(3))
//	user, err := client.User.Get(ctx, option.WithTimeout(5*time.Second))
//
// The root elevenlabs package re-exports these options.
package option
