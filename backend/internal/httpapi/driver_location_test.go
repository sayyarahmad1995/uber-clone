package httpapi

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sayyarahmad1995/uber-clone/backend/internal/driverlocation"
)

func TestDriverLocationResponseDoesNotExposeDriverIdentity(t *testing.T) {
	response := driverLocationResponse(driverlocation.Location{
		DriverUserID: uuid.New(),
		Latitude:     24.8607,
		Longitude:    67.0011,
		UpdatedAt:    time.Unix(100, 0).UTC(),
	})
	if _, ok := response["driver_user_id"]; ok {
		t.Fatal("driver_user_id must not be exposed")
	}
	if response["latitude"] != 24.8607 || response["longitude"] != 67.0011 {
		t.Fatalf("unexpected coordinates: %#v", response)
	}
}
