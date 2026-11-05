package user

import (
	"context"
	"net/http"
	"testing"

	"github.com/taigrr/elevenlabs/v2/internal/testutil"
)

func TestEndpoints(t *testing.T) {
	testutil.RunEndpoints(t, NewService, []testutil.Endpoint[*Service]{
		{
			Name:     "get",
			Call:     func(ctx context.Context, c *Service) (any, error) { return c.Get(ctx) },
			Method:   http.MethodGet,
			Path:     "/v1/user",
			RespBody: `{"user_id":"u1","subscription":{"tier":"creator","max_credit_limit_extension":"unlimited","character_count":5},"seat_type":"workspace_admin","created_at":17}`,
			Check: func(t *testing.T, r any) {
				u := r.(*User)
				if u.UserID != "u1" || u.Subscription.Tier != "creator" || !u.Subscription.MaxCreditLimitExtension.Unlimited ||
					u.Subscription.CharacterCount != 5 || u.SeatType != SeatTypeWorkspaceAdmin || u.CreatedAt != 17 {
					t.Errorf("user = %+v", u)
				}
			},
		},
		{
			Name:     "subscription.get",
			Call:     func(ctx context.Context, c *Service) (any, error) { return c.Subscription.Get(ctx) },
			Method:   http.MethodGet,
			Path:     "/v1/user/subscription",
			RespBody: `{"tier":"pro","max_credit_limit_extension":1000,"status":"active","open_invoices":[{"amount_due_cents":500,"discounts":[],"payment_intent_statusses":["succeeded"]}],"pending_change":{"kind":"cancellation","timestamp_seconds":9}}`,
			Check: func(t *testing.T, r any) {
				s := r.(*Subscription)
				if s.Tier != "pro" || s.MaxCreditLimitExtension.Value != 1000 || s.Status != SubscriptionStatusActive ||
					len(s.OpenInvoices) != 1 || s.OpenInvoices[0].PaymentIntentStatusses[0] != PaymentIntentStatusSucceeded ||
					s.PendingChange.Kind != "cancellation" {
					t.Errorf("subscription = %+v", s)
				}
			},
		},
	})
}
