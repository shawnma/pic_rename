package main

import (
	"os"

	"shawnma.com/pic_rename/pkg/index"
)

func main() {
	err := index.UpdateIndex(os.Args[1], os.Args[2])
	if err != nil {
		panic(err)
	}
}
