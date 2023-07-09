package index

import (
	"bufio"
	"crypto"
	_ "crypto/md5"
	"encoding/base32"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"shawnma.com/pic_rename/pkg/log"
)

var logger = log.New()

func UpdateIndex(dbPath string, dir string) error {
	db, err := NewHash(dbPath)
	if err != nil {
		return err
	}
	defer db.Close()
	filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			suffix := strings.ToLower(filepath.Ext(path))
			if suffix == ".jpg" || suffix == ".heic" || suffix == ".cr2" || suffix == ".jpeg" {
				h, err := hashFile(path)
				if err != nil {
					logger.Error("Hash file failed: %w", err)
					return err
				}
				existing, err := db.GetName(h)
				if err != nil {
					logger.Warn("Read failed: %w", err)
				} else if existing == "" {
					logger.Debug("No existing row, inserting (%s, %s)", path, h)
					err = db.Insert(path, h)
					if err != nil {
						logger.Warn("insert failed: %w", err)
					}
				} else {
					_, err = os.Stat(existing)
					if os.IsNotExist(err) {
						logger.Debug("Found existing path %s with same hash (%s, %s), existing file moved, updating...", existing, path, h)
						err := db.UpdatePath(path, h)
						if err != nil {
							logger.Warn("Update failed: %w", err)
						}
					} else if existing != path {
						logger.Warn("Found duplicated hash for %s and %s with hash %s", existing, path, h)
					} // else no update
				}
			}
		}
		return nil
	})
	return nil
}

func hashFile(f string) (string, error) {
	handle, err := os.Open(f)
	if err != nil {
		return "", err
	}
	hash := crypto.MD5.New()

	bs := make([]byte, 1024)
	r := bufio.NewReader(handle)
	for {
		n, err := r.Read(bs)
		if n > 0 {
			_, he := hash.Write(bs[0:n])
			if he != nil {
				return "", he
			}
		}
		if err != nil && err == io.EOF {
			var result []byte = make([]byte, 0)
			return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(hash.Sum(result)), nil
		}
	}
}
