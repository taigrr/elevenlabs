package elevenlabs

import (
	"context"
	"io"
	"iter"
	"net/http"
)

// PronunciationDictionariesService groups the /v1/pronunciation-dictionaries endpoints.
type PronunciationDictionariesService struct {
	cfg *config

	// Rules groups rule editing endpoints.
	Rules *PronunciationDictionaryRulesService
}

func newPronunciationDictionariesService(cfg *config) *PronunciationDictionariesService {
	return &PronunciationDictionariesService{cfg: cfg, Rules: &PronunciationDictionaryRulesService{cfg: cfg}}
}

// PronunciationRuleType distinguishes alias and phoneme rules.
type PronunciationRuleType string

// PronunciationRuleType values.
const (
	PronunciationRuleTypeAlias   PronunciationRuleType = "alias"
	PronunciationRuleTypePhoneme PronunciationRuleType = "phoneme"
)

// PronunciationRule is a pronunciation dictionary rule. Alias rules set
// Alias; phoneme rules set Phoneme and Alphabet. Use AliasRule and
// PhonemeRule to construct them.
type PronunciationRule struct {
	Type            PronunciationRuleType `json:"type"`
	StringToReplace string                `json:"string_to_replace"`
	CaseSensitive   *bool                 `json:"case_sensitive,omitempty"`
	WordBoundaries  *bool                 `json:"word_boundaries,omitempty"`
	Alias           string                `json:"alias,omitempty"`
	Phoneme         string                `json:"phoneme,omitempty"`
	Alphabet        string                `json:"alphabet,omitempty"`
}

// AliasRule returns a rule replacing text with an alias.
func AliasRule(stringToReplace, alias string) PronunciationRule {
	return PronunciationRule{Type: PronunciationRuleTypeAlias, StringToReplace: stringToReplace, Alias: alias}
}

// PhonemeRule returns a rule pronouncing text with a phoneme in the given
// alphabet (e.g. "ipa" or "cmu-arpabet").
func PhonemeRule(stringToReplace, phoneme, alphabet string) PronunciationRule {
	return PronunciationRule{Type: PronunciationRuleTypePhoneme, StringToReplace: stringToReplace, Phoneme: phoneme, Alphabet: alphabet}
}

// PronunciationDictionaryCreateFromFileRequest is the body of PronunciationDictionariesService.CreateFromFile.
type PronunciationDictionaryCreateFromFileRequest struct {
	// Name identifies the dictionary. Required.
	Name string `form:"name"`
	// File is a lexicon .pls file.
	File *File `form:"file"`
	// Description describes the dictionary.
	Description string `form:"description"`
	// WorkspaceAccess is the access granted to the workspace.
	WorkspaceAccess ResourcePermission `form:"workspace_access"`
}

// PronunciationDictionaryCreateFromRulesRequest is the body of PronunciationDictionariesService.CreateFromRules.
type PronunciationDictionaryCreateFromRulesRequest struct {
	// Rules are the dictionary's rules. Required.
	Rules []PronunciationRule `json:"rules"`
	// Name identifies the dictionary. Required.
	Name string `json:"name"`
	// Description describes the dictionary.
	Description string `json:"description,omitempty"`
	// WorkspaceAccess is the access granted to the workspace.
	WorkspaceAccess ResourcePermission `json:"workspace_access,omitempty"`
}

// AddPronunciationDictionaryResponse is the response of the create endpoints.
type AddPronunciationDictionaryResponse struct {
	ID                   string              `json:"id"`
	Name                 string              `json:"name"`
	CreatedBy            string              `json:"created_by"`
	CreationTimeUnix     int64               `json:"creation_time_unix"`
	VersionID            string              `json:"version_id"`
	VersionRulesNum      int64               `json:"version_rules_num"`
	Description          *string             `json:"description,omitempty"`
	PermissionOnResource *ResourcePermission `json:"permission_on_resource,omitempty"`
}

