package elevenlabs

import (
	"context"
	"encoding/json"
	"net/http"
)

// WorkspaceAuthConnectionsService groups the /v1/workspace/auth-connections endpoints.
type WorkspaceAuthConnectionsService struct{ cfg *config }

// AuthType discriminates auth connection variants.
type AuthType string

// AuthType values.
const (
	AuthTypeOAuth2ClientCredentials AuthType = "oauth2_client_credentials"
	AuthTypeCustomHeader            AuthType = "custom_header_auth"
	AuthTypeBasic                   AuthType = "basic_auth"
	AuthTypeBearer                  AuthType = "bearer_auth"
	AuthTypeOAuth2JWT               AuthType = "oauth2_jwt"
	AuthTypePrivateKeyJWT           AuthType = "private_key_jwt"
	AuthTypeMTLS                    AuthType = "mtls"
	AuthTypeRefreshToken            AuthType = "refresh_token_auth"
	AuthTypeOAuth2AuthCode          AuthType = "api_integration_oauth2_auth_code"
	AuthTypeOAuth2CustomApp         AuthType = "api_integration_oauth2_custom_app"
	AuthTypeWhatsApp                AuthType = "whatsapp_auth"
	AuthTypeSlackBot                AuthType = "slack_bot_auth"
	AuthTypeURLSecret               AuthType = "url_secret"
)

// AuthConnection is any auth connection returned by the API. Fields that do
// not apply to AuthType are empty; Raw holds the full object.
type AuthConnection struct {
	AuthType          AuthType          `json:"auth_type"`
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	Provider          string            `json:"provider"`
	Status            string            `json:"status,omitempty"`
	StatusDetail      *string           `json:"status_detail,omitempty"`
	StatusUpdatedAt   *string           `json:"status_updated_at,omitempty"`
	UsedBy            json.RawMessage   `json:"used_by,omitempty"`
	Username          string            `json:"username,omitempty"`
	HeaderName        string            `json:"header_name,omitempty"`
	ClientID          string            `json:"client_id,omitempty"`
	TokenURL          string            `json:"token_url,omitempty"`
	Scopes            []string          `json:"scopes,omitempty"`
	ExtraParams       map[string]string `json:"extra_params,omitempty"`
	CustomHeaders     map[string]string `json:"custom_headers,omitempty"`
	BasicAuthInHeader *bool             `json:"basic_auth_in_header,omitempty"`
	Algorithm         string            `json:"algorithm,omitempty"`
	KeyID             *string           `json:"key_id,omitempty"`
	Issuer            string            `json:"issuer,omitempty"`
	Audience          string            `json:"audience,omitempty"`
	Subject           string            `json:"subject,omitempty"`
	ExpirationSeconds *int              `json:"expiration_seconds,omitempty"`
	ExpiresAt         string            `json:"expires_at,omitempty"`
	IntegrationID     string            `json:"integration_id,omitempty"`
	CredentialID      string            `json:"credential_id,omitempty"`
	PhoneNumberID     string            `json:"phone_number_id,omitempty"`

	Raw json.RawMessage `json:"-"`
}

// UnmarshalJSON implements json.Unmarshaler, keeping the raw object.
func (a *AuthConnection) UnmarshalJSON(b []byte) error {
	type plain AuthConnection
	var p plain
	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}
	*a = AuthConnection(p)
	a.Raw = append(json.RawMessage(nil), b...)
	return nil
}

// AuthConnectionList is the response of WorkspaceAuthConnectionsService.List.
type AuthConnectionList struct {
	AuthConnections []AuthConnection `json:"auth_connections"`
}

