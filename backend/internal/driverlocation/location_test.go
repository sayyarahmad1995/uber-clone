package driverlocation

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	driverUserID uuid.UUID
	input        Input
	location     Location
	err          error
}

func (f *fakeRepository) UpsertCurrent(_ context.Context, driverUserID uuid.UUID, input Input) (Location, error) {
	f.driverUserID = driverUserID
	f.input = input
	return f.location, f.err
}

func TestSetCurrentUsesAuthenticatedDriverIdentity(t *testing.T) {
	driverUserID := uuid.New()
	repository := &fakeRepository{location: Location{DriverUserID: driverUserID, Latitude: 24.8607, Longitude: 67.0011}}
	service := NewService(repository)

	_, err := service.SetCurrent(context.Background(), driverUserID, Input{Latitude: 24.8607, Longitude: 67.0011})
	if err != nil {
		t.Fatalf("SetCurrent() error = %v", err)
	}
	if repository.driverUserID != driverUserID {
		t.Fatalf("driver user id = %s, want %s", repository.driverUserID, driverUserID)
	}
}

func TestSetCurrentRejectsOutOfRangeCoordinates(t *testing.T) {
	service := NewService(&fakeRepository{})
	tests := []Input{
		{Latitude: 90.0001, Longitude: 0},
		{Latitude: -90.0001, Longitude: 0},
		{Latitude: 0, Longitude: 180.0001},
		{Latitude: 0, Longitude: -180.0001},
	}
	for _, input := range tests {
		if _, err := service.SetCurrent(context.Background(), uuid.New(), input); err != ErrInvalidLocation {
			t.Fatalf("SetCurrent(%+v) error = %v, want %v", input, err, ErrInvalidLocation)
		}
	}
}

func TestSetCurrentAcceptsZeroCoordinates(t *testing.T) {
	service := NewService(&fakeRepository{location: Location{Latitude: 0, Longitude: 0}})
	if _, err := service.SetCurrent(context.Background(), uuid.New(), Input{Latitude: 0, Longitude: 0}); err != nil {
		t.Fatalf("SetCurrent() error = %v", err)
	}
}
