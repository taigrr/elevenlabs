package client

import (
	"context"
	"net/http"
	"os"
	"strings"
)

func (c Client) CreateVoice(ctx context.Context, name string, files []*os.File, description string, labels string) (AddVoiceResponseModel, *http.Response, error) {
	localVarHttpMethod = strings.ToUpper("Post")
	localVarPath := a.client.cfg.BasePath + "/v1/voices/add"
}

func (c Client) DeleteVoice(ctx context.Context, voiceId string) (Object, *http.Response, error) {
	localVarHttpMethod = strings.ToUpper("Delete")
	localVarPath := a.client.cfg.BasePath + "/v1/voices/{voice_id}"
}

func (c Client) EditVoiceSettings(ctx context.Context, body Settings, voiceId string) (Object, *http.Response, error) {
	localVarHttpMethod = strings.ToUpper("Post")
	localVarPath := a.client.cfg.BasePath + "/v1/voices/{voice_id}/settings/edit"
}

func (c Client) EditVoice(ctx context.Context, name string, files []*os.File, description string, labels string, voiceId string) (Object, *http.Response, error) {
	localVarHttpMethod = strings.ToUpper("Post")
	localVarPath := a.client.cfg.BasePath + "/v1/voices/{voice_id}/edit"
}

func (c Client) DefaultVoiceSettings(ctx context.Context) (VoiceSettingsResponseModel, *http.Response, error) {
	localVarHttpMethod = strings.ToUpper("Get")
	localVarPath := a.client.cfg.BasePath + "/v1/voices/settings/default"
}

func (c Client) GetVoiceSettings(ctx context.Context, voiceId string) (VoiceSettingsResponseModel, *http.Response, error) {
	localVarHttpMethod = strings.ToUpper("Get")
	localVarPath := a.client.cfg.BasePath + "/v1/voices/{voice_id}/settings"
}

func (c Client) GetVoice(ctx context.Context, voiceId string) (VoiceResponseModel, *http.Response, error) {
	localVarHttpMethod = strings.ToUpper("Get")
	localVarPath := a.client.cfg.BasePath + "/v1/voices/{voice_id}"
}

func (c Client) GetVoices(ctx context.Context) (GetVoicesResponseModel, *http.Response, error) {
	localVarHttpMethod = strings.ToUpper("Get")
	localVarPath := a.client.cfg.BasePath + "/v1/voices"
}
