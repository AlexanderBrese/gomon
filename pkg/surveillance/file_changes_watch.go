package surveillance

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexanderBrese/go-server-browser-reload/pkg/utils"
	"github.com/fsnotify/fsnotify"
)

func (w *FileChanges) watchDir(dir string) error {
	return filepath.WalkDir(dir, w.watchInsideDir)
}

func (w *FileChanges) watchInsideDir(path string, d os.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if !d.IsDir() {
		return nil
	}
	if w.isExcludedDir(path) {
		return filepath.SkipDir
	}
	if w.isIncludedDir(path) {
		isNotWatched := !utils.Contains(w.watchedDirPaths, path)
		if isNotWatched {
			w.addWatchedDirPath(path)
		} else {
			return nil
		}
		if err := w.addWatchedDir(path); err != nil {
			return err
		}
		go w.watchNewDir()
	}
	return nil
}

func (w *FileChanges) isExcludedDir(path string) bool {
	return w.isBuildDir(path) || w.isLogDir(path) || w.isHiddenDir(path) || w.isIgnoredDir(path)
}

func (w *FileChanges) isIncludedDir(dir string) bool {
	incDirs := w.config.IncludeDirs

	relDir, err := utils.RelPath(w.config.Root, dir)
	if err != nil {
		return false
	}
	if len(incDirs) == 0 || relDir == "." {
		return true
	}

	for _, d := range incDirs {
		incDir, err := utils.AbsolutePath(d)
		if err != nil {
			log.Printf("error: failed to get absolute path for %s: %s", d, err)
			return false
		}

		if dir == incDir {
			return true
		}

		if strings.HasPrefix(dir, incDir) {
			return true
		}
	}

	return false
}

func (w *FileChanges) addWatchedDirPath(path string) {
	w.watchedDirPaths[utils.Size(w.watchedDirPaths)] = path
}

func (w *FileChanges) addWatchedDir(path string) error {
	return w.watcher.Add(path)
}

func (w *FileChanges) watchNewDir() error {
	w.increaseWatchedDirCount()
	defer w.decreaseWatchedDirCount()
	for {
		select {
		case <-w.unwatchDirs:
			return nil
		case ev := <-w.watcher.Events:
			w.onChange(ev)
		case err := <-w.watcher.Errors:
			if err != nil {
				return err
			}
		}
	}
}

func (w *FileChanges) increaseWatchedDirCount() {
	utils.WithLock(&w.mu, func() {
		w.watchedDirCount++
	})
}

func (w *FileChanges) decreaseWatchedDirCount() {
	utils.WithLock(&w.mu, func() {
		w.watchedDirCount--
	})
}

func (w *FileChanges) onChange(changeEvent fsnotify.Event) error {
	isDir, err := utils.IsDir(changeEvent.Name)
	if err != nil {
		return err
	}

	if isDir {
		return w.onDirChange(changeEvent)
	}
	return w.onFileChange(changeEvent)
}

func (w *FileChanges) onFileChange(changeEvent fsnotify.Event) error {
	file := changeEvent.Name
	if !isFileChange(changeEvent) || isRemove(changeEvent) {
		return nil
	}
	hasNotChanged, err := w.hasNotChanged(file)
	if err != nil {
		return err
	}
	if !hasNotChanged {
		w.watchedFileChecksums.UpdateFileChecksum(file)
	}
	if w.isExcludedFile(file) || hasNotChanged {
		w.notifyNoFileChange()
		return nil
	}

	w.notifyFileChange(file)
	return nil
}

func isFileChange(ev fsnotify.Event) bool {
	return ev.Op&fsnotify.Write == fsnotify.Write
}

func isRemove(ev fsnotify.Event) bool {
	return ev.Op&fsnotify.Remove == fsnotify.Remove
}

func (w *FileChanges) hasNotChanged(path string) (bool, error) {
	hasChanged, err := w.watchedFileChecksums.HasChanged(path)
	if err != nil {
		return true, err
	}
	return !hasChanged, nil
}

func (w *FileChanges) notifyNoFileChange() {
	w.watchedFilesSubscription <- ""
}

func (w *FileChanges) notifyFileChange(changedFile string) {
	w.watchedFiles <- changedFile
	w.watchedFilesSubscription <- changedFile
}

func (w *FileChanges) onDirChange(changeEvent fsnotify.Event) error {
	dir := changeEvent.Name
	if w.isExcludedDir(dir) {
		if isWrite(changeEvent) {
			w.notifyNoFileChange()
		}

		return nil
	}

	if !isDirChange(changeEvent) {
		return nil
	}

	if isRemove(changeEvent) {
		if err := w.removeWatchedDir(dir); err != nil {
			return err
		}
		return nil
	}
	w.watchDir(dir)

	return nil
}

func (w *FileChanges) removeWatchedDir(path string) error {
	return w.watcher.Remove(path)
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
	for _, d := range w.config.ExcludeDirs {
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

func (w *FileChanges) isExcludedFile(path string) bool {
	return w.isIgnoredFile(path) || w.isIgnoredExt(path)
}

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

	for _, e := range w.config.IncludeExts {
		if ext == "."+strings.TrimSpace(e) {
			return false
		}
	}

	return true
}

func isDirChange(ev fsnotify.Event) bool {
	return ev.Op&fsnotify.Create == fsnotify.Create ||
		ev.Op&fsnotify.Remove == fsnotify.Remove
}

func isWrite(ev fsnotify.Event) bool {
	return ev.Op&fsnotify.Write == fsnotify.Write
}

func (w *FileChanges) stopWatchingDirs() {
	utils.WithLock(&w.mu, func() {
		for i := 0; i < int(w.watchedDirCount); i++ {
			w.unwatchDirs <- true
		}
	})
}

func (w *FileChanges) close() error {
	return w.watcher.Close()
}
