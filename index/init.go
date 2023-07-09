package index

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(dir string) error {
	db, err := sql.Open("sqlite3", "mydatabase.sqlite")
	if err != nil {
		return err
	}

	defer db.Close()
	db.Exec("CREATE TABLE ")
	return nil
}
