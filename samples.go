package elevenlabs

import (
	"context"
	"net/http"
)

// SamplesService groups the voice sample endpoints exposed at the top level
// of the official SDK (client.samples).
type SamplesService struct{ cfg *config }

// Delete removes a sample from a voice.
func (s *SamplesService) Delete(ctx context.Context, voiceID, sampleID string, opts ...RequestOption) (*StatusResponse, error) {
	return doValue[StatusResponse](ctx, s.cfg, &request{method: http.MethodDelete, path: pathf("/v1/voices/%s/samples/%s", voiceID, sampleID), opts: opts})
}
