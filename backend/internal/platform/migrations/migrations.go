package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"sort"
)

//go:embed sql/*.sql
var files embed.FS

func Apply(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW())`); err != nil { return err }
	entries, err := files.ReadDir("sql")
	if err != nil { return err }
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	for _, entry := range entries {
		if entry.IsDir() { continue }
		var applied bool
		if err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)", entry.Name()).Scan(&applied); err != nil { return err }
		if applied { continue }
		sqlBytes, err := files.ReadFile("sql/" + entry.Name())
		if err != nil { return err }
		tx, err := db.Begin()
		if err != nil { return err }
		if _, err := tx.Exec(string(sqlBytes)); err != nil { tx.Rollback(); return fmt.Errorf("apply %s: %w", entry.Name(), err) }
		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", entry.Name()); err != nil { tx.Rollback(); return err }
		if err := tx.Commit(); err != nil { return err }
	}
	return nil
}
