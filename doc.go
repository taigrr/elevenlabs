// Package elevenlabs is an unofficial Go client for the ElevenLabs API
// (https://elevenlabs.io), modeled on the official TypeScript SDK.
//
// # Client
//
// Create a client with functional options. Without WithAPIKey the key is read
// from the ELEVENLABS_API_KEY environment variable:
//
//	client := elevenlabs.NewClient(
//		elevenlabs.WithAPIKey(os.Getenv("XI_API_KEY")),
//		elevenlabs.WithEnvironment(elevenlabs.EnvironmentProductionEU),
//		elevenlabs.WithTimeout(30*time.Second),
//		elevenlabs.WithMaxRetries(3),
//	)
//
// Endpoints are grouped into services that mirror the resources of the
// official SDK, for example client.TextToSpeech.Convert,
// client.Voices.Settings.Get or client.History.List. Every method that
// performs I/O takes a context.Context first and accepts trailing
// RequestOption values (WithRequestHeader, WithQueryParam, WithBodyParam,
// WithRequestTimeout, WithRequestMaxRetries, WithRequestAPIKey).
//
// # Request parameters
//
// Request structs use plain strings for optional string and enum fields (the
// empty string means "not set") and pointers for optional booleans and
// numbers; Ptr helps build them. Enumerations are named string types with
// typed constants.
//
// # Audio and uploads
//
// Audio and other binary responses are returned as an io.ReadCloser that
// streams from the network and must be closed by the caller. Newline
// delimited JSON streams are returned as *Stream[T]. Uploads take a File
// wrapping any io.Reader and are streamed as multipart bodies; if the reader
// implements io.Seeker the request can be retried.
//
// # Errors and retries
//
// Non-2xx responses are returned as *APIError, which unwraps to a sentinel
// error (ErrBadRequest, ErrUnauthorized, ErrNotFound, ErrUnprocessableEntity,
// ErrTooManyRequests, ErrServer, ...). Requests that exceed the configured
// timeout match ErrTimeout. Responses with status 408, 409, 429 or 5xx are
// retried with exponential backoff and jitter, honoring Retry-After and
// X-RateLimit-Reset.
//
// # Pagination
//
// Cursor-paginated endpoints offer both the raw page call (e.g.
// History.List) and an iterator over every item (e.g. History.ListAll)
// usable with range-over-func.
//
// # Webhooks
//
// Workspace webhooks are managed through client.Webhooks. Deliveries are
// signed with the webhook secret; verify them with ConstructWebhookEvent (or
// client.Webhooks.ConstructEvent), passing the raw request body and the
// ElevenLabs-Signature header. Verification failures match
// ErrWebhookMissingSignature, ErrWebhookMissingSecret,
// ErrWebhookInvalidHeader, ErrWebhookTimestamp or ErrWebhookSignature.
//
// # Long-running work
//
// Flows generations and template runs are asynchronous. Each create call has
// a CreateAndWait variant that polls, honoring the server's Retry-After,
// until the work is completed or failed. A failed result is returned rather
// than reported as an error; exceeding FlowsWaitOptions.WaitTimeout returns a
// *FlowsWaitTimeoutError (matching ErrFlowsWaitTimeout). In production prefer
// a webhook over polling.
package elevenlabs
