package datastore

import (
	"database/sql"
	"fmt"

	"github.com/spf13/viper"

	_ "modernc.org/sqlite"
)

func startDb() (*sql.DB, error) {
	dbtype := viper.GetString("server.database.type")
	if dbtype == "sqlite" {
		dbname := viper.GetString("server.database.sqlite.name")
		db, err := connectSqliteDB(dbname)
		if err != nil {
			return nil, err
		}
		return db, nil
	}
	if dbtype == "mysql" {
		return nil, fmt.Errorf("mysql: Not Implemented")
	}
	return nil, fmt.Errorf("startDB: Not Implemented")
}

func connectSqliteDB(dbPath string) (*sql.DB, error) {
	// Note: the busy_timeout pragma must be first because
	// the connection needs to be set to block on busy before WAL mode
	// is set in case it hasn't been already set by another connection.
	pragmas := "?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=journal_size_limit(200000000)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)"

	db, err := sql.Open("sqlite", dbPath+pragmas)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// GetDB returns a database connection
func GetDB() (*sql.DB, error) {
	return startDb()
}

func InitAllTables() {
	db, err := startDb()
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		return
	}
	defer db.Close()

	// Apply schema migrations
	if err := ApplyMigrations(db); err != nil {
		fmt.Printf("Error applying migrations: %v\n", err)
		return
	}

	initAgentTable()
	fmt.Println("Database initialized successfully")
}
