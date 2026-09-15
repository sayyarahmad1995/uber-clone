package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakeReadinessChecker struct {
	err   error
	calls int
}

func (f *fakeReadinessChecker) Check(context.Context) error {
	f.calls++
	return f.err
}

func TestReadyUsesReadinessChecker(t *testing.T) {
	checker := &fakeReadinessChecker{}
	api := &API{readiness: checker}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/ready", nil)

	api.ready(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if checker.calls != 1 {
		t.Fatalf("checker calls = %d, want 1", checker.calls)
	}
}

func TestReadyReturnsUnavailableWhenCheckerFails(t *testing.T) {
	checker := &fakeReadinessChecker{err: errors.New("database unavailable")}
	api := &API{readiness: checker}
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/ready", nil)

	api.ready(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusServiceUnavailable, recorder.Body.String())
	}
	if checker.calls != 1 {
		t.Fatalf("checker calls = %d, want 1", checker.calls)
	}
}
