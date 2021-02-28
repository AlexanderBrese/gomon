package monitoring

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/AlexanderBrese/go-server-browser-reload/pkg/configuration"
	"github.com/AlexanderBrese/go-server-browser-reload/pkg/utils"
	"github.com/fsnotify/fsnotify"
)

type FileChanges struct {
	config               *configuration.Configuration
	watcher              *fsnotify.Watcher
	stopWatcherChan      chan bool
	stopWatchingChan     chan bool
	watchedFilesChan     chan string
	watchedDirs          uint
	watchedFileChecksums *FileChecksums
	mu                   sync.Mutex
}

func NewFileChanges(cfg *configuration.Configuration) (*FileChanges, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &FileChanges{
		config:           cfg,
		watcher:          watcher,
		watchedFilesChan: make(chan string, 1000),
		stopWatcherChan:  make(chan bool, 10),
		stopWatchingChan: make(chan bool),
		watchedDirs:      0,
	}

	w.watchedFileChecksums = &FileChecksums{storage: make(map[string]string)}

	return w, nil
}

func (w *FileChanges) Watch(path string) error {
	if err := w.watch(path); err != nil {
		return err
	}

	w.start()

	return w.cleanup()
}

func (w *FileChanges) StopWatching() {
	w.stopWatchingChan <- true
}

func (w *FileChanges) start() {
	for {
		select {
		case <-w.stopWatchingChan:
			return
		case file := <-w.watchedFilesChan:
			w.delay()
			if w.isExcludedFile(file) || w.hasNotChanged(file) {
				continue
			}
		}
	}
}

func (w *FileChanges) cleanup() error {
	w.stopWatchingDirs()

	if err := w.closeWatcher(); err != nil {
		return err
	}
	if err := w.removeBuildDir(); err != nil {
		return err
	}

	return nil
}

func (w *FileChanges) stopWatchingDirs() {
	utils.WithLock(&w.mu, func() {
		for i := 0; i < int(w.watchedDirs); i++ {
			w.stopWatcherChan <- true
		}
	})
}

func (w *FileChanges) closeWatcher() error {
	return w.watcher.Close()
}

func (w *FileChanges) removeBuildDir() error {
	return w.config.RemoveBuildDir()
}

func (w *FileChanges) delay() {
	time.Sleep(w.config.Delay())
	w.flushWatchedFiles()
}

func (w *FileChanges) flushWatchedFiles() {
	for {
		select {
		case <-w.watchedFilesChan:
		default:
			return
		}
	}
}

func (w *FileChanges) watch(path string) error {
	return filepath.WalkDir(path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if w.isExcludedDir(path) {
			return filepath.SkipDir
		}
		if w.isWatchedDir(path) {
			if err := w.addWatch(path); err != nil {
				return err
			}
			if err = w.updateFileChecksums(path); err != nil {
				return err
			}
			return w.watchDir(path)
		}
		return nil
	})
}

func (w *FileChanges) watchDir(path string) error {
	go func() {
		utils.WithLock(&w.mu, func() {
			w.watchedDirs++
		})
		defer func() {
			utils.WithLock(&w.mu, func() {
				w.watchedDirs--
			})
		}()

		for {
			select {
			case <-w.stopWatcherChan:
				return
			case ev := <-w.eventsChan():
				if !isValidWatchEvent(ev) {
					break
				}
				filePath := ev.Name
				if utils.IsDir(filePath) {
					// Directory was removed
					if isWatchRemoveEvent(ev) {
						if err := w.removeWatch(filePath); err != nil {
							log.Printf("error: failed to stop watching %s: %s", filePath, err)
						}
						break
					}
					// Watch recursively
					w.watch(filePath)
					break
				}
				if w.isExcludedFile(filePath) {
					break
				}
				w.watchedFilesChan <- filePath
			case err := <-w.watcher.Errors:
				log.Printf("error: during file watch at %s: %s", path, err)
			}
		}
	}()
	return nil
}

func (w *FileChanges) isExcludedDir(path string) bool {
	return w.isBuildDir(path) || w.isLogDir(path) || w.isHiddenDir(path) || w.isIgnoredDir(path)
}

func (w *FileChanges) isBuildDir(path string) bool {
	buildDir, err := w.config.BuildDir()
	if err != nil {
		log.Printf("error: failed to get build dir: %s", err)
		return false
	}
	return path == buildDir
}

func (w *FileChanges) isLogDir(path string) bool {
	logDir, err := w.config.LogDir()
	if err != nil {
		log.Printf("error: failed to get log dir: %s", err)
		return false
	}
	return path == logDir
}

func (w *FileChanges) isHiddenDir(path string) bool {
	return len(path) > 1 && strings.HasPrefix(filepath.Base(path), ".")
}

func (w *FileChanges) isIgnoredDir(path string) bool {
	for _, d := range w.config.IgnoreDirs {
		absIgnoredDirPath, err := utils.AbsolutePath(d)
		if err != nil {
			log.Printf("error: failed to get absolute path for %s: %s", d, err)
			return false
		}
		if path == absIgnoredDirPath {
			return true
		}
	}
	return false
}

func (w *FileChanges) isWatchedDir(path string) bool {
	iDirs := w.config.IgnoreDirs
	if len(iDirs) == 0 {
		return true
	}

	if path == utils.RootPath() {
		return false
	}

	for _, d := range iDirs {
		absWatchedDirPath, err := utils.AbsolutePath(d)
		if err != nil {
			log.Printf("error: failed to get absolute path for %s: %s", d, err)
			return false
		}

		if path == absWatchedDirPath {
			return true
		}

		if strings.HasPrefix(path, absWatchedDirPath) {
			return true
		}
	}

	return false
}

func (w *FileChanges) isExcludedFile(path string) bool {
	return w.isIgnoredFile(path) || w.isIgnoredExt(path)
}

// TODO: Refactor
func (w *FileChanges) isIgnoredFile(path string) bool {
	for _, f := range w.config.IgnoreFiles {
		absIgnoredFile, err := utils.AbsolutePath(f)
		if err != nil {
			log.Printf("error: failed to get absolute path for %s: %s", f, err)
			return false
		}

		if path == absIgnoredFile {
			return true
		}
	}
	return false
}

func (w *FileChanges) isIgnoredExt(path string) bool {
	ext := filepath.Ext(path)

	for _, e := range w.config.WatchExts {
		if ext == "."+strings.TrimSpace(e) {
			return true
		}
	}

	return false
}

func (w *FileChanges) hasNotChanged(path string) bool {
	return !w.watchedFileChecksums.HasChanged(path)
}

func (w *FileChanges) updateFileChecksums(path string) error {
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return filepath.SkipDir
		}

		if w.isExcludedFile(path) {
			return nil
		}

		w.watchedFileChecksums.UpdateFileChecksum(path)

		return nil
	})
}

func (w *FileChanges) addWatch(path string) error {
	return w.watcher.Add(path)
}

func (w *FileChanges) removeWatch(path string) error {
	return w.watcher.Remove(path)
}

func (w *FileChanges) eventsChan() chan fsnotify.Event {
	return w.watcher.Events
}

func isValidWatchEvent(ev fsnotify.Event) bool {
	return ev.Op&fsnotify.Create == fsnotify.Create ||
		ev.Op&fsnotify.Write == fsnotify.Write ||
		ev.Op&fsnotify.Remove == fsnotify.Remove
}

func isWatchRemoveEvent(ev fsnotify.Event) bool {
	return ev.Op&fsnotify.Remove == fsnotify.Remove
}
