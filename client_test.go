package elevenlabs

import (
	"context"
	"testing"
	"time"
)

func TestClientDefaultsAndOptions(t *testing.T) {
	t.Setenv(APIKeyEnvVar, "env-key")
	if got := NewClient().cfg.APIKey; got != "env-key" {
		t.Errorf("api key from env = %q", got)
	}
	if got := NewClient(WithAPIKey("explicit")).cfg.APIKey; got != "explicit" {
		t.Errorf("explicit api key = %q", got)
	}
	tests := []struct {
		env  Environment
		want string
	}{
		{EnvironmentProduction, "https://api.elevenlabs.io"},
		{EnvironmentProductionUS, "https://api.us.elevenlabs.io"},
		{EnvironmentProductionEU, "https://api.eu.residency.elevenlabs.io"},
		{EnvironmentProductionIndia, "https://api.in.residency.elevenlabs.io"},
		{EnvironmentProductionSingapore, "https://api.sg.residency.elevenlabs.io"},
	}
	for _, tt := range tests {
		if got := NewClient(WithEnvironment(tt.env)).BaseURL(); got != tt.want {
			t.Errorf("WithEnvironment(%q) = %q", tt.env, got)
		}
	}
	if got := NewClient(WithBaseURL("http://x/")).BaseURL(); got != "http://x" {
		t.Errorf("trailing slash not trimmed: %q", got)
	}
	c := NewClient(WithTimeout(time.Second), WithMaxRetries(-3), WithHTTPClient(nil), WithUserAgent("ua"))
	if c.cfg.Timeout != time.Second || c.cfg.MaxRetries != 0 || c.cfg.HTTPClient == nil || c.cfg.UserAgent != "ua" {
		t.Errorf("options not applied: %+v", c.cfg)
	}
	d := NewClient()
	if d.cfg.Timeout != DefaultTimeout || d.cfg.MaxRetries != DefaultMaxRetries || d.BaseURL() != string(EnvironmentProduction) {
		t.Errorf("unexpected defaults: %+v", d.cfg)
	}
}

func TestClientOptionsReachServices(t *testing.T) {
	client, ts := newTestServer(t, respond(200, contentTypeJSON, `{}`), WithHeader("X-Client", "c"))
	if _, err := client.User.Get(context.Background(), WithHeader("X-Call", "r")); err != nil {
		t.Fatal(err)
	}
	h := ts.last(t).Header
	if h.Get("X-Client") != "c" || h.Get("X-Call") != "r" || h.Get(headerAPIKey) != "test-key" || h.Get(headerUserAgent) != DefaultUserAgent {
		t.Errorf("headers = %v", h)
	}
}
