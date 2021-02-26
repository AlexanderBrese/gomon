package main

import (
	"path/filepath"
	"time"
)

const DFT = "go-server-browser-reload.toml"
const ROOT_PATH = "."

type Configuration struct {
	SourceDir   string   `toml:"relative_source_dir"`
	BuildDir    string   `toml:"relative_build_dir"`
	LogDir      string   `toml:"relative_log_dir"`
	WatchExt    []string `toml:"watch_ext"`
	IgnoreDir   []string `toml:"ignore_dir"`
	WatchDir    []string `toml:"include_dir"`
	IgnoreFiles []string `toml:"ignore_files"`
	Delay       int      `toml:"delay"`
	Port        int      `toml:"port"`
}

func (c *Configuration) delay() time.Duration {
	return time.Duration(c.Delay) * time.Millisecond
}

func (c *Configuration) srcPath() string {
	return _rel(c.SourceDir)
}

func (c *Configuration) buildPath() string {
	return _rel(c.BuildDir)
}

func (c *Configuration) logPath() string {
	return _rel(c.LogDir)
}

func _rel(path string) string {
	return filepath.Join(ROOT_PATH, path)
}
