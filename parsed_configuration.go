package main

import (
	"github.com/imdario/mergo"
	"github.com/pelletier/go-toml"
)

// ParsedConfiguration parses a configuration file and merges it with the default configuration
func ParsedConfiguration(path string) (*Configuration, error) {
	if path == "" {
		return defaultConfiguration, nil
	} else if err := checkPath(path); err != nil {
		return nil, err
	} else {
		cfg, err := _parse(path)
		if err != nil {
			return nil, err
		}
		err = _merge(cfg)
		if err != nil {
			return nil, err
		}
		return cfg, err
	}
}

func _parse(path string) (cfg *Configuration, err error) {
	cfgData, err := readFile(path)
	if err != nil {
		return nil, err
	}
	cfg, err = _unmarshal(cfgData)
	if err != nil {
		return nil, err
	}
	err = _validate(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, err
}

func _validate(cfg *Configuration) error {
	absPath, err := absolutePath(cfg.SourceDir)
	if err != nil {
		return err
	}
	return checkPath(absPath)
}

func _merge(cfg *Configuration) error {
	return mergo.Merge(cfg, defaultConfiguration)
}

func _unmarshal(cfgData []byte) (*Configuration, error) {
	cfg := new(Configuration)
	if err := toml.Unmarshal(cfgData, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
