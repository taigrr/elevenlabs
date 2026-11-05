package envvars

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
			Name: "environmentVariables.list", Call: func(ctx context.Context, c *Service) (any, error) {
				return c.List(ctx, &ListRequest{PageSize: types.Ptr(2), Label: "l", Environment: "prod", Type: TypeSecret})
			},
			Method: http.MethodGet, Path: "/v1/convai/environment-variables", Query: "environment=prod&label=l&page_size=2&type=secret",
			RespBody: `{"environment_variables":[{"id":"e","label":"l","type":"secret","workspace_id":"w","created_at_unix_secs":1,"updated_at_unix_secs":2,"values":{"prod":{"secret_id":"s"},"dev":"plain"}}],"has_more":false}`,
			Check: func(t *testing.T, r any) {
				v := r.(*Page).EnvironmentVariables[0]
				if v.Values["prod"].SecretID != "s" || v.Values["dev"].String != "plain" {
					t.Errorf("var = %+v", v)
				}
			},
		},
		{
			Name: "environmentVariables.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Create(ctx, &CreateRequest{Type: TypeAuthConnection, Label: "l", Values: map[string]Value{"prod": AuthConnectionValue("a")}})
			},
			Method: http.MethodPost, Path: "/v1/convai/environment-variables",
			Body: `{"type":"auth_connection","label":"l","values":{"prod":{"auth_connection_id":"a"}}}`, RespBody: `{"id":"e"}`,
		},
		{
			Name:   "environmentVariables.get",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Get(ctx, "e") },
			Method: http.MethodGet, Path: "/v1/convai/environment-variables/e", RespBody: `{"id":"e"}`,
		},
		{
			Name: "environmentVariables.delete",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.VoidCall(c.Delete(ctx, "e"))
			},
			Method: http.MethodDelete, Path: "/v1/convai/environment-variables/e",
		},
		{
			Name: "environmentVariables.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				v := StringValue("x")
				return c.Update(ctx, "e", &UpdateRequest{Values: map[string]*Value{"prod": &v, "dev": nil}})
			},
			Method: http.MethodPatch, Path: "/v1/convai/environment-variables/e",
			Body: `{"values":{"prod":"x","dev":null}}`, RespBody: `{"id":"e"}`,
		},
	})
}
