package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/drivertrip"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/identity"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/routing"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/trip"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/user"
)

func TestDriverRouteUsesAuthenticatedTripEndpoints(t *testing.T) {
	for _, status := range []trip.Status{trip.StatusAssigned, trip.StatusInProgress} {
		t.Run(string(status), func(t *testing.T) {
			driverID := uuid.New()
			view := drivertrip.View{
				RideRequestID: uuid.New(),
				Pickup:        ride.Location{Latitude: 24.86, Longitude: 67.01},
				Destination:   ride.Location{Latitude: 24.91, Longitude: 67.08},
				Status:        status,
			}
			trips := &routeTripRepository{driverID: driverID, view: view}
			provider := &routePreviewSpy{}
			api := routeTestAPI(driverID, true, trips, provider)
			body := map[string]any{
				"ride_request_id": view.RideRequestID.String(),
				"status":          string(status),
				"origin":          map[string]any{"latitude": 24.80, "longitude": 67.00},
				// Client-supplied trip endpoints must not be trusted.
				"pickup":      map[string]any{"latitude": 1, "longitude": 2},
				"destination": map[string]any{"latitude": 3, "longitude": 4},
			}
			response := routeRequest(api, body, true)
			if response.Code != http.StatusOK {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			wantFrom := routing.Point{Latitude: 24.80, Longitude: 67.00}
			wantTo := routing.Point{Latitude: 24.86, Longitude: 67.01}
			if status == trip.StatusInProgress {
				wantFrom = routing.Point{Latitude: 24.86, Longitude: 67.01}
				wantTo = routing.Point{Latitude: 24.91, Longitude: 67.08}
			}
			if provider.calls != 1 || provider.from.Point != wantFrom || provider.to.Point != wantTo {
				t.Fatalf("routing input: calls=%d from=%+v to=%+v", provider.calls, provider.from, provider.to)
			}
			var payload map[string]json.RawMessage
			if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
				t.Fatal(err)
			}
			var route struct {
				Polyline string `json:"encoded_polyline"`
			}
			if err := json.Unmarshal(payload["route"], &route); err != nil || route.Polyline != "_p~iF~ps|U_ulLnnqC" {
				t.Fatalf("missing driving polyline: %s", response.Body.String())
			}
		})
	}
}

func TestDriverRouteRejectsUnauthorizedStaleAndInvalidRequests(t *testing.T) {
	cases := []struct {
		name          string
		authorized    bool
		driver        bool
		status        trip.Status
		missingTrip   bool
		wrongID       bool
		wrongStatus   bool
		origin        map[string]any
		providerError error
		want          int
	}{
		{name: "unsigned", driver: true, status: trip.StatusAssigned, want: 401},
		{name: "rider only", authorized: true, status: trip.StatusAssigned, want: 403},
		{name: "no assigned trip", authorized: true, driver: true, missingTrip: true, want: 404},
		{name: "different trip", authorized: true, driver: true, status: trip.StatusAssigned, wrongID: true, want: 409},
		{name: "changed stage", authorized: true, driver: true, status: trip.StatusInProgress, wrongStatus: true, want: 409},
		{name: "completed", authorized: true, driver: true, status: trip.StatusCompleted, want: 409},
		{name: "cancelled", authorized: true, driver: true, status: trip.StatusCancelled, want: 409},
		{name: "missing GPS", authorized: true, driver: true, status: trip.StatusAssigned, want: 400},
		{name: "invalid GPS", authorized: true, driver: true, status: trip.StatusAssigned, origin: map[string]any{"latitude": 91, "longitude": 0}, want: 400},
		{name: "provider unavailable", authorized: true, driver: true, status: trip.StatusInProgress, providerError: routing.ErrUnavailable, want: 503},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			driverID := uuid.New()
			view := drivertrip.View{
				RideRequestID: uuid.New(), Status: tc.status,
				Pickup:      ride.Location{Latitude: 24.86, Longitude: 67.01},
				Destination: ride.Location{Latitude: 24.91, Longitude: 67.08},
			}
			trips := &routeTripRepository{driverID: driverID, view: view}
			if tc.missingTrip {
				trips.err = drivertrip.ErrNotFound
			}
			provider := &routePreviewSpy{err: tc.providerError}
			api := routeTestAPI(driverID, tc.driver, trips, provider)
			id := view.RideRequestID
			if tc.wrongID {
				id = uuid.New()
			}
			status := string(tc.status)
			if tc.wrongStatus {
				status = "assigned"
			}
			body := map[string]any{"ride_request_id": id.String(), "status": status}
			if tc.origin != nil {
				body["origin"] = tc.origin
			}
			response := routeRequest(api, body, tc.authorized)
			if response.Code != tc.want {
				t.Fatalf("status=%d want=%d body=%s", response.Code, tc.want, response.Body.String())
			}
			if tc.providerError == nil && provider.calls != 0 {
				t.Fatal("rejected request reached the route provider")
			}
		})
	}
}

func routeRequest(api *API, body map[string]any, authorized bool) *httptest.ResponseRecorder {
	data, _ := json.Marshal(body)
	request := httptest.NewRequest(http.MethodPost, "/v1/driver/trip/route-preview", strings.NewReader(string(data)))
	if authorized {
		request.Header.Set("Authorization", strings.Join([]string{"Bearer", "route-test"}, " "))
	}
	response := httptest.NewRecorder()
	api.Handler().ServeHTTP(response, request)
	return response
}

func routeTestAPI(id uuid.UUID, driver bool, trips *routeTripRepository, provider *routePreviewSpy) *API {
	capabilities := []user.Capability{user.CapabilityRider}
	if driver {
		capabilities = append(capabilities, user.CapabilityDriver)
	}
	return New(Dependencies{
		Identity:    routeIdentity{},
		Users:       user.NewService(routeUserRepository{value: user.User{ID: id, Capabilities: capabilities}}),
		DriverTrips: drivertrip.NewService(trips),
		Routing:     routing.NewService(provider),
	})
}

type routeIdentity struct{}

func (routeIdentity) AuthenticateVerified(context.Context, string) (identity.Principal, error) {
	return identity.Principal{Issuer: "route-test", Subject: "driver"}, nil
}

type routeUserRepository struct{ value user.User }

func (r routeUserRepository) CreateWithDefaultRider(context.Context, user.ExternalIdentity) (user.User, error) {
	return r.value, nil
}
func (r routeUserRepository) AddCapability(context.Context, uuid.UUID, user.Capability) (user.User, error) {
	return r.value, nil
}

type routeTripRepository struct {
	driverID uuid.UUID
	view     drivertrip.View
	err      error
}

func (r *routeTripRepository) GetCurrent(_ context.Context, id uuid.UUID) (drivertrip.View, error) {
	if id != r.driverID {
		return drivertrip.View{}, errors.New("wrong driver identity")
	}
	return r.view, r.err
}
func (*routeTripRepository) ListHistory(context.Context, uuid.UUID, int) ([]drivertrip.View, error) {
	return nil, nil
}

type routePreviewSpy struct {
	calls int
	from  routing.Endpoint
	to    routing.Endpoint
	err   error
}

func (s *routePreviewSpy) Preview(_ context.Context, from, to routing.Endpoint) (routing.Route, error) {
	s.calls++
	s.from, s.to = from, to
	return routing.Route{DistanceMeters: 2000, DurationSeconds: 300, EncodedPolyline: "_p~iF~ps|U_ulLnnqC"}, s.err
}
