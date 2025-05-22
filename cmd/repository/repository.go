package repository

import (
	"database/sql"
	"os/user"

	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const (
	databaseDirectory = ".config/semlink"
	databaseFilename  = "semlink.sqlite"
)

func getDBPath() string {
	var home string

	// If running with sudo, resolve SUDO_USER's home directory
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		usr, err := user.Lookup(sudoUser)
		if err == nil {
			home = usr.HomeDir
		}
	}

	// Fallback to current HOME
	if home == "" {
		home = os.Getenv("HOME")
	}

	if home == "" {
		fmt.Println("Could not determine user home directory")
		os.Exit(1)
	}

	dbPath := filepath.Join(home, databaseDirectory)
	return dbPath
}

func getDatabaseConnection() (*sql.DB, error) {
	dbPath := getDBPath()
	err := ensureDB(dbPath)
	if err != nil {
		return nil, err
	}

	// Join the database filename to the path
	dbFilePath := filepath.Join(dbPath, databaseFilename)
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		fmt.Printf("getDatabaseConnection(): Could not open database at %s\n", dbFilePath)
		return nil, err
	}

	return db, nil
}

type repository interface {
	GetAllFolders() ([]FolderInfo, error)
	AddFolder(FolderInfo) error
	RemoveFolder(FolderInfo) error
	AddTagsToFolder(FolderInfo, []string) error
	Obliterate() error /* completely wipes the database */
}

type SqliteRepo struct {
	conn *sql.DB
}

func NewSqliteRepo() (*SqliteRepo, error) {
	conn, err := getDatabaseConnection()
	if err != nil {
		return nil, err
	}

	return &SqliteRepo{conn: conn}, nil
}

func (repo *SqliteRepo) GetAllFolders() ([]FolderInfo, error) {
	query := `
		SELECT 
			f.inode, 
			f.filepath, 
			t.name 
		FROM folders f
		LEFT JOIN folder_tags ft ON f.id = ft.folder_id
		LEFT JOIN tags t ON ft.tag_id = t.id
		ORDER BY f.id
	`

	rows, err := repo.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch folders: %w", err)
	}
	defer rows.Close()

	foldersMap := make(map[uint64]*FolderInfo)

	for rows.Next() {
		var inode int64
		var path string
		var tag sql.NullString

		if err := rows.Scan(&inode, &path, &tag); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		uid := uint64(inode)
		folder, exists := foldersMap[uid]
		if !exists {
			folder = &FolderInfo{
				Inode:    uid,
				FullPath: path,
				tags:     []string{},
			}
			foldersMap[uid] = folder
		}

		if tag.Valid {
			folder.tags = append(folder.tags, tag.String)
		}
	}

	var result []FolderInfo
	for _, folder := range foldersMap {
		result = append(result, *folder)
	}

	return result, nil
}

func (folderRepo *SqliteRepo) AddFolder(folderInfo FolderInfo) error {

	query := `INSERT INTO folders (inode, filepath) VALUES (?, ?)`

	_, err := folderRepo.conn.Exec(query, folderInfo.Inode, folderInfo.FullPath)
	return err
}

func (repo *SqliteRepo) RemoveFolder(folderInfo FolderInfo) error {

	// TODO: remove also tags in folder_tags.

	stmt := `DELETE FROM folders WHERE inode = ?`
	_, err := repo.conn.Exec(stmt, folderInfo.Inode)

	return err
}

func (repo *SqliteRepo) HasFolder(folderInfo FolderInfo) (bool, error) {
	query := `SELECT * FROM folders WHERE inode = ? && filepath = ?`

	folder := folderInfo

	if err := repo.conn.QueryRow(query, folderInfo.Inode, folderInfo.FullPath).Scan(&folder); err == nil {
		return false, nil
	} else if err == sql.ErrNoRows {
		return true, nil
	} else {
		return false, err
	}
}

func (repo *SqliteRepo) AddTagsToFolder(folderInfo FolderInfo, tags []string) error {
	tx, err := repo.conn.Begin()
	if err != nil {
		return err
	}

	// Ensures that all operations are done atomically. If one fails, we can roll everything back:
	defer tx.Rollback()

	var folderID int
	err = tx.QueryRow(`SELECT id FROM folders WHERE filepath = ?`, folderInfo.FullPath).Scan(&folderID)
	if err != nil {
		return fmt.Errorf("folder not found: %w", err)
	}

	for _, tagName := range tags {
		_, err := tx.Exec(`INSERT OR IGNORE INTO tags (name) VALUES (?)`, tagName)
		if err != nil {
			return fmt.Errorf("failed to insert tag: %w", err)
		}

		var tagID int
		err = tx.QueryRow(`SELECT id FROM tags WHERE name = ?`, tagName).Scan(&tagID)
		if err != nil {
			return fmt.Errorf("failed to retrieve tag id: %w", err)
		}

		_, err = tx.Exec(`INSERT OR IGNORE INTO folder_tags (folder_id, tag_id) VALUES (?, ?)`, folderID, tagID)
		if err != nil {
			return fmt.Errorf("failed to link folder and tag: %w", err)
		}
	}

	return tx.Commit()
}

// TODO: implement
func (repo *SqliteRepo) RemoveTagsFromFolder(folderInfo FolderInfo, tags []string) error {
	tx, err := repo.conn.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	return tx.Commit()
}

func (repo *SqliteRepo) Obliterate() error {
	fmt.Println("Obliterate not yet implemented")
	return nil
}

func ensureDB(path string) error {

	dbFilePath := filepath.Join(path, databaseFilename)

	// Check if the database file exists
	fileInfo, err := os.Stat(dbFilePath)
	if err == nil {
		if fileInfo.IsDir() {
			return fmt.Errorf("database path exists but is a directory: %s", dbFilePath)
		}
		return nil
	}

	if os.IsNotExist(err) {
		// The database file does not exist, create it

		// Create necessary directories if they don't exist
		dir := filepath.Dir(dbFilePath)
		if err := os.MkdirAll(dir, 0775); err != nil {
			return fmt.Errorf("could not create directory for database: %v", err)
		}

		// Create the empty database file
		file, err := os.Create(dbFilePath)
		if err != nil {
			return fmt.Errorf("could not create database file: %v", err)
		}
		defer file.Close()

	}

	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		return fmt.Errorf("failed to open DB for schema init: %w", err)
	}
	defer db.Close()
	return initialiseDBschema(db)

}

func initialiseDBschema(db *sql.DB) error {
	_, err := db.Exec(schema)
	if err != nil {
		return err
	}

	return nil
}

const schema = `
CREATE TABLE IF NOT EXISTS folders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    inode INTEGER NOT NULL UNIQUE,
    filepath TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS folder_tags (
    folder_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY (folder_id, tag_id),
    FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);
`
