package usage

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
			Name: "get",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Get(ctx, &GetRequest{StartUnix: 1, EndUnix: 2, IncludeWorkspaceMetrics: types.Ptr(true), BreakdownType: BreakdownTypeVoice, AggregationInterval: AggregationIntervalDay, AggregationBucketSize: types.Ptr(60), Metric: MetricTypeCredits})
			},
			Method: http.MethodGet, Path: "/v1/usage/character-stats",
			Query:    "aggregation_bucket_size=60&aggregation_interval=day&breakdown_type=voice&end_unix=2&include_workspace_metrics=true&metric=credits&start_unix=1",
			RespBody: `{"time":[1],"usage":{"All":[5]}}`,
			Check: func(t *testing.T, r any) {
				if u := r.(*Characters); len(u.Time) != 1 || u.Usage["All"][0] != 5 {
					t.Errorf("usage = %+v", u)
				}
			},
		},
	})
}
