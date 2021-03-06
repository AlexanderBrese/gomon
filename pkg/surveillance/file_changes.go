package surveillance

import (
	"fmt"
	"sync"
	"time"

	"github.com/AlexanderBrese/go-server-browser-reload/pkg/configuration"
	"github.com/AlexanderBrese/go-server-browser-reload/pkg/reload"
	"github.com/AlexanderBrese/go-server-browser-reload/pkg/utils"
	"github.com/fsnotify/fsnotify"
)

const MAX_WATCHED_FILES = 1000
const MAX_WATCHED_DIRS = 10

type FileChanges struct {
	config       *configuration.Configuration
	watcher      *fsnotify.Watcher
	reloader     *reload.Reload
	mu           sync.Mutex
	stopWatching chan bool

	watchedFilesSubscription chan string
	watchedFiles             chan string
	watchedFileChecksums     *utils.FileChecksums

	watchedDirCount uint
	watchedDirPaths []string
	unwatchDirs     chan bool
}

func NewFileChanges(cfg *configuration.Configuration) (*FileChanges, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &FileChanges{
		config:  cfg,
		watcher: watcher,

		watchedFiles:         make(chan string, MAX_WATCHED_FILES),
		stopWatching:         make(chan bool),
		watchedFileChecksums: utils.NewFileChecksums(),

		unwatchDirs:     make(chan bool, MAX_WATCHED_DIRS),
		watchedDirCount: 0,
		watchedDirPaths: make([]string, MAX_WATCHED_DIRS),
		reloader:        reload.NewReload(cfg),
	}

	return w, nil
}

func (w *FileChanges) Subscribe(watchedFilesSubscription chan string) {
	w.watchedFilesSubscription = watchedFilesSubscription
}

func (w *FileChanges) Init() error {
	if err := w.checkRunEnvironment(); err != nil {
		return err
	}

	return w.watchDir(w.config.Root)
}

func (w *FileChanges) Surveil() error {
	if err := w.control(); err != nil {
		return err
	}

	return w.cleanup()
}

func (w *FileChanges) StopWatching() {
	w.stopWatching <- true
}

func (w *FileChanges) checkRunEnvironment() error {
	buildDir, err := w.config.BuildDir()
	if err != nil {
		return err
	}
	return utils.CreateBuildDir(buildDir)
}

func (w *FileChanges) control() error {
	for {
		select {
		case <-w.stopWatching:
			return nil
		case filePath := <-w.watchedFiles:
			relPath, err := utils.RelPath(w.config.Root, filePath)
			if err != nil {
				return err
			}
			fmt.Printf("%s has changed\n", relPath)
			w.buffer()
		}

		//w.reload()
	}
}

func (w *FileChanges) reload() {
	w.reloader.Reload()
}

func (w *FileChanges) buffer() {
	w.delay()
	w.flushWatchedFiles()
}

func (w *FileChanges) delay() {
	time.Sleep(w.config.BufferTime())
}

func (w *FileChanges) flushWatchedFiles() {
	for {
		select {
		case <-w.watchedFiles:
		default:
			return
		}
	}
}

func (w *FileChanges) cleanup() error {
	w.stopWatchingDirs()

	if err := w.stopWatcher(); err != nil {
		return err
	}

	w.reloader.Cleanup()

	return utils.RemoveBuildDir(w.config.RelBuildDir())
}
