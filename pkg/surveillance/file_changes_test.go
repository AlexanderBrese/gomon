package surveillance

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AlexanderBrese/go-server-browser-reload/pkg/configuration"
	"github.com/AlexanderBrese/go-server-browser-reload/pkg/utils"
)

const TEMP_FILE_CREATION_DELAY = 100
const TEMP_FILE_CONTENT = "test"

func TestFileChanges(t *testing.T) {
	defaultCfg := configuration.DefaultConfiguration()
	customExtsCfg := configuration.TestConfiguration()
	customExtsCfg.IncludeExts = []string{"custom"}
	customIgnoredDirCfg := configuration.TestConfiguration()
	customIgnoredDirCfg.ExcludeDirs = []string{"ignored"}
	customIgnoredDirCfg.IncludeExts = []string{"go"}
	customIgnoredFileCfg := configuration.TestConfiguration()
	customIgnoredFileCfg.IgnoreFiles = []string{"ignored.go"}
	customIgnoredFileCfg.IncludeExts = []string{"go"}
	customIncludeDirCfg := configuration.TestConfiguration()
	customIncludeDirCfg.IncludeDirs = []string{"watched"}
	customIncludeDirCfg.IncludeExts = []string{"go"}
	customIgnoredAndIncludedDirCfg := configuration.TestConfiguration()
	customIgnoredAndIncludedDirCfg.ExcludeDirs = []string{"watched"}
	customIgnoredAndIncludedDirCfg.IncludeDirs = []string{"watched"}
	customIgnoredAndIncludedDirCfg.IncludeExts = []string{"go"}
	customIgnoredFileAndWatchedDirCfg := configuration.TestConfiguration()
	customIgnoredFileAndWatchedDirCfg.IncludeDirs = []string{"watched"}
	customIgnoredFileAndWatchedDirCfg.IgnoreFiles = []string{"watched/ignored.go"}
	customIgnoredFileAndWatchedDirCfg.IncludeExts = []string{"go"}
	customWatchedDirAndWatchedExt := configuration.TestConfiguration()
	customWatchedDirAndWatchedExt.IncludeDirs = []string{"watched"}
	customWatchedDirAndWatchedExt.IncludeExts = []string{"go"}

	tests := []struct {
		name             string
		cfg              *configuration.Configuration
		relPath          string
		shouldBeDetected bool
	}{
		{"An ignored file inside a watched directory should not be detected.", customIgnoredFileAndWatchedDirCfg, "watched/ignored.go", false},
		{"A file with a valid extension should be detected.", defaultCfg, "test.go", true},

		{"A file with a valid extension inside a watched directory should be detected.", customWatchedDirAndWatchedExt, "watched/test.go", true},
		{"A file with an invalid extension inside a watched directory should not be detected.", customWatchedDirAndWatchedExt, "watched/test.custom", false},

		{"Files outside the ignored folder should be detected.", customIgnoredDirCfg, "test.go", true},
		{"Files in an ignored folder should not be detected.", customIgnoredDirCfg, "ignored/test.go", false},

		{"A file with a valid custom extension should be detected.", customExtsCfg, "test.custom", true},
		{"A file with an invalid custom extension should not be detected.", customExtsCfg, "test.go", false},

		{"A file in a watched directory should be detected.", customIncludeDirCfg, "watched/test.go", true},

		{"An ignored file should not be detected.", customIgnoredFileCfg, "ignored.go", false},

		{"A file in an ignored and watched dir should not be detected.", customIgnoredAndIncludedDirCfg, "watched/test.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := fileChanges(tt.cfg, tt.relPath, tt.shouldBeDetected); err != nil {
				t.Error(err)
			}
		})
	}
}

func fileChanges(cfg *configuration.Configuration, relChangedFile string, shouldBeDetected bool) error {
	changedFile := filepath.Join(cfg.Root, relChangedFile)
	if isInsideDir(relChangedFile) {
		dir, err := dir(relChangedFile)
		if err != nil {
			return err
		}
		defer delete(dir)
	} else {
		defer delete(changedFile)
	}

	fileChanges, err := NewFileChanges(cfg)
	if err != nil {
		return err
	}

	fileChangesSubscription := subscribe(fileChanges)
	if err := watch(fileChanges); err != nil {
		return err
	}
	if err := change(relChangedFile, changedFile); err != nil {
		return err
	}
	if err := result(fileChangesSubscription, fileChanges, changedFile, shouldBeDetected); err != nil {
		return err
	}

	return nil
}

func subscribe(fileChanges *FileChanges) chan string {
	fileChangesSubscription := make(chan string, 1)
	fileChanges.Subscribe(fileChangesSubscription)
	return fileChangesSubscription
}

func watch(fileChanges *FileChanges) error {
	if err := fileChanges.Init(); err != nil {
		return err
	}
	go fileChanges.Surveil()

	return nil
}

func delete(changedFile string) error {
	return utils.DeletePath(changedFile)
}

func change(relChangedFile string, changedFile string) error {
	if isInsideDir(relChangedFile) {
		dir, err := dir(relChangedFile)
		if err != nil {
			return err
		}
		go createTemporaryDirectoryAndFile(dir, changedFile)
	} else {
		go createTemporaryFile(changedFile)
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
	time.Sleep(TEMP_FILE_CREATION_DELAY * time.Millisecond)
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

func result(fileChangesSubscription chan string, fileChanges *FileChanges, changedFile string, shouldBeDetected bool) error {
	for {
		select {
		case watchedFile := <-fileChangesSubscription:
			clear(fileChangesSubscription, fileChanges)

			return check(watchedFile, changedFile, shouldBeDetected)
		}
	}
}

func clear(fileChangesSubscription chan string, fileChanges *FileChanges) {
	close(fileChangesSubscription)
	fileChanges.StopWatching()
}

func check(watchedFile string, changedFile string, shouldBeDetected bool) error {
	butWasDetected := watchedFile == changedFile
	if !shouldBeDetected && butWasDetected {
		return fmt.Errorf("error: a file change should not be detected at %s", changedFile)
	}
	if shouldBeDetected && !butWasDetected {
		return fmt.Errorf("error: a file change should be detected at %s", changedFile)
	}

	return nil
}
