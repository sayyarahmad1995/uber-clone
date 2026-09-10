package httpapi

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/driveronboarding"
)

type adminRejectDriverOnboardingRequest struct {
	Reason string `json:"reason"`
}

type adminDriverReviewDetailView struct {
	Application driveronboarding.Application
	Pending     bool
	Message     string
}

var adminDriverReviewListTemplate = template.Must(template.New("driver-review-list").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Driver onboarding review</title>
<style>
body{font-family:system-ui,sans-serif;max-width:960px;margin:40px auto;padding:0 20px;color:#1f2937}table{width:100%;border-collapse:collapse}th,td{text-align:left;padding:10px;border-bottom:1px solid #e5e7eb}a{color:#1d4ed8}code{font-size:.9em}.muted{color:#6b7280}
</style>
</head>
<body>
<h1>Driver onboarding review</h1>
<p class="muted">Pending applications only.</p>
{{if .}}
<table>
<thead><tr><th>Driver</th><th>Service</th><th>Vehicle</th><th>Submitted</th><th></th></tr></thead>
<tbody>
{{range .}}
<tr>
<td>{{.DisplayName}}</td>
<td>{{.Service.DisplayName}}</td>
<td>{{.Vehicle.Make}} {{.Vehicle.Model}} {{.Vehicle.ModelYear}} · {{.Vehicle.LicensePlate}}</td>
<td>{{.SubmittedAt}}</td>
<td><a href="/admin/driver-onboarding/{{.ID}}">Review</a></td>
</tr>
{{end}}
</tbody>
</table>
{{else}}
<p>No pending Driver onboarding applications.</p>
{{end}}
</body>
</html>`))

var adminDriverReviewDetailTemplate = template.Must(template.New("driver-review-detail").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Driver onboarding application</title>
<style>
body{font-family:system-ui,sans-serif;max-width:760px;margin:40px auto;padding:0 20px;color:#1f2937}.card{border:1px solid #e5e7eb;border-radius:10px;padding:18px;margin:18px 0}.grid{display:grid;grid-template-columns:160px 1fr;gap:8px 18px}.label{color:#6b7280}button{padding:10px 16px;border:0;border-radius:7px;cursor:pointer}.approve{background:#166534;color:white}.reject{background:#b91c1c;color:white}textarea{width:100%;min-height:90px;box-sizing:border-box;margin:8px 0 12px;padding:8px}a{color:#1d4ed8}.message{padding:10px;background:#eff6ff;border-radius:7px}
</style>
</head>
<body>
<p><a href="/admin/driver-onboarding">← Pending applications</a></p>
<h1>Driver onboarding application</h1>
{{if .Message}}<p class="message">{{.Message}}</p>{{end}}
<div class="card grid">
<div class="label">Application</div><div>{{.Application.ID}}</div>
<div class="label">Driver user</div><div>{{.Application.DriverUserID}}</div>
<div class="label">Display name</div><div>{{.Application.DisplayName}}</div>
<div class="label">Status</div><div>{{.Application.Status}}</div>
<div class="label">Service</div><div>{{.Application.Service.DisplayName}} ({{.Application.Service.Code}})</div>
<div class="label">Vehicle</div><div>{{.Application.Vehicle.Make}} {{.Application.Vehicle.Model}} {{.Application.Vehicle.ModelYear}}</div>
<div class="label">Color</div><div>{{.Application.Vehicle.Color}}</div>
<div class="label">License plate</div><div>{{.Application.Vehicle.LicensePlate}}</div>
<div class="label">Submitted</div><div>{{.Application.SubmittedAt}}</div>
{{if .Application.DecidedAt}}<div class="label">Decided</div><div>{{.Application.DecidedAt}} by {{.Application.DecidedBy}}</div>{{end}}
{{if .Application.RejectionReason}}<div class="label">Rejection reason</div><div>{{.Application.RejectionReason}}</div>{{end}}
</div>
{{if .Pending}}
<div class="card">
<h2>Approve</h2>
<p>Approval promotes this exact submitted Driver, vehicle, and selected service snapshot into approved records. It does not yet make the Driver operationally active.</p>
<form method="post" action="/admin/driver-onboarding/{{.Application.ID}}/approve">
<button class="approve" type="submit">Approve application</button>
</form>
</div>
<div class="card">
<h2>Reject</h2>
<form method="post" action="/admin/driver-onboarding/{{.Application.ID}}/reject">
<label for="reason">Reason</label>
<textarea id="reason" name="reason" required></textarea>
<button class="reject" type="submit">Reject application</button>
</form>
</div>
{{end}}
</body>
</html>`))

