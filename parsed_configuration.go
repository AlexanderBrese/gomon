// assumptions:
// - type toml
// - predefined file path
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"
)

func ParsedConfiguration(path string) (*Configuration, error) {
	if path == "" {
		return _default(), nil
	} else if err := _check(path); err != nil {
		return nil, err
	} else {
		return _parse(path)
	}
}

func _parse(path string) (cfg *Configuration, err error) {
	cfgData, err := _read(path)
	if err != nil {
		return nil, err
	}
	cfg, err = _unmarshal(cfgData)
	if err != nil {
		return nil, err
	}
	return cfg, err
}

func _unmarshal(cfgData []byte) (cfg *Configuration, err error) {
	if err = toml.Unmarshal(cfgData, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func _default() *Configuration {
	return &Configuration{
		SourceDir:   "cmd/web",
		BuildDir:    "tmp/build",
		LogDir:      "tmp/go-server-browser-reload.log",
		WatchExt:    []string{"go", "tpl", "tmpl", "html", "css", "js", "env", "yaml"},
		IgnoreDir:   []string{"assets", "tmp", "vendor", "node_modules", "build"},
		WatchDir:    []string{},
		IgnoreFiles: []string{},
		Delay:       1000,
		Port:        3000,
	}
}

func _check(path string) error {
	absPath, err := _absolutePath(path)
	if err != nil {
		return fmt.Errorf("Failed to construct an absolute path for: %s", path)
	}
	_, err = _read(absPath)
	if err != nil {
		return err
	}
	return nil
}

func _absolutePath(path string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(wd, path), nil
}

func _read(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}
