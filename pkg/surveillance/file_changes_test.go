package surveillance

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/AlexanderBrese/go-server-browser-reload/pkg/configuration"
	"github.com/AlexanderBrese/go-server-browser-reload/pkg/utils"
)

func TestFileChanges(t *testing.T) {
	defaultCfg := configuration.DefaultConfiguration()
	customExtsCfg := configuration.DefaultConfiguration()
	customExtsCfg.WatchExts = []string{"custom"}
	customIgnoredDirCfg := configuration.DefaultConfiguration()
	customIgnoredDirCfg.IgnoreDirs = []string{"ignored_dir"}
	customIgnoredFileCfg := configuration.DefaultConfiguration()
	customIgnoredFileCfg.IgnoreFiles = []string{"ignored"}

	tests := []struct {
		name             string
		cfg              *configuration.Configuration
		tmpPath          string
		shouldBeDetected bool
	}{
		{".go file with default configuration", defaultCfg, "test.go", true},
		{".custom file with custom ext configuration", customExtsCfg, "test.custom", true},
		{".go file with custom ext configuration", customExtsCfg, "test.go", false},
		{".go file in ignored dir with ignored dir configuration", customIgnoredDirCfg, "ignored_dir/test.go", false},
		{".go file not in ignored dir with ignored dir configuration", customIgnoredDirCfg, "test.go", true},
		{".go file in ignored files with ignored files configuration", customIgnoredFileCfg, "ignored.go", false},
		{".go file not in ignored files with ignored files configuration", customIgnoredFileCfg, "test.go", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := watch(tt.cfg, tt.tmpPath, tt.shouldBeDetected); err != nil {
				t.Error(err)
			}
		})
	}
}

func watch(cfg *configuration.Configuration, tmpPath string, shouldBeDetected bool) error {
	fileChanges, err := NewFileChanges(cfg)
	if err != nil {
		return fmt.Errorf("error during file changes creation: %s", err)
	}
	watchedFilesSubscription := make(chan string, 1)
	fileChanges.Subscribe(watchedFilesSubscription)

	if err := fileChanges.Init(); err != nil {
		return err
	}

	go func() {
		fileChanges.Watch()
	}()

	if !shouldBeDetected {
		defer fileChanges.StopWatching()
		go delayed(func() {
			close(watchedFilesSubscription)
		})
	}

	filePath := filepath.Join(cfg.Root, tmpPath)
	if _, err = utils.CreateFile(filePath, []byte("test")); err != nil {
		return fmt.Errorf("error during temp file creation: %s", err)
	}
	defer utils.DeleteFile(filePath)

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
func delayed(f func()) {
	const DELAY = 1000
	time.Sleep(time.Duration(DELAY) * time.Millisecond)
	f()
}
