package configuration

import (
	"runtime"
	"strings"

	"github.com/AlexanderBrese/gomon/pkg/utils"
	"github.com/imdario/mergo"
	"github.com/pelletier/go-toml"
)

// ParsedConfiguration is a parsed configuration merged with the default configuration and adapted to the OS
func ParsedConfiguration(path string) (*Configuration, error) {
	if path == "" {
		cfg := DefaultConfiguration()
		if err := adapt(cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	} else if err := utils.CheckPath(path); err != nil {
		return nil, err
	} else {
		cfg, err := parse(path)
		if err != nil {
			return nil, err
		}
		err = merge(cfg)
		if err != nil {
			return nil, err
		}
		if err := adapt(cfg); err != nil {
			return nil, err
		}
		return cfg, err
	}
}

func parse(path string) (cfg *Configuration, err error) {
	cfgData, err := utils.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg, err = unmarshal(cfgData)
	if err != nil {
		return nil, err
	}
	err = validate(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, err
}

func validate(cfg *Configuration) error {
	absPath, err := utils.CurrentAbsolutePath(cfg.RelSrcDir)
	if err != nil {
		return err
	}
	return utils.CheckPath(absPath)
}

func merge(cfg *Configuration) error {
	return mergo.Merge(cfg, DefaultConfiguration())
}

// Adapt to OS
func adapt(cfg *Configuration) error {
	if runtime.GOOS == PlatformWindows {
		extName := ".exe"
		if !strings.HasSuffix(cfg.BuildName, extName) {
			cfg.BuildName += extName
		}
	}
	binary, err := cfg.Binary()
	if err != nil {
		return err
	}
	cfg.ExecutionCommand = binary
	return nil
}

func unmarshal(cfgData []byte) (*Configuration, error) {
	cfg := new(Configuration)
	if err := toml.Unmarshal(cfgData, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
