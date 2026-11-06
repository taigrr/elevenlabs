//go:build integration

package integration_test

import (
	"context"
	"os"
	"testing"

	"github.com/taigrr/elevenlabs/v2/integration/e2e"
)

func TestLiveE2E(t *testing.T) {
	c := liveClient(t)
	outDir := os.Getenv("ELEVENLABS_E2E_OUT")
	h, err := e2e.New(e2e.Config{Client: c, OutDir: outDir, Logf: t.Logf})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(h.Cleanup)
	var results []e2e.Result
	for _, area := range e2e.Areas() {
		t.Run(area.Name, func(t *testing.T) {
			r := h.RunArea(context.Background(), area)
			results = append(results, r)
			switch r.Status {
			case e2e.Fail:
				t.Error(r.Detail)
			case e2e.Skipped:
				t.Skip(r.Detail)
			default:
				t.Log(r.Detail)
			}
		})
	}
	if err := h.WriteResults(results); err != nil {
		t.Error(err)
	}
}
