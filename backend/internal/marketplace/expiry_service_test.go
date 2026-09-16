package marketplace

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
)

type expiryTestExecer struct {
	statements []string
	failAt     int
}

func (e *expiryTestExecer) ExecContext(_ context.Context, query string, _ ...any) (sql.Result, error) {
	e.statements = append(e.statements, query)
	if e.failAt > 0 && len(e.statements) == e.failAt {
		return nil, errors.New("sweep failed")
	}
	return expiryTestResult(0), nil
}

type expiryTestResult int64

func (r expiryTestResult) LastInsertId() (int64, error) { return 0, nil }
func (r expiryTestResult) RowsAffected() (int64, error) { return int64(r), nil }

func TestExpiryServiceSweepOwnsAllMarketplaceTerminalTransitions(t *testing.T) {
	executor := &expiryTestExecer{}
	service := NewExpiryService(executor)

	if err := service.Sweep(context.Background()); err != nil {
		t.Fatalf("Sweep returned error: %v", err)
	}

	if len(executor.statements) != 6 {
		t.Fatalf("Sweep executed %d statements, want 6", len(executor.statements))
	}

	for _, transition := range []string{
		"SET status = 'expired'",
		"SET status = 'offer_expired'",
		"SET status = 'window_expired'",
		"SET status = 'expired'",
		"SET status = 'closed'",
		"SET status = 'closed'",
	} {
		found := false
		for _, statement := range executor.statements {
			if strings.Contains(statement, transition) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Sweep did not materialize transition containing %q", transition)
		}
	}
}

func TestExpiryServiceSweepStopsOnFailure(t *testing.T) {
	executor := &expiryTestExecer{failAt: 3}
	service := NewExpiryService(executor)

	if err := service.Sweep(context.Background()); err == nil {
		t.Fatal("Sweep returned nil error")
	}
	if len(executor.statements) != 3 {
		t.Fatalf("Sweep executed %d statements after failure, want 3", len(executor.statements))
	}
}
