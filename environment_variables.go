package elevenlabs

import (
	"context"
	"encoding/json"
	"iter"
	"net/http"
)

// EnvironmentVariablesService groups the /v1/convai/environment-variables endpoints.
type EnvironmentVariablesService struct{ cfg *config }

// EnvironmentVariableType is the kind of value an environment variable holds.
type EnvironmentVariableType string

// EnvironmentVariableType values.
const (
	EnvironmentVariableTypeString         EnvironmentVariableType = "string"
	EnvironmentVariableTypeSecret         EnvironmentVariableType = "secret"
	EnvironmentVariableTypeAuthConnection EnvironmentVariableType = "auth_connection"
)

// EnvironmentVariableValue is the value of a variable in one environment:
// a plain string, a reference to a workspace secret, or a reference to an
// auth connection. Build one with StringValue, SecretValue or
// AuthConnectionValue.
type EnvironmentVariableValue struct {
	String           string
	SecretID         string
	AuthConnectionID string
}

// StringValue is a plain string environment variable value.
func StringValue(s string) EnvironmentVariableValue { return EnvironmentVariableValue{String: s} }

// SecretValue references a workspace secret.
func SecretValue(secretID string) EnvironmentVariableValue {
	return EnvironmentVariableValue{SecretID: secretID}
}

// AuthConnectionValue references a workspace auth connection.
func AuthConnectionValue(authConnectionID string) EnvironmentVariableValue {
	return EnvironmentVariableValue{AuthConnectionID: authConnectionID}
}

// MarshalJSON implements json.Marshaler.
func (v EnvironmentVariableValue) MarshalJSON() ([]byte, error) {
	switch {
	case v.SecretID != "":
		return json.Marshal(map[string]string{"secret_id": v.SecretID})
	case v.AuthConnectionID != "":
		return json.Marshal(map[string]string{"auth_connection_id": v.AuthConnectionID})
	}
	return json.Marshal(v.String)
}

// UnmarshalJSON implements json.Unmarshaler.
func (v *EnvironmentVariableValue) UnmarshalJSON(b []byte) error {
	*v = EnvironmentVariableValue{}
	if json.Unmarshal(b, &v.String) == nil {
		return nil
	}
	var ref struct {
		SecretID         string `json:"secret_id"`
		AuthConnectionID string `json:"auth_connection_id"`
	}
	if err := json.Unmarshal(b, &ref); err != nil {
		return err
	}
	v.SecretID, v.AuthConnectionID = ref.SecretID, ref.AuthConnectionID
	return nil
}

// EnvironmentVariable is a conversational AI environment variable.
type EnvironmentVariable struct {
	ID                string                              `json:"id"`
	Label             string                              `json:"label"`
	Type              EnvironmentVariableType             `json:"type"`
	WorkspaceID       string                              `json:"workspace_id"`
	CreatedAtUnixSecs int64                               `json:"created_at_unix_secs"`
	UpdatedAtUnixSecs int64                               `json:"updated_at_unix_secs"`
	CreatedByUserID   *string                             `json:"created_by_user_id,omitempty"`
	Values            map[string]EnvironmentVariableValue `json:"values"`
}

// EnvironmentVariablesListRequest holds the optional parameters of EnvironmentVariablesService.List.
type EnvironmentVariablesListRequest struct {
	Cursor   string `query:"cursor"`
	PageSize *int   `query:"page_size"`
	// Label filters by exact label.
	Label string `query:"label"`
	// Environment limits values to one environment.
	Environment string                  `query:"environment"`
	Type        EnvironmentVariableType `query:"type"`
}

// EnvironmentVariablesPage is a page of environment variables.
type EnvironmentVariablesPage struct {
	EnvironmentVariables []EnvironmentVariable `json:"environment_variables"`
	NextCursor           *string               `json:"next_cursor,omitempty"`
	HasMore              bool                  `json:"has_more"`
}

