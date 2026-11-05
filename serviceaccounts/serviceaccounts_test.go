package serviceaccounts

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
			Name: "serviceAccounts.list", Call: func(ctx context.Context, c *Service) (any, error) { return c.List(ctx) }, Method: http.MethodGet, Path: "/v1/service-accounts",
			RespBody: `{"service-accounts":[{"service_account_user_id":"s","name":"n","api-keys":[{"name":"k","hint":"h","key_id":"id","service_account_user_id":"s","hashed_xi_api_key":"x"}]}]}`,
			Check: func(t *testing.T, r any) {
				sa := r.(*ListResponse).ServiceAccounts[0]
				if sa.ServiceAccountUserID != "s" || sa.APIKeys[0].KeyID != "id" {
					t.Errorf("sa = %+v", sa)
				}
			},
		},
		{
			Name: "serviceAccounts.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Create(ctx, &CreateRequest{Name: "n", DefaultSharingGroups: []DefaultSharingGroupConfig{{GroupID: "g", PermissionLevel: SharingPermissionEditor}}})
			},
			Method: http.MethodPost, Path: "/v1/service-accounts",
			Body:     `{"name":"n","default_sharing_groups":[{"group_id":"g","permission_level":"editor"}]}`,
			RespBody: `{"service-account-user-id":"s"}`,
			Check: func(t *testing.T, r any) {
				if r.(*CreateResponse).ServiceAccountUserID != "s" {
					t.Errorf("resp = %+v", r)
				}
			},
		},
		{
			Name:   "serviceAccounts.apiKeys.list",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.APIKeys.List(ctx, "s") },
			Method: http.MethodGet, Path: "/v1/service-accounts/s/api-keys", RespBody: `{"api-keys":[]}`,
		},
		{
			Name: "serviceAccounts.apiKeys.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.APIKeys.Create(ctx, "s", &APIKeyCreateRequest{Name: "k", Permissions: AllPermissions(), CharacterLimit: types.Ptr[int64](5)})
			},
			Method: http.MethodPost, Path: "/v1/service-accounts/s/api-keys",
			Body: `{"name":"k","permissions":"all","character_limit":5}`, RespBody: `{"xi-api-key":"sk","key_id":"id"}`,
			Check: func(t *testing.T, r any) {
				if k := r.(*APIKeyCreateResponse); k.XiAPIKey != "sk" || k.KeyID != "id" {
					t.Errorf("resp = %+v", k)
				}
			},
		},
		{
			Name: "serviceAccounts.apiKeys.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.APIKeys.Delete(ctx, "s", "k")
			},
			Method: http.MethodDelete, Path: "/v1/service-accounts/s/api-keys/k", RespBody: `{}`,
		},
		{
			Name: "serviceAccounts.apiKeys.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.APIKeys.Update(ctx, "s", "k", &APIKeyUpdateRequest{
					IsEnabled: types.Ptr(true), Permissions: types.Ptr(Permissions(PermissionTextToSpeech)),
					CharacterLimit: types.ClearField[int64](), AllowedIPs: types.SetTo([]string{"1.2.3.4"}), TTSConcurrencyLimit: types.SetTo(3),
				})
			},
			Method: http.MethodPatch, Path: "/v1/service-accounts/s/api-keys/k",
			Body:     `{"is_enabled":true,"permissions":["text_to_speech"],"character_limit":"clear","allowed_ips":["1.2.3.4"],"tts_concurrency_limit":3}`,
			RespBody: `{}`,
		},
		{
			Name: "serviceAccounts.apiKeys.disable",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.APIKeys.Disable(ctx, "s", "k")
			},
			Method: http.MethodPost, Path: "/v1/service-accounts/s/api-keys/k/disable", RespBody: `{}`,
		},
	})
}

func TestClearableAndPermissionsJSON(t *testing.T) {
	var c types.Clearable[int]
	if err := c.UnmarshalJSON([]byte(`"clear"`)); err != nil || !c.IsClear() {
		t.Errorf("clear = %+v, %v", c, err)
	}
	if err := c.UnmarshalJSON([]byte(`4`)); err != nil {
		t.Fatal(err)
	}
	if v, ok := c.Value(); !ok || v != 4 {
		t.Errorf("value = %d, %v", v, ok)
	}
	var p APIKeyPermissions
	if err := p.UnmarshalJSON([]byte(`["voices_read"]`)); err != nil || p.All || p.Scopes[0] != PermissionVoicesRead {
		t.Errorf("perms = %+v, %v", p, err)
	}
	if b, _ := Permissions().MarshalJSON(); string(b) != "[]" {
		t.Errorf("empty perms = %s", b)
	}
}
