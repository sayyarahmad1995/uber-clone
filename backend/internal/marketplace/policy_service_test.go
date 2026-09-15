package marketplace

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeTimingPolicyStore struct {
	loaded      TimingPolicy
	loadErr     error
	updated     TimingPolicy
	updateErr   error
	updateCalls int
	updateActor string
	updateInput TimingPolicy
}

func (f *fakeTimingPolicyStore) Load(context.Context) (TimingPolicy, error) {
	return f.loaded, f.loadErr
}

func (f *fakeTimingPolicyStore) Update(_ context.Context, policy TimingPolicy, actor string) (TimingPolicy, error) {
	f.updateCalls++
	f.updateInput = policy
	f.updateActor = actor
	return f.updated, f.updateErr
}

func validPolicyForServiceTest() TimingPolicy {
	return TimingPolicy{
		RideRequestTTLSeconds:       180,
		DriverOpportunityTTLSeconds: 30,
		OfferDecisionTTLSeconds:     10,
	}
}

func TestPolicyServiceLoadReturnsStorePolicy(t *testing.T) {
	want := validPolicyForServiceTest()
	want.UpdatedAt = time.Date(2026, 9, 15, 9, 0, 0, 0, time.UTC)
	want.UpdatedBy = "reviewer"
	store := &fakeTimingPolicyStore{loaded: want}

	got, err := NewPolicyService(store).Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got != want {
		t.Fatalf("Load() = %#v, want %#v", got, want)
	}
}

func TestPolicyServiceRejectsBlankActorWithoutStoreUpdate(t *testing.T) {
	store := &fakeTimingPolicyStore{}

	_, err := NewPolicyService(store).Update(context.Background(), validPolicyForServiceTest(), "   ")
	if !errors.Is(err, ErrInvalidTimingPolicy) {
		t.Fatalf("Update() error = %v, want ErrInvalidTimingPolicy", err)
	}
	if store.updateCalls != 0 {
		t.Fatalf("store Update() calls = %d, want 0", store.updateCalls)
	}
}

func TestPolicyServiceRejectsOutOfRangePolicyWithoutStoreUpdate(t *testing.T) {
	store := &fakeTimingPolicyStore{}
	invalid := validPolicyForServiceTest()
	invalid.OfferDecisionTTLSeconds = MinOfferDecisionTTLSeconds - 1

	_, err := NewPolicyService(store).Update(context.Background(), invalid, "reviewer")
	if !errors.Is(err, ErrInvalidTimingPolicy) {
		t.Fatalf("Update() error = %v, want ErrInvalidTimingPolicy", err)
	}
	if store.updateCalls != 0 {
		t.Fatalf("store Update() calls = %d, want 0", store.updateCalls)
	}
}

func TestPolicyServiceUpdatesStoreWithTrimmedActor(t *testing.T) {
	input := validPolicyForServiceTest()
	want := input
	want.UpdatedAt = time.Date(2026, 9, 15, 9, 1, 0, 0, time.UTC)
	want.UpdatedBy = "reviewer"
	store := &fakeTimingPolicyStore{updated: want}

	got, err := NewPolicyService(store).Update(context.Background(), input, "  reviewer  ")
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if got != want {
		t.Fatalf("Update() = %#v, want %#v", got, want)
	}
	if store.updateCalls != 1 {
		t.Fatalf("store Update() calls = %d, want 1", store.updateCalls)
	}
	if store.updateActor != "reviewer" {
		t.Fatalf("store actor = %q, want reviewer", store.updateActor)
	}
	if store.updateInput != input {
		t.Fatalf("store policy = %#v, want %#v", store.updateInput, input)
	}
}
