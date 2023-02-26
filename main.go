package main

import (
	"errors"
	"flag"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/evanoberholster/imagemeta"
	"github.com/fatih/color"
)

func main() {
	monthOnly := flag.Bool("m", false, "use month in path only")
	dest := flag.String("dest", "output", "Destination dir")
	src := flag.String("src", ".", "source dir")
	flag.Parse()
	r := renamer{
		dirsCache: make(map[string]bool),
		monthOnly: *monthOnly,
		src:       *src,
		dest:      *dest,
	}
	r.rename()
}

const (
	nameFormat      = "20060102_150405"
	folderFomrat    = "2006.01.02"
	folderMonthOnly = "2006.01"
)

type renamer struct {
	dirsCache map[string]bool
	monthOnly bool
	dest      string
	src       string
}

func (r *renamer) rename() {
	log.Printf("Renaming pictures, src=%s, dest=%s, month only=%v\n", r.src, r.dest, r.monthOnly)
	filepath.Walk(r.src, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			suffix := strings.ToLower(filepath.Ext(path))
			if suffix == ".jpg" || suffix == ".heic" {
				t, e := getDate(path)
				if e != nil {
					color.Red("Error getting date from %s: %v\n", path, e)
				} else {
					dest := r.getDest(t, suffix)
					e = os.Rename(path, dest)
					if e != nil {
						color.Red("Failed to rename %s->%s: %s", path, dest, e)
					} else {
						color.Cyan("Success: %s->%s", path, dest)
					}
				}
			}
		}
		return nil
	})
}

func (r *renamer) getDest(t time.Time, suffix string) string {
	var folder string
	if r.monthOnly {
		folder = t.Format(folderMonthOnly)
	} else {
		folder = t.Format(folderFomrat)
	}
	destDir := path.Join(r.dest, folder)
	if _, ok := r.dirsCache[destDir]; !ok {
		os.MkdirAll(destDir, 0700)
		r.dirsCache[destDir] = true
	}
	file := t.Format(nameFormat) + suffix
	return path.Join(destDir, file)
}

func getDate(path string) (time.Time, error) {
	f, err := os.Open(path)
	if err != nil {
		return time.Time{}, err
	}
	defer f.Close()
	m, err := imagemeta.Parse(f)
	if err != nil {
		return time.Time{}, err
	}
	if m == nil {
		return time.Time{}, errors.New("parse succeeded, but no metadata found")
	}
	e, err := m.Exif()
	if err != nil {
		return time.Time{}, err
	}
	if e == nil {
		return time.Time{}, errors.New("no exif data found")
	}
	return e.DateTime(time.Local)
}
