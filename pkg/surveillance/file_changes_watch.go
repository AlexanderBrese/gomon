package surveillance

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexanderBrese/go-server-browser-reload/pkg/utils"
	"github.com/fsnotify/fsnotify"
)

func (w *FileChangesDetection) watchDir(dir string) error {
	return filepath.WalkDir(dir, w.watchInsideDir)
}

func (w *FileChangesDetection) watchInsideDir(path string, d os.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if !d.IsDir() {
		return nil
	}
	isExcluded, err := w.isExcludedDir(path)
	if err != nil {
		return err
	}
	if isExcluded {
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

func (w *FileChangesDetection) isExcludedDir(path string) (bool, error) {
	isIgnored, err := w.isIgnoredDir(path)
	if err != nil {
		return false, err
	}
	isExcluded := w.isBuildDir(path) || w.isLogDir(path) || w.isHiddenDir(path) || isIgnored
	return isExcluded, nil
}

func (w *FileChangesDetection) isIncludedDir(dir string) bool {
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

func (w *FileChangesDetection) addWatchedDirPath(path string) {
	w.watchedDirPaths[utils.Size(w.watchedDirPaths)] = path
}

func (w *FileChangesDetection) addWatchedDir(path string) error {
	return w.watcher.Add(path)
}

func (w *FileChangesDetection) watchNewDir() error {
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

func (w *FileChangesDetection) increaseWatchedDirCount() {
	utils.WithLock(&w.mu, func() {
		w.watchedDirCount++
	})
}

func (w *FileChangesDetection) decreaseWatchedDirCount() {
	utils.WithLock(&w.mu, func() {
		w.watchedDirCount--
	})
}

func (w *FileChangesDetection) onChange(changeEvent fsnotify.Event) error {
	isDir, err := utils.IsDir(changeEvent.Name)
	if err != nil {
		return err
	}

	if isDir {
		return w.onDirChange(changeEvent)
	}
	return w.onFileChange(changeEvent)
}

func (w *FileChangesDetection) onFileChange(changeEvent fsnotify.Event) error {
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

func (w *FileChangesDetection) hasNotChanged(path string) (bool, error) {
	hasChanged, err := w.watchedFileChecksums.HasChanged(path)
	if err != nil {
		return true, err
	}
	return !hasChanged, nil
}

func (w *FileChangesDetection) notifyNoFileChange() {
	w.watchedFilesSubscription <- ""
}

func (w *FileChangesDetection) notifyFileChange(changedFile string) {
	w.watchedFiles <- changedFile
	w.watchedFilesSubscription <- changedFile
}

func (w *FileChangesDetection) onDirChange(changeEvent fsnotify.Event) error {
	dir := changeEvent.Name
	isExcluded, err := w.isExcludedDir(dir)
	if err != nil {
		return err
	}
	if isExcluded {
		if isWrite(changeEvent) {
			select {
			case <-w.watchedFilesSubscription:
			default:
				w.notifyNoFileChange()
			}
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

func (w *FileChangesDetection) removeWatchedDir(path string) error {
	return w.watcher.Remove(path)
}

func (w *FileChangesDetection) isBuildDir(path string) bool {
	buildDir, err := w.config.BuildDir()
	if err != nil {
		log.Printf("error: failed to get build dir: %s", err)
		return false
	}
	return path == buildDir
}

func (w *FileChangesDetection) isLogDir(path string) bool {
	logDir, err := w.config.LogDir()
	if err != nil {
		log.Printf("error: failed to get log dir: %s", err)
		return false
	}
	return path == logDir
}

func (w *FileChangesDetection) isHiddenDir(path string) bool {
	return len(path) > 1 && strings.HasPrefix(filepath.Base(path), ".")
}

func (w *FileChangesDetection) isIgnoredDir(path string) (bool, error) {
	relPath, err := utils.RelPath(w.config.Root, path)
	if err != nil {
		return false, err
	}

	rootParent := strings.Split(relPath, "/")[0]
	for _, ignoredDir := range w.config.ExcludeDirs {
		if rootParent == ignoredDir {
			return true, nil
		}
	}

	return false, nil
}

func (w *FileChangesDetection) isExcludedFile(path string) bool {
	return w.isIgnoredFile(path) || w.isIgnoredExt(path)
}

func (w *FileChangesDetection) isIgnoredFile(path string) bool {
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

func (w *FileChangesDetection) isIgnoredExt(path string) bool {
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

func (w *FileChangesDetection) stopWatchingDirs() {
	utils.WithLock(&w.mu, func() {
		for i := 0; i < int(w.watchedDirCount); i++ {
			w.unwatchDirs <- true
		}
	})
}

func (w *FileChangesDetection) stopWatcher() error {
	return w.watcher.Close()
}
