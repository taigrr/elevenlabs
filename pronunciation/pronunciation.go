package pronunciation

import (
	"context"
	"io"
	"iter"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/stream"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

// Service groups the /v1/pronunciation-dictionaries endpoints.
type Service struct {
	cfg *requestconfig.Config

	// Rules groups rule editing endpoints.
	Rules *RulesService
}

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	cfg := requestconfig.Build(opts...)
	return &Service{cfg: cfg, Rules: &RulesService{cfg: cfg}}
}

// RuleType distinguishes alias and phoneme rules.
type RuleType string

// RuleType values.
const (
	RuleTypeAlias   RuleType = "alias"
	RuleTypePhoneme RuleType = "phoneme"
)

// Rule is a pronunciation dictionary rule. Alias rules set
// Alias; phoneme rules set Phoneme and Alphabet. Use AliasRule and
// PhonemeRule to construct them.
type Rule struct {
	Type            RuleType `json:"type"`
	StringToReplace string   `json:"string_to_replace"`
	CaseSensitive   *bool    `json:"case_sensitive,omitempty"`
	WordBoundaries  *bool    `json:"word_boundaries,omitempty"`
	Alias           string   `json:"alias,omitempty"`
	Phoneme         string   `json:"phoneme,omitempty"`
	Alphabet        string   `json:"alphabet,omitempty"`
}

// AliasRule returns a rule replacing text with an alias.
func AliasRule(stringToReplace, alias string) Rule {
	return Rule{Type: RuleTypeAlias, StringToReplace: stringToReplace, Alias: alias}
}

// PhonemeRule returns a rule pronouncing text with a phoneme in the given
// alphabet (e.g. "ipa" or "cmu-arpabet").
func PhonemeRule(stringToReplace, phoneme, alphabet string) Rule {
	return Rule{Type: RuleTypePhoneme, StringToReplace: stringToReplace, Phoneme: phoneme, Alphabet: alphabet}
}

// CreateFromFileRequest is the body of Service.CreateFromFile.
type CreateFromFileRequest struct {
	// Name identifies the dictionary. Required.
	Name string `form:"name"`
	// File is a lexicon .pls file.
	File *types.File `form:"file"`
	// Description describes the dictionary.
	Description string `form:"description"`
	// WorkspaceAccess is the access granted to the workspace.
	WorkspaceAccess types.ResourcePermission `form:"workspace_access"`
}

// CreateFromRulesRequest is the body of Service.CreateFromRules.
type CreateFromRulesRequest struct {
	// Rules are the dictionary's rules. Required.
	Rules []Rule `json:"rules"`
	// Name identifies the dictionary. Required.
	Name string `json:"name"`
	// Description describes the dictionary.
	Description string `json:"description,omitempty"`
	// WorkspaceAccess is the access granted to the workspace.
	WorkspaceAccess types.ResourcePermission `json:"workspace_access,omitempty"`
}

// AddResponse is the response of the create endpoints.
type AddResponse struct {
	ID                   string                    `json:"id"`
	Name                 string                    `json:"name"`
	CreatedBy            string                    `json:"created_by"`
	CreationTimeUnix     int64                     `json:"creation_time_unix"`
	VersionID            string                    `json:"version_id"`
	VersionRulesNum      int64                     `json:"version_rules_num"`
	Description          *string                   `json:"description,omitempty"`
	PermissionOnResource *types.ResourcePermission `json:"permission_on_resource,omitempty"`
}

// Metadata describes a pronunciation dictionary.
type Metadata struct {
	ID                    string                    `json:"id"`
	LatestVersionID       string                    `json:"latest_version_id"`
	LatestVersionRulesNum int64                     `json:"latest_version_rules_num"`
	Name                  string                    `json:"name"`
	PermissionOnResource  *types.ResourcePermission `json:"permission_on_resource,omitempty"`
	CreatedBy             string                    `json:"created_by"`
	CreationTimeUnix      int64                     `json:"creation_time_unix"`
	ArchivedTimeUnix      *int64                    `json:"archived_time_unix,omitempty"`
	Description           *string                   `json:"description,omitempty"`
}

// DictionaryWithRules is a dictionary with the rules of its latest version.
type DictionaryWithRules struct {
	Metadata
	Rules []Rule `json:"rules"`
}

// CreateFromFile creates a pronunciation dictionary from a lexicon file.
func (s *Service) CreateFromFile(ctx context.Context, req *CreateFromFileRequest, opts ...option.RequestOption) (*AddResponse, error) {
	return transport.DoValue[AddResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/pronunciation-dictionaries/add-from-file", Form: transport.OrEmpty(req), Opts: opts})
}

// CreateFromRules creates a pronunciation dictionary from rules.
func (s *Service) CreateFromRules(ctx context.Context, req *CreateFromRulesRequest, opts ...option.RequestOption) (*AddResponse, error) {
	body := *transport.OrEmpty(req)
	body.Rules = transport.NonNil(body.Rules)
	return transport.DoValue[AddResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: "/v1/pronunciation-dictionaries/add-from-rules", Body: &body, Opts: opts})
}

