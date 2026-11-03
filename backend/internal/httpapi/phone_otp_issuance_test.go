package httpapi

import (
    "context"
    "encoding/json"
    "errors"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "time"

    "github.com/google/uuid"
    "github.com/sayyarahmad1995/uber-clone/backend/internal/auth"
)

const testPhone = "+923001234567"
const testIdempotencyKey = "7bbd313a-49da-40eb-8440-24a5ab7855a6"

type stubChallengeIssuer struct {
    calls int
    phone string
    source string
    idempotency string
    result auth.PhoneChallengeCreation
    err error
}

func (s *stubChallengeIssuer) Issue(_ context.Context, req auth.PhoneChallengeRequest, source, idempotency string) (auth.PhoneChallengeCreation, error) {
    s.calls++
    s.phone = req.PhoneNumber
    s.source = source
    s.idempotency = idempotency
    return s.result, s.err
}

func issueTestAPI(s *stubChallengeIssuer) *API {
    return New(Dependencies{
        PhoneSMSIssuanceEnabled: true,
        PhoneOTPIssuer: s,
        PhoneSMSTestAllowedNumbers: map[string]struct{}{testPhone: {}},
    })
}

func issueTestRequest(body string, withKey bool) *http.Request {
    r := httptest.NewRequest(http.MethodPost, "/v1/auth/phone/challenges", strings.NewReader(body))
    if withKey {
        r.Header.Set("Idempotency-Key", testIdempotencyKey)
    }
    return r
}

func TestPhoneChallengeIssuanceIsClosedUnlessExplicitlyConfigured(t *testing.T) {
    stub := &stubChallengeIssuer{}
    for _, deps := range []Dependencies{
        {},
        {PhoneSMSIssuanceEnabled: true, PhoneOTPIssuer: stub},
        {PhoneOTPIssuer: stub, PhoneSMSTestAllowedNumbers: map[string]struct{}{testPhone: {}}},
        {PhoneSMSIssuanceEnabled: true, PhoneSMSTestAllowedNumbers: map[string]struct{}{testPhone: {}}},
    } {
        resp := httptest.NewRecorder()
        New(deps).Handler().ServeHTTP(resp, issueTestRequest(`{"phone_number":"+923001234567","purpose":"sign_in_or_register"}`, true))
        if resp.Code != http.StatusNotFound {
            t.Fatalf("unexpected route status %d for deps %+v", resp.Code, deps)
        }
    }
    if stub.calls != 0 {
        t.Fatal("SMS issuer was invoked despite disabled route")
    }
}

func TestPhoneChallengeIssuanceRejectsBadInputsBeforeSMS(t *testing.T) {
    stub := &stubChallengeIssuer{}
    api := issueTestAPI(stub)
    for _, tc := range []struct {
        name string
        body string
        key bool
        remote string
        want int
    }{
        {"invalid JSON", "{", true, "", 400},
        {"empty JSON", "{}", true, "", 400},
        {"unknown field", `{"phone_number":"+923001234567","purpose":"sign_in_or_register","extra":1}`, true, "", 400},
        {"trailing JSON", `{"phone_number":"+923001234567","purpose":"sign_in_or_register"}{}`, true, "", 400},
        {"wrong purpose", `{"phone_number":"+923001234567","purpose":"password_reset"}`, true, "", 400},
        {"unapproved number", `{"phone_number":"+923009876543","purpose":"sign_in_or_register"}`, true, "", 403},
        {"missing idempotency", `{"phone_number":"+923001234567","purpose":"sign_in_or_register"}`, false, "", 400},
        {"untrusted peer", `{"phone_number":"+923001234567","purpose":"sign_in_or_register"}`, true, "bad", 503},
    } {
        t.Run(tc.name, func(t *testing.T) {
            req := issueTestRequest(tc.body, tc.key)
            if tc.remote != "" {
                req.RemoteAddr = tc.remote
            }
            resp := httptest.NewRecorder()
            api.Handler().ServeHTTP(resp, req)
            if resp.Code != tc.want {
                t.Fatalf("status %d, expected %d: %s", resp.Code, tc.want, resp.Body.String())
            }
            if resp.Header().Get("Cache-Control") != "no-store" {
                t.Fatal("OTP response may be cached")
            }
        })
    }
    if stub.calls != 0 {
        t.Fatalf("invalid requests reached SMS issuer %d times", stub.calls)
    }
}

