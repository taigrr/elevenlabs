package usage

import (
	"context"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
)

// Service groups the /v1/usage endpoints.
type Service struct{ cfg *requestconfig.Config }

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	return &Service{cfg: requestconfig.Build(opts...)}
}

// BreakdownType selects how usage is broken down.
type BreakdownType string

// BreakdownType values.
const (
	BreakdownTypeNone                 BreakdownType = "none"
	BreakdownTypeVoice                BreakdownType = "voice"
	BreakdownTypeVoiceMultiplier      BreakdownType = "voice_multiplier"
	BreakdownTypeUser                 BreakdownType = "user"
	BreakdownTypeGroups               BreakdownType = "groups"
	BreakdownTypeAPIKeys              BreakdownType = "api_keys"
	BreakdownTypeAllAPIKeys           BreakdownType = "all_api_keys"
	BreakdownTypeProductType          BreakdownType = "product_type"
	BreakdownTypeModel                BreakdownType = "model"
	BreakdownTypeResource             BreakdownType = "resource"
	BreakdownTypeRequestQueue         BreakdownType = "request_queue"
	BreakdownTypeRegion               BreakdownType = "region"
	BreakdownTypeSubresourceID        BreakdownType = "subresource_id"
	BreakdownTypeReportingWorkspaceID BreakdownType = "reporting_workspace_id"
	BreakdownTypeHasAPIKey            BreakdownType = "has_api_key"
	BreakdownTypeRequestSource        BreakdownType = "request_source"
)

// AggregationInterval is the time bucket of usage statistics.
type AggregationInterval string

// AggregationInterval values.
const (
	AggregationIntervalHour       AggregationInterval = "hour"
	AggregationIntervalDay        AggregationInterval = "day"
	AggregationIntervalWeek       AggregationInterval = "week"
	AggregationIntervalMonth      AggregationInterval = "month"
	AggregationIntervalCumulative AggregationInterval = "cumulative"
)

// MetricType selects the usage metric.
type MetricType string

// MetricType values.
const (
	MetricTypeCredits            MetricType = "credits"
	MetricTypeTTSCharacters      MetricType = "tts_characters"
	MetricTypeMinutesUsed        MetricType = "minutes_used"
	MetricTypeRequestCount       MetricType = "request_count"
	MetricTypeTTFBAvg            MetricType = "ttfb_avg"
	MetricTypeTTFBP95            MetricType = "ttfb_p95"
	MetricTypeFiatUnitsSpent     MetricType = "fiat_units_spent"
	MetricTypeConcurrency        MetricType = "concurrency"
	MetricTypeConcurrencyAverage MetricType = "concurrency_average"
)

// GetRequest is the request of Service.Get.
type GetRequest struct {
	// StartUnix is the start of the range in unix milliseconds. Required.
	StartUnix int64 `query:"start_unix"`
	// EndUnix is the end of the range in unix milliseconds. Required.
	EndUnix int64 `query:"end_unix"`
	// IncludeWorkspaceMetrics includes the whole workspace's usage.
	IncludeWorkspaceMetrics *bool `query:"include_workspace_metrics"`
	// BreakdownType selects how usage is broken down.
	BreakdownType BreakdownType `query:"breakdown_type"`
	// AggregationInterval is the time bucket (default day).
	AggregationInterval AggregationInterval `query:"aggregation_interval"`
	// AggregationBucketSize is a custom bucket size in seconds.
	AggregationBucketSize *int `query:"aggregation_bucket_size"`
	// Metric selects the usage metric.
	Metric MetricType `query:"metric"`
}

// Characters is a usage time series keyed by breakdown value.
type Characters struct {
	// Time holds the start of each bucket in unix milliseconds.
	Time []int64 `json:"time"`
	// Usage maps each breakdown value to one value per bucket.
	Usage map[string][]float64 `json:"usage"`
}

// Get returns usage statistics. Deprecated upstream in favor of the
// workspace analytics endpoints.
func (s *Service) Get(ctx context.Context, req *GetRequest, opts ...option.RequestOption) (*Characters, error) {
	return transport.DoValue[Characters](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/usage/character-stats", Query: transport.OrEmpty(req), Opts: opts})
}
