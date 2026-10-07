package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/database"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/locationsearch"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/platform/migrations"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/pricing"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/rideservice"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/routing"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/user"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

func pricedDB(t *testing.T) (*sql.DB, string, uuid.UUID, *pricing.PostgresRepository) {
	t.Helper()
	raw := os.Getenv("TEST_DATABASE_URL")
	if raw == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	u, e := url.Parse(raw)
	if e != nil || !strings.HasSuffix(strings.TrimPrefix(u.Path, "/"), "_test") {
		t.Fatal("test database required")
	}
	db, e := database.Open(raw)
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Close() })
	if e := migrations.Apply(db); e != nil {
		t.Fatal(e)
	}
	code := "http_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	id := uuid.New()
	if _, e := db.Exec("INSERT INTO users(id) VALUES($1)", id); e != nil {
		t.Fatal(e)
	}
	if _, e := db.Exec("INSERT INTO driver_service_catalog(code,display_name,is_active,rider_visible) VALUES($1,'HTTP third service',true,true)", code); e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { db.Exec("UPDATE driver_service_catalog SET rider_visible=false WHERE code=$1", code) })
	return db, code, id, pricing.NewPostgresRepository(db)
}
func pricedDraft(code string) pricing.Draft {
	return pricing.Draft{ServiceCode: code, Currency: "PKR", BaseFareMinor: 10000, RateMinorPerKm: 1000, RateMinorPerMinute: 100, MinimumFareMinor: 10000, RoundingIncrementMinor: 100}
}
func pricedAPI(db *sql.DB, id uuid.UUID, p *pricing.PostgresRepository, enabled bool, provider *routePreviewSpy) *API {
	return New(Dependencies{Identity: routeIdentity{}, Users: user.NewService(routeUserRepository{value: user.User{ID: id, Capabilities: []user.Capability{user.CapabilityRider}}}), Pricing: p, SuggestedFaresEnabled: enabled, Routing: routing.NewService(provider), Rides: ride.NewService(ride.NewPostgresRepositoryWithPricing(db, p, enabled)), RideServices: rideservice.NewService(rideservice.NewPostgresRepositoryWithPricing(db, enabled))})
}
func pricedHTTP(api *API, path string, body any, auth bool) *httptest.ResponseRecorder {
	data, _ := json.Marshal(body)
	method := http.MethodPost
	if body == nil {
		method = http.MethodGet
	}
	req := httptest.NewRequest(method, path, strings.NewReader(string(data)))
	if auth {
		req.Header.Set("Authorization", "Bearer test")
	}
	res := httptest.NewRecorder()
	api.Handler().ServeHTTP(res, req)
	return res
}
func previewBody(code string) map[string]any {
	return map[string]any{"service_code": code, "pickup": map[string]any{"latitude": 24.86, "longitude": 67.01}, "destination": map[string]any{"latitude": 24.90, "longitude": 67.05}}
}
func TestPricedPreview(t *testing.T) {
	db, code, id, p := pricedDB(t)
	policy, e := p.Publish(context.Background(), pricedDraft(code), "owner")
	if e != nil {
		t.Fatal(e)
	}
	spy := &routePreviewSpy{}
	api := pricedAPI(db, id, p, true, spy)
	res := pricedHTTP(api, "/v1/ride-previews", previewBody(code), true)
	if res.Code != 200 {
		t.Fatalf("%d %s", res.Code, res.Body.String())
	}
	var v struct {
		Fare struct {
			Amount   int64  `json:"amount_minor"`
			Currency string `json:"currency"`
		} `json:"suggested_fare"`
		Version string `json:"pricing_policy_version"`
	}
	if e := json.Unmarshal(res.Body.Bytes(), &v); e != nil {
		t.Fatal(e)
	}
	if v.Fare.Amount != 12500 || v.Fare.Currency != "PKR" || v.Version != policy.ID.String() || spy.calls != 1 {
		t.Fatalf("preview %+v calls %d", v, spy.calls)
	}
	if res := pricedHTTP(api, "/v1/ride-previews", previewBody(code), false); res.Code != 401 {
		t.Fatalf("unsigned %d", res.Code)
	}
	if e := p.Disable(context.Background(), code, "PKR", policy.ID, "owner"); e != nil {
		t.Fatal(e)
	}
	res = pricedHTTP(api, "/v1/ride-previews", previewBody(code), true)
	if res.Code != 503 || spy.calls != 1 {
		t.Fatalf("unpriced %d calls %d: %s", res.Code, spy.calls, res.Body.String())
	}
	res = pricedHTTP(api, "/v1/ride-previews", previewBody("missing"), true)
	if res.Code != 400 || spy.calls != 1 {
		t.Fatalf("unknown %d", res.Code)
	}
}
func TestPricingRollout(t *testing.T) {
	db, code, id, p := pricedDB(t)
	spy := &routePreviewSpy{}
	api := pricedAPI(db, id, p, false, spy)
	res := pricedHTTP(api, "/v1/ride-previews", previewBody(code), true)
	if res.Code != 200 || strings.Contains(res.Body.String(), "suggested_fare") {
		t.Fatalf("disabled %d %s", res.Code, res.Body.String())
	}
	res = pricedHTTP(api, "/v1/ride-services", nil, true)
	if res.Code != 200 || !strings.Contains(res.Body.String(), code) || !strings.Contains(res.Body.String(), "\"pricing_required\":false") {
		t.Fatalf("manual catalog %s", res.Body.String())
	}
	enabled := pricedAPI(db, id, p, true, spy)
	res = pricedHTTP(enabled, "/v1/ride-services", nil, true)
	if strings.Contains(res.Body.String(), code) {
		t.Fatal("unpriced service exposed")
	}
	policy, e := p.Publish(context.Background(), pricedDraft(code), "owner")
	if e != nil {
		t.Fatal(e)
	}
	res = pricedHTTP(enabled, "/v1/ride-services", nil, true)
	if !strings.Contains(res.Body.String(), code) || !strings.Contains(res.Body.String(), "\"pricing_required\":true") {
		t.Fatalf("priced catalog %s", res.Body.String())
	}
	p.Disable(context.Background(), code, "PKR", policy.ID, "owner")
	res = pricedHTTP(enabled, "/v1/ride-services", nil, true)
	if strings.Contains(res.Body.String(), code) {
		t.Fatal("disabled price exposed")
	}
}
func TestPricedRideRequestHTTP(t *testing.T) {
	db, code, id, p := pricedDB(t)
	a, e := p.Publish(context.Background(), pricedDraft(code), "A")
	if e != nil {
		t.Fatal(e)
	}
	api := pricedAPI(db, id, p, true, &routePreviewSpy{})
	body := previewBody(code)
	body["proposed_fare"] = map[string]any{"amount_minor": 110000, "currency": "PKR"}
	res := pricedHTTP(api, "/v1/ride-requests", body, true)
	if res.Code != 400 {
		t.Fatalf("missing version %d %s", res.Code, res.Body.String())
	}
	body["pricing_policy_version"] = a.ID.String()
	p.Publish(context.Background(), pricedDraft(code), "B")
	res = pricedHTTP(api, "/v1/ride-requests", body, true)
	if res.Code != 409 || !strings.Contains(res.Body.String(), "pricing_policy_changed") {
		t.Fatalf("conflict %d %s", res.Code, res.Body.String())
	}
	b, e := p.Current(context.Background(), code, "PKR")
	if e != nil {
		t.Fatal(e)
	}
	body["pricing_policy_version"] = b.ID.String()
	body["proposed_fare"] = map[string]any{"amount_minor": 12500, "currency": "PKR"}
	res = pricedHTTP(api, "/v1/ride-requests", body, true)
	if res.Code != 201 || !strings.Contains(res.Body.String(), "12500") {
		t.Fatalf("create %d %s", res.Code, res.Body.String())
	}
	var n int
	if e := db.QueryRow("SELECT count(*) FROM ride_requests WHERE rider_user_id=$1", id).Scan(&n); e != nil || n != 1 {
		t.Fatalf("requests %d %v", n, e)
	}
}

