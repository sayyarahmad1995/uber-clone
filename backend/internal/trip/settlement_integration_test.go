package trip

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestPostgresRepositoryConfirmCashCollectedRequiresCompletedTrip(t *testing.T) {
	db := openTripIntegrationDB(t)
	riderID := createTripIntegrationUser(t, db, "rider")
	driverID := createTripIntegrationDriver(t, db)
	rideID := createTripIntegrationRide(t, db, riderID)
	createAcceptedTripFixture(t, db, rideID, riderID, driverID)

	_, err := NewPostgresRepository(db).ConfirmCashCollected(context.Background(), rideID, driverID)
	if !errors.Is(err, ErrTripNotCompleted) {
		t.Fatalf("expected ErrTripNotCompleted, got %v", err)
	}

	var status string
	var collectedAt sql.NullTime
	if err := db.QueryRow(`SELECT settlement_status, cash_collected_at FROM trips WHERE ride_request_id = $1`, rideID).Scan(&status, &collectedAt); err != nil {
		t.Fatalf("select settlement: %v", err)
	}
	if status != string(SettlementUnsettled) || collectedAt.Valid {
		t.Fatalf("settlement mutated before completion: status=%s collected=%v", status, collectedAt.Valid)
	}
}

func TestPostgresRepositoryConfirmCashCollectedRecordsCashSettlement(t *testing.T) {
	db := openTripIntegrationDB(t)
	riderID := createTripIntegrationUser(t, db, "rider")
	driverID := createTripIntegrationDriver(t, db)
	rideID := createTripIntegrationRide(t, db, riderID)
	createAcceptedTripFixture(t, db, rideID, riderID, driverID)

	repository := NewPostgresRepository(db)
	if _, err := repository.Start(context.Background(), rideID, driverID); err != nil {
		t.Fatalf("start: %v", err)
	}
	if _, err := repository.Complete(context.Background(), rideID, driverID); err != nil {
		t.Fatalf("complete: %v", err)
	}

	first, err := repository.ConfirmCashCollected(context.Background(), rideID, driverID)
	if err != nil {
		t.Fatalf("first cash confirmation: %v", err)
	}
	if first.Settlement.Status != SettlementCashCollected {
		t.Fatalf("expected cash_collected, got %s", first.Settlement.Status)
	}
	if first.Settlement.Method == nil || *first.Settlement.Method != "cash" {
		t.Fatalf("expected cash method, got %v", first.Settlement.Method)
	}
	if first.Settlement.CashCollectedAt == nil {
		t.Fatal("expected cash_collected_at")
	}

	second, err := repository.ConfirmCashCollected(context.Background(), rideID, driverID)
	if err != nil {
		t.Fatalf("second cash confirmation: %v", err)
	}
	if second.Settlement.CashCollectedAt == nil || !second.Settlement.CashCollectedAt.Equal(*first.Settlement.CashCollectedAt) {
		t.Fatalf("expected stable cash_collected_at, first=%v second=%v", first.Settlement.CashCollectedAt, second.Settlement.CashCollectedAt)
	}

	var collectedBy uuid.NullUUID
	if err := db.QueryRow(`SELECT cash_collected_by FROM trips WHERE ride_request_id = $1`, rideID).Scan(&collectedBy); err != nil {
		t.Fatalf("select cash_collected_by: %v", err)
	}
	if !collectedBy.Valid || collectedBy.UUID != driverID {
		t.Fatalf("expected cash collection by assigned driver, got %v", collectedBy)
	}
}

func TestPostgresRepositoryConfirmCashCollectedRejectsOtherDriver(t *testing.T) {
	db := openTripIntegrationDB(t)
	riderID := createTripIntegrationUser(t, db, "rider")
	driverID := createTripIntegrationDriver(t, db)
	otherDriverID := createTripIntegrationDriver(t, db)
	rideID := createTripIntegrationRide(t, db, riderID)
	createAcceptedTripFixture(t, db, rideID, riderID, driverID)

	repository := NewPostgresRepository(db)
	if _, err := repository.Start(context.Background(), rideID, driverID); err != nil {
		t.Fatalf("start: %v", err)
	}
	if _, err := repository.Complete(context.Background(), rideID, driverID); err != nil {
		t.Fatalf("complete: %v", err)
	}

	_, err := repository.ConfirmCashCollected(context.Background(), rideID, otherDriverID)
	if !errors.Is(err, ErrTripNotFound) {
		t.Fatalf("expected ErrTripNotFound for non-assigned driver, got %v", err)
	}
}