// List returns the workspace auth connections.
func (s *WorkspaceAuthConnectionsService) List(ctx context.Context, opts ...Option) (*AuthConnectionList, error) {
	return doValue[AuthConnectionList](ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/workspace/auth-connections", opts: opts})
}

// AuthConnectionCreateRequest is a create body for one auth connection
// variant: *BearerAuthCreate, *BasicAuthCreate, *CustomHeaderAuthCreate,
// *OAuth2ClientCredentialsCreate, *OAuth2JWTCreate, *PrivateKeyJWTCreate or
// *MTLSAuthCreate.
type AuthConnectionCreateRequest interface {
	authConnectionAuthType() AuthType
}

// AuthConnectionUpdateRequest is an update body for one auth connection
// variant: *BearerAuthUpdate, *BasicAuthUpdate, *OAuth2ClientCredentialsUpdate
// or *OAuth2JWTUpdate.
type AuthConnectionUpdateRequest interface {
	authConnectionUpdateAuthType() AuthType
}

// withAuthType encodes v with an added auth_type discriminator.
func withAuthType(authType AuthType, v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	m := map[string]json.RawMessage{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	m["auth_type"], _ = json.Marshal(authType)
	return json.Marshal(m)
}

// BearerAuthCreate creates a bearer token auth connection.
type BearerAuthCreate struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Token    string `json:"token"`
}

// BasicAuthCreate creates an HTTP basic auth connection.
type BasicAuthCreate struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// CustomHeaderAuthCreate creates a custom header auth connection.
type CustomHeaderAuthCreate struct {
	Name       string `json:"name"`
	Provider   string `json:"provider"`
	HeaderName string `json:"header_name"`
	Token      string `json:"token"`
}

// OAuth2ClientCredentialsCreate creates an OAuth2 client credentials connection.
type OAuth2ClientCredentialsCreate struct {
	Name              string            `json:"name"`
	Provider          string            `json:"provider"`
	ClientID          string            `json:"client_id"`
	ClientSecret      string            `json:"client_secret"`
	TokenURL          string            `json:"token_url"`
	Scopes            []string          `json:"scopes,omitempty"`
	ExtraParams       map[string]string `json:"extra_params,omitempty"`
	BasicAuthInHeader *bool             `json:"basic_auth_in_header,omitempty"`
	CustomHeaders     map[string]string `json:"custom_headers,omitempty"`
}

// JWTClaims are the claims shared by JWT-based auth connections.
type JWTClaims struct {
	// Algorithm is one of HS256, HS384, HS512, RS256, RS384 or RS512.
	Algorithm         string            `json:"algorithm,omitempty"`
	KeyID             string            `json:"key_id,omitempty"`
	Issuer            string            `json:"issuer"`
	Audience          string            `json:"audience"`
	Subject           string            `json:"subject"`
	ExpirationSeconds *int              `json:"expiration_seconds,omitempty"`
	ExtraParams       map[string]string `json:"extra_params,omitempty"`
}

// OAuth2JWTCreate creates an OAuth2 JWT bearer grant connection.
type OAuth2JWTCreate struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	JWTClaims
	TokenURL string   `json:"token_url"`
	Scopes   []string `json:"scopes,omitempty"`
	// TokenResponseField is "access_token" or "id_token".
	TokenResponseField string `json:"token_response_field,omitempty"`
	SecretKey          string `json:"secret_key"`
}

// PrivateKeyJWTCreate creates a private key JWT connection.
type PrivateKeyJWTCreate struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	JWTClaims
	SecretKey string `json:"secret_key"`
}

// MTLSAuthCreate creates a mutual TLS connection.
type MTLSAuthCreate struct {
	Name              string `json:"name"`
	Provider          string `json:"provider"`
	ClientCertificate string `json:"client_certificate"`
	ClientKey         string `json:"client_key"`
	CACertificate     string `json:"ca_certificate,omitempty"`
	KeyPassphrase     string `json:"key_passphrase,omitempty"`
}

func (*BearerAuthCreate) authConnectionAuthType() AuthType       { return AuthTypeBearer }
func (*BasicAuthCreate) authConnectionAuthType() AuthType        { return AuthTypeBasic }
func (*CustomHeaderAuthCreate) authConnectionAuthType() AuthType { return AuthTypeCustomHeader }
func (*OAuth2ClientCredentialsCreate) authConnectionAuthType() AuthType {
	return AuthTypeOAuth2ClientCredentials
}
func (*OAuth2JWTCreate) authConnectionAuthType() AuthType     { return AuthTypeOAuth2JWT }
func (*PrivateKeyJWTCreate) authConnectionAuthType() AuthType { return AuthTypePrivateKeyJWT }
func (*MTLSAuthCreate) authConnectionAuthType() AuthType      { return AuthTypeMTLS }