func TestReadOnlySuggestedFareHTTP(t *testing.T) {
	db, code, id, policies := pricedDB(t)
	policy, err := policies.Publish(context.Background(), pricedDraft(code), "owner")
	if err != nil {
		t.Fatal(err)
	}
	spy := &readOnlyRouteSpy{}
	api := pricedAPI(db, id, policies, true, &routePreviewSpy{})
	api.routing = routing.NewService(spy)
	api.locationSearch = &pricingPlaces{}
	body := previewBody(code)
	body["pickup"].(map[string]any)["place_id"] = "pickup-poi"
	body["destination"].(map[string]any)["place_id"] = "destination-poi"
	body["pricing_policy_version"] = policy.ID.String()
	body["proposed_fare"] = map[string]any{"amount_minor": 110000, "currency": "PKR"}
	res := pricedHTTP(api, "/v1/ride-requests", body, true)
	if res.Code != 409 || !strings.Contains(res.Body.String(), "suggested_fare_changed") {
		t.Fatalf("edited fare accepted: %d %s", res.Code, res.Body.String())
	}
	var count int
	if err := db.QueryRow("SELECT count(*) FROM ride_requests WHERE rider_user_id=$1", id).Scan(&count); err != nil || count != 0 {
		t.Fatalf("rejected fare wrote request: %d %v", count, err)
	}
	if spy.pickup.PlaceID != "pickup-poi" || spy.destination.PlaceID != "destination-poi" {
		t.Fatalf("lost routing Place IDs: %+v %+v", spy.pickup, spy.destination)
	}
	body["proposed_fare"] = map[string]any{"amount_minor": 12500, "currency": "PKR"}
	spy.err = routing.ErrUnavailable
	res = pricedHTTP(api, "/v1/ride-requests", body, true)
	if res.Code != 503 {
		t.Fatalf("provider failure created request: %d %s", res.Code, res.Body.String())
	}
	spy.err = nil
	res = pricedHTTP(api, "/v1/ride-requests", body, true)
	if res.Code != 201 || !strings.Contains(res.Body.String(), "12500") || spy.calls != 3 {
		t.Fatalf("suggestion booking: %d %s calls=%d", res.Code, res.Body.String(), spy.calls)
	}
}