// List returns one page of environment variables. req may be nil.
func (s *EnvironmentVariablesService) List(ctx context.Context, req *EnvironmentVariablesListRequest, opts ...RequestOption) (*EnvironmentVariablesPage, error) {
	return doValue[EnvironmentVariablesPage](ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/convai/environment-variables", query: req, opts: opts})
}

// ListAll iterates over every environment variable, fetching pages on demand.
func (s *EnvironmentVariablesService) ListAll(ctx context.Context, req *EnvironmentVariablesListRequest, opts ...RequestOption) iter.Seq2[EnvironmentVariable, error] {
	params := *orEmpty(req)
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[EnvironmentVariable], error) {
		params.Cursor = cursor
		resp, err := s.List(ctx, &params, opts...)
		if err != nil {
			return page[EnvironmentVariable]{}, err
		}
		return page[EnvironmentVariable]{items: resp.EnvironmentVariables, next: deref(resp.NextCursor), hasMore: resp.HasMore}, nil
	})
}

// EnvironmentVariableCreateRequest is the body of EnvironmentVariablesService.Create.
type EnvironmentVariableCreateRequest struct {
	// Type selects the value kind. Required.
	Type EnvironmentVariableType `json:"type"`
	// Label names the variable. Required.
	Label string `json:"label"`
	// Values maps environment names to values; they must match Type.
	Values map[string]EnvironmentVariableValue `json:"values"`
}

// Create creates an environment variable.
func (s *EnvironmentVariablesService) Create(ctx context.Context, req *EnvironmentVariableCreateRequest, opts ...RequestOption) (*EnvironmentVariable, error) {
	body := *orEmpty(req)
	if body.Values == nil {
		body.Values = map[string]EnvironmentVariableValue{}
	}
	return doValue[EnvironmentVariable](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/convai/environment-variables", body: &body, opts: opts})
}

// Get returns an environment variable.
func (s *EnvironmentVariablesService) Get(ctx context.Context, envVarID string, opts ...RequestOption) (*EnvironmentVariable, error) {
	return doValue[EnvironmentVariable](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/convai/environment-variables/%s", envVarID), opts: opts})
}

// Delete deletes an environment variable.
func (s *EnvironmentVariablesService) Delete(ctx context.Context, envVarID string, opts ...RequestOption) error {
	return doJSON(ctx, s.cfg, &request{method: http.MethodDelete, path: pathf("/v1/convai/environment-variables/%s", envVarID), opts: opts}, nil)
}

// EnvironmentVariableUpdateRequest is the body of EnvironmentVariablesService.Update.
type EnvironmentVariableUpdateRequest struct {
	// Values maps environment names to new values; a nil value removes the
	// variable from that environment.
	Values map[string]*EnvironmentVariableValue `json:"values"`
}

// Update edits the values of an environment variable.
func (s *EnvironmentVariablesService) Update(ctx context.Context, envVarID string, req *EnvironmentVariableUpdateRequest, opts ...RequestOption) (*EnvironmentVariable, error) {
	body := *orEmpty(req)
	if body.Values == nil {
		body.Values = map[string]*EnvironmentVariableValue{}
	}
	return doValue[EnvironmentVariable](ctx, s.cfg, &request{method: http.MethodPatch, path: pathf("/v1/convai/environment-variables/%s", envVarID), body: &body, opts: opts})
}

// EnvironmentsService groups the /v1/convai/environments endpoints.
type EnvironmentsService struct{ cfg *config }

// AgentEnvironment is a named conversational AI deployment environment.
type AgentEnvironment struct {
	Name              string  `json:"name"`
	UpdatedAtUnixSecs *int64  `json:"updated_at_unix_secs,omitempty"`
	CreatedByUserID   *string `json:"created_by_user_id,omitempty"`
}

// AgentEnvironmentList is the response of EnvironmentsService.List.
type AgentEnvironmentList struct {
	Environments []AgentEnvironment `json:"environments"`
}

// List returns the workspace environments.
func (s *EnvironmentsService) List(ctx context.Context, opts ...RequestOption) (*AgentEnvironmentList, error) {
	return doValue[AgentEnvironmentList](ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/convai/environments", opts: opts})
}

type nameBody struct {
	Name string `json:"name"`
}

// Create creates an environment named name.
func (s *EnvironmentsService) Create(ctx context.Context, name string, opts ...RequestOption) (*AgentEnvironment, error) {
	return doValue[AgentEnvironment](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/convai/environments", body: nameBody{name}, opts: opts})
}

// Delete deletes an environment.
func (s *EnvironmentsService) Delete(ctx context.Context, environment string, opts ...RequestOption) error {
	return doJSON(ctx, s.cfg, &request{method: http.MethodDelete, path: pathf("/v1/convai/environments/%s", environment), opts: opts}, nil)
}
