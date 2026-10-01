package httpapi

import (
	"errors"
	"testing"
	"time"

	"github.com/sayyarahmad1995/uber-clone/backend/internal/locationsearch"
)

func TestPlaceResponseKeepsReadableLabelAndCoordinates(t *testing.T) {
	response := placeResponse(locationsearch.Place{
		PlaceID: "p1",
		Label:   "Test Road",
		Location: locationsearch.Point{
			Latitude:  24.86,
			Longitude: 67.01,
		},
	})
	if response["place_id"] != "p1" || response["label"] != "Test Road" {
		t.Fatalf("unexpected place response: %#v", response)
	}
	if response["latitude"] != 24.86 || response["longitude"] != 67.01 {
		t.Fatalf("unexpected coordinates: %#v", response)
	}
}

func TestLocationSearchErrorStatusContract(t *testing.T) {
	cases := []struct {
		err    error
		status int
	}{
		{locationsearch.ErrInvalidInput, 400},
		{locationsearch.ErrNotFound, 404},
		{locationsearch.ErrUnavailable, 503},
		{locationsearch.ErrProvider, 502},
	}
	for _, tc := range cases {
		status := locationSearchStatus(tc.err)
		if status != tc.status {
			t.Fatalf("error %v: expected %d, got %d", tc.err, tc.status, status)
		}
	}
	if locationSearchStatus(errors.New("other")) != 502 {
		t.Fatal("unknown provider errors must map to 502")
	}
}

func TestUserWindowLimiterResetsAfterWindow(t *testing.T) {
	limiter := newUserWindowLimiter(2)
	now := time.Unix(100, 0)
	if !limiter.allow("user", now) || !limiter.allow("user", now) || limiter.allow("user", now) {
		t.Fatal("expected two allowed calls then a rate-limit rejection")
	}
	if !limiter.allow("user", now.Add(time.Minute)) {
		t.Fatal("expected limiter to reset after one minute")
	}
}
