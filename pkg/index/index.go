package index

import (
	"bufio"
	"crypto"
	_ "crypto/md5"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/exp/slices"
	"shawnma.com/pic_rename/pkg/log"
)

var (
	logger        = log.New()
	videosSuffix  = []string{".mp4", ".mov"}
	pictureSuffix = []string{".jpg", ".heic", ".cr2", ".jpeg", ".png", ".gif"}
)

func isSupported(ext string) bool {
	return isVideo(ext) || slices.Contains(pictureSuffix, ext)
}

func isVideo(f string) bool {
	return slices.Contains(videosSuffix, f)
}

// Duplicate represents a set of files that have been duplicated from one folder to another
type Duplicate struct {
	OriginalFolder string
	NewFolder      string
	OriginalFiles  []string
	NewFiles       []string
}

func UpdateIndex(dbPath string, dir string) ([]Duplicate, error) {
	absDb, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, err
	}
	db, err := NewHash(absDb)
	if err != nil {
		return nil, err
	}
	logger.Info("Opened DB at %s", absDb)
	defer db.Close()
	count := 1

	// Map to store duplicates by folder pairs
	folderPairsToDuplicates := make(map[string]*Duplicate)

	filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			path = strings.ToLower(path)
			suffix := filepath.Ext(path)
			if isSupported(suffix) {
				absPath, _ := filepath.Abs(path)
				relPath, _ := filepath.Rel(absDb, absPath)
				h, err := db.GetHash(relPath)
				if err != nil {
					logger.Error("Failed to get hash from db for %s", path)
					return nil
				}
				if h != "" {
					return nil // assuming everyone is fine and file is immutable.
				}
				h, err = hashFile(relPath)
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
					} else if existing != relPath {
						logger.Warn("DUP: %s - %s (%s)", existing, relPath, h)
						// Get the directories of both files
						dir1 := filepath.Dir(existing)
						dir2 := filepath.Dir(relPath)

						// Create a unique key for this folder pair
						var key string
						if dir1 < dir2 {
							key = dir1 + "|" + dir2
						} else {
							key = dir2 + "|" + dir1
						}

						// Get or create the Duplicate entry for this folder pair
						dup, exists := folderPairsToDuplicates[key]
						if !exists {
							dup = &Duplicate{
								OriginalFolder: dir1,
								NewFolder:      dir2,
							}
							folderPairsToDuplicates[key] = dup
						}

						// Add the files to their respective folders
						dup.OriginalFiles = append(dup.OriginalFiles, existing)
						dup.NewFiles = append(dup.NewFiles, relPath)
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

	// Convert the map to a slice of Duplicates
	var duplicates []Duplicate
	for _, dup := range folderPairsToDuplicates {
		duplicates = append(duplicates, *dup)
	}

	return duplicates, nil
}

func hashFile(f string) (string, error) {
	handle, err := os.Open(f)
	if err != nil {
		return "", err
	}
	defer handle.Close()
	hash := crypto.MD5.New()

	count := 1
	video := isVideo(filepath.Ext(f))

	if video {
		// for video, we'll only first 10K bytes. add the size as a factor as well.
		info, err := os.Stat(f)
		if err != nil {
			return "", fmt.Errorf("unabled to stat %s: %w", f, err)
		}
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(info.Size()))
		hash.Write(b)
	}

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
		count += 1
		if err == io.EOF || (count == 10 && video) {
			var result []byte = make([]byte, 0)
			return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(hash.Sum(result)), nil
		}
	}
}
