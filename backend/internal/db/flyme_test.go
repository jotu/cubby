package db

import (
	"database/sql"
	"testing"
)

func TestMigrateUp_TracksFlymeSchemaMigrations(t *testing.T) {
	database, err := OpenWithoutMigrations(t.TempDir() + "/flyme-up.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	if err := MigrateUp(database); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	var count int
	if err := database.QueryRow("SELECT COUNT(1) FROM flyme_schema_migrations").Scan(&count); err != nil {
		t.Fatalf("count flyme schema migrations: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}

	var checksum string
	if err := database.QueryRow("SELECT checksum FROM flyme_schema_migrations WHERE version = 1").Scan(&checksum); err != nil {
		t.Fatalf("read migration checksum: %v", err)
	}
	if checksum != migrationChecksum(legacyMigrationSQL(t)) {
		t.Fatalf("checksum = %q, want checksum for initial migration", checksum)
	}
}

func TestMigrateUp_BackfillsChecksumForExistingFlymeTable(t *testing.T) {
	database, err := OpenWithoutMigrations(t.TempDir() + "/flyme-existing.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	if _, err := database.Exec(`CREATE TABLE flyme_schema_migrations (
		version INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		t.Fatalf("create existing flyme table: %v", err)
	}
	if _, err := database.Exec("INSERT INTO flyme_schema_migrations (version, name) VALUES (1, '001_initial')"); err != nil {
		t.Fatalf("insert existing migration: %v", err)
	}

	if err := MigrateUp(database); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	var checksum string
	if err := database.QueryRow("SELECT checksum FROM flyme_schema_migrations WHERE version = 1").Scan(&checksum); err != nil {
		t.Fatalf("read backfilled checksum: %v", err)
	}
	if checksum != migrationChecksum(legacyMigrationSQL(t)) {
		t.Fatalf("checksum = %q, want checksum for initial migration", checksum)
	}
}

func TestMigrateUp_RejectsChangedAppliedMigration(t *testing.T) {
	database, err := OpenWithoutMigrations(t.TempDir() + "/flyme-checksum.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	if err := MigrateUp(database); err != nil {
		t.Fatalf("migrate up: %v", err)
	}
	if _, err := database.Exec("UPDATE flyme_schema_migrations SET checksum = 'changed' WHERE version = 1"); err != nil {
		t.Fatalf("change migration checksum: %v", err)
	}

	if err := MigrateUp(database); err == nil {
		t.Fatal("expected changed migration checksum to be rejected")
	}
}

func TestValidateAppliedMigrations_RejectsMissingEarlierVersion(t *testing.T) {
	database, err := OpenWithoutMigrations(t.TempDir() + "/flyme-order.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	if err := MigrateUp(database); err != nil {
		t.Fatalf("migrate up: %v", err)
	}
	if _, err := database.Exec("DELETE FROM flyme_schema_migrations WHERE version = 1"); err != nil {
		t.Fatalf("delete first migration: %v", err)
	}
	second := upMigration{version: 2, name: "002_second", sql: "CREATE TABLE second (id INTEGER)"}
	if _, err := database.Exec("INSERT INTO flyme_schema_migrations (version, name, checksum) VALUES (?, ?, ?)", second.version, second.name, migrationChecksum(second.sql)); err != nil {
		t.Fatalf("insert second migration: %v", err)
	}

	if err := validateAppliedMigrations(database, []upMigration{{version: 1, name: "001_initial", sql: legacyMigrationSQL(t)}, second}); err == nil {
		t.Fatal("expected missing earlier migration to be rejected")
	}
}

func TestMigrateUp_RejectsUnknownAppliedMigration(t *testing.T) {
	database, err := OpenWithoutMigrations(t.TempDir() + "/flyme-unknown.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	if err := MigrateUp(database); err != nil {
		t.Fatalf("migrate up: %v", err)
	}
	if _, err := database.Exec(`INSERT INTO flyme_schema_migrations (version, name, checksum) VALUES (99, '099_missing', 'missing')`); err != nil {
		t.Fatalf("insert unknown migration: %v", err)
	}

	if err := MigrateUp(database); err == nil {
		t.Fatal("expected unknown applied migration to be rejected")
	}
}

func TestContainsMigrationSQL(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{name: "sql", content: "-- comment\nDROP TABLE items;\n", want: true},
		{name: "comments only", content: "-- generated\n\n", want: false},
		{name: "empty", content: "\n", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := containsMigrationSQL(test.content); got != test.want {
				t.Fatalf("containsMigrationSQL() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestMigrateDownOne_RevertsLatest(t *testing.T) {
	database, err := OpenWithoutMigrations(t.TempDir() + "/flyme-down.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	if err := MigrateUp(database); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	reverted, err := MigrateDownOne(database)
	if err != nil {
		t.Fatalf("migrate down: %v", err)
	}
	if reverted == nil {
		t.Fatal("expected reverted migration")
	}
	if reverted.Name != "001_initial" {
		t.Fatalf("name = %q, want %q", reverted.Name, "001_initial")
	}

	var exists int
	if err := database.QueryRow("SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='items')").Scan(&exists); err != nil {
		t.Fatalf("check items table existence: %v", err)
	}
	if exists != 0 {
		t.Fatalf("items table should be dropped, exists = %d", exists)
	}
}

func TestMigrateUp_BootstrapsLegacySchemaMigrations(t *testing.T) {
	database, err := OpenWithoutMigrations(t.TempDir() + "/flyme-legacy.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	if _, err := database.Exec(`CREATE TABLE schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		t.Fatalf("create legacy schema migrations table: %v", err)
	}

	legacyMigrations, err := loadUpMigrations(migrationsFS)
	if err != nil {
		t.Fatalf("load migrations: %v", err)
	}
	if _, err := database.Exec(legacyMigrations[0].sql); err != nil {
		t.Fatalf("apply legacy schema: %v", err)
	}

	if _, err := database.Exec("INSERT INTO schema_migrations (version) VALUES (1)"); err != nil {
		t.Fatalf("insert legacy migration row: %v", err)
	}

	if err := MigrateUp(database); err != nil {
		t.Fatalf("migrate up with legacy bootstrap: %v", err)
	}

	var name string
	if err := database.QueryRow("SELECT name FROM flyme_schema_migrations WHERE version = 1").Scan(&name); err != nil {
		t.Fatalf("read bootstrapped migration: %v", err)
	}
	if name != "001_legacy" {
		t.Fatalf("name = %q, want %q", name, "001_legacy")
	}

	if _, err := MigrateDownOne(database); err != nil {
		t.Fatalf("migrate down legacy database: %v", err)
	}
	if err := MigrateUp(database); err != nil {
		t.Fatalf("migrate up after legacy rollback: %v", err)
	}

	var itemsExists int
	if err := database.QueryRow("SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='items')").Scan(&itemsExists); err != nil {
		t.Fatalf("check restored items table: %v", err)
	}
	if itemsExists != 1 {
		t.Fatal("items table was not restored after legacy rollback")
	}
}

func TestFlymeStatus_ShowsAppliedAndPending(t *testing.T) {
	database, err := OpenWithoutMigrations(t.TempDir() + "/flyme-status.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	statuses, err := FlymeStatus(database)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("status count = %d, want 1", len(statuses))
	}
	if statuses[0].Applied {
		t.Fatal("expected migration to be pending before apply")
	}

	if err := MigrateUp(database); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	statuses, err = FlymeStatus(database)
	if err != nil {
		t.Fatalf("status after migrate: %v", err)
	}
	if !statuses[0].Applied {
		t.Fatal("expected migration to be applied after migrate")
	}
	if statuses[0].AppliedAt == nil {
		t.Fatal("expected applied timestamp")
	}
}

func TestLatestAppliedMigration_NoRows(t *testing.T) {
	database, err := OpenWithoutMigrations(t.TempDir() + "/flyme-empty.db")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	if err := ensureFlymeSchemaMigrationsTable(database); err != nil {
		t.Fatalf("ensure migration table: %v", err)
	}

	latest, err := latestAppliedMigration(database)
	if err != nil {
		t.Fatalf("latest applied migration: %v", err)
	}
	if latest != nil {
		t.Fatalf("expected nil latest migration, got %+v", *latest)
	}
}

func TestParseMigrationVersion_Invalid(t *testing.T) {
	tests := []string{"invalid", "0_zero", "-1_negative"}
	for _, name := range tests {
		if _, err := parseMigrationVersion(name); err == nil {
			t.Fatalf("parseMigrationVersion(%q) expected error", name)
		}
	}
}

func legacyMigrationSQL(t *testing.T) string {
	t.Helper()

	migrations, err := loadUpMigrations(migrationsFS)
	if err != nil {
		t.Fatalf("load migrations: %v", err)
	}
	return migrations[0].sql
}

func TestTableExists_False(t *testing.T) {
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	exists, err := tableExists(database, "missing_table")
	if err != nil {
		t.Fatalf("table exists check: %v", err)
	}
	if exists {
		t.Fatal("expected missing table to not exist")
	}
}
