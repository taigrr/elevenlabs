package environments

import (
	"context"
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/internal/testutil"
)

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, []testutil.Endpoint[*Service]{
		{
			Name: "environments.list", Call: func(ctx context.Context, c *Service) (any, error) { return c.List(ctx) }, Method: http.MethodGet, Path: "/v1/convai/environments", RespBody: `{"environments":[{"name":"prod"}]}`},
		{
			Name:   "environments.create",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Create(ctx, "prod") },
			Method: http.MethodPost, Path: "/v1/convai/environments", Body: `{"name":"prod"}`, RespBody: `{"name":"prod"}`,
		},
		{
			Name:   "environments.delete",
			Call:   func(ctx context.Context, c *Service) (any, error) { return testutil.VoidCall(c.Delete(ctx, "prod")) },
			Method: http.MethodDelete, Path: "/v1/convai/environments/prod",
		},
	})
}
