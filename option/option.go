package option

import (
	"net/http"
	"strings"
	"time"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
)

// RequestOption configures a client, a service or a single API call.
// Construct options with the With* functions of this package.
type RequestOption func(*requestconfig.Config)

// Environment is the base URL of an ElevenLabs API deployment.
type Environment string

// Environments mirror the ElevenLabsEnvironment values of the official SDK.
const (
	EnvironmentProduction          Environment = "https://api.elevenlabs.io"
	EnvironmentProductionUS        Environment = "https://api.us.elevenlabs.io"
	EnvironmentProductionEU        Environment = "https://api.eu.residency.elevenlabs.io"
	EnvironmentProductionIndia     Environment = "https://api.in.residency.elevenlabs.io"
	EnvironmentProductionSingapore Environment = "https://api.sg.residency.elevenlabs.io"
)

const (
	// DefaultUserAgent is sent as the User-Agent header unless overridden with WithUserAgent.
	DefaultUserAgent = requestconfig.DefaultUserAgent
	// DefaultTimeout is the per-attempt request timeout, matching the official SDK.
	DefaultTimeout = requestconfig.DefaultTimeout
	// DefaultMaxRetries is the number of retries for retryable failures, matching the official SDK.
	DefaultMaxRetries = requestconfig.DefaultMaxRetries
	// APIKeyEnvVar is the environment variable consulted for a default API key.
	APIKeyEnvVar = requestconfig.APIKeyEnvVar
)

// WithAPIKey sets the xi-api-key used to authenticate requests.
// It overrides the ELEVENLABS_API_KEY environment variable.
func WithAPIKey(apiKey string) RequestOption {
	return func(c *requestconfig.Config) { c.APIKey = apiKey }
}

// WithBaseURL points requests at a custom base URL (for example a proxy or a test server).
func WithBaseURL(baseURL string) RequestOption {
	return func(c *requestconfig.Config) { c.BaseURL = strings.TrimRight(baseURL, "/") }
}

// WithEnvironment selects one of the predefined ElevenLabs deployments, such as a data residency region.
func WithEnvironment(env Environment) RequestOption {
	return WithBaseURL(string(env))
}

// WithHTTPClient sets the *http.Client used to send requests.
// A nil client restores the default.
func WithHTTPClient(hc *http.Client) RequestOption {
	return func(c *requestconfig.Config) {
		if hc == nil {
			hc = &http.Client{}
		}
		c.HTTPClient = hc
	}
}

// WithTimeout sets the per-attempt timeout. Zero or negative disables the timeout.
func WithTimeout(d time.Duration) RequestOption {
	return func(c *requestconfig.Config) { c.Timeout = d }
}

// WithMaxRetries sets the number of retries for retryable responses
// (408, 409, 429 and 5xx). Negative values are treated as zero.
func WithMaxRetries(n int) RequestOption {
	return func(c *requestconfig.Config) { c.MaxRetries = max(n, 0) }
}

// WithHeader sets a header sent with requests, replacing any value set by
// an earlier option for the same key.
func WithHeader(key, value string) RequestOption {
	return func(c *requestconfig.Config) { c.Headers.Set(key, value) }
}

// WithUserAgent overrides the User-Agent header.
func WithUserAgent(ua string) RequestOption {
	return func(c *requestconfig.Config) { c.UserAgent = ua }
}

// WithQueryParam adds an extra query string parameter to requests.
func WithQueryParam(key, value string) RequestOption {
	return func(c *requestconfig.Config) { c.Query.Add(key, value) }
}

// WithBodyParam merges an extra top-level property into JSON request bodies.
// It has no effect on multipart or bodiless requests.
func WithBodyParam(key string, value any) RequestOption {
	return func(c *requestconfig.Config) {
		if c.BodyParams == nil {
			c.BodyParams = make(map[string]any)
		}
		c.BodyParams[key] = value
	}
}