func (api *API) listAdminDriverOnboardingApplications(w http.ResponseWriter, r *http.Request) {
	applications, err := api.driverOnboardingReview.ListPending(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "unable to load pending Driver applications"})
		return
	}
	items := make([]map[string]any, 0, len(applications))
	for _, application := range applications {
		items = append(items, adminDriverOnboardingResponse(application))
	}
	writeJSON(w, http.StatusOK, map[string]any{"applications": items})
}

func (api *API) getAdminDriverOnboardingApplication(w http.ResponseWriter, r *http.Request) {
	applicationID, ok := parseAdminApplicationID(w, r)
	if !ok {
		return
	}
	application, err := api.driverOnboardingReview.Get(r.Context(), applicationID)
	if err != nil {
		writeDriverReviewError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, adminDriverOnboardingResponse(application))
}

func (api *API) approveAdminDriverOnboardingApplication(w http.ResponseWriter, r *http.Request) {
	applicationID, ok := parseAdminApplicationID(w, r)
	if !ok {
		return
	}
	application, err := api.driverOnboardingReview.Approve(r.Context(), applicationID, adminReviewer(r))
	if err != nil {
		writeDriverReviewError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, adminDriverOnboardingResponse(application))
}

func (api *API) rejectAdminDriverOnboardingApplication(w http.ResponseWriter, r *http.Request) {
	applicationID, ok := parseAdminApplicationID(w, r)
	if !ok {
		return
	}
	var body adminRejectDriverOnboardingRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	application, err := api.driverOnboardingReview.Reject(r.Context(), applicationID, adminReviewer(r), body.Reason)
	if err != nil {
		writeDriverReviewError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, adminDriverOnboardingResponse(application))
}

