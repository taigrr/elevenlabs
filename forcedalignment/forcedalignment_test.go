package forcedalignment

import (
	"context"
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/internal/testutil"
)

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, []testutil.Endpoint[*Service]{
		{
			Name: "forcedAlignment.create",
			Call: func(ctx context.Context, c *Service) (any, error) {
				return c.Create(ctx, &CreateRequest{File: testutil.File("a.mp3", "mp3"), Text: "hi"})
			},
			Method: http.MethodPost, Path: "/v1/forced-alignment",
			Form:     map[string][]string{"file": {"file:a.mp3:mp3"}, "text": {"hi"}},
			RespBody: `{"characters":[{"text":"h","start":0,"end":0.1}],"words":[{"text":"hi","start":0,"end":0.2,"loss":0.3}],"loss":0.3}`,
			Check: func(t *testing.T, r any) {
				a := r.(*Alignment)
				if len(a.Characters) != 1 || a.Words[0].Loss != 0.3 || a.Loss != 0.3 {
					t.Errorf("alignment = %+v", a)
				}
			},
		},
	})
}
