package pronunciation

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/taigrr/elevenlabs/v2/internal/testutil"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
	"github.com/taigrr/elevenlabs/v2/types"
)

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, []testutil.Endpoint[*Service]{
		{
			Name: "pronunciationDictionaries.createFromFile",
			Call: func(ctx context.Context, c *Service) (any, error) {
				f := testutil.File("d.pls", "PLS")
				return c.CreateFromFile(ctx, &CreateFromFileRequest{Name: "n", File: &f, Description: "d", WorkspaceAccess: types.ResourcePermissionEditor})
			},
			Method: http.MethodPost, Path: "/v1/pronunciation-dictionaries/add-from-file",
			Form:     map[string][]string{"name": {"n"}, "file": {"file:d.pls:PLS"}, "description": {"d"}, "workspace_access": {"editor"}},
			RespBody: `{"id":"pd","name":"n","version_id":"v1","version_rules_num":3,"permission_on_resource":"admin"}`,
			Check: func(t *testing.T, r any) {
				if p := r.(*AddResponse); p.ID != "pd" || p.VersionRulesNum != 3 || *p.PermissionOnResource != types.ResourcePermissionAdmin {
					t.Errorf("resp = %+v", p)
				}
			},
		},
		{
			Name: "pronunciationDictionaries.createFromRules",
			Call: func(ctx context.Context, c *Service) (any, error) {
				alias := AliasRule("Thailand", "tie-land")
				alias.CaseSensitive = types.Ptr(true)
				return c.CreateFromRules(ctx, &CreateFromRulesRequest{
					Rules: []Rule{alias, PhonemeRule("tomato", "təˈmeɪtoʊ", "ipa")},
					Name:  "My Dictionary",
				})
			},
			Method: http.MethodPost, Path: "/v1/pronunciation-dictionaries/add-from-rules",
			Body:     `{"rules":[{"type":"alias","string_to_replace":"Thailand","case_sensitive":true,"alias":"tie-land"},{"type":"phoneme","string_to_replace":"tomato","phoneme":"təˈmeɪtoʊ","alphabet":"ipa"}],"name":"My Dictionary"}`,
			RespBody: `{"id":"pd"}`,
		},
		{
			Name:   "pronunciationDictionaries.get",
			Call:   func(ctx context.Context, c *Service) (any, error) { return c.Get(ctx, "pd") },
			Method: http.MethodGet, Path: "/v1/pronunciation-dictionaries/pd",
			RespBody: `{"id":"pd","latest_version_id":"v","name":"n","rules":[{"type":"alias","string_to_replace":"a","alias":"b"},{"type":"phoneme","string_to_replace":"c","phoneme":"d","alphabet":"ipa"}]}`,
			Check: func(t *testing.T, r any) {
				d := r.(*DictionaryWithRules)
				if d.ID != "pd" || d.LatestVersionID != "v" || len(d.Rules) != 2 || d.Rules[1].Type != RuleTypePhoneme || d.Rules[0].Alias != "b" {
					t.Errorf("dictionary = %+v", d)
				}
			},
		},
		{
			Name: "pronunciationDictionaries.update",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Update(ctx, "pd", &UpdateRequest{Archived: types.Ptr(true)})
			},
			Method: http.MethodPatch, Path: "/v1/pronunciation-dictionaries/pd", Body: `{"archived":true}`, RespBody: `{"id":"pd"}`,
		},
		{
			Name: "pronunciationDictionaries.download",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return testutil.ReadAll(c.Download(ctx, "pd", "v"))
			},
			Method: http.MethodGet, Path: "/v1/pronunciation-dictionaries/pd/v/download", Accept: transport.AcceptAny,
			RespType: "application/pls+xml", RespBody: "<lexicon/>",
			Check: testutil.CheckString("<lexicon/>"),
		},
		{
			Name: "pronunciationDictionaries.list",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.List(ctx, &ListRequest{PageSize: types.Ptr(2), Sort: SortName, IncludeArchived: types.Ptr(false)})
			},
			Method: http.MethodGet, Path: "/v1/pronunciation-dictionaries",
			Query:    "include_archived=false&page_size=2&sort=name",
			RespBody: `{"pronunciation_dictionaries":[{"id":"a"}],"has_more":false}`,
		},
		{
			Name: "pronunciationDictionaries.rules.set",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Rules.Set(ctx, "pd", []Rule{AliasRule("a", "b")})
			},
			Method: http.MethodPost, Path: "/v1/pronunciation-dictionaries/pd/set-rules",
			Body:     `{"rules":[{"type":"alias","string_to_replace":"a","alias":"b"}]}`,
			RespBody: `{"id":"pd","version_id":"v2","version_rules_num":1}`,
			Check: func(t *testing.T, r any) {
				if p := r.(*RulesResponse); p.VersionID != "v2" || p.VersionRulesNum != 1 {
					t.Errorf("resp = %+v", p)
				}
			},
		},
		{
			Name: "pronunciationDictionaries.rules.add",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Rules.Add(ctx, "pd", nil)
			},
			Method: http.MethodPost, Path: "/v1/pronunciation-dictionaries/pd/add-rules", Body: `{"rules":[]}`, RespBody: `{}`,
		},
		{
			Name: "pronunciationDictionaries.rules.remove",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Rules.Remove(ctx, "pd", []string{"a"})
			},
			Method: http.MethodPost, Path: "/v1/pronunciation-dictionaries/pd/remove-rules", Body: `{"rule_strings":["a"]}`, RespBody: `{}`,
		},
	})
}

func TestListAll(t *testing.T) {
	client, ts := newTestServer(t, testutil.Sequence(
		testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"pronunciation_dictionaries":[{"id":"a"}],"next_cursor":"c2","has_more":true}`),
		testutil.Respond(http.StatusOK, transport.ContentTypeJSON, `{"pronunciation_dictionaries":[{"id":"b"}],"has_more":false}`),
	))
	var ids []string
	for d, err := range client.ListAll(context.Background(), nil) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, d.ID)
	}
	if strings.Join(ids, ",") != "a,b" || ts.Last(t).RawQuery != "cursor=c2" {
		t.Errorf("ids = %v, last query = %q", ids, ts.Last(t).RawQuery)
	}
}

func newTestServer(t *testing.T, h http.HandlerFunc, opts ...option.RequestOption) (*Service, *testutil.Server) {
	t.Helper()
	ts, o := testutil.NewServer(t, h, opts...)
	return NewService(o...), ts
}
