package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"sync"
)

// FileChecksums is a thread-safe map that stores file checksums
type FileChecksums struct {
	lock    sync.Mutex
	storage map[string]string
}

// NewFileChecksums creates a new file checksums map
func NewFileChecksums() *FileChecksums {
	return &FileChecksums{storage: make(map[string]string)}
}

// UpdateFileChecksum updates the checksum for the given path in a thread-safe manner
func (c *FileChecksums) UpdateFileChecksum(path string, checksum string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.storage[path] = checksum
}

// HasChanged checks if the checksum for the given path has changed in a thread safe manner
func (c *FileChecksums) HasChanged(path string, checksum string) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	return checksum != c.storage[path]
}

// FileChecksum calculates a new checksum for the given path
func FileChecksum(path string) (string, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	/*
		if len(contents) == 0 {
			return "", fmt.Errorf("error: empty file, could not update checksum for %s", path)
		}
	*/

	h := sha256.New()
	if _, err := h.Write(contents); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
