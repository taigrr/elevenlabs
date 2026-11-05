package envvars

import (
	"context"
	"encoding/json"
	"iter"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/stream"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
)

// Service groups the /v1/convai/environment-variables endpoints.
type Service struct{ cfg *requestconfig.Config }

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	return &Service{cfg: requestconfig.Build(opts...)}
}

// Type is the kind of value an environment variable holds.
type Type string

// Type values.
const (
	TypeString         Type = "string"
	TypeSecret         Type = "secret"
	TypeAuthConnection Type = "auth_connection"
)

// Value is the value of a variable in one environment:
// a plain string, a reference to a workspace secret, or a reference to an
// auth connection. Build one with StringValue, SecretValue or
// AuthConnectionValue.
type Value struct {
	String           string
	SecretID         string
	AuthConnectionID string
}

// StringValue is a plain string environment variable value.
func StringValue(s string) Value { return Value{String: s} }

// SecretValue references a workspace secret.
func SecretValue(secretID string) Value {
	return Value{SecretID: secretID}
}

// AuthConnectionValue references a workspace auth connection.
func AuthConnectionValue(authConnectionID string) Value {
	return Value{AuthConnectionID: authConnectionID}
}

// MarshalJSON implements json.Marshaler.
func (v Value) MarshalJSON() ([]byte, error) {
	switch {
	case v.SecretID != "":
		return json.Marshal(map[string]string{"secret_id": v.SecretID})
	case v.AuthConnectionID != "":
		return json.Marshal(map[string]string{"auth_connection_id": v.AuthConnectionID})
	}
	return json.Marshal(v.String)
}

// UnmarshalJSON implements json.Unmarshaler.
func (v *Value) UnmarshalJSON(b []byte) error {
	*v = Value{}
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

// Variable is a conversational AI environment variable.
type Variable struct {
	ID                string           `json:"id"`
	Label             string           `json:"label"`
	Type              Type             `json:"type"`
	WorkspaceID       string           `json:"workspace_id"`
	CreatedAtUnixSecs int64            `json:"created_at_unix_secs"`
	UpdatedAtUnixSecs int64            `json:"updated_at_unix_secs"`
	CreatedByUserID   *string          `json:"created_by_user_id,omitempty"`
	Values            map[string]Value `json:"values"`
}

// ListRequest holds the optional parameters of Service.List.
type ListRequest struct {
	Cursor   string `query:"cursor"`
	PageSize *int   `query:"page_size"`
	// Label filters by exact label.
	Label string `query:"label"`
	// Environment limits values to one environment.
	Environment string `query:"environment"`
	Type        Type   `query:"type"`
}

// Page is a page of environment variables.
type Page struct {
	EnvironmentVariables []Variable `json:"environment_variables"`
	NextCursor           *string    `json:"next_cursor,omitempty"`
	HasMore              bool       `json:"has_more"`
}

// List returns one page of environment variables. req may be nil.
func (s *Service) List(ctx context.Context, req *ListRequest, opts ...option.RequestOption) (*Page, error) {
	return transport.DoValue[Page](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/convai/environment-variables", Query: req, Opts: opts})
}

// ListAll iterates over every environment variable, fetching pages on demand.
func (s *Service) ListAll(ctx context.Context, req *ListRequest, opts ...option.RequestOption) iter.Seq2[Variable, error] {
	params := *transport.OrEmpty(req)
	return stream.Paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (stream.Page[Variable], error) {
		params.Cursor = cursor
		resp, err := s.List(ctx, &params, opts...)
		if err != nil {
			return stream.Page[Variable]{}, err
		}
		return stream.Page[Variable]{Items: resp.EnvironmentVariables, Next: transport.Deref(resp.NextCursor), HasMore: resp.HasMore}, nil
	})
}

// CreateRequest is the body of Service.Create.
type CreateRequest struct {
	// Type selects the value kind. Required.
	Type Type `json:"type"`
	// Label names the variable. Required.
	Label string `json:"label"`
	// Values maps environment names to values; they must match Type.
	Values map[string]Value `json:"values"`
}

// Create creates an environment variable.
func (s *Service) Create(ctx context.Context, req *CreateRequest, opts ...option.RequestOption) (*Variable, error) {
	body := *transport.OrEmpty(req)
	if body.Values == nil {
		body.Values = map[string]Value{}
	}
	return transport.DoValue[Variable](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/convai/environment-variables", Body: &body, Opts: opts})
}

// Get returns an environment variable.
func (s *Service) Get(ctx context.Context, envVarID string, opts ...option.RequestOption) (*Variable, error) {
	return transport.DoValue[Variable](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/convai/environment-variables/%s", envVarID), Opts: opts})
}

// Delete deletes an environment variable.
func (s *Service) Delete(ctx context.Context, envVarID string, opts ...option.RequestOption) error {
	return transport.DoJSON(ctx, s.cfg, &transport.Request{Method: http.MethodDelete, Path: transport.Pathf("/v1/convai/environment-variables/%s", envVarID), Opts: opts}, nil)
}

// UpdateRequest is the body of Service.Update.
type UpdateRequest struct {
	// Values maps environment names to new values; a nil value removes the
	// variable from that environment.
	Values map[string]*Value `json:"values"`
}

// Update edits the values of an environment variable.
func (s *Service) Update(ctx context.Context, envVarID string, req *UpdateRequest, opts ...option.RequestOption) (*Variable, error) {
	body := *transport.OrEmpty(req)
	if body.Values == nil {
		body.Values = map[string]*Value{}
	}
	return transport.DoValue[Variable](ctx, s.cfg, &transport.Request{Method: http.MethodPatch, Path: transport.Pathf("/v1/convai/environment-variables/%s", envVarID), Body: &body, Opts: opts})
}
