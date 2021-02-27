package utils

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func CheckPath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}
	return nil
}

func AbsolutePath(path string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, path), nil
}

func ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func CreateDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
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
