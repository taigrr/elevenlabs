package samples

import (
	"context"
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/types"
)

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, []testutil.Endpoint[*Service]{
		{
			Name:     "samples.delete",
			Call:     func(ctx context.Context, c *Service) (any, error) { return c.Delete(ctx, "v1", "s1") },
			Method:   http.MethodDelete,
			Path:     "/v1/voices/v1/samples/s1",
			RespBody: `{"status":"ok"}`,
			Check:    checkStatusOK,
		},
	})
}

func checkStatusOK(t *testing.T, r any) {
	t.Helper()
	if s := r.(*types.StatusResponse); s.Status != "ok" {
		t.Errorf("status = %q", s.Status)
	}
}
