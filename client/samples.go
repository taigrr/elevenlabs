package client

import (
	"context"
	"net/http"
	"strings"
)

func (c Client) DeleteSampleV1VoicesVoiceIdSamplesSampleIdDelete(ctx context.Context, voiceId string, sampleId string, localVarOptionals *SamplesApiDeleteSampleV1VoicesVoiceIdSamplesSampleIdDeleteOpts) (Object, *http.Response, error) {
	localVarHttpMethod = strings.ToUpper("Delete")
	localVarPath := a.client.cfg.BasePath + "/v1/voices/{voice_id}/samples/{sample_id}"
}

func (c Client) GetAudioFromSampleV1VoicesVoiceIdSamplesSampleIdAudioGet(ctx context.Context, voiceId string, sampleId string, localVarOptionals *SamplesApiGetAudioFromSampleV1VoicesVoiceIdSamplesSampleIdAudioGetOpts) (*http.Response, error) {
	localVarHttpMethod = strings.ToUpper("Get")
	localVarPath := a.client.cfg.BasePath + "/v1/voices/{voice_id}/samples/{sample_id}/audio"
}
