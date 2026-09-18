package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreate_NextMigrationVersion_EmptyDirStartsAtOne(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	got, err := nextMigrationVersion(dir)
	if err != nil {
		t.Fatalf("nextMigrationVersion() error = %v", err)
	}

	if got != 1 {
		t.Errorf("nextMigrationVersion() = %d, want %d", got, 1)
	}
}

func TestCreate_NextMigrationVersion_UsesMaxExistingPlusOne(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	seedFiles := []string{
		"001_initial.up.sql",
		"001_initial.down.sql",
		"004_add_items.up.sql",
		"004_add_items.down.sql",
		"007_add_tags.up.sql",
		"README.md",
		"bad_prefix.up.sql",
	}

	for _, name := range seedFiles {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("-- seed\n"), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	got, err := nextMigrationVersion(dir)
	if err != nil {
		t.Fatalf("nextMigrationVersion() error = %v", err)
	}

	if got != 8 {
		t.Errorf("nextMigrationVersion() = %d, want %d", got, 8)
	}
}

func TestCreate_ValidateMigrationDescription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		desc    string
		wantErr bool
	}{
		{name: "valid", desc: "add_items", wantErr: false},
		{name: "valid with numbers", desc: "a1_b2", wantErr: false},
		{name: "invalid uppercase", desc: "Bad Name", wantErr: true},
		{name: "invalid hyphen", desc: "bad-name", wantErr: true},
		{name: "invalid empty", desc: "", wantErr: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validateMigrationDescription(tc.desc)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateMigrationDescription(%q) error = %v, wantErr %v", tc.desc, err, tc.wantErr)
			}

			if tc.wantErr && err != nil && !strings.Contains(err.Error(), "invalid migration name") {
				t.Errorf("validateMigrationDescription(%q) error = %q, want contains %q", tc.desc, err.Error(), "invalid migration name")
			}
		})
	}
}

func TestCreate_CreateMigrationPair_GeneratesUpAndDownFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	name := "001_create_items"

	err := createMigrationPair(dir, name)
	if err != nil {
		t.Fatalf("createMigrationPair() error = %v", err)
	}

	upPath := filepath.Join(dir, name+".up.sql")
	downPath := filepath.Join(dir, name+".down.sql")

	upContent, err := os.ReadFile(upPath)
	if err != nil {
		t.Fatalf("read up migration: %v", err)
	}

	downContent, err := os.ReadFile(downPath)
	if err != nil {
		t.Fatalf("read down migration: %v", err)
	}

	wantUp := "-- " + name + ".up.sql\n\n"
	if string(upContent) != wantUp {
		t.Errorf("up migration content = %q, want %q", string(upContent), wantUp)
	}

	wantDown := "-- " + name + ".down.sql\n\n"
	if string(downContent) != wantDown {
		t.Errorf("down migration content = %q, want %q", string(downContent), wantDown)
	}
}
