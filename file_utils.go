package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

func checkPath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}
	return nil
}

func absolutePath(path string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, path), nil
}

func readFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func createDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

func createFile(path string, content []byte) (*os.File, error) {
	var (
		f   *os.File
		err error
	)
	if err = checkPath(path); err != nil {
		if f, err = openFile(path); err != nil {
			return nil, err
		}
	}

	defer closeFile(f)

	if err = writeFile(f, content); err != nil {
		return nil, err
	}
	return f, nil
}

func writeFile(file *os.File, content []byte) error {
	if _, err := file.Write(content); err != nil {
		return err
	}
	return nil
}

func closeFile(file *os.File) {
	file.Close()
}

func deletePath(path string) {
	os.RemoveAll(path)
}

func openFile(path string) (*os.File, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return f, err
}
