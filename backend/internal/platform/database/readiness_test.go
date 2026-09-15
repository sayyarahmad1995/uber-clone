package database

import (
	"context"
	"errors"
	"testing"
)

type fakePinger struct {
	err   error
	calls int
}

func (f *fakePinger) PingContext(context.Context) error {
	f.calls++
	return f.err
}

func TestReadinessCheckerDelegatesToDatabasePing(t *testing.T) {
	pinger := &fakePinger{}
	checker := NewReadinessChecker(pinger)

	if err := checker.Check(context.Background()); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if pinger.calls != 1 {
		t.Fatalf("PingContext() calls = %d, want 1", pinger.calls)
	}
}

func TestReadinessCheckerReturnsDatabasePingError(t *testing.T) {
	want := errors.New("database unavailable")
	pinger := &fakePinger{err: want}
	checker := NewReadinessChecker(pinger)

	if err := checker.Check(context.Background()); !errors.Is(err, want) {
		t.Fatalf("Check() error = %v, want %v", err, want)
	}
}
