package index

import (
	"database/sql"
	"errors"
	"fmt"
	"path"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbName    = "image_hash.sqlite3"
	tableName = "ImageHash"
)

type ImageHash struct {
	db *sql.DB
}

func NewHash(path string) (*ImageHash, error) {
	ih := &ImageHash{}
	db, err := openOrCreate(path)
	if err != nil {
		return nil, err
	}
	ih.db = db
	return ih, nil
}

func (i *ImageHash) GetName(h string) (string, error) {
	var existing string
	err := i.db.QueryRow("SELECT path from "+tableName+" where hash=?", h).Scan(&existing)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		err = fmt.Errorf("read failed: %w", err)
	}
	return existing, err
}

func (i *ImageHash) Insert(path, hash string) error {
	_, err := i.db.Exec("Insert into "+tableName+" values (?, ?)", path, hash)
	if err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}
	return nil
}

func (i *ImageHash) UpdatePath(path, hash string) error {
	_, err := i.db.Exec("Update "+tableName+" set path=? where hash=?", path, hash)
	return err
}

func (i *ImageHash) Close() error {
	return i.db.Close()
}

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
