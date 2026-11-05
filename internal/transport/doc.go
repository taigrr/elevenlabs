// Package transport sends API requests: it builds URLs and bodies, applies
// headers and per-call options, retries retryable failures with the official
// SDK's backoff, maps non-2xx responses to *apierror.APIError and decodes
// responses.
package transport
