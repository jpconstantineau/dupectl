package datastore

import (
	"database/sql"
)

const CreateFoldersTableSQL = `
CREATE TABLE IF NOT EXISTS folders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL UNIQUE,
    parent_folder_id INTEGER,
    root_folder_id INTEGER NOT NULL,
    error_status TEXT,
    first_scanned_at INTEGER NOT NULL,
    last_scanned_at INTEGER NOT NULL,
    removed INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (parent_folder_id) REFERENCES folders(id) ON DELETE CASCADE,
    FOREIGN KEY (root_folder_id) REFERENCES root_folders(id) ON DELETE CASCADE
);`

const CreateFoldersIndexesSQL = `
CREATE INDEX IF NOT EXISTS idx_folders_parent ON folders(parent_folder_id);
CREATE INDEX IF NOT EXISTS idx_folders_root ON folders(root_folder_id);
CREATE INDEX IF NOT EXISTS idx_folders_path ON folders(path);
CREATE INDEX IF NOT EXISTS idx_folders_removed ON folders(removed);
`

// Folder represents a folder record in the database
type Folder struct {
	ID             int64
	Path           string
	ParentFolderID *int64
	RootFolderID   int64
	ErrorStatus    *string
	FirstScannedAt int64
	LastScannedAt  int64
	Removed        bool
}

// InsertFolder inserts a new folder record
func InsertFolder(db *sql.DB, folder *Folder) (int64, error) {
	query := `
	INSERT INTO folders (path, parent_folder_id, root_folder_id, error_status, 
	                     first_scanned_at, last_scanned_at, removed)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(path) DO UPDATE SET
		parent_folder_id = excluded.parent_folder_id,
		error_status = excluded.error_status,
		last_scanned_at = excluded.last_scanned_at,
		removed = 0
	RETURNING id
	`

	removed := 0
	if folder.Removed {
		removed = 1
	}

	var id int64
	err := db.QueryRow(query, folder.Path, folder.ParentFolderID, folder.RootFolderID,
		folder.ErrorStatus, folder.FirstScannedAt, folder.LastScannedAt, removed).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetFolderByPath retrieves a folder by path
func GetFolderByPath(db *sql.DB, path string) (*Folder, error) {
	query := `
	SELECT id, path, parent_folder_id, root_folder_id, error_status,
	       first_scanned_at, last_scanned_at, removed
	FROM folders
	WHERE path = ?
	`

	folder := &Folder{}
	var removed int
	err := db.QueryRow(query, path).Scan(&folder.ID, &folder.Path, &folder.ParentFolderID,
		&folder.RootFolderID, &folder.ErrorStatus, &folder.FirstScannedAt,
		&folder.LastScannedAt, &removed)
	if err != nil {
		return nil, err
	}
	folder.Removed = removed != 0
	return folder, nil
}

// SetFolderRemoved marks a folder as removed
func SetFolderRemoved(db *sql.DB, path string, removed bool) error {
	removedInt := 0
	if removed {
		removedInt = 1
	}
	query := `UPDATE folders SET removed = ?, last_scanned_at = strftime('%s', 'now') WHERE path = ?`
	_, err := db.Exec(query, removedInt, path)
	return err
}

// GetFoldersByRootID returns all folders for a root folder
func GetFoldersByRootID(db *sql.DB, rootFolderID int64) ([]*Folder, error) {
	query := `
	SELECT id, path, parent_folder_id, root_folder_id, error_status,
	       first_scanned_at, last_scanned_at, removed
	FROM folders
	WHERE root_folder_id = ? AND removed = 0
	ORDER BY path
	`

	rows, err := db.Query(query, rootFolderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []*Folder
	for rows.Next() {
		folder := &Folder{}
		var removed int
		err := rows.Scan(&folder.ID, &folder.Path, &folder.ParentFolderID, &folder.RootFolderID,
			&folder.ErrorStatus, &folder.FirstScannedAt, &folder.LastScannedAt, &removed)
		if err != nil {
			return nil, err
		}
		folder.Removed = removed != 0
		folders = append(folders, folder)
	}

	return folders, rows.Err()
}
