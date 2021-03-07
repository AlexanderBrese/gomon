package surveillance

import (
	"errors"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AlexanderBrese/go-server-browser-reload/pkg/configuration"
	"github.com/AlexanderBrese/go-server-browser-reload/pkg/utils"
)

const TEMP_FILE_CREATION_DELAY = 100
const TEMP_FILE_CONTENT = "test"

func TestFileChangesDetection(t *testing.T) {
	defaultCfg := configuration.DefaultConfiguration()
	defaultCfg.Reload = false
	defaultCfg.Sync = false
	customExtsCfg, _ := configuration.TestConfiguration()
	customExtsCfg.IncludeExts = append(customExtsCfg.IncludeExts, "custom")
	customIgnoredDirCfg, _ := configuration.TestConfiguration()
	customIgnoredDirCfg.ExcludeDirs = append(customIgnoredDirCfg.ExcludeDirs, "ignored")
	customIgnoredDirCfg.IncludeExts = append(customIgnoredDirCfg.IncludeExts, "go")
	customIgnoredFileCfg, _ := configuration.TestConfiguration()
	customIgnoredFileCfg.IgnoreFiles = append(customIgnoredFileCfg.IgnoreFiles, "ignored.go")
	customIgnoredFileCfg.IncludeExts = append(customIgnoredFileCfg.IncludeExts, "go")
	customIncludeDirCfg, _ := configuration.TestConfiguration()
	customIncludeDirCfg.IncludeDirs = append(customIncludeDirCfg.IncludeDirs, "watched")
	customIncludeDirCfg.IncludeExts = append(customIncludeDirCfg.IncludeExts, "go")
	customIgnoredAndIncludedDirCfg, _ := configuration.TestConfiguration()
	customIgnoredAndIncludedDirCfg.ExcludeDirs = append(customIgnoredAndIncludedDirCfg.ExcludeDirs, "watched")
	customIgnoredAndIncludedDirCfg.IncludeDirs = append(customIgnoredAndIncludedDirCfg.IncludeDirs, "watched")
	customIgnoredAndIncludedDirCfg.IncludeExts = append(customIgnoredAndIncludedDirCfg.IncludeExts, "go")
	customIgnoredFileAndWatchedDirCfg, _ := configuration.TestConfiguration()
	customIgnoredFileAndWatchedDirCfg.IncludeDirs = append(customIgnoredFileAndWatchedDirCfg.IncludeDirs, "watched")
	customIgnoredFileAndWatchedDirCfg.IgnoreFiles = append(customIgnoredFileAndWatchedDirCfg.IgnoreFiles, "watched/ignored.go")
	customIgnoredFileAndWatchedDirCfg.IncludeExts = append(customIgnoredFileAndWatchedDirCfg.IncludeExts, "go")
	customWatchedDirAndWatchedExt, _ := configuration.TestConfiguration()
	customWatchedDirAndWatchedExt.IncludeDirs = append(customWatchedDirAndWatchedExt.IncludeDirs, "watched")
	customWatchedDirAndWatchedExt.IncludeExts = append(customWatchedDirAndWatchedExt.IncludeExts, "go")

	tests := []struct {
		name             string
		cfg              *configuration.Configuration
		relPath          string
		shouldBeDetected bool
	}{

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
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(tests), func(i, j int) { tests[i], tests[j] = tests[j], tests[i] })
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := fileChanges(tt.cfg, tt.relPath, tt.shouldBeDetected); err != nil {
				t.Error(err)
			}
		})
	}
}

func fileChanges(cfg *configuration.Configuration, relFile string, shouldBeDetected bool) error {
	file := filepath.Join(cfg.Root, relFile)
	fileChanges, err := NewFileChangesDetection(cfg)
	if err != nil {
		return err
	}
	if err := prepare(fileChanges); err != nil {
		return err
	}
	changeDetectedSubscription := subscribe(fileChanges)

	defer cleanup(file, relFile, changeDetectedSubscription, fileChanges)

	if err := do(relFile, file); err != nil {
		return err
	}
	if err := check(changeDetectedSubscription, shouldBeDetected); err != nil {
		return err
	}

	return nil
}

func cleanup(file string, relFile string, changeDetectedSubscription chan bool, fileChanges *FileChangesDetection) error {
	fileChanges.StopWatching()
	close(changeDetectedSubscription)
	if isInsideDir(relFile) {
		dir, err := dir(relFile)
		if err != nil {
			return err
		}
		delete(dir)
	} else {
		delete(file)
	}
	return nil
}

func subscribe(fileChanges *FileChangesDetection) chan bool {
	fileChangesSubscription := make(chan bool, 1)
	fileChanges.Subscribe(fileChangesSubscription)
	return fileChangesSubscription
}

func prepare(fileChanges *FileChangesDetection) error {
	if err := fileChanges.Init(); err != nil {
		return err
	}

	go fileChanges.Surveil()

	return nil
}

func delete(changedFile string) error {
	return utils.RemoveDir(changedFile)
}

func do(relFile string, file string) error {
	if isInsideDir(relFile) {
		dir, err := dir(relFile)
		if err != nil {
			return err
		}
		go createTemporaryDirectoryAndFile(dir, file)
	} else {
		go createTemporaryFile(file)
	}

	return nil
}

func dir(path string) (string, error) {
	relDir := strings.Split(path, "/")[0]
	dir, err := utils.AbsolutePath(relDir)
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
	time.Sleep(300 * time.Millisecond)
	if err := createTemporaryFile(file); err != nil {
		return err
	}

	return nil
}

func createTemporaryDirectory(path string) error {
	if err := utils.CreateDir(path); err != nil {
		return err
	}

	return nil
}

func createTemporaryFile(path string) error {
	if _, err := utils.CreateFile(path, []byte(TEMP_FILE_CONTENT)); err != nil {
		return err
	}

	return nil
}

func check(changeDetectedSubscription chan bool, shouldBeDetected bool) error {
	for {
		select {
		case changeDetected, ok := <-changeDetectedSubscription:
			if !ok {
				return nil
			}
			if !shouldBeDetected && changeDetected {
				return errors.New("error: expected no change detection got change detection")
			}
			if shouldBeDetected && !changeDetected {
				return errors.New("error: expected change detection got no change detection")
			}
			return nil
		case <-time.After(800 * time.Millisecond):
			return nil
		}
	}
}
