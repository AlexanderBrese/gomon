package surveillance

import (
	"fmt"
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
	config *configuration.Configuration

	watcher *fsnotify.Watcher

	watchedFilesSubscriberChan chan string
	stopWatcherChan            chan bool
	stopWatchingChan           chan bool
	watchedFilesChan           chan string

	watchedDirs uint

	watchedFileChecksums *utils.FileChecksums

	mu sync.Mutex
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

	w.watchedFileChecksums = utils.NewFileChecksums()

	return w, nil
}

func (w *FileChanges) Subscribe(subscriberChan chan string) {
	w.watchedFilesSubscriberChan = subscriberChan
}

func (w *FileChanges) Init() error {
	return w.watch(w.config.Root)
}

func (w *FileChanges) Watch() error {
	w.control()

	return w.cleanup()
}

func (w *FileChanges) StopWatching() {
	w.stopWatchingChan <- true
}

func (w *FileChanges) control() {
	for {
		select {
		case <-w.stopWatchingChan:
			return
		case filePath := <-w.watchedFilesChan:
			relPath, err := utils.RelPath(w.config.Root, filePath)
			if err != nil {
				log.Print(err)
			}
			fmt.Printf("%s has changed\n", relPath)
			w.delay()
		}
	}
}

func (w *FileChanges) cleanup() error {
	w.stopWatchingDirs()

	if err := w.closeWatcher(); err != nil {
		return err
	}
	/* TODO: implement
	if err := w.removeBuildDir(); err != nil {
		return err
	}
	*/

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

func (w *FileChanges) watch(rootPath string) error {
	return filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
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
			go w.watchDir()
		}
		return nil
	})
}

func (w *FileChanges) watchDir() error {
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
			return nil
		case ev := <-w.watcher.Events:
			path := ev.Name
			isDir, err := utils.IsDir(path)
			if err != nil {
				log.Printf("error: during file watch: %s", err)
				break
			}
			if isDir && w.isExcludedDir(path) {
				if isWriteEvent(ev) {
					w.watchedFilesSubscriberChan <- ""
				}

				break
			}

			if !isValidEvent(ev, isDir) {
				break
			}

			if isDir {
				if isRemoveEvent(ev) {
					if err := w.removeWatch(path); err != nil {
						log.Printf("error: during file watch: %s", err)
					}
					break
				}
				w.watch(path)
				break
			} else if isRemoveEvent(ev) {
				break
			}

			hasNotChanged, err := w.hasNotChanged(path)
			if err != nil {
				log.Printf("error: during file watch: %s", err)
			}
			if !hasNotChanged {
				w.watchedFileChecksums.UpdateFileChecksum(path)
			}
			if w.isExcludedFile(path) || hasNotChanged {
				w.watchedFilesSubscriberChan <- ""
				break
			}

			w.watchedFilesChan <- path
			w.watchedFilesSubscriberChan <- path
		case err := <-w.watcher.Errors:
			if err != nil {
				log.Printf("error: during file watch: %s", err)
			}
		}
	}
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
	iDirs := w.config.WatchDirs

	relPath, err := utils.RelPath(w.config.Root, path)
	if err != nil {
		return false
	}
	if len(iDirs) == 0 || relPath == "." {
		return true
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
			return false
		}
	}

	return true
}

func (w *FileChanges) hasNotChanged(path string) (bool, error) {
	hasChanged, err := w.watchedFileChecksums.HasChanged(path)
	if err != nil {
		return true, err
	}
	return !hasChanged, nil
}

func (w *FileChanges) addWatch(path string) error {
	return w.watcher.Add(path)
}

func (w *FileChanges) removeWatch(path string) error {
	return w.watcher.Remove(path)
}

func isValidEvent(ev fsnotify.Event, isDir bool) bool {
	if isDir {
		return ev.Op&fsnotify.Create == fsnotify.Create ||
			ev.Op&fsnotify.Remove == fsnotify.Remove
	}
	return ev.Op&fsnotify.Write == fsnotify.Write
}

func isWriteEvent(ev fsnotify.Event) bool {
	return ev.Op&fsnotify.Write == fsnotify.Write
}

func isRemoveEvent(ev fsnotify.Event) bool {
	return ev.Op&fsnotify.Remove == fsnotify.Remove
}
