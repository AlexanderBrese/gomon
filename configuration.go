package main

import (
	"time"
)

// Configuration is a in-memory representation of the expected configuration file
type Configuration struct {
	SourceDir   string   `toml:"relative_source_dir"`
	BuildDir    string   `toml:"relative_build_dir"`
	LogDir      string   `toml:"relative_log_dir"`
	WatchExt    []string `toml:"watch_relative_ext"`
	IgnoreDir   []string `toml:"ignore_relative_dir"`
	WatchDir    []string `toml:"watch_relative_dir"`
	IgnoreFiles []string `toml:"ignore_relative_files"`
	Delay       int      `toml:"delay"`
	Port        int      `toml:"port"`
}

var defaultConfiguration = &Configuration{
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

func (c *Configuration) delay() time.Duration {
	return time.Duration(c.Delay) * time.Millisecond
}

func (c *Configuration) srcPath() (string, error) {
	return absolutePath(c.SourceDir)
}

func (c *Configuration) buildPath() (string, error) {
	return absolutePath(c.BuildDir)
}

func (c *Configuration) logPath() (string, error) {
	return absolutePath(c.LogDir)
}
