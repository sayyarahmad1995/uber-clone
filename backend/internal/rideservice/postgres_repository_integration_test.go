package rideservice

import (
    "context"
    "database/sql"
    "errors"
    "net/url"
    "os"
    "strings"
    "testing"

    "github.com/google/uuid"
    "github.com/sayyarahmad1995/uber-clone/backend/internal/platform/database"
    "github.com/sayyarahmad1995/uber-clone/backend/internal/platform/migrations"
    "github.com/sayyarahmad1995/uber-clone/backend/internal/ride"
)

func openCatalogTestDB(t *testing.T) *sql.DB {
    t.Helper()
    databaseURL := os.Getenv("TEST_DATABASE_URL")
    if databaseURL == "" {
        t.Skip("TEST_DATABASE_URL is not set")
    }
    parsed, err := url.Parse(databaseURL)
    if err != nil {
        t.Fatal(err)
    }
    if !strings.HasSuffix(strings.TrimPrefix(parsed.Path, "/"), "_test") {
        t.Fatalf("integration database must end with _test: %s", parsed.Path)
    }
    db, err := database.Open(databaseURL)
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { _ = db.Close() })
    if err := migrations.Apply(db); err != nil {
        t.Fatal(err)
    }
    return db
}

func TestRiderCatalogThirdServiceVisibilityAndRequestGuard(t *testing.T) {
    db := openCatalogTestDB(t)
    code := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
    _, err := db.Exec(`
        INSERT INTO driver_service_catalog
            (code, display_name, description, is_active, sort_order, rider_visible, presentation_token)
        VALUES ($1, 'Third category', 'A test-only car service', TRUE, 15, TRUE, 'unknown-new-icon')
    `, code)
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() {
        _, _ = db.Exec(`DELETE FROM ride_requests WHERE service_code=$1`, code)
        _, _ = db.Exec(`DELETE FROM driver_service_catalog WHERE code=$1`, code)
    })
    userID := uuid.New()
    if _, err := db.Exec(`INSERT INTO users (id) VALUES ($1)`, userID); err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { _, _ = db.Exec(`DELETE FROM users WHERE id=$1`, userID) })

    catalog := NewService(NewPostgresRepository(db))
    list, err := catalog.List(context.Background())
    if err != nil {
        t.Fatal(err)
    }
    found := false
    for _, item := range list {
        if item.Code == code {
            found = item.DisplayName == "Third category" && item.DisplayOrder == 15 && item.PresentationToken == "unknown-new-icon"
        }
    }
    if !found {
        t.Fatal("third service not exposed or presentation metadata lost")
    }

    input := ride.CreateInput{
        ServiceCode: code,
        Pickup: ride.Location{Latitude: 24.86, Longitude: 67.01},
        Destination: ride.Location{Latitude: 24.92, Longitude: 67.08},
        ProposedFare: &ride.Money{AmountMinor: 10000, Currency: "PKR"},
    }
    rides := ride.NewService(ride.NewPostgresRepository(db))
    created, err := rides.Create(context.Background(), userID, input)
    if err != nil || created.ServiceCode != code {
        t.Fatalf("book new service: request=%+v err=%v", created, err)
    }

    if _, err := db.Exec(`UPDATE driver_service_catalog SET display_name='Updated name', sort_order=5 WHERE code=$1`, code); err != nil {
        t.Fatal(err)
    }
    list, err = catalog.List(context.Background())
    if err != nil {
        t.Fatal(err)
    }
    if list[0].Code != code || list[0].DisplayName != "Updated name" {
        t.Fatalf("renamed service did not appear first after reorder: %+v", list)
    }

    if _, err := db.Exec(`UPDATE driver_service_catalog SET rider_visible=FALSE WHERE code=$1`, code); err != nil {
        t.Fatal(err)
    }
    list, err = catalog.List(context.Background())
    if err != nil {
        t.Fatal(err)
    }
    for _, item := range list {
        if item.Code == code {
            t.Fatal("hidden service remained visible")
        }
    }
    if _, err := rides.Create(context.Background(), userID, input); !errors.Is(err, ride.ErrInvalidService) {
        t.Fatalf("expected hidden service to reject new Ride Request, got %v", err)
    }
    var historicalCode string
    if err := db.QueryRow(`SELECT service_code FROM ride_requests WHERE id=$1`, created.ID).Scan(&historicalCode); err != nil || historicalCode != code {
        t.Fatalf("historical service snapshot changed: code=%q err=%v", historicalCode, err)
    }

    if _, err := db.Exec(`UPDATE driver_service_catalog SET rider_visible=TRUE, is_active=FALSE WHERE code=$1`, code); err != nil {
        t.Fatal(err)
    }
    if _, err := rides.Create(context.Background(), userID, input); !errors.Is(err, ride.ErrInvalidService) {
        t.Fatalf("expected inactive service to reject new Ride Request, got %v", err)
    }
    input.ServiceCode = "unknown_service_code"
    if _, err := rides.Create(context.Background(), userID, input); !errors.Is(err, ride.ErrInvalidService) {
        t.Fatalf("expected unknown service to reject new Ride Request, got %v", err)
    }
}
