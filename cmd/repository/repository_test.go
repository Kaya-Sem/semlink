package repository

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

// This test will not pass for others
func TestGetDBPath(t *testing.T) {
	want := "/home/kayasem/.config/semlink"
	have := getDBPath()

	if want != have {
		t.Errorf(`want and have are not the same. want: %s, have: %s`, want, have)
	}
}

func TestEnsureDB(t *testing.T) {
	t.Run("creates db if it doesn't exist", func(t *testing.T) {
		tempDir := t.TempDir()

		err := ensureDB(tempDir)
		if err != nil {
			t.Errorf("Could not ensure db: %v", err)
		}

		dbPath := filepath.Join(tempDir, databaseFilename)
		info, err := os.Stat(dbPath)
		if os.IsNotExist(err) {
			t.Errorf("expected database file to be created, but it does not exist")
		} else if err != nil {
			t.Errorf("unexpected error while checking database file: %v", err)
		} else if info.IsDir() {
			t.Errorf("expected a file, but found a directory at %s", dbPath)
		}
	})

	t.Run("does not overwrite if db already exists", func(t *testing.T) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, databaseFilename)

		// Create dummy file before calling ensureDB
		originalContent := []byte("do not erase")
		if err := os.WriteFile(dbPath, originalContent, 0644); err != nil {
			t.Fatalf("failed to create dummy db file: %v", err)
		}

		err := ensureDB(tempDir)
		if err != nil {
			t.Errorf("ensureDB failed on existing file: %v", err)
		}

		content, err := os.ReadFile(dbPath)
		if err != nil {
			t.Errorf("failed to read existing file after ensureDB: %v", err)
		}
		if string(content) != string(originalContent) {
			t.Errorf("ensureDB overwrote existing file: got %q, want %q", string(content), string(originalContent))
		}
	})

	t.Run("fails if parent directory is not writable", func(t *testing.T) {
		parent := t.TempDir()
		badDir := filepath.Join(parent, "readonly")

		err := os.Mkdir(badDir, 0500) // read and execute only
		if err != nil {
			t.Fatalf("could not create restricted dir: %v", err)
		}
		defer os.Chmod(badDir, 0700) // so TempDir cleanup works

		err = ensureDB(badDir)
		if err == nil {
			t.Errorf("expected error when creating DB in unwritable dir, got nil")
		}
	})

	t.Run("fails gracefully if file can't be created", func(t *testing.T) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, databaseFilename)

		// Create a *directory* instead of file
		if err := os.Mkdir(dbPath, 0755); err != nil {
			t.Fatalf("failed to create dummy directory in place of db file: %v", err)
		}

		err := ensureDB(tempDir)
		if err == nil {
			t.Errorf("expected error when DB file path is a directory, got nil")
		}
	})

	t.Run("initializes schema correctly", func(t *testing.T) {
		tempDir := t.TempDir()
		err := ensureDB(tempDir)
		if err != nil {
			t.Fatalf("Could not ensure db: %v", err)
		}

		dbPath := filepath.Join(tempDir, databaseFilename)
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			t.Fatalf("Failed to open database: %v", err)
		}
		defer db.Close()

		tables := []string{"folders", "tags", "folder_tags"}
		for _, table := range tables {
			var name string
			err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
			if err != nil {
				if err == sql.ErrNoRows {
					t.Errorf("Table %s was not created", table)
				} else {
					t.Errorf("Error checking for table %s: %v", table, err)
				}
			}
		}

		rows, err := db.Query("PRAGMA table_info(folders)")
		if err != nil {
			t.Fatalf("Failed to query table structure: %v", err)
		}
		defer rows.Close()

		expectedColumns := map[string]bool{
			"id":       false,
			"inode":    false,
			"filepath": false,
		}

		for rows.Next() {
			var cid int
			var name string
			var dataType string
			var notNull bool
			var dfltValue interface{}
			var pk int
			err := rows.Scan(&cid, &name, &dataType, &notNull, &dfltValue, &pk)
			if err != nil {
				t.Fatalf("Failed to scan row: %v", err)
			}

			if _, exists := expectedColumns[name]; exists {
				expectedColumns[name] = true
			}
		}

		for col, found := range expectedColumns {
			if !found {
				t.Errorf("Expected column %s in folders table was not found", col)
			}
		}
	})
}

