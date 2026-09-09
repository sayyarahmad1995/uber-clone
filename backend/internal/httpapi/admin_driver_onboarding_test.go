package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/driveronboarding"
)

func TestAdminReviewerRoutesDisabledWithoutCredentials(t *testing.T) {
	api := &API{}
	request := httptest.NewRequest(http.MethodGet, "/admin/driver-onboarding", nil)
	response := httptest.NewRecorder()

	api.routes().ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected admin route to be absent, got %d", response.Code)
	}
}

func TestAdminReviewerRequiresBasicAuthentication(t *testing.T) {
	repository := &fakeDriverReviewRepository{application: pendingAdminReviewApplication()}
	api := adminReviewTestAPI(repository)
	request := httptest.NewRequest(http.MethodGet, "/admin/driver-onboarding", nil)
	response := httptest.NewRecorder()

	api.routes().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
	if response.Header().Get("WWW-Authenticate") == "" {
		t.Fatal("expected Basic authentication challenge")
	}
}

func TestAdminReviewerListsPendingApplications(t *testing.T) {
	repository := &fakeDriverReviewRepository{application: pendingAdminReviewApplication()}
	api := adminReviewTestAPI(repository)
	request := httptest.NewRequest(http.MethodGet, "/admin/driver-onboarding", nil)
	request.SetBasicAuth("reviewer", "secret")
	response := httptest.NewRecorder()

	api.routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "Test Driver") || !strings.Contains(response.Body.String(), "Comfort") {
		t.Fatalf("reviewer page did not render pending application: %s", response.Body.String())
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("expected no-store admin response, got %q", response.Header().Get("Cache-Control"))
	}
}

func TestAdminReviewerApproveUsesAuthenticatedReviewer(t *testing.T) {
	application := pendingAdminReviewApplication()
	repository := &fakeDriverReviewRepository{application: application}
	api := adminReviewTestAPI(repository)
	request := httptest.NewRequest(
		http.MethodPost,
		"http://application.test/v1/admin/driver-onboarding-applications/"+application.ID.String()+"/approve",
		nil,
	)
	request.Host = "application.test"
	request.Header.Set("Origin", "http://application.test")
	request.SetBasicAuth("reviewer", "secret")
	response := httptest.NewRecorder()

	api.routes().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}
	if repository.application.Status != driveronboarding.StatusApproved || repository.application.DecidedBy != "reviewer" {
		t.Fatalf("approval did not use authenticated reviewer: %#v", repository.application)
	}
}

func TestAdminReviewerRejectRequiresReason(t *testing.T) {
	application := pendingAdminReviewApplication()
	repository := &fakeDriverReviewRepository{application: application}
	api := adminReviewTestAPI(repository)
	request := httptest.NewRequest(
		http.MethodPost,
		"http://application.test/v1/admin/driver-onboarding-applications/"+application.ID.String()+"/reject",
		strings.NewReader("{\"reason\":\"\"}"),
	)
	request.Host = "application.test"
	request.Header.Set("Origin", "http://application.test")
	request.Header.Set("Content-Type", "application/json")
	request.SetBasicAuth("reviewer", "secret")
	response := httptest.NewRecorder()

	api.routes().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", response.Code, response.Body.String())
	}
	if repository.application.Status != driveronboarding.StatusPending {
		t.Fatalf("invalid rejection mutated application: %#v", repository.application)
	}
}

func TestAdminReviewerRejectsCrossOriginMutation(t *testing.T) {
	application := pendingAdminReviewApplication()
	repository := &fakeDriverReviewRepository{application: application}
	api := adminReviewTestAPI(repository)
	request := httptest.NewRequest(
		http.MethodPost,
		"http://application.test/v1/admin/driver-onboarding-applications/"+application.ID.String()+"/approve",
		nil,
	)
	request.Host = "application.test"
	request.Header.Set("Origin", "https://malicious.example")
	request.SetBasicAuth("reviewer", "secret")
	response := httptest.NewRecorder()

	api.routes().ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", response.Code)
	}
	if repository.application.Status != driveronboarding.StatusPending {
		t.Fatalf("cross-origin request mutated application: %#v", repository.application)
	}
}

func TestAdminReviewerRejectsCrossSchemeMutation(t *testing.T) {
	application := pendingAdminReviewApplication()
	repository := &fakeDriverReviewRepository{application: application}
	api := adminReviewTestAPI(repository)
	request := httptest.NewRequest(
		http.MethodPost,
		"https://application.test/v1/admin/driver-onboarding-applications/"+application.ID.String()+"/approve",
		nil,
	)
	request.Host = "application.test"
	request.Header.Set("Origin", "http://application.test")
	request.SetBasicAuth("reviewer", "secret")
	response := httptest.NewRecorder()

	api.routes().ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", response.Code)
	}
	if repository.application.Status != driveronboarding.StatusPending {
		t.Fatalf("cross-scheme request mutated application: %#v", repository.application)
	}
}