type readOnlyRouteSpy struct {
	calls       int
	pickup      routing.Endpoint
	destination routing.Endpoint
	err         error
}

func (s *readOnlyRouteSpy) Preview(_ context.Context, pickup, destination routing.Endpoint) (routing.Route, error) {
	s.calls++
	s.pickup, s.destination = pickup, destination
	return routing.Route{DistanceMeters: 2000, DurationSeconds: 300, EncodedPolyline: "??_ibE_ibE"}, s.err
}


func TestSuggestedFareRejectsMismatchedPlaceCoordinates(t *testing.T) {
	db, code, id, policies := pricedDB(t)
	policy, err := policies.Publish(context.Background(), pricedDraft(code), "owner")
	if err != nil {
		t.Fatal(err)
	}
	spy := &routePreviewSpy{}
	api := pricedAPI(db, id, policies, true, spy)
	api.locationSearch = &pricingPlaces{}
	body := previewBody(code)
	body["pickup"] = map[string]any{"latitude": 25.86, "longitude": 68.01, "place_id": "pickup-poi"}
	body["destination"].(map[string]any)["place_id"] = "destination-poi"
	body["pricing_policy_version"] = policy.ID.String()
	body["proposed_fare"] = map[string]any{"amount_minor": 12500, "currency": "PKR"}
	res := pricedHTTP(api, "/v1/ride-requests", body, true)
	if res.Code != 400 || !strings.Contains(res.Body.String(), "location_selection_changed") || spy.calls != 0 {
		t.Fatalf("mismatched endpoint accepted: %d %s route calls %d", res.Code, res.Body.String(), spy.calls)
	}
	var count int
	if err := db.QueryRow("SELECT count(*) FROM ride_requests WHERE rider_user_id=$1", id).Scan(&count); err != nil || count != 0 {
		t.Fatalf("mismatch wrote request: %d %v", count, err)
	}
}

type pricingPlaces struct { locationsearch.Searcher }

func (s *pricingPlaces) Details(_ context.Context, id, _ string) (locationsearch.Place, error) {
	point := locationsearch.Point{Latitude: 24.86, Longitude: 67.01}
	if id == "destination-poi" {
		point = locationsearch.Point{Latitude: 24.90, Longitude: 67.05}
	}
	return locationsearch.Place{PlaceID: id, Location: point}, nil
}
