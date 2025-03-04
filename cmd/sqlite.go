package cmd

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"

	"github.com/Kaya-Sem/oopsie"
)

func getDBPath() string {
	home := os.Getenv("HOME")

	// FIX: not working correctly I think. No need for sudo anyway?

	// If running with sudo, get the original user's home
	if os.Getenv("SUDO_USER") != "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			home = homeDir
		}
	}

	if home == "" {
		fmt.Println("Could not determine user home directory")
		os.Exit(1)
	}

	dbDir := filepath.Join(home, databaseDir)
	return filepath.Join(dbDir, databaseFile)
}

func getDatabaseConnection() *sql.DB {
	err := ensureDB()
	if err != nil {
		fmt.Print(oopsie.CreateOopsie().Title("Database error").IndicatorMessage("DATABASE").IndicatorColors(oopsie.GREEN, oopsie.BRIGHT_BLACK).Render())
		os.Exit(1)
	}

	path := getDBPath() // Use the correct path now
	db, err := sql.Open("sqlite3", path)
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
		if err := os.MkdirAll(dir, 0666); err != nil {
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

func AddFolder(fi FolderInfo) error {

	return nil
}

func addFolderToDatabase(path string, inode int, folderType string) error {
	db := getDatabaseConnection()
	if db == nil {
		return fmt.Errorf("failed to connect to database")
	}

	defer db.Close()

	query := `INSERT INTO folders (inode, path, type) VALUES(?,?,?)`

	rows, err := db.Query(query, inode, path, folderType)
	if err != nil {
		return fmt.Errorf("Database query failed: %w", err)
	}

	defer rows.Close()

	return nil
}

func GetTypedFolders(t string) ([]FolderInfo, error) {
	db := getDatabaseConnection()
	if db == nil {
		return nil, fmt.Errorf("failed to connect to database")
	}

	defer db.Close()

	query := `SELECT inode, path FROM folders WHERE type = ?`
	rows, err := db.Query(query, t)
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	defer rows.Close()

	var results []FolderInfo

	for rows.Next() {
		var fileInfo FolderInfo
		if err := rows.Scan(&fileInfo.Inode, &fileInfo.FullPath); err != nil {
			fmt.Printf("Error scanning row: %v\n", err)
			continue
		}
		results = append(results, fileInfo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return results, nil
}
