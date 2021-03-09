package configuration

import (
	"path/filepath"
	"time"

	"github.com/AlexanderBrese/gomon/pkg/utils"
)

var root string

func init() {
	var err error
	root, err = utils.CurrentRootPath()
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
	IncludeExts      []string `toml:"include_exts"`
	ExcludeDirs      []string `toml:"exclude_relative_dirs"`
	IncludeDirs      []string `toml:"include_relative_dirs"`
	ExcludeFiles     []string `toml:"exclude_relative_files"`
	EventBufferTime  int
	KillDelay        int
	Port             int `toml:"port"`
	Root             string
	ExecutionCommand string `toml:"execution_command"`
	BuildCommand     string `toml:"build_command"`
	Reload           bool
	Sync             bool
}

// DefaultConfiguration is the default configuration if none is provided
func DefaultConfiguration() *Configuration {
	return &Configuration{
		BuildName:        "main",
		LogName:          "gomon.log",
		RelSrcDir:        "",
		RelBuildDir:      "tmp/build",
		RelLogDir:        "tmp",
		IncludeExts:      []string{"go", "tpl", "tmpl", "html", "css", "js", "env", "yaml"},
		ExcludeDirs:      []string{"assets", "tmp", "vendor", "node_modules", "build"},
		IncludeDirs:      []string{},
		ExcludeFiles:     []string{},
		EventBufferTime:  100,
		KillDelay:        100,
		ExecutionCommand: "",
		Port:             3000,
		Root:             root,
		BuildCommand:     "go build -o",
		Reload:           true,
		Sync:             true,
	}
}

// TestConfiguration is the configuration used for internal tests
func TestConfiguration() (*Configuration, error) {
	cfg := DefaultConfiguration()
	cfg.IncludeExts = []string{}
	cfg.ExcludeDirs = []string{"tmp", "build"}

	cfg.Reload = false
	cfg.Sync = false

	if err := adapt(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// BufferTime is the event buffer time in milliseconds
func (c *Configuration) BufferTime() time.Duration {
	return time.Duration(c.EventBufferTime) * time.Millisecond
}

// SrcDir is the current absolute source directory path
func (c *Configuration) SrcDir() (string, error) {
	return utils.CurrentAbsolutePath(c.RelSrcDir)
}

// BuildDir is the current absolute build directory path
func (c *Configuration) BuildDir() (string, error) {
	return utils.CurrentAbsolutePath(c.RelBuildDir)
}

// LogDir is the current absolute log directory path
func (c *Configuration) LogDir() (string, error) {
	return utils.CurrentAbsolutePath(c.RelLogDir)
}

// Binary is the current absolute binary path
func (c *Configuration) Binary() (string, error) {
	return utils.CurrentAbsolutePath(filepath.Join(c.RelBuildDir, c.BuildName))
}

// Log is the current absolute log path
func (c *Configuration) Log() (string, error) {
	return utils.CurrentAbsolutePath(filepath.Join(c.RelLogDir, c.LogName))
}
