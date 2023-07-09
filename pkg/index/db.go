package index

import (
	"database/sql"
	"fmt"
	"path"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbName    = "image_hash.sqlite3"
	tableName = "ImageHash"
)

func openOrCreate(dbPath string) (*sql.DB, error) {
	dbf := path.Join(dbPath, dbName)
	db, err := sql.Open("sqlite3", dbf)
	if err == nil {
		err = initDB(db)
		if err != nil {
			return nil, fmt.Errorf("unable to create sqlite file: %w", err)
		}
	}
	return db, err
}

func initDB(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS " + tableName + " (path varchar primary key, hash varchar)")
	if err != nil {
		return err
	}
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS Index_hash on " + tableName + "(hash)")
	return err
}
