package datastore

import (
	"database/sql"
	"fmt"
)

// CreateRootFoldersTableSQL creates the root_folders table
const CreateRootFoldersTableSQL = `
CREATE TABLE IF NOT EXISTS root_folders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    path TEXT NOT NULL UNIQUE,
    agent_id INTEGER,
    traverse_links INTEGER NOT NULL DEFAULT 0,
    last_scan_date INTEGER,
    folder_count INTEGER NOT NULL DEFAULT 0,
    file_count INTEGER NOT NULL DEFAULT 0,
    total_size INTEGER NOT NULL DEFAULT 0
);`

const CreateRootFoldersIndexesSQL = `
CREATE INDEX IF NOT EXISTS idx_root_folders_path ON root_folders(path);
CREATE INDEX IF NOT EXISTS idx_root_folders_agent ON root_folders(agent_id);
`

// RootFolder represents a root folder entity
type RootFolder struct {
	ID            int64
	Path          string
	AgentID       *int64
	TraverseLinks bool
	LastScanDate  *int64
	FolderCount   int64
	FileCount     int64
	TotalSize     int64
}

// InsertRootFolder registers a new root folder
func InsertRootFolder(db *sql.DB, folder *RootFolder) (int64, error) {
	query := `
	INSERT INTO root_folders (path, agent_id, traverse_links, last_scan_date, 
	                          folder_count, file_count, total_size)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(path) DO UPDATE SET
		agent_id = excluded.agent_id,
		traverse_links = excluded.traverse_links
	RETURNING id
	`

	traverseLinks := 0
	if folder.TraverseLinks {
		traverseLinks = 1
	}

	var id int64
	err := db.QueryRow(query, folder.Path, folder.AgentID, traverseLinks,
		folder.LastScanDate, folder.FolderCount, folder.FileCount, folder.TotalSize).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetRootFolderByPath retrieves a root folder by its path
func GetRootFolderByPath(db *sql.DB, path string) (*RootFolder, error) {
	query := `
	SELECT id, path, agent_id, traverse_links, last_scan_date, 
	       folder_count, file_count, total_size
	FROM root_folders
	WHERE path = ?
	`

	var folder RootFolder
	var traverseLinks int
	err := db.QueryRow(query, path).Scan(
		&folder.ID, &folder.Path, &folder.AgentID, &traverseLinks,
		&folder.LastScanDate, &folder.FolderCount, &folder.FileCount, &folder.TotalSize,
	)
	if err != nil {
		return nil, err
	}

	folder.TraverseLinks = traverseLinks == 1
	return &folder, nil
}

// UpdateRootFolderStats updates the scan statistics for a root folder
func UpdateRootFolderStats(db *sql.DB, id int64, folderCount, fileCount, totalSize, lastScanDate int64) error {
	query := `
	UPDATE root_folders
	SET folder_count = ?,
	    file_count = ?,
	    total_size = ?,
	    last_scan_date = ?
	WHERE id = ?
	`

	_, err := db.Exec(query, folderCount, fileCount, totalSize, lastScanDate, id)
	return err
}

// GetAllRootFolders retrieves all root folders
func GetAllRootFolders(db *sql.DB) ([]*RootFolder, error) {
	query := `
	SELECT id, path, agent_id, traverse_links, last_scan_date,
	       folder_count, file_count, total_size
	FROM root_folders
	ORDER BY path
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []*RootFolder
	for rows.Next() {
		var folder RootFolder
		var traverseLinks int
		err := rows.Scan(
			&folder.ID, &folder.Path, &folder.AgentID, &traverseLinks,
			&folder.LastScanDate, &folder.FolderCount, &folder.FileCount, &folder.TotalSize,
		)
		if err != nil {
			return nil, err
		}
		folder.TraverseLinks = traverseLinks == 1
		folders = append(folders, &folder)
	}

	return folders, rows.Err()
}

// DeleteRootFolder removes a root folder and all associated scan data (CASCADE)
func DeleteRootFolder(db *sql.DB, id int64) error {
	query := `DELETE FROM root_folders WHERE id = ?`
	result, err := db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("root folder not found")
	}

	return nil
}
