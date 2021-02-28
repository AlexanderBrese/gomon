package utils

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func CheckPath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}
	return nil
}

func IsDir(path string) bool {
	f, err := os.Stat(path)
	if err != nil {
		return false
	}
	return f.IsDir()
}

func RootPath() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("error: could not get root path: %s", err)
		return ""
	}
	return wd
}

func AbsolutePath(relPath string) (string, error) {
	root := RootPath()
	return filepath.Join(root, relPath), nil
}

func ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func CreateDir(path string) error {
	return os.Mkdir(path, os.ModePerm)
}

func CreateDirIfNotExist(path string) error {
	if err := CheckPath(path); err != nil {
		if err = CreateDir(path); err != nil {
			return err
		}
	}
	return nil
}

func CreateFile(path string, content []byte) (*os.File, error) {
	var (
		f   *os.File
		err error
	)
	if err = CheckPath(path); err != nil {
		if f, err = OpenFile(path); err != nil {
			return nil, err
		}
	}

	defer CloseFile(f)

	if err = WriteFile(f, content); err != nil {
		return nil, err
	}
	return f, nil
}

func WriteFile(file *os.File, content []byte) error {
	if _, err := file.Write(content); err != nil {
		return err
	}
	return nil
}

func CloseFile(file *os.File) {
	file.Close()
}

func DeletePath(path string) {
	os.RemoveAll(path)
}

func OpenFile(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return f, err
}
