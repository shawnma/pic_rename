package main

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
	"shawnma.com/pic_rename/pkg/index"
)

func main() {
	var dbPath, dir string
	flag.StringVarP(&dbPath, "db", "d", ".", "path to the database file")
	flag.StringVarP(&dir, "dir", "s", ".", "directory to index")
	flag.Parse()

	if dbPath == "" || dir == "" {
		fmt.Printf("%s: required --db and --dir\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	err := index.UpdateIndex(dbPath, dir)
	if err != nil {
		panic(err)
	}
}