func (api *API) adminDriverOnboardingIndex(w http.ResponseWriter, r *http.Request) {
	applications, err := api.driverOnboardingReview.ListPending(r.Context())
	if err != nil {
		http.Error(w, "Unable to load pending Driver applications", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := adminDriverReviewListTemplate.Execute(w, applications); err != nil {
		return
	}
}

func (api *API) adminDriverOnboardingDetail(w http.ResponseWriter, r *http.Request) {
	applicationID, ok := parseAdminApplicationIDHTML(w, r)
	if !ok {
		return
	}
	application, err := api.driverOnboardingReview.Get(r.Context(), applicationID)
	if err != nil {
		http.Error(w, driverReviewErrorMessage(err), driverReviewStatus(err))
		return
	}
	view := adminDriverReviewDetailView{
		Application: application,
		Pending:     application.Status == driveronboarding.StatusPending,
		Message:     r.URL.Query().Get("message"),
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := adminDriverReviewDetailTemplate.Execute(w, view); err != nil {
		return
	}
}

func (api *API) adminApproveDriverOnboarding(w http.ResponseWriter, r *http.Request) {
	if !api.adminSameOrigin(r) {
		http.Error(w, "Invalid request origin", http.StatusForbidden)
		return
	}
	applicationID, ok := parseAdminApplicationIDHTML(w, r)
	if !ok {
		return
	}
	if _, err := api.driverOnboardingReview.Approve(r.Context(), applicationID, adminReviewer(r)); err != nil {
		http.Error(w, driverReviewErrorMessage(err), driverReviewStatus(err))
		return
	}
	http.Redirect(w, r, "/admin/driver-onboarding/"+applicationID.String()+"?message="+url.QueryEscape("Application approved"), http.StatusSeeOther)
}

func (api *API) adminRejectDriverOnboarding(w http.ResponseWriter, r *http.Request) {
	if !api.adminSameOrigin(r) {
		http.Error(w, "Invalid request origin", http.StatusForbidden)
		return
	}
	applicationID, ok := parseAdminApplicationIDHTML(w, r)
	if !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if _, err := api.driverOnboardingReview.Reject(r.Context(), applicationID, adminReviewer(r), r.FormValue("reason")); err != nil {
		http.Error(w, driverReviewErrorMessage(err), driverReviewStatus(err))
		return
	}
	http.Redirect(w, r, "/admin/driver-onboarding/"+applicationID.String()+"?message="+url.QueryEscape("Application rejected"), http.StatusSeeOther)
}

func parseAdminApplicationID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	applicationID, err := uuid.Parse(r.PathValue("application_id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid Driver application id"})
		return uuid.Nil, false
	}
	return applicationID, true
}

func parseAdminApplicationIDHTML(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	applicationID, err := uuid.Parse(r.PathValue("application_id"))
	if err != nil {
		http.Error(w, "Invalid Driver application id", http.StatusBadRequest)
		return uuid.Nil, false
	}
	return applicationID, true
}

func adminDriverOnboardingResponse(application driveronboarding.Application) map[string]any {
	response := driverOnboardingResponse(application)
	response["driver_user_id"] = application.DriverUserID
	response["decided_by"] = application.DecidedBy
	return response
}

func writeDriverReviewError(w http.ResponseWriter, err error) {
	writeJSON(w, driverReviewStatus(err), map[string]string{"error": driverReviewErrorMessage(err)})
}

func driverReviewStatus(err error) int {
	switch {
	case errors.Is(err, driveronboarding.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, driveronboarding.ErrInvalidReviewDecision):
		return http.StatusBadRequest
	case errors.Is(err, driveronboarding.ErrApplicationNotPending), errors.Is(err, driveronboarding.ErrAlreadyOnboarded):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func driverReviewErrorMessage(err error) string {
	switch {
	case errors.Is(err, driveronboarding.ErrNotFound):
		return "Driver onboarding application not found"
	case errors.Is(err, driveronboarding.ErrInvalidReviewDecision):
		return "reviewer and rejection reason are required"
	case errors.Is(err, driveronboarding.ErrApplicationNotPending):
		return "Driver onboarding application is no longer pending"
	case errors.Is(err, driveronboarding.ErrAlreadyOnboarded):
		return "Driver already has approved operational records"
	default:
		return "unable to review Driver onboarding application"
	}
}

func (api *API) adminSameOrigin(r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return true
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return false
	}
	if api.adminReviewOrigin != "" {
		// Use the configured public origin behind TLS termination, never client-supplied forwarding headers.
		expected, err := url.Parse(api.adminReviewOrigin)
		if err != nil || (expected.Scheme != "http" && expected.Scheme != "https") || expected.Host == "" || expected.User != nil || expected.Path != "" || expected.RawQuery != "" || expected.Fragment != "" {
			return false
		}
		return strings.EqualFold(parsed.Scheme, expected.Scheme) && strings.EqualFold(parsed.Host, expected.Host) && strings.EqualFold(expected.Host, r.Host)
	}
	return strings.EqualFold(parsed.Scheme, adminRequestScheme(r)) && strings.EqualFold(parsed.Host, r.Host)
}

func adminRequestScheme(r *http.Request) string {
	if r.URL.Scheme != "" {
		return strings.ToLower(r.URL.Scheme)
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}