func TestAdminReviewerBehindHTTPSProxy(t *testing.T) {
	for _, tc := range []struct {
		name, action, origin, host, configured string
		want                                   int
	}{
		{"reject", "reject", "https://application.test", "application.test", "https://application.test", http.StatusSeeOther},
		{"approve", "approve", "https://application.test", "application.test", "https://application.test", http.StatusSeeOther},
		{"other origin", "reject", "https://evil.test", "application.test", "https://application.test", http.StatusForbidden},
		{"wrong scheme", "reject", "http://application.test", "application.test", "https://application.test", http.StatusForbidden},
		{"wrong host", "reject", "https://application.test", "evil.test", "https://application.test", http.StatusForbidden},
		{"invalid configuration", "reject", "https://application.test", "application.test", ":invalid", http.StatusForbidden},
		{"untrusted forwarding header", "reject", "https://application.test", "application.test", "", http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repository := &fakeDriverReviewRepository{application: pendingAdminReviewApplication()}
			api := adminReviewTestAPI(repository)
			api.adminReviewOrigin = tc.configured
			request := httptest.NewRequest(http.MethodPost, "http://"+tc.host+"/admin/driver-onboarding/"+repository.application.ID.String()+"/"+tc.action, strings.NewReader("reason=Incomplete+documents"))
			request.Header.Set("Origin", tc.origin)
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set("X-Forwarded-Proto", "https")
			request.SetBasicAuth("reviewer", "secret")
			response := httptest.NewRecorder()
			api.routes().ServeHTTP(response, request)
			if response.Code != tc.want {
				t.Fatalf("expected %d, got %d: %s", tc.want, response.Code, response.Body.String())
			}
			if tc.want == http.StatusForbidden && repository.application.Status != driveronboarding.StatusPending {
				t.Fatal("blocked request changed application")
			}
			if tc.want == http.StatusSeeOther && repository.application.Status == driveronboarding.StatusPending {
				t.Fatal("accepted request did not apply decision")
			}
		})
	}
}

func adminReviewTestAPI(repository *fakeDriverReviewRepository) *API {
	return &API{
		driverOnboardingReview: driveronboarding.NewReviewService(repository),
		adminReviewUsername:    "reviewer",
		adminReviewPassword:    "secret",
	}
}

type fakeDriverReviewRepository struct {
	application driveronboarding.Application
}

func (r *fakeDriverReviewRepository) ListPendingApplications(context.Context) ([]driveronboarding.Application, error) {
	if r.application.Status == driveronboarding.StatusPending {
		return []driveronboarding.Application{r.application}, nil
	}
	return []driveronboarding.Application{}, nil
}

func (r *fakeDriverReviewRepository) FindApplicationByID(_ context.Context, applicationID uuid.UUID) (driveronboarding.Application, error) {
	if r.application.ID != applicationID {
		return driveronboarding.Application{}, driveronboarding.ErrNotFound
	}
	return r.application, nil
}

func (r *fakeDriverReviewRepository) ApproveApplication(_ context.Context, applicationID uuid.UUID, reviewer string) (driveronboarding.Application, error) {
	if r.application.ID != applicationID {
		return driveronboarding.Application{}, driveronboarding.ErrNotFound
	}
	if r.application.Status != driveronboarding.StatusPending {
		return driveronboarding.Application{}, driveronboarding.ErrApplicationNotPending
	}
	now := time.Now().UTC()
	r.application.Status = driveronboarding.StatusApproved
	r.application.DecidedAt = &now
	r.application.DecidedBy = reviewer
	return r.application, nil
}

func (r *fakeDriverReviewRepository) RejectApplication(_ context.Context, applicationID uuid.UUID, reviewer, reason string) (driveronboarding.Application, error) {
	if r.application.ID != applicationID {
		return driveronboarding.Application{}, driveronboarding.ErrNotFound
	}
	if r.application.Status != driveronboarding.StatusPending {
		return driveronboarding.Application{}, driveronboarding.ErrApplicationNotPending
	}
	now := time.Now().UTC()
	r.application.Status = driveronboarding.StatusRejected
	r.application.RejectionReason = reason
	r.application.DecidedAt = &now
	r.application.DecidedBy = reviewer
	return r.application, nil
}

func pendingAdminReviewApplication() driveronboarding.Application {
	return driveronboarding.Application{
		ID:           uuid.New(),
		DriverUserID: uuid.New(),
		DisplayName:  "Test Driver",
		Service: driveronboarding.ServiceOption{
			Code:        "comfort",
			DisplayName: "Comfort",
		},
		Vehicle: driveronboarding.VehicleInput{
			Make:         "Toyota",
			Model:        "Corolla",
			ModelYear:    2024,
			Color:        "White",
			LicensePlate: "ABC-123",
		},
		Status:      driveronboarding.StatusPending,
		SubmittedAt: time.Now().UTC(),
	}
}
