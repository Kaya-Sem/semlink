package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

func getDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	dbDir := filepath.Join(home, databaseDir)
	return filepath.Join(dbDir, databaseFile)
}

func getDatabaseConnection() *sql.DB {
	path := getDBPath() // Use the correct path now
	db, err := sql.Open("sqlite", path)
	if err != nil {
		fmt.Printf("Could not open database at %s\n", path)
		return nil
	}

	return db
}

func ensureDB() error {
	dbPath := getDBPath()

	// Check if the database file exists
	_, err := os.Stat(dbPath)
	if err == nil {
		// File exists, no need to create it
		return nil
	}

	if os.IsNotExist(err) {
		// The database file does not exist, create it

		// Create necessary directories if they don't exist
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("could not create directory for database: %v", err)
		}

		// Create the empty database file
		file, err := os.Create(dbPath)
		if err != nil {
			return fmt.Errorf("could not create database file: %v", err)
		}
		defer file.Close()

		// Optionally, initialize the database schema after creating the file
		fmt.Println("Database file created.")
	}

	// Ensure that the database schema is initialized
	db := getDatabaseConnection()
	if db != nil {
		defer db.Close()
		initialiseDBschema(db)
	}

	return nil
}

func initialiseDBschema(db *sql.DB) {
	sql := `
	CREATE TABLE IF NOT EXISTS folders (
		inode BIGINT PRIMARY KEY, 
		path TEXT NOT NULL,
		type TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS mounts (
		source BIGINT NOT NULL,
		target BIGINT NOT NULL,
		FOREIGN KEY (source) REFERENCES folders (inode),
		FOREIGN KEY (target) REFERENCES folders (inode)
	);`

	_, err := db.Exec(sql)
	if err != nil {
		fmt.Printf("Could not initialise schema for database:\n%s\n", err)
	}
}
