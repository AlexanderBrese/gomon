package surveillance

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/AlexanderBrese/go-server-browser-reload/pkg/configuration"
	"github.com/AlexanderBrese/go-server-browser-reload/pkg/utils"
)

const DELAY = 1000

func TestGoFileHasChangedWithDefaultSettings(t *testing.T) {
	cfg := configuration.DefaultConfiguration()

	fileChanges, err := NewFileChanges(cfg)
	if err != nil {
		t.Fatalf("error during file changes creation: %s", err)
	}
	watchedFilesSubscription := make(chan string)
	fileChanges.Subscribe(watchedFilesSubscription)

	if err := fileChanges.Init(); err != nil {
		t.Fatal(err)
	}

	go func() {
		fileChanges.Watch()
	}()

	filePath := filepath.Join(cfg.Root, "test.go")
	if _, err = utils.CreateFile(filePath, []byte("test")); err != nil {
		t.Fatalf("error during temp file creation: %s", err)
	}
	defer utils.DeletePath(filePath)

	watchedFilePath := <-watchedFilesSubscription
	if watchedFilePath != filePath {
		t.Errorf("error: failed to detect a created file with content at %s", watchedFilePath)
	}
}

func TestCustomExtFileHasChangedWithCustomSettings(t *testing.T) {
	cfg := configuration.DefaultConfiguration()
	cfg.WatchExts = []string{"custom"}

	fileChanges, err := NewFileChanges(cfg)
	if err != nil {
		t.Fatalf("error during file changes creation: %s", err)
	}
	watchedFilesSubscription := make(chan string)
	fileChanges.Subscribe(watchedFilesSubscription)

	if err := fileChanges.Init(); err != nil {
		t.Fatal(err)
	}

	go func() {
		fileChanges.Watch()
	}()

	filePath := filepath.Join(cfg.Root, "test.custom")
	if _, err = utils.CreateFile(filePath, []byte("test")); err != nil {
		t.Fatalf("error during temp file creation: %s", err)
	}
	defer utils.DeletePath(filePath)
	watchedFilePath := <-watchedFilesSubscription
	if watchedFilePath != filePath {
		t.Errorf("error: failed to detect a created file with content at %s", watchedFilePath)
	}
}

func TestFileChanges(t *testing.T) {
	defaultCfg := configuration.DefaultConfiguration()
	customExtsCfg := configuration.DefaultConfiguration()
	customExtsCfg.WatchExts = []string{"custom"}

	tests := []struct {
		name             string
		cfg              *configuration.Configuration
		tmpPath          string
		shouldBeDetected bool
	}{
		{"should be detected: .go file with default configuration", defaultCfg, "test.go", true},
		{"should be detected: .custom file with custom ext configuration", customExtsCfg, "test.custom", true},
		{"should not be detected: .go file with custom ext configuration", customExtsCfg, "test.go", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := watch(tt.cfg, tt.tmpPath, tt.shouldBeDetected); err != nil {
				t.Error(err)
			}
		})
	}
}

/*
func TestFileHasNotChangedAndDirIsIgnored() {

}
func TestFileHasChangedAndDirIsNotIgnored() {

}

func TestFileHasChangedAndDirIsWatched() {

}

func TestFileHasNotChangedAndDirIsNotWatched() {

}
*/
func watch(cfg *configuration.Configuration, tmpPath string, shouldBeDetected bool) error {
	fileChanges, err := NewFileChanges(cfg)
	if err != nil {
		return fmt.Errorf("error during file changes creation: %s", err)
	}
	watchedFilesSubscription := make(chan string)
	fileChanges.Subscribe(watchedFilesSubscription)
	go unsubscribeDelayed(DELAY, watchedFilesSubscription)

	if err := fileChanges.Init(); err != nil {
		return err
	}

	go func() {
		fileChanges.Watch()
	}()
	defer fileChanges.StopWatching()

	filePath := filepath.Join(cfg.Root, tmpPath)
	if _, err = utils.CreateFile(filePath, []byte("test")); err != nil {
		return fmt.Errorf("error during temp file creation: %s", err)
	}
	defer utils.DeletePath(filePath)

	watchedFilePath := <-watchedFilesSubscription
	butWasDetected := watchedFilePath == filePath
	if !shouldBeDetected && butWasDetected {
		return fmt.Errorf("error: a file change should not be detected at %s", filePath)
	}
	if shouldBeDetected && !butWasDetected {
		return fmt.Errorf("error: a file change should be detected at %s", filePath)
	}

	return nil
}

// assumption: a successful file change detection takes no longer than a certain (small) duration
func unsubscribeDelayed(delay int, subscription chan string) {
	time.Sleep(time.Duration(delay) * time.Millisecond)
	_, closed := <-subscription
	if !closed {
		close(subscription)
	}
}
