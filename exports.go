package elevenlabs

import (
	"net/http"
	"time"

	"github.com/taigrr/elevenlabs/v2/apierror"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

// Option configures the client, a service or a single API call. It is an
// alias of option.RequestOption.
type Option = option.RequestOption

// Environment is the base URL of an ElevenLabs API deployment.
type Environment = option.Environment

// Environments mirror the ElevenLabsEnvironment values of the official SDK.
const (
	EnvironmentProduction          = option.EnvironmentProduction
	EnvironmentProductionUS        = option.EnvironmentProductionUS
	EnvironmentProductionEU        = option.EnvironmentProductionEU
	EnvironmentProductionIndia     = option.EnvironmentProductionIndia
	EnvironmentProductionSingapore = option.EnvironmentProductionSingapore
)

const (
	// DefaultUserAgent is sent as the User-Agent header unless overridden with WithUserAgent.
	DefaultUserAgent = option.DefaultUserAgent
	// DefaultTimeout is the per-attempt request timeout, matching the official SDK.
	DefaultTimeout = option.DefaultTimeout
	// DefaultMaxRetries is the number of retries for retryable failures, matching the official SDK.
	DefaultMaxRetries = option.DefaultMaxRetries
	// APIKeyEnvVar is the environment variable consulted by NewClient for a default API key.
	APIKeyEnvVar = option.APIKeyEnvVar
)

// WithAPIKey sets the xi-api-key used to authenticate requests.
// It overrides the ELEVENLABS_API_KEY environment variable.
func WithAPIKey(apiKey string) Option { return option.WithAPIKey(apiKey) }

// WithBaseURL points requests at a custom base URL (for example a proxy or a test server).
func WithBaseURL(baseURL string) Option { return option.WithBaseURL(baseURL) }

// WithEnvironment selects one of the predefined ElevenLabs deployments, such as a data residency region.
func WithEnvironment(env Environment) Option { return option.WithEnvironment(env) }

// WithHTTPClient sets the *http.Client used to send requests.
// A nil client restores the default.
func WithHTTPClient(hc *http.Client) Option { return option.WithHTTPClient(hc) }

// WithTimeout sets the per-attempt timeout. Zero or negative disables the timeout.
func WithTimeout(d time.Duration) Option { return option.WithTimeout(d) }

// WithMaxRetries sets the number of retries for retryable responses
// (408, 409, 429 and 5xx). Negative values are treated as zero.
func WithMaxRetries(n int) Option { return option.WithMaxRetries(n) }

// WithHeader sets a header sent with requests, replacing any value set by
// an earlier option for the same key.
func WithHeader(key, value string) Option { return option.WithHeader(key, value) }

// WithUserAgent overrides the User-Agent header.
func WithUserAgent(ua string) Option { return option.WithUserAgent(ua) }

// WithQueryParam adds an extra query string parameter to requests.
func WithQueryParam(key, value string) Option { return option.WithQueryParam(key, value) }

// WithBodyParam merges an extra top-level property into JSON request bodies.
// It has no effect on multipart or bodiless requests.
func WithBodyParam(key string, value any) Option { return option.WithBodyParam(key, value) }

// APIError is returned for any non-2xx response. See apierror.APIError.
type APIError = apierror.APIError

// ValidationError is one entry of an HTTP validation error response.
type ValidationError = apierror.ValidationError

// Sentinel errors, re-exported from package apierror. Every *APIError
// unwraps to the sentinel matching its status code.
var (
	ErrBadRequest          = apierror.ErrBadRequest
	ErrUnauthorized        = apierror.ErrUnauthorized
	ErrForbidden           = apierror.ErrForbidden
	ErrNotFound            = apierror.ErrNotFound
	ErrConflict            = apierror.ErrConflict
	ErrTooEarly            = apierror.ErrTooEarly
	ErrUnprocessableEntity = apierror.ErrUnprocessableEntity
	ErrTooManyRequests     = apierror.ErrTooManyRequests
	ErrServer              = apierror.ErrServer
	ErrUnexpectedStatus    = apierror.ErrUnexpectedStatus
	ErrTimeout             = apierror.ErrTimeout
)

// Ptr returns a pointer to v. It is a convenience for populating optional
// request fields, e.g. Seed: elevenlabs.Ptr(42).
func Ptr[T any](v T) *T { return types.Ptr(v) }
