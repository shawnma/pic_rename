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
	"unicode"

	"shawnma.com/pic_rename/pkg/log"
)

var logger = log.New()

func UpdateIndex(dbPath string, dir string) error {
	absDb, err := filepath.Abs(dbPath)
	if err != nil {
		return err
	}
	db, err := NewHash(absDb)
	if err != nil {
		return err
	}
	logger.Info("Opened DB at %s", absDb)
	defer db.Close()
	count := 1
	filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			suffix := strings.ToLower(filepath.Ext(path))
			if suffix == ".jpg" || suffix == ".heic" || suffix == ".cr2" || suffix == ".jpeg" {
				absPath, _ := filepath.Abs(path)
				relPath, _ := filepath.Rel(absDb, absPath)
				relPath = strings.ToLower(relPath)
				h, err := db.GetHash(relPath)
				if err != nil {
					logger.Error("Failed to get hash from db for %s", path)
					return nil
				}
				if h != "" {
					return nil // assuming everyone is fine and file is immutable.
				}
				h, err = hashFile(path)
				if err != nil {
					logger.Error("Hash file failed: %w", err)
					return err
				}
				existing, err := db.GetName(h)
				existing = strings.ToLower(existing)
				if err != nil {
					logger.Warn("Read name for hash failed: %w", err)
				} else if existing == "" {
					logger.Debug("No existing row, inserting (%s, %s)", relPath, h)
					err = db.Insert(relPath, h)
					if err != nil {
						logger.Warn("insert failed: %w", err)
					}
				} else {
					_, err = os.Stat(filepath.Join(absDb, existing))
					if os.IsNotExist(err) {
						logger.Debug("Found existing path %s with same hash (%s, %s), existing file moved, updating...", existing, relPath, h)
						err := db.UpdatePath(relPath, h)
						if err != nil {
							logger.Warn("Update failed: %w", err)
						}
					} else if existing != relPath && unicode.IsDigit([]rune(relPath)[0]) && unicode.IsDigit([]rune(existing)[0]) {
						logger.Warn("DUP: %s - %s (%s)", existing, relPath, h)
					} // else no update
				}
				count += 1
				if count%100 == 0 {
					logger.Error("%d - %s", count, path)
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
	defer handle.Close()
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
