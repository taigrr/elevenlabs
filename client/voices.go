package client

import (
	"context"
	"net/http"
	"os"
	"strings"
)

func (c Client) AddVoiceV1VoicesAddPost(ctx context.Context, name string, files []*os.File, description string, labels string, localVarOptionals *VoicesApiAddVoiceV1VoicesAddPostOpts) (AddVoiceResponseModel, *http.Response, error) {
	localVarHttpMethod = strings.ToUpper("Post")
	localVarPath := a.client.cfg.BasePath + "/v1/voices/add"
}

func (c Client) DeleteVoiceV1VoicesVoiceIdDelete(ctx context.Context, voiceId string, localVarOptionals *VoicesApiDeleteVoiceV1VoicesVoiceIdDeleteOpts) (Object, *http.Response, error) {
	localVarHttpMethod = strings.ToUpper("Delete")
	localVarPath := a.client.cfg.BasePath + "/v1/voices/{voice_id}"
}

func (c Client) EditVoiceSettingsV1VoicesVoiceIdSettingsEditPost(ctx context.Context, body Settings, voiceId string, localVarOptionals *VoicesApiEditVoiceSettingsV1VoicesVoiceIdSettingsEditPostOpts) (Object, *http.Response, error) {
	localVarHttpMethod = strings.ToUpper("Post")
	localVarPath := a.client.cfg.BasePath + "/v1/voices/{voice_id}/settings/edit"
}

func (c Client) EditVoiceV1VoicesVoiceIdEditPost(ctx context.Context, name string, files []*os.File, description string, labels string, voiceId string, localVarOptionals *VoicesApiEditVoiceV1VoicesVoiceIdEditPostOpts) (Object, *http.Response, error) {
	localVarHttpMethod = strings.ToUpper("Post")
	localVarPath := a.client.cfg.BasePath + "/v1/voices/{voice_id}/edit"
}

func (c Client) GetDefaultVoiceSettingsV1VoicesSettingsDefaultGet(ctx context.Context) (VoiceSettingsResponseModel, *http.Response, error) {
	localVarHttpMethod = strings.ToUpper("Get")
	localVarPath := a.client.cfg.BasePath + "/v1/voices/settings/default"
}

func (c Client) GetVoiceSettingsV1VoicesVoiceIdSettingsGet(ctx context.Context, voiceId string, localVarOptionals *VoicesApiGetVoiceSettingsV1VoicesVoiceIdSettingsGetOpts) (VoiceSettingsResponseModel, *http.Response, error) {
	localVarHttpMethod = strings.ToUpper("Get")
	localVarPath := a.client.cfg.BasePath + "/v1/voices/{voice_id}/settings"
}

func (c Client) GetVoiceV1VoicesVoiceIdGet(ctx context.Context, voiceId string, localVarOptionals *VoicesApiGetVoiceV1VoicesVoiceIdGetOpts) (VoiceResponseModel, *http.Response, error) {
	localVarHttpMethod = strings.ToUpper("Get")
	localVarPath := a.client.cfg.BasePath + "/v1/voices/{voice_id}"
}

func (c Client) GetVoicesV1VoicesGet(ctx context.Context, localVarOptionals *VoicesApiGetVoicesV1VoicesGetOpts) (GetVoicesResponseModel, *http.Response, error) {
	localVarHttpMethod = strings.ToUpper("Get")
	localVarPath := a.client.cfg.BasePath + "/v1/voices"
}
