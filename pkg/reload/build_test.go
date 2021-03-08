package reload

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/AlexanderBrese/GOATmon/pkg/configuration"
	"github.com/AlexanderBrese/GOATmon/pkg/utils"
)

const (
	testFile        = "test.go"
	testFileContent = `package main
	import "fmt"
	func main() {
		fmt.Println("hello world")
	} 
`
)

func TestBuild(t *testing.T) {
	cfg, err := configuration.TestConfiguration()
	cfg.RelSrcDir = "cmd/web"

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

func prepareBuild(srcDir string, buildDir string) error {
	if err := createSourceDir(srcDir); err != nil {
		return err
	}
	if err := createSourceFile(srcDir); err != nil {
		return err
	}
	return utils.CreateBuildDirIfNotExist(buildDir)
}

func cleanupBuild(relSrcDir string, relBuildDir string) error {
	if err := removeSourceDir(relSrcDir); err != nil {
		return err
	}
	return utils.RemoveRootBuildDir(relBuildDir)
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

	return prepareBuild(srcDir, buildDir)
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
	cfg := reloader.config
	return cleanupBuild(cfg.RelSrcDir, cfg.RelBuildDir)
}

func removeSourceDir(relSrcDir string) error {
	return utils.RemoveRootDir(relSrcDir)
}

func createSourceDir(srcDir string) error {
	return utils.CreateAllDir(srcDir)
}

func createSourceFile(srcDir string) error {
	_, err := utils.CreateFile(filepath.Join(srcDir, testFile), []byte(testFileContent))
	return err
}
