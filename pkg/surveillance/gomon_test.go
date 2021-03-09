package surveillance

import (
	"errors"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AlexanderBrese/gomon/pkg/configuration"
	"github.com/AlexanderBrese/gomon/pkg/utils"
)

const (
	tempFileCreationDelay  = 300
	tempFileContent        = "test"
	changeDetectionTimeout = 800
)

type Test struct {
	name             string
	cfg              *configuration.Configuration
	relPath          string
	shouldBeDetected bool
}

func TestGomon(t *testing.T) {
	tests := testSuite()
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(tests), func(i, j int) { tests[i], tests[j] = tests[j], tests[i] })
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := detect(tt.cfg, tt.relPath, tt.shouldBeDetected); err != nil {
				t.Error(err)
			}
		})
	}
}

func testSuite() []Test {
	defaultCfg := configuration.DefaultConfiguration()
	defaultCfg.Reload = false
	defaultCfg.Sync = false
	customExtsCfg, _ := configuration.TestConfiguration()
	customExtsCfg.Filter.IncludeExts = append(customExtsCfg.Filter.IncludeExts, "custom")
	customIgnoredDirCfg, _ := configuration.TestConfiguration()
	customIgnoredDirCfg.Filter.ExcludeDirs = append(customIgnoredDirCfg.Filter.ExcludeDirs, "ignored")
	customIgnoredDirCfg.Filter.IncludeExts = append(customIgnoredDirCfg.Filter.IncludeExts, "go")
	customIgnoredFileCfg, _ := configuration.TestConfiguration()
	customIgnoredFileCfg.Filter.ExcludeFiles = append(customIgnoredFileCfg.Filter.ExcludeFiles, "ignored.go")
	customIgnoredFileCfg.Filter.IncludeExts = append(customIgnoredFileCfg.Filter.IncludeExts, "go")
	customIncludeDirCfg, _ := configuration.TestConfiguration()
	customIncludeDirCfg.Filter.IncludeDirs = append(customIncludeDirCfg.Filter.IncludeDirs, "watched")
	customIncludeDirCfg.Filter.IncludeExts = append(customIncludeDirCfg.Filter.IncludeExts, "go")
	customIgnoredAndIncludedDirCfg, _ := configuration.TestConfiguration()
	customIgnoredAndIncludedDirCfg.Filter.ExcludeDirs = append(customIgnoredAndIncludedDirCfg.Filter.ExcludeDirs, "watched")
	customIgnoredAndIncludedDirCfg.Filter.IncludeDirs = append(customIgnoredAndIncludedDirCfg.Filter.IncludeDirs, "watched")
	customIgnoredAndIncludedDirCfg.Filter.IncludeExts = append(customIgnoredAndIncludedDirCfg.Filter.IncludeExts, "go")
	customIgnoredFileAndWatchedDirCfg, _ := configuration.TestConfiguration()
	customIgnoredFileAndWatchedDirCfg.Filter.IncludeDirs = append(customIgnoredFileAndWatchedDirCfg.Filter.IncludeDirs, "watched")
	customIgnoredFileAndWatchedDirCfg.Filter.ExcludeFiles = append(customIgnoredFileAndWatchedDirCfg.Filter.ExcludeFiles, "watched/ignored.go")
	customIgnoredFileAndWatchedDirCfg.Filter.IncludeExts = append(customIgnoredFileAndWatchedDirCfg.Filter.IncludeExts, "go")
	customWatchedDirAndWatchedExt, _ := configuration.TestConfiguration()
	customWatchedDirAndWatchedExt.Filter.IncludeDirs = append(customWatchedDirAndWatchedExt.Filter.IncludeDirs, "watched")
	customWatchedDirAndWatchedExt.Filter.IncludeExts = append(customWatchedDirAndWatchedExt.Filter.IncludeExts, "go")

	return []Test{
		{"Files in an ignored folder should not be detected.", customIgnoredDirCfg, "ignored/test.go", false},
		{"A file in an ignored and watched dir should not be detected.", customIgnoredAndIncludedDirCfg, "watched/test.go", false},
		{"An ignored file inside a watched directory should not be detected.", customIgnoredFileAndWatchedDirCfg, "watched/ignored.go", false},
		{"A file with a valid extension should be detected.", defaultCfg, "test.go", true},
		{"A file with a valid extension inside a watched directory should be detected.", customWatchedDirAndWatchedExt, "watched/test.go", true},
		{"A file with an invalid extension inside a watched directory should not be detected.", customWatchedDirAndWatchedExt, "watched/test.custom", false},
		{"Files outside the ignored folder should be detected.", customIgnoredDirCfg, "test.go", true},
		{"A file with a valid custom extension should be detected.", customExtsCfg, "test.custom", true},
		{"A file with an invalid custom extension should not be detected.", customExtsCfg, "test.go", false},
		{"A file in a watched directory should be detected.", customIncludeDirCfg, "watched/test.go", true},
		{"A file outside of a watched directory should not be detected.", customIncludeDirCfg, "other/test.go", false},
		{"An ignored file should not be detected.", customIgnoredFileCfg, "ignored.go", false},
	}
}

