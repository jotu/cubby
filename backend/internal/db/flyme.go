package db

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const flymeSchemaTable = "flyme_schema_migrations"

type MigrationStatus struct {
	Version   int64
	Name      string
	Checksum  string
	Applied   bool
	AppliedAt *time.Time
}

type upMigration struct {
	version int64
	name    string
	sql     string
}

type appliedMigration struct {
	Name      string
	Checksum  string
	AppliedAt *time.Time
}

func MigrateUp(db *sql.DB) error {
	return runFlymeMigrations(db, migrationsFS)
}

func ValidateMigrations() error {
	migrations, err := loadUpMigrations(migrationsFS)
	if err != nil {
		return err
	}

	return validateMigrationPairs(migrationsFS, migrations)
}

func MigrateDownOne(db *sql.DB) (*MigrationStatus, error) {
	if err := ensureFlymeSchemaMigrationsTable(db); err != nil {
		return nil, err
	}

	migrations, err := loadUpMigrations(migrationsFS)
	if err != nil {
		return nil, err
	}
	if err := validateAppliedMigrations(db, migrations); err != nil {
		return nil, err
	}
	if err := validateMigrationPairs(migrationsFS, migrations); err != nil {
		return nil, err
	}

	latest, err := latestAppliedMigration(db)
	if err != nil {
		return nil, err
	}
	if latest == nil {
		return nil, nil
	}

	downBase := migrationNameByVersion(migrations, latest.Version)
	if downBase == "" {
		downBase = latest.Name
	}

	downPath := fmt.Sprintf("migrations/%s.down.sql", downBase)
	content, err := fs.ReadFile(migrationsFS, downPath)
	if err != nil {
		return nil, fmt.Errorf("read down migration %s: %w", downPath, err)
	}
	if !containsMigrationSQL(string(content)) {
		return nil, fmt.Errorf("down migration %s is empty", downPath)
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin tx for down migration %s: %w", latest.Name, err)
	}

	if _, err := tx.Exec(string(content)); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("execute down migration %s: %w", latest.Name, err)
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE version = ?", flymeSchemaTable)
	if _, err := tx.Exec(query, latest.Version); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("delete migration record %s: %w", latest.Name, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit down migration %s: %w", latest.Name, err)
	}

	latest.Applied = false
	latest.AppliedAt = nil
	latest.Name = downBase
	return latest, nil
}

