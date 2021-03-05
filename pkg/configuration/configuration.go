package configuration

import (
	"time"

	"github.com/AlexanderBrese/go-server-browser-reload/pkg/utils"
)

var root string

func init() {
	var err error
	root, err = utils.RootPath()
	if err != nil {
		panic(err)
	}
}

// Configuration is a in-memory representation of the expected configuration file
type Configuration struct {
	sourceDir   string   `toml:"relative_source_dir"`
	buildDir    string   `toml:"relative_build_dir"`
	logDir      string   `toml:"relative_log_dir"`
	IncludeExts []string `toml:"watch_relative_ext"`
	ExcludeDirs []string `toml:"ignore_relative_dir"`
	IncludeDirs []string `toml:"watch_relative_dir"`
	IgnoreFiles []string `toml:"ignore_relative_files"`
	bufferTime  int      `toml:"delay"`
	Port        int      `toml:"port"`
	Root        string
}

func DefaultConfiguration() *Configuration {
	return &Configuration{
		sourceDir:   "cmd/web",
		buildDir:    "tmp",
		logDir:      "tmp",
		IncludeExts: []string{"go", "tpl", "tmpl", "html", "css", "js", "env", "yaml"},
		ExcludeDirs: []string{"assets", "tmp", "vendor", "node_modules", "build"},
		IncludeDirs: []string{},
		IgnoreFiles: []string{},
		bufferTime:  1000,
		Port:        3000,
		Root:        root,
	}
}

func TestConfiguration() *Configuration {
	cfg := DefaultConfiguration()
	cfg.IncludeExts = []string{}
	cfg.ExcludeDirs = []string{}
	return cfg
}

func (c *Configuration) BufferTime() time.Duration {
	return time.Duration(c.bufferTime) * time.Millisecond
}

func (c *Configuration) SrcDir() (string, error) {
	return utils.AbsolutePath(c.sourceDir)
}

func (c *Configuration) BuildDir() (string, error) {
	return utils.AbsolutePath(c.buildDir)
}

func (c *Configuration) LogDir() (string, error) {
	return utils.AbsolutePath(c.logDir)
}

func (c *Configuration) RemoveBuildDir() error {
	buildDir, err := c.BuildDir()
	if err != nil {
		return err
	}

	if err := utils.DeletePath(buildDir); err != nil {
		return err
	}

	return nil
}