// BearerAuthUpdate edits a bearer token connection.
type BearerAuthUpdate struct {
	Provider string `json:"provider,omitempty"`
	Token    string `json:"token,omitempty"`
}

// BasicAuthUpdate edits a basic auth connection.
type BasicAuthUpdate struct {
	Provider string `json:"provider,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// OAuth2ClientCredentialsUpdate edits an OAuth2 client credentials connection.
type OAuth2ClientCredentialsUpdate struct {
	Provider          string             `json:"provider,omitempty"`
	ClientID          string             `json:"client_id,omitempty"`
	ClientSecret      string             `json:"client_secret,omitempty"`
	Scopes            []string           `json:"scopes,omitempty"`
	ExtraParams       map[string]*string `json:"extra_params,omitempty"`
	BasicAuthInHeader *bool              `json:"basic_auth_in_header,omitempty"`
	CustomHeaders     map[string]*string `json:"custom_headers,omitempty"`
}

// OAuth2JWTUpdate edits an OAuth2 JWT connection.
type OAuth2JWTUpdate struct {
	Provider           string             `json:"provider,omitempty"`
	Algorithm          string             `json:"algorithm,omitempty"`
	KeyID              string             `json:"key_id,omitempty"`
	Issuer             string             `json:"issuer,omitempty"`
	Audience           string             `json:"audience,omitempty"`
	Subject            string             `json:"subject,omitempty"`
	ExpirationSeconds  *int               `json:"expiration_seconds,omitempty"`
	ExtraParams        map[string]*string `json:"extra_params,omitempty"`
	Scopes             []string           `json:"scopes,omitempty"`
	TokenResponseField string             `json:"token_response_field,omitempty"`
	SecretKey          string             `json:"secret_key,omitempty"`
}

func (*BearerAuthUpdate) authConnectionUpdateAuthType() AuthType { return AuthTypeBearer }
func (*BasicAuthUpdate) authConnectionUpdateAuthType() AuthType  { return AuthTypeBasic }
func (*OAuth2ClientCredentialsUpdate) authConnectionUpdateAuthType() AuthType {
	return AuthTypeOAuth2ClientCredentials
}
func (*OAuth2JWTUpdate) authConnectionUpdateAuthType() AuthType { return AuthTypeOAuth2JWT }

type authBody struct {
	authType AuthType
	v        any
}

func (a authBody) MarshalJSON() ([]byte, error) { return withAuthType(a.authType, a.v) }

// Create creates an auth connection.
func (s *WorkspaceAuthConnectionsService) Create(ctx context.Context, req AuthConnectionCreateRequest, opts ...Option) (*AuthConnection, error) {
	var body any = struct{}{}
	if req != nil {
		body = authBody{req.authConnectionAuthType(), req}
	}
	return doValue[AuthConnection](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/workspace/auth-connections", body: body, opts: opts})
}

// Update edits an auth connection.
func (s *WorkspaceAuthConnectionsService) Update(ctx context.Context, authConnectionID string, req AuthConnectionUpdateRequest, opts ...Option) (*AuthConnection, error) {
	var body any = struct{}{}
	if req != nil {
		body = authBody{req.authConnectionUpdateAuthType(), req}
	}
	return doValue[AuthConnection](ctx, s.cfg, &request{method: http.MethodPatch, path: pathf("/v1/workspace/auth-connections/%s", authConnectionID), body: body, opts: opts})
}

// Delete deletes an auth connection.
func (s *WorkspaceAuthConnectionsService) Delete(ctx context.Context, authConnectionID string, opts ...Option) (json.RawMessage, error) {
	return doRaw(ctx, s.cfg, &request{method: http.MethodDelete, path: pathf("/v1/workspace/auth-connections/%s", authConnectionID), opts: opts})
}
