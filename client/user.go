package client

import (
	"context"
	"net/http"
	"strings"
)

func (c Client) GetUserInfoV1UserGet(ctx context.Context) (UserResponseModel, *http.Response, error) {
	localVarHttpMethod = strings.ToUpper("Get")
	localVarPath := a.client.cfg.BasePath + "/v1/user"
}

func (c Client) GetSubscriptionInfo(ctx context.Context) error {
	localVarHttpMethod = strings.ToUpper("Get")
	localVarPath := a.client.cfg.BasePath + "/v1/user/subscription"
}
