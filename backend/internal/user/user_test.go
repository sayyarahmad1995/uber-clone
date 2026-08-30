package user

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type fakeRepository struct {
	user             User
	addedCapability  Capability
	createCallCount  int
	addCallCount     int
}

func (f *fakeRepository) CreateWithDefaultRider(context.Context, ExternalIdentity) (User, error) {
	f.createCallCount++
	return f.user, nil
}

func (f *fakeRepository) AddCapability(_ context.Context, userID uuid.UUID, capability Capability) (User, error) {
	f.addCallCount++
	f.addedCapability = capability
	result := f.user
	result.ID = userID
	result.Capabilities = append(append([]Capability{}, result.Capabilities...), capability)
	return result, nil
}

func TestEnableCapabilityAddsDriverToExistingUser(t *testing.T) {
	userID := uuid.New()
	repository := &fakeRepository{user: User{ID: userID, Capabilities: []Capability{CapabilityRider}}}
	service := NewService(repository)

	result, err := service.EnableCapability(context.Background(), ExternalIdentity{Issuer: "primary-identity-v1", Subject: "subject-1"}, CapabilityDriver)
	if err != nil {
		t.Fatalf("EnableCapability returned error: %v", err)
	}
	if repository.createCallCount != 1 {
		t.Fatalf("expected one user lookup/create, got %d", repository.createCallCount)
	}
	if repository.addCallCount != 1 {
		t.Fatalf("expected one capability add, got %d", repository.addCallCount)
	}
	if repository.addedCapability != CapabilityDriver {
		t.Fatalf("expected driver capability, got %q", repository.addedCapability)
	}
	if result.ID != userID {
		t.Fatalf("expected user ID %s, got %s", userID, result.ID)
	}
	if len(result.Capabilities) != 2 || result.Capabilities[0] != CapabilityRider || result.Capabilities[1] != CapabilityDriver {
		t.Fatalf("unexpected capabilities: %#v", result.Capabilities)
	}
}