// PronunciationDictionaryMetadata describes a pronunciation dictionary.
type PronunciationDictionaryMetadata struct {
	ID                    string              `json:"id"`
	LatestVersionID       string              `json:"latest_version_id"`
	LatestVersionRulesNum int64               `json:"latest_version_rules_num"`
	Name                  string              `json:"name"`
	PermissionOnResource  *ResourcePermission `json:"permission_on_resource,omitempty"`
	CreatedBy             string              `json:"created_by"`
	CreationTimeUnix      int64               `json:"creation_time_unix"`
	ArchivedTimeUnix      *int64              `json:"archived_time_unix,omitempty"`
	Description           *string             `json:"description,omitempty"`
}

// PronunciationDictionaryWithRules is a dictionary with the rules of its latest version.
type PronunciationDictionaryWithRules struct {
	PronunciationDictionaryMetadata
	Rules []PronunciationRule `json:"rules"`
}

// CreateFromFile creates a pronunciation dictionary from a lexicon file.
func (s *PronunciationDictionariesService) CreateFromFile(ctx context.Context, req *PronunciationDictionaryCreateFromFileRequest, opts ...RequestOption) (*AddPronunciationDictionaryResponse, error) {
	return doValue[AddPronunciationDictionaryResponse](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/pronunciation-dictionaries/add-from-file", form: orEmpty(req), opts: opts})
}

// CreateFromRules creates a pronunciation dictionary from rules.
func (s *PronunciationDictionariesService) CreateFromRules(ctx context.Context, req *PronunciationDictionaryCreateFromRulesRequest, opts ...RequestOption) (*AddPronunciationDictionaryResponse, error) {
	body := *orEmpty(req)
	body.Rules = nonNil(body.Rules)
	return doValue[AddPronunciationDictionaryResponse](ctx, s.cfg, &request{method: http.MethodPost, path: "/v1/pronunciation-dictionaries/add-from-rules", body: &body, opts: opts})
}

// Get returns a pronunciation dictionary and its latest rules.
func (s *PronunciationDictionariesService) Get(ctx context.Context, dictionaryID string, opts ...RequestOption) (*PronunciationDictionaryWithRules, error) {
	return doValue[PronunciationDictionaryWithRules](ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/pronunciation-dictionaries/%s", dictionaryID), opts: opts})
}

// PronunciationDictionaryUpdateRequest is the body of PronunciationDictionariesService.Update.
type PronunciationDictionaryUpdateRequest struct {
	// Archived archives or unarchives the dictionary.
	Archived *bool `json:"archived,omitempty"`
	// Name renames the dictionary.
	Name string `json:"name,omitempty"`
}

// Update edits a pronunciation dictionary's name or archived state.
func (s *PronunciationDictionariesService) Update(ctx context.Context, dictionaryID string, req *PronunciationDictionaryUpdateRequest, opts ...RequestOption) (*PronunciationDictionaryMetadata, error) {
	return doValue[PronunciationDictionaryMetadata](ctx, s.cfg, &request{method: http.MethodPatch, path: pathf("/v1/pronunciation-dictionaries/%s", dictionaryID), body: orEmpty(req), opts: opts})
}

// Download streams the PLS file of a dictionary version. The caller must close the returned reader.
func (s *PronunciationDictionariesService) Download(ctx context.Context, dictionaryID, versionID string, opts ...RequestOption) (io.ReadCloser, error) {
	return doStream(ctx, s.cfg, &request{method: http.MethodGet, path: pathf("/v1/pronunciation-dictionaries/%s/%s/download", dictionaryID, versionID), accept: acceptAny, opts: opts})
}

// PronunciationDictionarySort is the field pronunciation dictionaries are sorted by.
type PronunciationDictionarySort string

// PronunciationDictionarySort values.
const (
	PronunciationDictionarySortCreationTimeUnix PronunciationDictionarySort = "creation_time_unix"
	PronunciationDictionarySortName             PronunciationDictionarySort = "name"
)

// PronunciationDictionariesListRequest holds the optional parameters of PronunciationDictionariesService.List.
type PronunciationDictionariesListRequest struct {
	// Cursor continues a previous listing.
	Cursor string `query:"cursor"`
	// PageSize is the number of dictionaries per page (max 100, default 30).
	PageSize *int `query:"page_size"`
	// Sort is the sort field.
	Sort PronunciationDictionarySort `query:"sort"`
	// SortDirection is "ascending" or "descending".
	SortDirection string `query:"sort_direction"`
	// IncludeArchived includes archived dictionaries.
	IncludeArchived *bool `query:"include_archived"`
}

// GetPronunciationDictionariesResponse is a page of pronunciation dictionaries.
type GetPronunciationDictionariesResponse struct {
	PronunciationDictionaries []PronunciationDictionaryMetadata `json:"pronunciation_dictionaries"`
	NextCursor                *string                           `json:"next_cursor,omitempty"`
	HasMore                   bool                              `json:"has_more"`
}

// List returns one page of pronunciation dictionaries. req may be nil.
func (s *PronunciationDictionariesService) List(ctx context.Context, req *PronunciationDictionariesListRequest, opts ...RequestOption) (*GetPronunciationDictionariesResponse, error) {
	return doValue[GetPronunciationDictionariesResponse](ctx, s.cfg, &request{method: http.MethodGet, path: "/v1/pronunciation-dictionaries", query: req, opts: opts})
}

// ListAll iterates over every pronunciation dictionary, fetching pages on demand.
func (s *PronunciationDictionariesService) ListAll(ctx context.Context, req *PronunciationDictionariesListRequest, opts ...RequestOption) iter.Seq2[PronunciationDictionaryMetadata, error] {
	var params PronunciationDictionariesListRequest
	if req != nil {
		params = *req
	}
	return paginate(ctx, params.Cursor, func(ctx context.Context, cursor string) (page[PronunciationDictionaryMetadata], error) {
		params.Cursor = cursor
		resp, err := s.List(ctx, &params, opts...)
		if err != nil {
			return page[PronunciationDictionaryMetadata]{}, err
		}
		next := ""
		if resp.NextCursor != nil {
			next = *resp.NextCursor
		}
		return page[PronunciationDictionaryMetadata]{items: resp.PronunciationDictionaries, next: next, hasMore: resp.HasMore}, nil
	})
}

// PronunciationDictionaryRulesService groups rule editing endpoints.
type PronunciationDictionaryRulesService struct{ cfg *config }

// PronunciationDictionaryRulesResponse is the dictionary version produced by a rules edit.
type PronunciationDictionaryRulesResponse struct {
	ID              string `json:"id"`
	VersionID       string `json:"version_id"`
	VersionRulesNum int64  `json:"version_rules_num"`
}

type pronunciationRulesBody struct {
	Rules []PronunciationRule `json:"rules"`
}

type removeRulesBody struct {
	RuleStrings []string `json:"rule_strings"`
}

func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

// Set replaces all rules of a dictionary, creating a new version.
func (s *PronunciationDictionaryRulesService) Set(ctx context.Context, dictionaryID string, rules []PronunciationRule, opts ...RequestOption) (*PronunciationDictionaryRulesResponse, error) {
	return doValue[PronunciationDictionaryRulesResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/pronunciation-dictionaries/%s/set-rules", dictionaryID), body: pronunciationRulesBody{Rules: nonNil(rules)}, opts: opts})
}

// Add adds rules to a dictionary, creating a new version.
func (s *PronunciationDictionaryRulesService) Add(ctx context.Context, dictionaryID string, rules []PronunciationRule, opts ...RequestOption) (*PronunciationDictionaryRulesResponse, error) {
	return doValue[PronunciationDictionaryRulesResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/pronunciation-dictionaries/%s/add-rules", dictionaryID), body: pronunciationRulesBody{Rules: nonNil(rules)}, opts: opts})
}

// Remove removes the rules matching ruleStrings, creating a new version.
func (s *PronunciationDictionaryRulesService) Remove(ctx context.Context, dictionaryID string, ruleStrings []string, opts ...RequestOption) (*PronunciationDictionaryRulesResponse, error) {
	return doValue[PronunciationDictionaryRulesResponse](ctx, s.cfg, &request{method: http.MethodPost, path: pathf("/v1/pronunciation-dictionaries/%s/remove-rules", dictionaryID), body: removeRulesBody{RuleStrings: nonNil(ruleStrings)}, opts: opts})
}
