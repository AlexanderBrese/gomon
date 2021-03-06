package reload

import (
	"fmt"
	"testing"

	"github.com/AlexanderBrese/go-server-browser-reload/pkg/configuration"
	"github.com/AlexanderBrese/go-server-browser-reload/pkg/utils"
)

func TestBuild(t *testing.T) {
	cfg, err := configuration.TestConfiguration()
	if err != nil {
		t.Error(err)
	}
	reloader := NewReload(cfg)

	if err := buildPrepare(cfg); err != nil {
		t.Error(err)
	}

	defer buildCleanup(reloader)

	if err := buildStart(reloader); err != nil {
		t.Error(err)
	}

	if err := buildPassed(reloader.config); err != nil {
		t.Error(err)
	}
}

func buildPrepare(cfg *configuration.Configuration) error {
	srcDir, err := cfg.SrcDir()
	if err != nil {
		return err
	}
	buildDir, err := cfg.BuildDir()
	if err != nil {
		return err
	}

	return utils.PrepareBuild(srcDir, buildDir)
}

func buildStart(reloader *Reload) error {
	return reloader.build()
}

func buildPassed(cfg *configuration.Configuration) error {
	binary, err := cfg.Binary()
	if err != nil {
		return err
	}
	if err := utils.CheckPath(binary); err != nil {
		return fmt.Errorf("There was no built binary found at %s", binary)
	}
	return nil
}

func buildCleanup(reloader *Reload) error {
	reloader.BuildCleanup()
	cfg := reloader.Configuration()
	return utils.CleanupBuild(cfg.RelSrcDir(), cfg.RelBuildDir())
}
