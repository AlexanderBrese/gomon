package configuration

import (
	"testing"

	"github.com/AlexanderBrese/gomon/pkg/utils"
	"github.com/pelletier/go-toml"
)

func TestNoConfigProvided(t *testing.T) {
	parsed, err := ParsedConfiguration("")
	if err != nil {
		t.Errorf("want: config, got: %q", err)
	}
	if parsed.Root == "" {
		t.Error("want: root to contain a value, got: nothing")
	}
}

func TestWrongConfigName(t *testing.T) {
	if _, err := ParsedConfiguration("no/file"); err == nil {
		t.Errorf("want: error, got: %q", err)
	}
}

func TestInvalidSourcePathProvided(t *testing.T) {
	testCfg := &Configuration{
		Build: &BuildConfiguration{
			RelSrcDir: "wrong_source_dir",
		},
	}
	cfgData, err := toml.Marshal(testCfg)
	if err != nil {
		t.Error(err)
	}

	dir := "test"
	absDir, err := utils.CurrentAbsolutePath(dir)
	if err != nil {
		t.Error(err)
	}
	err = utils.CreateAllDir(absDir)
	if err != nil {
		t.Error(err)
	}
	defer func() {
		if err := utils.RemoveAllDir(absDir); err != nil {
			t.Error(err)
		}
	}()

	path := dir + "/test.toml"
	absPath, err := utils.CurrentAbsolutePath(path)
	if err != nil {
		t.Error(err)
	}
	if _, err = utils.CreateFile(absPath, cfgData); err != nil {
		t.Error(err)
	}
	if _, err = ParsedConfiguration(absPath); err == nil {
		t.Errorf("want: error, got: %q", err)
	}
}

func TestConfigMerge(t *testing.T) {
	port := 4000
	testCfg := &Configuration{
		Build: &BuildConfiguration{
			Port: port,
		},
	}
	testCfgData, err := toml.Marshal(testCfg)
	if err != nil {
		t.Error(err)
	}

	path := "test.toml"
	absPath, err := utils.CurrentAbsolutePath(path)
	if err != nil {
		t.Error(err)
	}
	if _, err := utils.CreateFile(absPath, testCfgData); err != nil {
		t.Error(err)
	}

	defer func() {
		if err := utils.RemoveAllDir(absPath); err != nil {
			t.Error(err)
		}
	}()

	cfg, err := ParsedConfiguration(absPath)
	if err != nil {
		t.Error(err)
	}
	if cfg.Build.Port != port {
		t.Errorf("want: %q, got: %q", port, cfg.Build.Port)
	}
}