func FlymeStatus(db *sql.DB) ([]MigrationStatus, error) {
	if err := ensureFlymeSchemaMigrationsTable(db); err != nil {
		return nil, err
	}

	migrations, err := loadUpMigrations(migrationsFS)
	if err != nil {
		return nil, err
	}

	if err := validateAppliedMigrations(db, migrations); err != nil {
		return nil, err
	}
	applied, err := loadAppliedFlymeVersions(db)
	if err != nil {
		return nil, err
	}

	statuses := make([]MigrationStatus, 0, len(migrations))
	for _, migration := range migrations {
		status := MigrationStatus{
			Version: migration.version,
			Name:    migration.name,
		}

		if appliedMigration, ok := applied[migration.version]; ok {
			status.Applied = true
			status.Checksum = appliedMigration.Checksum
			status.AppliedAt = appliedMigration.AppliedAt
			if appliedMigration.Name != "" {
				status.Name = appliedMigration.Name
			}
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

func runFlymeMigrations(db *sql.DB, migrationFS fs.FS) error {
	if err := ensureFlymeSchemaMigrationsTable(db); err != nil {
		return err
	}

	migrations, err := loadUpMigrations(migrationFS)
	if err != nil {
		return err
	}
	if err := validateAppliedMigrations(db, migrations); err != nil {
		return err
	}
	applied, err := loadAppliedFlymeVersions(db)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if _, ok := applied[migration.version]; ok {
			continue
		}

		if err := applyUpMigration(db, migration.version, migration.name, migration.sql); err != nil {
			return err
		}
	}

	return nil
}

func loadUpMigrations(migrationFS fs.FS) ([]upMigration, error) {
	entries, err := fs.ReadDir(migrationFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("read migrations directory: %w", err)
	}

	migrations := make([]upMigration, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		base := strings.TrimSuffix(name, ".up.sql")
		version, err := parseMigrationVersion(base)
		if err != nil {
			return nil, fmt.Errorf("parse migration version from %s: %w", name, err)
		}

		path := filepath.Join("migrations", name)
		content, err := fs.ReadFile(migrationFS, path)
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", name, err)
		}

		migrations = append(migrations, upMigration{
			version: version,
			name:    base,
			sql:     string(content),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	for i := 1; i < len(migrations); i++ {
		if migrations[i].version == migrations[i-1].version {
			return nil, fmt.Errorf("duplicate migration version: %d", migrations[i].version)
		}
	}

	return migrations, nil
}

func parseMigrationVersion(migrationName string) (int64, error) {
	parts := strings.SplitN(migrationName, "_", 2)
	if len(parts) != 2 || parts[0] == "" {
		return 0, fmt.Errorf("invalid migration filename: %s", migrationName)
	}

	version, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid migration version %q: %w", parts[0], err)
	}
	if version < 1 {
		return 0, fmt.Errorf("migration version must be positive: %d", version)
	}

	return version, nil
}

func ensureFlymeSchemaMigrationsTable(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin flyme schema setup: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	existed, err := tableExistsTx(tx, flymeSchemaTable)
	if err != nil {
		return err
	}

	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		version INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		checksum TEXT NOT NULL DEFAULT '',
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`, flymeSchemaTable)
	if _, err := tx.Exec(query); err != nil {
		return fmt.Errorf("create flyme schema migrations table: %w", err)
	}

	columns, err := migrationTableColumns(tx)
	if err != nil {
		return err
	}
	if !columns["checksum"] {
		if _, err := tx.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN checksum TEXT NOT NULL DEFAULT ''", flymeSchemaTable)); err != nil {
			return fmt.Errorf("add migration checksum column: %w", err)
		}
	}

	if !existed {
		if err := bootstrapFromLegacySchemaMigrations(tx); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit flyme schema setup: %w", err)
	}
	committed = true
	return nil
}

func loadAppliedFlymeVersions(db *sql.DB) (map[int64]appliedMigration, error) {
	query := fmt.Sprintf("SELECT version, name, checksum, applied_at FROM %s", flymeSchemaTable)

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query applied flyme migrations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	applied := make(map[int64]appliedMigration)
	for rows.Next() {
		var version int64
		var name string
		var checksum string
		var appliedAtRaw string
		if err := rows.Scan(&version, &name, &checksum, &appliedAtRaw); err != nil {
			return nil, fmt.Errorf("scan applied flyme migration: %w", err)
		}

		appliedAt := parseSQLiteTime(appliedAtRaw)
		applied[version] = appliedMigration{Name: name, Checksum: checksum, AppliedAt: appliedAt}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied flyme migrations: %w", err)
	}

	return applied, nil
}

func validateAppliedMigrations(db *sql.DB, migrations []upMigration) error {
	applied, err := loadAppliedFlymeVersions(db)
	if err != nil {
		return err
	}

	byVersion := make(map[int64]upMigration, len(migrations))
	for _, migration := range migrations {
		byVersion[migration.version] = migration
	}

	var highestApplied int64
	hasApplied := false
	type checksumUpdate struct {
		version  int64
		checksum string
	}
	updates := make([]checksumUpdate, 0)

	for version, appliedMigration := range applied {
		migration, ok := byVersion[version]
		if !ok {
			return fmt.Errorf("applied migration version %d has no migration file", version)
		}

		if !hasApplied || version > highestApplied {
			highestApplied = version
			hasApplied = true
		}

		expected := migrationChecksum(migration.sql)
		if appliedMigration.Checksum == "" {
			updates = append(updates, checksumUpdate{version: version, checksum: expected})
			continue
		}
		if appliedMigration.Checksum != expected {
			return fmt.Errorf("migration %s checksum mismatch", migration.name)
		}
	}

	if hasApplied {
		for _, migration := range migrations {
			if migration.version <= highestApplied {
				if _, ok := applied[migration.version]; !ok {
					return fmt.Errorf("migration version %d is missing before applied version %d", migration.version, highestApplied)
				}
			}
		}
	}

	query := fmt.Sprintf("UPDATE %s SET checksum = ? WHERE version = ?", flymeSchemaTable)
	for _, update := range updates {
		if _, err := db.Exec(query, update.checksum, update.version); err != nil {
			return fmt.Errorf("record checksum for migration version %d: %w", update.version, err)
		}
	}

	return nil
}

func validateMigrationPairs(migrationFS fs.FS, migrations []upMigration) error {
	for _, migration := range migrations {
		path := fmt.Sprintf("migrations/%s.down.sql", migration.name)
		content, err := fs.ReadFile(migrationFS, path)
		if err != nil {
			return fmt.Errorf("read down migration %s: %w", path, err)
		}
		if !containsMigrationSQL(string(content)) {
			return fmt.Errorf("down migration %s is empty", path)
		}
	}

	return nil
}

func containsMigrationSQL(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "--") {
			return true
		}
	}

	return false
}

func migrationChecksum(content string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
}

func latestAppliedMigration(db *sql.DB) (*MigrationStatus, error) {
	query := fmt.Sprintf("SELECT version, name, checksum, applied_at FROM %s ORDER BY version DESC LIMIT 1", flymeSchemaTable)

	row := db.QueryRow(query)
	var version int64
	var name string
	var checksum string
	var appliedAtRaw string
	if err := row.Scan(&version, &name, &checksum, &appliedAtRaw); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get latest applied migration: %w", err)
	}

	appliedAt := parseSQLiteTime(appliedAtRaw)
	return &MigrationStatus{Version: version, Name: name, Checksum: checksum, Applied: true, AppliedAt: appliedAt}, nil
}

func bootstrapFromLegacySchemaMigrations(tx *sql.Tx) error {
	exists, err := tableExistsTx(tx, "schema_migrations")
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	rows, err := tx.Query("SELECT version, applied_at FROM schema_migrations ORDER BY version")
	if err != nil {
		return fmt.Errorf("load legacy schema migrations: %w", err)
	}

	type legacyMigration struct {
		version   int64
		appliedAt string
	}
	legacyMigrations := make([]legacyMigration, 0)
	for rows.Next() {
		var migration legacyMigration
		if err := rows.Scan(&migration.version, &migration.appliedAt); err != nil {
			_ = rows.Close()
			return fmt.Errorf("scan legacy schema migration: %w", err)
		}
		legacyMigrations = append(legacyMigrations, migration)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("close legacy schema migrations: %w", err)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate legacy schema migrations: %w", err)
	}

	insertQuery := fmt.Sprintf("INSERT INTO %s (version, name, checksum, applied_at) VALUES (?, ?, ?, ?)", flymeSchemaTable)
	for _, migration := range legacyMigrations {
		name := migrationNameForVersion(migration.version)
		if _, err := tx.Exec(insertQuery, migration.version, name, "", migration.appliedAt); err != nil {
			return fmt.Errorf("insert flyme schema migration for version %d: %w", migration.version, err)
		}
	}

	return nil
}

func tableExists(db *sql.DB, name string) (bool, error) {
	var exists int
	query := "SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name = ?)"
	if err := db.QueryRow(query, name).Scan(&exists); err != nil {
		return false, fmt.Errorf("check table %s existence: %w", name, err)
	}

	return exists == 1, nil
}

func tableExistsTx(tx *sql.Tx, name string) (bool, error) {
	var exists int
	query := "SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name = ?)"
	if err := tx.QueryRow(query, name).Scan(&exists); err != nil {
		return false, fmt.Errorf("check table %s existence: %w", name, err)
	}

	return exists == 1, nil
}

func migrationTableColumns(tx *sql.Tx) (map[string]bool, error) {
	rows, err := tx.Query(fmt.Sprintf("PRAGMA table_info(%s)", flymeSchemaTable))
	if err != nil {
		return nil, fmt.Errorf("inspect flyme schema migrations table: %w", err)
	}
	defer func() { _ = rows.Close() }()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, columnType string
		var notNull, primaryKey int
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return nil, fmt.Errorf("scan flyme schema migration column: %w", err)
		}
		columns[name] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate flyme schema migration columns: %w", err)
	}

	return columns, nil
}

func migrationNameForVersion(version int64) string {
	return fmt.Sprintf("%03d_legacy", version)
}

func migrationNameByVersion(migrations []upMigration, version int64) string {
	for _, migration := range migrations {
		if migration.version == version {
			return migration.name
		}
	}

	return ""
}

func parseSQLiteTime(raw string) *time.Time {
	layouts := []string{
		time.RFC3339Nano,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04:05.999999999-07:00",
	}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, raw)
		if err == nil {
			return &parsed
		}
	}

	return nil
}

func applyUpMigration(db *sql.DB, version int64, name string, sqlContent string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx for migration %s: %w", name, err)
	}

	if _, err := tx.Exec(sqlContent); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("execute migration %s: %w", name, err)
	}

	query := fmt.Sprintf("INSERT INTO %s (version, name, checksum) VALUES (?, ?, ?)", flymeSchemaTable)
	if _, err := tx.Exec(query, version, name, migrationChecksum(sqlContent)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("record migration %s: %w", name, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", name, err)
	}

	return nil
}
