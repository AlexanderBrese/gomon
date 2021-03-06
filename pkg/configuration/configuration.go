package configuration

import (
	"path/filepath"
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
	BuildName        string   `toml:"build_name"`
	LogName          string   `toml:"log_name"`
	RelSrcDir        string   `toml:"relative_source_dir"`
	RelBuildDir      string   `toml:"relative_build_dir"`
	RelLogDir        string   `toml:"relative_log_dir"`
	IncludeExts      []string `toml:"watch_relative_ext"`
	ExcludeDirs      []string `toml:"ignore_relative_dir"`
	IncludeDirs      []string `toml:"watch_relative_dir"`
	IgnoreFiles      []string `toml:"ignore_relative_files"`
	bufferTime       int
	Port             int `toml:"port"`
	Root             string
	ExecutionCommand string `toml:"execution_command"`
	BuildCommand     string `toml:"build_command"`
}

func DefaultConfiguration() *Configuration {
	return &Configuration{
		BuildName:        "main",
		LogName:          "GOATmon.log",
		RelSrcDir:        "cmd/web",
		RelBuildDir:      "tmp/build",
		RelLogDir:        "tmp",
		IncludeExts:      []string{"go", "tpl", "tmpl", "html", "css", "js", "env", "yaml"},
		ExcludeDirs:      []string{"assets", "tmp", "vendor", "node_modules", "build"},
		IncludeDirs:      []string{},
		IgnoreFiles:      []string{},
		bufferTime:       1000,
		ExecutionCommand: "",
		Port:             3000,
		Root:             root,
		BuildCommand:     "go build -o",
	}
}

func TestConfiguration() (*Configuration, error) {
	cfg := DefaultConfiguration()
	cfg.IncludeExts = []string{}
	cfg.ExcludeDirs = []string{"tmp", "build"}
	if err := adapt(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Configuration) BufferTime() time.Duration {
	return time.Duration(c.bufferTime) * time.Millisecond
}

func (c *Configuration) SrcDir() (string, error) {
	return utils.AbsolutePath(c.RelSrcDir)
}

func (c *Configuration) Binary() (string, error) {
	return utils.AbsolutePath(filepath.Join(c.RelBuildDir, c.BuildName))
}

func (c *Configuration) BuildDir() (string, error) {
	return utils.AbsolutePath(c.RelBuildDir)
}

func (c *Configuration) LogDir() (string, error) {
	return utils.AbsolutePath(c.RelLogDir)
}

func (c *Configuration) Log() (string, error) {
	return utils.AbsolutePath(filepath.Join(c.RelLogDir, c.LogName))
}
