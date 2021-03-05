package surveillance

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlexanderBrese/go-server-browser-reload/pkg/configuration"
	"github.com/AlexanderBrese/go-server-browser-reload/pkg/utils"
)

func TestFileChanges(t *testing.T) {

	defaultCfg := configuration.DefaultConfiguration()
	customExtsCfg := configuration.TestConfiguration()
	customExtsCfg.WatchExts = []string{"custom"}
	customIgnoredDirCfg := configuration.TestConfiguration()
	customIgnoredDirCfg.IgnoreDirs = []string{"ignored"}
	customIgnoredDirCfg.WatchExts = []string{"go"}
	customIgnoredFileCfg := configuration.TestConfiguration()
	customIgnoredFileCfg.IgnoreFiles = []string{"ignored.go"}
	customIgnoredFileCfg.WatchExts = []string{"go"}
	customIncludeDirCfg := configuration.TestConfiguration()
	customIncludeDirCfg.WatchDirs = []string{"watched"}
	customIncludeDirCfg.WatchExts = []string{"go"}
	customIgnoredAndIncludedDirCfg := configuration.TestConfiguration()
	customIgnoredAndIncludedDirCfg.IgnoreDirs = []string{"watched"}
	customIgnoredAndIncludedDirCfg.WatchDirs = []string{"watched"}
	customIgnoredAndIncludedDirCfg.WatchExts = []string{"go"}
	customIgnoredFileAndWatchedDirCfg := configuration.TestConfiguration()
	customIgnoredFileAndWatchedDirCfg.WatchDirs = []string{"watched"}
	customIgnoredFileAndWatchedDirCfg.IgnoreFiles = []string{"watched/ignored.go"}
	customIgnoredFileAndWatchedDirCfg.WatchExts = []string{"go"}
	customWatchedDirAndWatchedExt := configuration.TestConfiguration()
	customWatchedDirAndWatchedExt.WatchDirs = []string{"watched"}
	customWatchedDirAndWatchedExt.WatchExts = []string{"go"}

	tests := []struct {
		name             string
		cfg              *configuration.Configuration
		relPath          string
		shouldBeDetected bool
	}{
		{"A file with a valid extension should be detected.", defaultCfg, "test.go", true},

		{"An ignored file inside a watched directory should not be detected.", customIgnoredFileAndWatchedDirCfg, "watched/ignored.go", false},

		{"A file with an invalid extension inside a watched directory should not be detected.", customWatchedDirAndWatchedExt, "watched/test.custom", false},
		{"A file with a valid extension inside a watched directory should be detected.", customWatchedDirAndWatchedExt, "watched/test.go", true},

		{"A file with a valid custom extension should be detected.", customExtsCfg, "test.custom", true},
		{"A file with an invalid custom extension should not be detected.", customExtsCfg, "test.go", false},

		{"A file in a watched directory should be detected.", customIncludeDirCfg, "watched/test.go", true},

		{"Files outside the ignored folder should be detected.", customIgnoredDirCfg, "test.go", true},
		{"Files in an ignored folder should not be detected.", customIgnoredDirCfg, "ignored/test.go", false},

		{"An ignored file should not be detected.", customIgnoredFileCfg, "ignored.go", false},

		{"A file in an ignored and watched dir should not be detected.", customIgnoredAndIncludedDirCfg, "watched/test.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := watch(tt.cfg, tt.relPath, tt.shouldBeDetected); err != nil {
				t.Error(err)
			}
		})
	}
}

func watch(cfg *configuration.Configuration, relPath string, shouldBeDetected bool) error {
	fullPath := filepath.Join(cfg.Root, relPath)
	fileChangesSubscription := make(chan string, 1)
	fileChanges, err := NewFileChanges(cfg)
	if err != nil {
		return fmt.Errorf("error during file changes creation: %s", err)
	}

	fileChanges.Subscribe(fileChangesSubscription)
	if err := fileChanges.Init(); err != nil {
		return err
	}

	isInsideDir := strings.Contains(relPath, "/")
	go fileChanges.Watch()
	if isInsideDir {
		dir := strings.Split(relPath, "/")[0]
		var dirFullPath string
		if dirFullPath, err = utils.AbsolutePath(dir); err != nil {
			return err
		}
		if err := utils.CreateDir(dirFullPath); err != nil {
			return err
		}
		if err := createTempFile(fullPath); err != nil {
			return err
		}
		defer utils.DeletePath(dirFullPath)
	} else {
		if err := createTempFile(fullPath); err != nil {
			return err
		}
		defer utils.DeleteFile(fullPath)
	}

	for {
		select {
		case watchedFilePath := <-fileChangesSubscription:
			close(fileChangesSubscription)
			fileChanges.StopWatching()
			butWasDetected := watchedFilePath == fullPath
			if !shouldBeDetected && butWasDetected {
				return fmt.Errorf("error: a file change should not be detected at %s", fullPath)
			}
			if shouldBeDetected && !butWasDetected {
				return fmt.Errorf("error: a file change should be detected at %s", fullPath)
			}

			return nil
		}

	}
}

func createTempFile(path string) error {
	if _, err := utils.CreateFile(path, []byte("test")); err != nil {
		return fmt.Errorf("error: can not create temporary file: %s", err)
	}

	return nil
}
