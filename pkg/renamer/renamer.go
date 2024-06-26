package renamer

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/evanoberholster/imagemeta"
	flag "github.com/spf13/pflag"
	"shawnma.com/pic_rename/pkg/log"
)

const (
	nameFormat      = "20060102_150405"
	folderFomrat    = "06/2006.01.02"
	folderMonthOnly = "06/2006.01"
)

var suffixes = []string{".jpg", ".jpeg", ".heic", ".cr2", ".mp4", ".mov"}

var logger = log.New()

type Renamer struct {
	dirsCache map[string]bool
	monthOnly bool
	overwrite bool
	dryRun    bool
	seq       bool // whether to rename the files using a seq number if already exists
	dest      string
	src       string
}

func (r *Renamer) parseFlags() {
	flag.BoolVarP(&r.monthOnly, "month-only", "m", false, "use month in path only")
	flag.BoolVarP(&r.overwrite, "overwrite", "o", false, "overwrite the file if the file exists")
	flag.BoolVarP(&r.seq, "seq", "q", false, "if the dest exits, using a seq number suffix")
	flag.BoolVar(&r.dryRun, "dry-run", false, "dry run only")

	flag.StringVarP(&r.dest, "dest", "d", "output", "destination dir")
	flag.StringVarP(&r.src, "src", "s", "", "source dir")
	flag.Parse()
}

func (r *Renamer) Rename() {
	r.parseFlags()
	if r.src == "" {
		fmt.Printf("%s: required --src\n", os.Args[0])
		flag.PrintDefaults()
		return
	}
	logger.Info("Renaming pictures, src=%s, dest=%s, month only=%v dry run=%v override=%v seq=%v\n",
		r.src, r.dest, r.monthOnly, r.dryRun, r.overwrite, r.seq)
	suffixSet := map[string]bool{}
	for _, s := range suffixes {
		suffixSet[s] = true
	}
	filepath.Walk(r.src, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			suffix := strings.ToLower(filepath.Ext(path))
			if _, ok := suffixSet[suffix]; ok {
				r.renameOne(path, suffix)
			}
		}
		return nil
	})
}

func (r *Renamer) renameOne(path string, suffix string) {
	t, e := getDate(path)
	if e != nil || t.Year() < 2000 {
		logger.Error("Error getting date from %s: %v, t=%q\n", path, e, t)
		return
	}

	for i := 0; ; i++ {
		dest := r.getDest(t, suffix, i)
		if _, err := os.Stat(dest); err == nil {
			logger.Info("%s -> %s: File already exists", path, dest)
			if r.seq {
				// get next file name
				logger.Info("Getting next file name")
				continue
			}
			if !r.overwrite {
				return
			}
		}
		if r.dryRun {
			logger.Info("Dry run: mv %s %s", path, dest)
			return
		}
		e = os.Rename(path, dest)
		if e != nil {
			ex := copy(path, dest)
			if ex != nil {
				logger.Error("Failed to rename and copy %s->%s: %s; %s", path, dest, e, ex)
			} else {
				os.Remove(path)
				logger.Debug("Success: %s->%s", path, dest)
			}
		} else {
			logger.Debug("Success: %s->%s", path, dest)
		}
		return
	}
}

func copy(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

func (r *Renamer) getDest(t time.Time, suffix string, seq int) string {
	var folder string
	if r.monthOnly {
		folder = t.Format(folderMonthOnly)
	} else {
		folder = t.Format(folderFomrat)
	}
	destDir := path.Join(r.dest, folder)
	if _, ok := r.dirsCache[destDir]; !ok && !r.dryRun {
		os.MkdirAll(destDir, 0700)
		r.dirsCache[destDir] = true
	}
	var file string
	if seq == 0 {
		file = t.Format(nameFormat) + suffix
	} else {
		file = fmt.Sprintf("%s_%d%s", t.Format(nameFormat), seq, suffix)
	}
	return path.Join(destDir, file)
}

func getDate(path string) (time.Time, error) {
	f, err := os.Open(path)
	if err != nil {
		return time.Time{}, err
	}
	defer f.Close()
	e, err := imagemeta.Decode(f)
	if err != nil {
		return time.Time{}, err
	}
	return e.CreateDate(), nil
}

func New() *Renamer {
	return &Renamer{
		dirsCache: make(map[string]bool),
	}
}