func TestPhoneChallengeIssuanceQueuesOnceAndReturnsOnlyMetadata(t *testing.T) {
    id := uuid.New()
    stub := &stubChallengeIssuer{result: auth.PhoneChallengeCreation{
        ID: id,
        ShouldSend: true,
        ResendAfter: time.Now().Add(30 * time.Second),
    }}
    api := issueTestAPI(stub)
    body := `{"phone_number":"+923001234567","purpose":"sign_in_or_register"}`
    resp := httptest.NewRecorder()
    api.Handler().ServeHTTP(resp, issueTestRequest(body, true))
    if resp.Code != http.StatusAccepted {
        t.Fatalf("expected accepted, got %d: %s", resp.Code, resp.Body.String())
    }
    if stub.calls != 1 || stub.phone != testPhone || stub.source != "192.0.2.1" || stub.idempotency != testIdempotencyKey {
        t.Fatalf("issuer inputs incorrect: calls=%d, phone=%q, source=%q", stub.calls, stub.phone, stub.source)
    }
    var payload struct {
        ChallengeID string `json:"challenge_id"`
        ResendAfterSeconds int `json:"resend_after_seconds"`
    }
    if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
        t.Fatal(err)
    }
    if payload.ChallengeID != id.String() || payload.ResendAfterSeconds < 1 || payload.ResendAfterSeconds > 30 {
        t.Fatalf("bad challenge response %+v", payload)
    }
    if strings.Contains(resp.Body.String(), testPhone) || strings.Contains(resp.Body.String(), "otp") {
        t.Fatal("challenge response disclosed sensitive data")
    }
}

func TestPhoneChallengeIssuanceBackoffAndFailures(t *testing.T) {
    stub := &stubChallengeIssuer{}
    api := issueTestAPI(stub)
    body := `{"phone_number":"+923001234567","purpose":"sign_in_or_register"}`
    for _, tc := range []struct {
        err error
        result auth.PhoneChallengeCreation
        code int
        retry string
        exposed string
    }{
        {auth.ErrPhoneChallengeCooldown, auth.PhoneChallengeCreation{ResendAfter: time.Now().Add(25 * time.Second)}, 429, "25", "resend_cooldown"},
        {auth.ErrPhoneOTPRateLimited, auth.PhoneChallengeCreation{}, 429, "3600", "rate_limited"},
        {errors.New("redis secret detail"), auth.PhoneChallengeCreation{}, 503, "", "issuance_unavailable"},
    } {
        stub.err = tc.err
        stub.result = tc.result
        resp := httptest.NewRecorder()
        api.Handler().ServeHTTP(resp, issueTestRequest(body, true))
        if resp.Code != tc.code || !strings.Contains(resp.Body.String(), tc.exposed) {
            t.Fatalf("unexpected error response %d %s", resp.Code, resp.Body.String())
        }
        if tc.retry != "" && resp.Header().Get("Retry-After") == "" {
            t.Fatal("missing retry hint")
        }
        if strings.Contains(resp.Body.String(), "redis secret detail") {
            t.Fatal("internal error was exposed")
        }
    }
}

func TestSMSIssueModeAllowsProofVerificationWithoutRegistration(t *testing.T) {
    issuer := &stubChallengeIssuer{}
    verifier := &stubProofIssuer{}
    api := New(Dependencies{
        PhoneSMSIssuanceEnabled: true,
        PhoneOTPIssuer: issuer,
        PhoneSMSTestAllowedNumbers: map[string]struct{}{testPhone: {}},
        PhoneOTPVerifier: verifier,
    })
    req := httptest.NewRequest(http.MethodPost, "/v1/auth/phone/challenges/verify",
        strings.NewReader(`{"challenge_id":"fd19cba6-d5f9-4c1a-b5e4-81c191baae4f","code":"123456"}`))
    resp := httptest.NewRecorder()
    api.Handler().ServeHTTP(resp, req)
    if resp.Code != http.StatusOK || verifier.calls != 1 {
        t.Fatalf("proof endpoint unavailable in SMS test mode: %d", resp.Code)
    }
}
