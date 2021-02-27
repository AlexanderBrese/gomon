package main

import (
	"testing"

	"github.com/pelletier/go-toml"
)

func TestNoConfigProvided(t *testing.T) {
	cfg, err := ParsedConfiguration("")
	if err != nil {
		t.Errorf("want: error, got: %q", err)
	}
	if cfg != defaultConfiguration {
		t.Errorf("want: %q, got: %q", defaultConfiguration, cfg)
	}
}

func TestWrongConfigName(t *testing.T) {
	if _, err := ParsedConfiguration("no/file"); err == nil {
		t.Errorf("want: error, got: %q", err)
	}
}

func TestInvalidSourcePathProvided(t *testing.T) {
	testCfg := &Configuration{
		SourceDir: "wrong_source_dir",
	}
	cfgData, err := toml.Marshal(testCfg)
	if err != nil {
		t.Error(err)
	}

	dir := "test"
	absDir, err := absolutePath(dir)
	if err != nil {
		t.Error(err)
	}
	err = createDir(absDir)
	defer deletePath(absDir)
	if err != nil {
		t.Error(err)
	}

	path := dir + "/test.toml"
	absPath, err := absolutePath(path)
	if err != nil {
		t.Error(err)
	}
	if _, err = createFile(absPath, cfgData); err != nil {
		t.Error(err)
	}
	if _, err = ParsedConfiguration(absPath); err == nil {
		t.Errorf("want: error, got: %q", err)
	}
}

func TestConfigMerge(t *testing.T) {
	port := 4000
	testCfg := &Configuration{
		Port: port,
	}
	testCfgData, err := toml.Marshal(testCfg)
	if err != nil {
		t.Error(err)
	}

	path := "test.toml"
	absPath, err := absolutePath(path)
	if err != nil {
		t.Error(err)
	}
	if _, err := createFile(absPath, testCfgData); err != nil {
		t.Error(err)
	}

	defer deletePath(absPath)

	cfg, err := ParsedConfiguration(absPath)
	if err != nil {
		t.Error(err)
	}
	if cfg.Port != port {
		t.Errorf("want: %q, got: %q", port, cfg.Port)
	}
}
