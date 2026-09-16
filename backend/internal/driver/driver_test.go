package driver

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	profile Profile
	online  *bool
}

func (f *fakeRepository) FindByUserID(_ context.Context, _ uuid.UUID) (Profile, error) {
	if f.profile.UserID == uuid.Nil {
		return Profile{}, ErrNotFound
	}
	return f.profile, nil
}

func (f *fakeRepository) ListVehicles(_ context.Context, _ uuid.UUID) ([]Vehicle, error) {
	if f.profile.UserID == uuid.Nil {
		return nil, ErrNotFound
	}
	return []Vehicle{f.profile.Vehicle}, nil
}

func (f *fakeRepository) SetOnline(_ context.Context, userID uuid.UUID, online bool) (Profile, error) {
	if f.profile.UserID == uuid.Nil {
		return Profile{}, ErrNotFound
	}
	f.online = &online
	f.profile.UserID = userID
	f.profile.IsOnline = online
	return f.profile, nil
}

func TestSetOnlineRequiresExistingDriverProfile(t *testing.T) {
	service := NewService(&fakeRepository{})
	_, err := service.SetOnline(context.Background(), uuid.New(), true)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func (f *fakeRepository) OperatingState(context.Context, uuid.UUID) (OperatingState, error) {
	return OperatingState{}, nil
}
func (f *fakeRepository) SelectOperation(context.Context, uuid.UUID, uuid.UUID, string) (OperatingState, error) {
	return OperatingState{}, nil
}
