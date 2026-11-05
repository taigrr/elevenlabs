package requestconfig

import (
	"context"
	"maps"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"time"
)

// Defaults shared with the public option package.
const (
	DefaultBaseURL    = "https://api.elevenlabs.io"
	DefaultUserAgent  = "github.com/taigrr/elevenlabs/v2"
	DefaultTimeout    = 240 * time.Second
	DefaultMaxRetries = 2
	APIKeyEnvVar      = "ELEVENLABS_API_KEY"
)

// Config is the resolved configuration of a client or a single call.
type Config struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	// Headers are sent with every request, overriding defaults per key.
	Headers http.Header
	// Query holds extra query parameters appended to every request.
	Query url.Values
	// BodyParams are merged into JSON request bodies.
	BodyParams map[string]any
	// Timeout is the per-attempt timeout; zero or negative disables it.
	Timeout    time.Duration
	MaxRetries int
	UserAgent  string

	// Sleep, Now and Jitter are seams for retry timing; tests replace them.
	Sleep  func(ctx context.Context, d time.Duration) error
	Now    func() time.Time
	Jitter func() float64
}

// New returns a Config with the library defaults; the API key is read from
// the ELEVENLABS_API_KEY environment variable.
func New() *Config {
	return &Config{
		BaseURL:    DefaultBaseURL,
		APIKey:     os.Getenv(APIKeyEnvVar),
		HTTPClient: &http.Client{},
		Headers:    make(http.Header),
		Query:      make(url.Values),
		Timeout:    DefaultTimeout,
		MaxRetries: DefaultMaxRetries,
		UserAgent:  DefaultUserAgent,
		Sleep:      SleepContext,
		Now:        time.Now,
		Jitter:     rand.Float64,
	}
}

// Build returns New() with opts applied.
func Build[O ~func(*Config)](opts ...O) *Config {
	cfg := New()
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

// Clone returns a deep copy of c (the HTTP client is shared).
func (c *Config) Clone() *Config {
	out := *c
	out.Headers = c.Headers.Clone()
	if out.Headers == nil {
		out.Headers = make(http.Header)
	}
	out.Query = cloneValues(c.Query)
	out.BodyParams = maps.Clone(c.BodyParams)
	return &out
}

// With returns c unchanged when opts is empty, otherwise a clone of c with
// opts applied.
func With[O ~func(*Config)](c *Config, opts ...O) *Config {
	if len(opts) == 0 {
		return c
	}
	out := c.Clone()
	for _, opt := range opts {
		opt(out)
	}
	return out
}

func cloneValues(v url.Values) url.Values {
	out := make(url.Values, len(v))
	for k, vs := range v {
		out[k] = append([]string(nil), vs...)
	}
	return out
}

// SleepContext sleeps for d or until ctx is done.
func SleepContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
