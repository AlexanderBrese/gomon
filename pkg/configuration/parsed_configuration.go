package configuration

import (
	"github.com/AlexanderBrese/go-server-browser-reload/pkg/utils"
	"github.com/imdario/mergo"
	"github.com/pelletier/go-toml"
)

// ParsedConfiguration parses a configuration file and merges it with the default configuration
func ParsedConfiguration(path string) (*Configuration, error) {
	if path == "" {
		return defaultConfiguration, nil
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
	absPath, err := utils.AbsolutePath(cfg.sourceDir)
	if err != nil {
		return err
	}
	return utils.CheckPath(absPath)
}

func merge(cfg *Configuration) error {
	return mergo.Merge(cfg, defaultConfiguration)
}

func unmarshal(cfgData []byte) (*Configuration, error) {
	cfg := new(Configuration)
	if err := toml.Unmarshal(cfgData, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}