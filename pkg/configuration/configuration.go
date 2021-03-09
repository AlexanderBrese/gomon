package configuration

import (
	"path/filepath"
	"time"

	"github.com/AlexanderBrese/gomon/pkg/utils"
	"github.com/fatih/color"
)

var root string

func init() {
	var err error
	root, err = utils.CurrentRootPath()
	if err != nil {
		panic(err)
	}
}

type LogConfiguration struct {
	BuildLog       string `toml:"build_log_name"`
	RelBuildLogDir string `toml:"relative_build_log_dir"`
	Time           bool   `toml:"time"`
	Main           bool   `toml:"main"`
	Detection      bool   `toml:"detection"`
	Build          bool   `toml:"build"`
	Run            bool   `toml:"run"`
	Sync           bool   `toml:"sync"`
	App            bool   `toml:"app"`
}

type ColorConfiguration struct {
	Main      string `toml:"main"`
	Detection string `toml:"detection"`
	Build     string `toml:"build"`
	Run       string `toml:"run"`
	Sync      string `toml:"sync"`
	App       string `toml:"app"`
}

type BuildConfiguration struct {
	Name             string `toml:"build_name"`
	RelDir           string `toml:"relative_build_dir"`
	RelSrcDir        string `toml:"relative_source_dir"`
	ExecutionCommand string `toml:"execution_command"`
	Command          string `toml:"build_command"`
	EventBufferTime  int
	KillDelay        int
	Port             int `toml:"port"`
}

type FilterConfiguration struct {
	IncludeExts  []string `toml:"include_exts"`
	ExcludeDirs  []string `toml:"exclude_relative_dirs"`
	IncludeDirs  []string `toml:"include_relative_dirs"`
	ExcludeFiles []string `toml:"exclude_relative_files"`
}

// Configuration is a in-memory representation of the expected configuration file
type Configuration struct {
	Root   string
	Reload bool
	Sync   bool
	Build  *BuildConfiguration  `toml:"build"`
	Log    *LogConfiguration    `toml:"log"`
	Color  *ColorConfiguration  `toml:"color"`
	Filter *FilterConfiguration `toml:"filter"`
}

// DefaultConfiguration is the default configuration if none is provided
func DefaultConfiguration() *Configuration {
	return &Configuration{
		Root:   root,
		Reload: true,
		Sync:   true,
		Build: &BuildConfiguration{
			Name:             "main",
			RelDir:           "tmp/build",
			RelSrcDir:        "",
			EventBufferTime:  100,
			KillDelay:        100,
			ExecutionCommand: "",
			Port:             3000,
			Command:          "go build -o",
		},
		Log: &LogConfiguration{
			BuildLog:       "gomon.log",
			RelBuildLogDir: "tmp",
			Main:           true,
			Detection:      false,
			Build:          true,
			Run:            false,
			Sync:           false,
			App:            true,
			Time:           true,
		},
		Color: &ColorConfiguration{
			Main:      "red",
			Detection: "magenta",
			Build:     "yellow",
			Run:       "green",
			Sync:      "cyan",
			App:       "blue",
		},
		Filter: &FilterConfiguration{
			IncludeExts:  []string{"go", "tpl", "tmpl", "html", "css", "js", "env", "yaml"},
			ExcludeDirs:  []string{"assets", "tmp", "vendor", "node_modules", "build"},
			IncludeDirs:  []string{},
			ExcludeFiles: []string{},
		},
	}
}

// TestConfiguration is the configuration used for internal tests
func TestConfiguration() (*Configuration, error) {
	cfg := DefaultConfiguration()
	cfg.Filter.IncludeExts = []string{}
	cfg.Filter.ExcludeDirs = []string{"tmp", "build"}

	cfg.Reload = false
	cfg.Sync = false

	if err := adapt(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Configuration) Colors() map[string]*color.Color {
	return map[string]*color.Color{
		"Main":      utils.Color(c.Color.Main),
		"Build":     utils.Color(c.Color.Build),
		"Run":       utils.Color(c.Color.Run),
		"Detection": utils.Color(c.Color.Detection),
		"Sync":      utils.Color(c.Color.Sync),
		"App":       utils.Color(c.Color.App),
	}
}

// BufferTime is the event buffer time in milliseconds
func (c *Configuration) BufferTime() time.Duration {
	return time.Duration(c.Build.EventBufferTime) * time.Millisecond
}

// SrcDir is the current absolute source directory path
func (c *Configuration) SrcDir() (string, error) {
	return utils.CurrentAbsolutePath(c.Build.RelSrcDir)
}

// BuildDir is the current absolute build directory path
func (c *Configuration) BuildDir() (string, error) {
	return utils.CurrentAbsolutePath(c.Build.RelDir)
}

// BuildLogDir is the current absolute log directory path
func (c *Configuration) BuildLogDir() (string, error) {
	return utils.CurrentAbsolutePath(c.Log.RelBuildLogDir)
}

// Binary is the current absolute binary path
func (c *Configuration) Binary() (string, error) {
	return utils.CurrentAbsolutePath(filepath.Join(c.Build.RelDir, c.Build.Name))
}

// Log is the current absolute log path
func (c *Configuration) BuildLog() (string, error) {
	return utils.CurrentAbsolutePath(filepath.Join(c.Log.RelBuildLogDir, c.Log.BuildLog))
}
