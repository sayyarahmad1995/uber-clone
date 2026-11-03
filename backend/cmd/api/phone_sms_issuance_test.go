package main

import (
    "strings"
    "testing"
)

func TestPhoneSMSTestIssuanceFlagsDisabledByDefault(t *testing.T) {
    t.Setenv("PHONE_SMS_ISSUANCE_ENABLED", "")
    t.Setenv("PHONE_SMS_TEST_ALLOWED_NUMBERS", "")
    cfg := loadConfig()
    if cfg.PhoneSMSIssuanceEnabled || cfg.PhoneSMSTestAllowedNumbers != "" {
        t.Fatal("SMS test issuance must be disabled by default")
    }
}

func TestPhoneSMSTestAllowlistIsExplicitAndNarrow(t *testing.T) {
    for _, raw := range []string{
        "", "all", "*", "+92", "+92300BAD123", "923001234567",
        "+923001234567,", strings.Repeat("+923001234567,", 10) + "+923001234567",
    } {
        if _, err := parsePhoneSMSTestAllowlist(raw); err == nil {
            t.Fatalf("unsafe SMS allowlist accepted: %q", raw)
        }
    }
    allowed, err := parsePhoneSMSTestAllowlist(" +923001234567, +923009876543 ")
    if err != nil || len(allowed) != 2 {
        t.Fatalf("valid allowlist rejected: count=%d, err=%v", len(allowed), err)
    }
    if _, ok := allowed["+923001234567"]; !ok {
        t.Fatal("expected normalized test number missing")
    }
}

func TestSMSIssuanceCannotStartWithoutWorker(t *testing.T) {
    _, closeFn, err := newApplication(config{PhoneSMSIssuanceEnabled: true})
    if err == nil || closeFn != nil || !strings.Contains(err.Error(), "PHONE_SMS_WORKER_ENABLED") {
        t.Fatalf("expected fail-closed SMS issue startup, got %v", err)
    }
}
