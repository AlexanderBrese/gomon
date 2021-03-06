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

type FileChangesDetection struct {
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

func NewFileChangesDetection(cfg *configuration.Configuration) (*FileChangesDetection, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &FileChangesDetection{
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

func (w *FileChangesDetection) Subscribe(watchedFilesSubscription chan string) {
	w.watchedFilesSubscription = watchedFilesSubscription
}

func (w *FileChangesDetection) Init() error {
	if err := w.checkRunEnvironment(); err != nil {
		return err
	}

	return w.watchDir(w.config.Root)
}

func (w *FileChangesDetection) Surveil() error {
	if err := w.control(); err != nil {
		return err
	}

	return w.cleanup()
}

func (w *FileChangesDetection) StopWatching() {
	w.stopWatching <- true
}

func (w *FileChangesDetection) checkRunEnvironment() error {
	buildDir, err := w.config.BuildDir()
	if err != nil {
		return err
	}
	return utils.CreateBuildDir(buildDir)
}

func (w *FileChangesDetection) control() error {
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

		w.reload()
	}
}

func (w *FileChangesDetection) reload() {
	w.reloader.Reload()
}

func (w *FileChangesDetection) buffer() {
	w.delay()
	w.flushWatchedFiles()
}

func (w *FileChangesDetection) delay() {
	time.Sleep(w.config.BufferTime())
}

func (w *FileChangesDetection) flushWatchedFiles() {
	for {
		select {
		case <-w.watchedFiles:
		default:
			return
		}
	}
}

func (w *FileChangesDetection) cleanup() error {
	w.stopWatchingDirs()

	if err := w.stopWatcher(); err != nil {
		return err
	}

	w.reloader.Cleanup()

	return utils.RemoveBuildDir(w.config.RelBuildDir)
}
