package user

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/taigrr/elevenlabs/v2/internal/requestconfig"
	"github.com/taigrr/elevenlabs/v2/internal/transport"
	"github.com/taigrr/elevenlabs/v2/option"
)

// Service groups the /v1/user endpoints.
type Service struct {
	cfg *requestconfig.Config

	// Subscription groups the /v1/user/subscription endpoints.
	Subscription *SubscriptionService
}

// NewService returns a Service configured by opts. Without option.WithAPIKey
// the key is read from the ELEVENLABS_API_KEY environment variable.
func NewService(opts ...option.RequestOption) *Service {
	return newService(requestconfig.Build(opts...))
}

func newService(cfg *requestconfig.Config) *Service {
	return &Service{cfg: cfg, Subscription: &SubscriptionService{cfg: cfg}}
}

// Get returns information about the authenticated user.
func (s *Service) Get(ctx context.Context, opts ...option.RequestOption) (*User, error) {
	return transport.DoValue[User](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/user", Opts: opts})
}

// SubscriptionService groups the /v1/user/subscription endpoints.
type SubscriptionService struct{ cfg *requestconfig.Config }

// Get returns extended subscription information for the authenticated user.
func (s *SubscriptionService) Get(ctx context.Context, opts ...option.RequestOption) (*Subscription, error) {
	return transport.DoValue[Subscription](ctx, s.cfg, &transport.Request{Method: http.MethodGet, Path: "/v1/user/subscription", Opts: opts})
}

// SeatType is a workspace seat type.
type SeatType string

// SeatType values.
const (
	SeatTypeWorkspaceAdmin      SeatType = "workspace_admin"
	SeatTypeWorkspaceMember     SeatType = "workspace_member"
	SeatTypeWorkspaceLiteMember SeatType = "workspace_lite_member"
)

// User is the response of Service.Get.
type User struct {
	UserID                         string               `json:"user_id"`
	WorkspaceID                    string               `json:"workspace_id"`
	Subscription                   SubscriptionResponse `json:"subscription"`
	IsNewUser                      bool                 `json:"is_new_user"`
	CanUseDelayedPaymentMethods    bool                 `json:"can_use_delayed_payment_methods"`
	IsOnboardingCompleted          bool                 `json:"is_onboarding_completed"`
	IsOnboardingChecklistCompleted bool                 `json:"is_onboarding_checklist_completed"`
	ShowComplianceTerms            *bool                `json:"show_compliance_terms,omitempty"`
	FirstName                      *string              `json:"first_name,omitempty"`
	IsAPIKeyHashed                 *bool                `json:"is_api_key_hashed,omitempty"`
	XiAPIKeyPreview                *string              `json:"xi_api_key_preview,omitempty"`
	ReferralLinkCode               *string              `json:"referral_link_code,omitempty"`
	PartnerstackPartnerDefaultLink *string              `json:"partnerstack_partner_default_link,omitempty"`
	CreatedAt                      int64                `json:"created_at"`
	SeatType                       SeatType             `json:"seat_type"`
}

// Currency is an ISO currency code used in billing responses.
type Currency string

// Currency values.
const (
	CurrencyUSD Currency = "usd"
	CurrencyEUR Currency = "eur"
	CurrencyINR Currency = "inr"
	CurrencyPLN Currency = "pln"
	CurrencyGBP Currency = "gbp"
)

// SubscriptionStatus is the status of a subscription.
type SubscriptionStatus string

// SubscriptionStatus values.
const (
	SubscriptionStatusTrialing     SubscriptionStatus = "trialing"
	SubscriptionStatusActive       SubscriptionStatus = "active"
	SubscriptionStatusIncomplete   SubscriptionStatus = "incomplete"
	SubscriptionStatusPastDue      SubscriptionStatus = "past_due"
	SubscriptionStatusFree         SubscriptionStatus = "free"
	SubscriptionStatusFreeDisabled SubscriptionStatus = "free_disabled"
)

// BillingPeriod is a billing or character refresh period.
type BillingPeriod string

// BillingPeriod values.
const (
	BillingPeriodMonthly    BillingPeriod = "monthly_period"
	BillingPeriodThreeMonth BillingPeriod = "3_month_period"
	BillingPeriodSixMonth   BillingPeriod = "6_month_period"
	BillingPeriodAnnual     BillingPeriod = "annual_period"
)

// PaymentIntentStatus is the status of an invoice payment intent.
type PaymentIntentStatus string

// PaymentIntentStatus values.
const (
	PaymentIntentStatusCanceled              PaymentIntentStatus = "canceled"
	PaymentIntentStatusProcessing            PaymentIntentStatus = "processing"
	PaymentIntentStatusRequiresAction        PaymentIntentStatus = "requires_action"
	PaymentIntentStatusRequiresCapture       PaymentIntentStatus = "requires_capture"
	PaymentIntentStatusRequiresConfirmation  PaymentIntentStatus = "requires_confirmation"
	PaymentIntentStatusRequiresPaymentMethod PaymentIntentStatus = "requires_payment_method"
	PaymentIntentStatusSucceeded             PaymentIntentStatus = "succeeded"
)

// CreditLimitExtension is either a number of credits or unlimited.
type CreditLimitExtension struct {
	Unlimited bool
	Value     float64
}

// UnmarshalJSON accepts a number or the string "unlimited".
func (c *CreditLimitExtension) UnmarshalJSON(data []byte) error {
	var s string
	if json.Unmarshal(data, &s) == nil {
		*c = CreditLimitExtension{Unlimited: s == "unlimited"}
		return nil
	}
	*c = CreditLimitExtension{}
	return json.Unmarshal(data, &c.Value)
}

// MarshalJSON implements json.Marshaler.
func (c CreditLimitExtension) MarshalJSON() ([]byte, error) {
	if c.Unlimited {
		return []byte(`"unlimited"`), nil
	}
	return json.Marshal(c.Value)
}

// Price is an amount in a currency.
type Price struct {
	Amount   string   `json:"amount"`
	Currency Currency `json:"currency"`
}

// Discount is a discount applied to an invoice.
type Discount struct {
	DiscountPercentOff *float64 `json:"discount_percent_off,omitempty"`
	DiscountAmountOff  *float64 `json:"discount_amount_off,omitempty"`
}

// Invoice describes an upcoming or open invoice.
type Invoice struct {
	AmountDueCents         int64                 `json:"amount_due_cents"`
	SubtotalCents          *int64                `json:"subtotal_cents,omitempty"`
	TaxCents               *int64                `json:"tax_cents,omitempty"`
	DiscountPercentOff     *float64              `json:"discount_percent_off,omitempty"`
	DiscountAmountOff      *float64              `json:"discount_amount_off,omitempty"`
	Discounts              []Discount            `json:"discounts"`
	NextPaymentAttemptUnix int64                 `json:"next_payment_attempt_unix"`
	PaymentIntentStatus    *PaymentIntentStatus  `json:"payment_intent_status,omitempty"`
	PaymentIntentStatusses []PaymentIntentStatus `json:"payment_intent_statusses"`
}

// PendingSubscriptionChange is a scheduled tier switch ("change") or
// cancellation ("cancellation"), discriminated by Kind.
type PendingSubscriptionChange struct {
	Kind              string        `json:"kind,omitempty"`
	NextTier          string        `json:"next_tier,omitempty"`
	NextBillingPeriod BillingPeriod `json:"next_billing_period,omitempty"`
	TimestampSeconds  int64         `json:"timestamp_seconds"`
}

// SubscriptionResponse is the subscription embedded in User.
type SubscriptionResponse struct {
	Tier                                  string               `json:"tier"`
	CharacterCount                        int64                `json:"character_count"`
	CharacterLimit                        int64                `json:"character_limit"`
	MaxCharacterLimitExtension            *int64               `json:"max_character_limit_extension,omitempty"`
	MaxCreditLimitExtension               CreditLimitExtension `json:"max_credit_limit_extension"`
	CanExtendCharacterLimit               bool                 `json:"can_extend_character_limit"`
	AllowedToExtendCharacterLimit         bool                 `json:"allowed_to_extend_character_limit"`
	NextCharacterCountResetUnix           *int64               `json:"next_character_count_reset_unix,omitempty"`
	VoiceSlotsUsed                        int64                `json:"voice_slots_used"`
	ProfessionalVoiceSlotsUsed            int64                `json:"professional_voice_slots_used"`
	ProfessionalVoiceSlotsUsedInWorkspace int64                `json:"professional_voice_slots_used_in_workspace"`
	VoiceLimit                            int64                `json:"voice_limit"`
	MaxVoiceAddEdits                      *int64               `json:"max_voice_add_edits,omitempty"`
	VoiceAddEditCounter                   int64                `json:"voice_add_edit_counter"`
	ProfessionalVoiceLimit                int64                `json:"professional_voice_limit"`
	CanExtendVoiceLimit                   bool                 `json:"can_extend_voice_limit"`
	CanUseInstantVoiceCloning             bool                 `json:"can_use_instant_voice_cloning"`
	CanUseProfessionalVoiceCloning        bool                 `json:"can_use_professional_voice_cloning"`
	Currency                              *Currency            `json:"currency,omitempty"`
	CurrentOverage                        Price                `json:"current_overage"`
	Status                                SubscriptionStatus   `json:"status"`
	BillingPeriod                         *BillingPeriod       `json:"billing_period,omitempty"`
	CharacterRefreshPeriod                *BillingPeriod       `json:"character_refresh_period,omitempty"`
}

// Subscription is the response of SubscriptionService.Get.
type Subscription struct {
	SubscriptionResponse
	NextInvoice                   *Invoice                   `json:"next_invoice,omitempty"`
	OpenInvoices                  []Invoice                  `json:"open_invoices"`
	HasOpenInvoices               bool                       `json:"has_open_invoices"`
	PendingChange                 *PendingSubscriptionChange `json:"pending_change,omitempty"`
	HasUsedStarterCouponOnAccount *bool                      `json:"has_used_starter_coupon_on_account,omitempty"`
	HasUsedCreatorCouponOnAccount *bool                      `json:"has_used_creator_coupon_on_account,omitempty"`
	IsEligibleForStarterPromo     *bool                      `json:"is_eligible_for_starter_promo,omitempty"`
}
