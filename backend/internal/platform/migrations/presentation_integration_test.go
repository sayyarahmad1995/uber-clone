package migrations

import (
	"database/sql"
	"testing"

	"github.com/google/uuid"
)

func TestPresentationUpgradePreservesLegacyDriver(t *testing.T) {
	db := retirementDB(t, true)
	previous, err := files.ReadFile("sql/016_retire_candidate_schema.sql")
	if err != nil {
		t.Fatal(err)
	}
	execRetirement(t, db, string(previous))
	execRetirement(t, db, `INSERT INTO schema_migrations(version) VALUES ('016_retire_candidate_schema.sql')`)
	id, vehicleID := uuid.New(), uuid.New()
	execRetirement(t, db, `INSERT INTO users(id) VALUES ($1)`, id)
	execRetirement(t, db, `INSERT INTO driver_profiles(user_id,status,is_online) VALUES ($1,'active',true)`, id)
	execRetirement(t, db, `INSERT INTO driver_vehicles(id,driver_user_id,make,model,color,license_plate) VALUES ($1,$2,'Toyota','Corolla','White','ABC-123')`, vehicleID, id)
	if err := Apply(db); err != nil {
		t.Fatal(err)
	}
	if err := Apply(db); err != nil {
		t.Fatalf("repeat migration: %v", err)
	}
	var name sql.NullString
	var year sql.NullInt64
	var online bool
	var make, plate string
	if err := db.QueryRow(`SELECT p.display_name,v.model_year,p.is_online,v.make,v.license_plate FROM driver_profiles p JOIN driver_vehicles v ON v.driver_user_id=p.user_id WHERE p.user_id=$1`, id).Scan(&name, &year, &online, &make, &plate); err != nil {
		t.Fatal(err)
	}
	if name.Valid || year.Valid || !online || make != "Toyota" || plate != "ABC-123" {
		t.Fatal("legacy Driver data changed during upgrade")
	}
	if _, err := db.Exec(`UPDATE driver_profiles SET display_name=' ' WHERE user_id=$1`, id); err == nil {
		t.Fatal("blank name accepted")
	}
	if _, err := db.Exec(`UPDATE driver_vehicles SET model_year=0 WHERE driver_user_id=$1`, id); err == nil {
		t.Fatal("zero year accepted")
	}
	execRetirement(t, db, `UPDATE driver_profiles SET display_name='Driver' WHERE user_id=$1`, id)
	execRetirement(t, db, `UPDATE driver_vehicles SET model_year=2024 WHERE driver_user_id=$1`, id)
}
