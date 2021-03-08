package utils

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// CheckPath checks if the path exists
func CheckPath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}
	return nil
}

// IsDir checks if the path is a directory
func IsDir(path string) (bool, error) {
	f, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return f.IsDir(), nil
}

// RelPath is a relative representation of the absolute path provided
func RelPath(root string, path string) (string, error) {
	s, err := filepath.Rel(root, path)
	if err != nil {
		return "", err
	}
	return s, nil
}

// CurrentRootPath is the current root path
func CurrentRootPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("error: could not get root path: %s", err)
		return "", err
	}
	return wd, nil
}

// CurrentAbsolutePath is a absolute representation of the relative path provided by taking the current root path into account
func CurrentAbsolutePath(relPath string) (string, error) {
	root, err := CurrentRootPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, relPath), nil
}

// ReadFile is the file read at the path provided
func ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

// CreateAllDir creates all directories up to the path provided
func CreateAllDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

// CreateAllDirIfNotExist creates all directories if they do not already exist up to the path provided
func CreateAllDirIfNotExist(path string) error {
	if err := CheckPath(path); err != nil {
		if err = CreateAllDir(path); err != nil {
			return err
		}
	}
	return nil
}

// RemoveFileIfExist removes a file if it does already exist at the path provided
func RemoveFileIfExist(path string) error {
	if err := CheckPath(path); err != nil {
		return err
	}
	return RemoveFile(path)
}

// CreateFile creates a file with the given content at the path provided
func CreateFile(path string, content []byte) (*os.File, error) {
	var (
		f   *os.File
		err error
	)

	if f, err = OpenFile(path); err != nil {
		return nil, err
	}
	if err = WriteFile(f, content); err != nil {
		return nil, err
	}
	defer CloseFile(f)
	return f, nil
}

// WriteFile writes to a file the content provided
func WriteFile(file *os.File, content []byte) error {
	if _, err := file.Write(content); err != nil {
		return err
	}
	return nil
}

// CloseFile closes the file provided
func CloseFile(file *os.File) error {
	return file.Close()
}

// RemoveAllDir removes all directories up to the path provided
func RemoveAllDir(path string) error {
	return os.RemoveAll(path)
}

// RemoveFile removes a file at the path provided
func RemoveFile(filePath string) error {
	return os.Remove(filePath)
}

// OpenFile opens a file at the path provided
func OpenFile(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return f, err
}

// RemoveRootDir removes all directories up to the relative path provided
func RemoveRootDir(relPath string) error {
	relParent := strings.Split(relPath, "/")[0]
	parent, err := CurrentAbsolutePath(relParent)
	if err != nil {
		return err
	}
	return RemoveAllDir(parent)
}

// IsWrite checks if the fsnotify event is a write
func IsWrite(ev fsnotify.Event) bool {
	return ev.Op&fsnotify.Write == fsnotify.Write
}

// IsRemove checks if the fsnotify event is a remove
func IsRemove(ev fsnotify.Event) bool {
	return ev.Op&fsnotify.Remove == fsnotify.Remove
}

// IsCreate checks if the fsnotify event is a create
func IsCreate(ev fsnotify.Event) bool {
	return ev.Op&fsnotify.Create == fsnotify.Create
}
