package drivertrip

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	gotDriverUserID uuid.UUID
	view            View
	err             error
}

func (f *fakeRepository) GetCurrent(_ context.Context, driverUserID uuid.UUID) (View, error) {
	f.gotDriverUserID = driverUserID
	return f.view, f.err
}

func TestServiceGetCurrentUsesAuthenticatedDriverIdentity(t *testing.T) {
	driverUserID := uuid.New()
	expected := View{RideRequestID: uuid.New()}
	repository := &fakeRepository{view: expected}
	service := NewService(repository)

	result, err := service.GetCurrent(context.Background(), driverUserID)
	if err != nil {
		t.Fatalf("GetCurrent() error = %v", err)
	}
	if repository.gotDriverUserID != driverUserID {
		t.Fatalf("driver user id = %s, want %s", repository.gotDriverUserID, driverUserID)
	}
	if result.RideRequestID != expected.RideRequestID {
		t.Fatalf("ride request id = %s, want %s", result.RideRequestID, expected.RideRequestID)
	}
}