// Get returns a pronunciation dictionary and its latest rules.
func (s *Service) Get(ctx context.Context, dictionaryID string, opts ...option.RequestOption) (*DictionaryWithRules, error) {
	return transport.DoValue[DictionaryWithRules](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/pronunciation-dictionaries/%s", dictionaryID), Opts: opts})
}

// UpdateRequest is the body of Service.Update.
type UpdateRequest struct {
	// Archived archives or unarchives the dictionary.
	Archived *bool `json:"archived,omitempty"`
	// Name renames the dictionary.
	Name string `json:"name,omitempty"`
}

// Update edits a pronunciation dictionary's name or archived state.
func (s *Service) Update(ctx context.Context, dictionaryID string, req *UpdateRequest, opts ...option.RequestOption) (*Metadata, error) {
	return transport.DoValue[Metadata](ctx, s.cfg, &transport.Request{Method: http.MethodPatch, Path: transport.Pathf("/v1/pronunciation-dictionaries/%s", dictionaryID), Body: transport.OrEmpty(req), Opts: opts})
}

// Download streams the PLS file of a dictionary version. The caller must close the returned reader.
func (s *Service) Download(ctx context.Context, dictionaryID, versionID string, opts ...option.RequestOption) (io.ReadCloser, error) {
	return transport.DoStream(ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: transport.Pathf("/v1/pronunciation-dictionaries/%s/%s/download", dictionaryID, versionID), Accept: transport.AcceptAny, Opts: opts})
}

// Sort is the field pronunciation dictionaries are sorted by.
type Sort string

// Sort values.
const (
	SortCreationTimeUnix Sort = "creation_time_unix"
	SortName             Sort = "name"
)

// ListRequest holds the optional parameters of Service.List.
type ListRequest struct {
	// Cursor continues a previous listing.
	Cursor string `query:"cursor"`
	// PageSize is the number of dictionaries per page (max 100, default 30).
	PageSize *int `query:"page_size"`
	// Sort is the sort field.
	Sort Sort `query:"sort"`
	// SortDirection is "ascending" or "descending".
	SortDirection string `query:"sort_direction"`
	// IncludeArchived includes archived dictionaries.
	IncludeArchived *bool `query:"include_archived"`
}

// ListResponse is a page of pronunciation dictionaries.
type ListResponse struct {
	PronunciationDictionaries []Metadata `json:"pronunciation_dictionaries"`
	NextCursor                *string    `json:"next_cursor,omitempty"`
	HasMore                   bool       `json:"has_more"`
}

// List returns one page of pronunciation dictionaries. req may be nil.
func (s *Service) List(ctx context.Context, req *ListRequest, opts ...option.RequestOption) (*ListResponse, error) {
	return transport.DoValue[ListResponse](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/pronunciation-dictionaries", Query: req, Opts: opts})
}

// ListAll iterates over every pronunciation dictionary, fetching pages on demand.
func (s *Service) ListAll(ctx context.Context, req *ListRequest, opts ...option.RequestOption) iter.Seq2[Metadata, error] {
	var params ListRequest
	if req != nil {
		params = *req
	}
	return stream.Paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (stream.Page[Metadata], error) {
		params.Cursor = cursor
		resp, err := s.List(ctx, &params, opts...)
		if err != nil {
			return stream.Page[Metadata]{}, err
		}
		next := ""
		if resp.NextCursor != nil {
			next = *resp.NextCursor
		}
		return stream.Page[Metadata]{Items: resp.PronunciationDictionaries, Next: next, HasMore: resp.HasMore}, nil
	})
}

// RulesService groups rule editing endpoints.
type RulesService struct{ cfg *requestconfig.Config }

// RulesResponse is the dictionary version produced by a rules edit.
type RulesResponse struct {
	ID              string `json:"id"`
	VersionID       string `json:"version_id"`
	VersionRulesNum int64  `json:"version_rules_num"`
}

type pronunciationRulesBody struct {
	Rules []Rule `json:"rules"`
}

type removeRulesBody struct {
	RuleStrings []string `json:"rule_strings"`
}

// Set replaces all rules of a dictionary, creating a new version.
func (s *RulesService) Set(ctx context.Context, dictionaryID string, rules []Rule, opts ...option.RequestOption) (*RulesResponse, error) {
	return transport.DoValue[RulesResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/pronunciation-dictionaries/%s/set-rules", dictionaryID), Body: pronunciationRulesBody{Rules: transport.NonNil(rules)}, Opts: opts})
}

// Add adds rules to a dictionary, creating a new version.
func (s *RulesService) Add(ctx context.Context, dictionaryID string, rules []Rule, opts ...option.RequestOption) (*RulesResponse, error) {
	return transport.DoValue[RulesResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/pronunciation-dictionaries/%s/add-rules", dictionaryID), Body: pronunciationRulesBody{Rules: transport.NonNil(rules)}, Opts: opts})
}

// Remove removes the rules matching ruleStrings, creating a new version.
func (s *RulesService) Remove(ctx context.Context, dictionaryID string, ruleStrings []string, opts ...option.RequestOption) (*RulesResponse, error) {
	return transport.DoValue[RulesResponse](ctx, s.cfg, &transport.Request{Method: http.MethodPost, Path: transport.Pathf("/v1/pronunciation-dictionaries/%s/remove-rules", dictionaryID), Body: removeRulesBody{RuleStrings: transport.NonNil(ruleStrings)}, Opts: opts})
}
