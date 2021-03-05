package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"sync"
)

// FileChecksums is a thread-safe map that stores file checksums
type FileChecksums struct {
	lock    sync.Mutex
	storage map[string]string
}

func NewFileChecksums() *FileChecksums {
	return &FileChecksums{storage: make(map[string]string)}
}

// UpdateFileChecksum updates the checksum for the given path in a thread-safe manner
func (c *FileChecksums) UpdateFileChecksum(path string) (bool, error) {
	newChecksum, err := fileChecksum(path)
	if err != nil {
		return false, err
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	oldChecksum, ok := c.storage[path]
	if !ok || oldChecksum != newChecksum {
		c.storage[path] = newChecksum
		return true, nil
	}

	return false, nil
}

func (c *FileChecksums) HasChanged(path string) (bool, error) {
	newChecksum, err := fileChecksum(path)
	if err != nil {
		return false, err
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	return newChecksum != c.storage[path], nil
}

func fileChecksum(path string) (string, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	if len(contents) == 0 {
		return "", fmt.Errorf("error: empty file, could not update checksum for %s", path)
	}

	h := sha256.New()
	if _, err := h.Write(contents); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