func detect(cfg *configuration.Configuration, relFile string, shouldBeDetected bool) error {
	gomon := NewGomon(cfg)
	if gomon == nil {
		return errors.New("error: during gomon initialization")
	}
	subscription := subscribe(gomon)
	go gomon.Start()

	file := filepath.Join(cfg.Root, relFile)
	defer func() {
		if err := cleanup(file, relFile, subscription, gomon); err != nil {
			// TODO: log
			return
		}
	}()

	if err := do(relFile, file); err != nil {
		return err
	}
	if err := check(subscription, shouldBeDetected); err != nil {
		return err
	}

	return nil
}

func cleanup(file string, relFile string, sub chan bool, g *Gomon) error {
	g.Stop()
	close(sub)
	if isInsideDir(relFile) {
		dir, err := dir(relFile)
		if err != nil {
			return err
		}
		if err := delete(dir); err != nil {
			return err
		}
	} else {
		if err := delete(file); err != nil {
			return err
		}
	}
	return nil
}

func subscribe(cd *Gomon) chan bool {
	subscription := make(chan bool, 1)
	cd.Subscribe(subscription)
	return subscription
}

func delete(changedFile string) error {
	return utils.RemoveAllDir(changedFile)
}

func do(relFile string, file string) error {
	if isInsideDir(relFile) {
		dir, err := dir(relFile)
		if err != nil {
			return err
		}
		go func() {
			if err := createTemporaryDirectoryAndFile(dir, file); err != nil {
				//TODO: log
				return
			}
		}()
	} else {
		go func() {
			if err := createTemporaryFile(file); err != nil {
				//TODO: log
				return
			}
		}()
	}

	return nil
}

func dir(path string) (string, error) {
	relDir := strings.Split(path, "/")[0]
	dir, err := utils.CurrentAbsolutePath(relDir)
	if err != nil {
		return "", err
	}
	return dir, nil
}

func isInsideDir(path string) bool {
	return strings.Contains(path, "/")
}

func createTemporaryDirectoryAndFile(dir string, file string) error {
	if err := createTemporaryDirectory(dir); err != nil {
		return err
	}
	time.Sleep(tempFileCreationDelay * time.Millisecond)
	if err := createTemporaryFile(file); err != nil {
		return err
	}

	return nil
}

func createTemporaryDirectory(path string) error {
	if err := utils.CreateAllDir(path); err != nil {
		return err
	}

	return nil
}

func createTemporaryFile(path string) error {
	if _, err := utils.CreateFile(path, []byte(tempFileContent)); err != nil {
		return err
	}

	return nil
}

func check(sub chan bool, shouldBeDetected bool) error {
	for {
		select {
		case detected, ok := <-sub:
			if !ok {
				return nil
			}
			if !shouldBeDetected && detected {
				return errors.New("error: expected no change detection got change detection")
			}
			if shouldBeDetected && !detected {
				return errors.New("error: expected change detection got no change detection")
			}
			return nil
		case <-time.After(changeDetectionTimeout * time.Millisecond):
			return nil
		}
	}
}
