package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/joacim/cubby/internal/db"
)

var migrationDescriptionPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

const migrationDirPath = "internal/db/migrations"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "create":
		if len(os.Args) < 3 {
			log.Fatalf("flyme create: migration name is required")
		}

		if err := runCreate(os.Args[2]); err != nil {
			log.Fatalf("flyme create: %v", err)
		}
	case "up", "status", "down", "redo":
		database, closeDatabase, err := openDatabaseForFlyme()
		if err != nil {
			log.Fatalf("open database: %v", err)
		}
		defer closeDatabase()

		switch command {
		case "up":
			if err := db.MigrateUp(database); err != nil {
				log.Fatalf("flyme up: %v", err)
			}
			fmt.Println("flyme: migrations applied")
		case "status":
			statuses, err := db.FlymeStatus(database)
			if err != nil {
				log.Fatalf("flyme status: %v", err)
			}
			printStatus(statuses)
		case "down":
			downFlags := flag.NewFlagSet("down", flag.ExitOnError)
			steps := downFlags.Int("steps", 1, "number of migrations to roll back")
			_ = downFlags.Parse(os.Args[2:])

			if *steps < 1 {
				log.Fatalf("flyme down: steps must be >= 1")
			}

			for i := 0; i < *steps; i++ {
				reverted, err := db.MigrateDownOne(database)
				if err != nil {
					log.Fatalf("flyme down: %v", err)
				}
				if reverted == nil {
					fmt.Println("flyme: no migrations to roll back")
					return
				}
				fmt.Printf("flyme: rolled back %s\n", reverted.Name)
			}
		case "redo":
			reverted, err := db.MigrateDownOne(database)
			if err != nil {
				log.Fatalf("flyme redo (down): %v", err)
			}
			if reverted == nil {
				fmt.Println("flyme: no migration to redo")
				return
			}

			if err := db.MigrateUp(database); err != nil {
				log.Fatalf("flyme redo (up): %v", err)
			}
			fmt.Printf("flyme: redone %s\n", reverted.Name)
		}
	case "check":
		if err := db.ValidateMigrations(); err != nil {
			log.Fatalf("flyme check: %v", err)
		}
		fmt.Println("flyme: check passed (read-only)")
	default:
		usage()
		os.Exit(1)
	}
}

func runCreate(description string) error {
	if err := validateMigrationDescription(description); err != nil {
		return err
	}

	nextVersion, err := nextMigrationVersion(migrationDirPath)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("%03d_%s", nextVersion, description)
	if err := createMigrationPair(migrationDirPath, name); err != nil {
		return err
	}

	fmt.Printf("flyme: created %s.up.sql and %s.down.sql\n", name, name)
	return nil
}

func validateMigrationDescription(description string) error {
	if !migrationDescriptionPattern.MatchString(description) {
		return fmt.Errorf("invalid migration name %q (expected ^[a-z][a-z0-9_]*$)", description)
	}

	return nil
}

func nextMigrationVersion(migrationsDir string) (int, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return 0, fmt.Errorf("read migrations directory: %w", err)
	}

	maxVersion := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		base := strings.TrimSuffix(name, ".up.sql")
		parts := strings.SplitN(base, "_", 2)
		if len(parts) != 2 || parts[0] == "" {
			continue
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		if version > maxVersion {
			maxVersion = version
		}
	}

	return maxVersion + 1, nil
}

func createMigrationPair(migrationsDir, name string) error {
	upFilename := name + ".up.sql"
	downFilename := name + ".down.sql"

	upPath := filepath.Join(migrationsDir, upFilename)
	downPath := filepath.Join(migrationsDir, downFilename)

	upContent := fmt.Sprintf("-- %s\n\n", upFilename)
	downContent := fmt.Sprintf("-- %s\n\n", downFilename)

	upFile, err := os.OpenFile(upPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create %s: %w", upFilename, err)
	}

	if _, err := upFile.WriteString(upContent); err != nil {
		_ = upFile.Close()
		_ = os.Remove(upPath)
		return fmt.Errorf("write %s: %w", upFilename, err)
	}

	if err := upFile.Close(); err != nil {
		_ = os.Remove(upPath)
		return fmt.Errorf("close %s: %w", upFilename, err)
	}

	downFile, err := os.OpenFile(downPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		_ = os.Remove(upPath)
		return fmt.Errorf("create %s: %w", downFilename, err)
	}

	if _, err := downFile.WriteString(downContent); err != nil {
		_ = downFile.Close()
		_ = os.Remove(downPath)
		_ = os.Remove(upPath)
		return fmt.Errorf("write %s: %w", downFilename, err)
	}

	if err := downFile.Close(); err != nil {
		_ = os.Remove(downPath)
		_ = os.Remove(upPath)
		return fmt.Errorf("close %s: %w", downFilename, err)
	}

	return nil
}

func openDatabaseForFlyme() (*sql.DB, func(), error) {
	dbPath, err := resolveDBPath()
	if err != nil {
		return nil, nil, fmt.Errorf("resolve db path: %w", err)
	}

	database, err := db.OpenWithoutMigrations(dbPath)
	if err != nil {
		return nil, nil, err
	}

	closeDatabase := func() { _ = database.Close() }
	return database, closeDatabase, nil
}

func resolveDBPath() (string, error) {
	if dbPath := os.Getenv("CUBBY_DB_PATH"); dbPath != "" {
		return dbPath, nil
	}

	return "", fmt.Errorf("CUBBY_DB_PATH is required")
}

func printStatus(statuses []db.MigrationStatus) {
	writer := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	_, _ = fmt.Fprintln(writer, "VERSION\tSTATE\tAPPLIED_AT\tNAME")

	for _, status := range statuses {
		state := "pending"
		appliedAt := "-"
		if status.Applied {
			state = "applied"
			if status.AppliedAt != nil {
				appliedAt = status.AppliedAt.UTC().Format(time.RFC3339)
			}
		}

		_, _ = fmt.Fprintf(writer, "%03d\t%s\t%s\t%s\n", status.Version, state, appliedAt, status.Name)
	}

	_ = writer.Flush()
}

func usage() {
	fmt.Println("usage: go run ./cmd/flyme <up|down|status|redo|check|create>")
}
