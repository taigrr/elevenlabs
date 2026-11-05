package tokens

import (
	"context"
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/internal/testutil"
)

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, []testutil.Endpoint[*Service]{
		{
			Name: "singleUse.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.SingleUse.Create(ctx, SingleUseTokenTypeRealtimeScribe)
			},
			Method: http.MethodPost, Path: "/v1/single-use-token/realtime_scribe", RespBody: `{"token":"tk"}`,
			Check: func(t *testing.T, r any) {
				if r.(*SingleUseToken).Token != "tk" {
					t.Errorf("token = %+v", r)
				}
			},
		},
	})
}
