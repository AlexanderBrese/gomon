package surveillance

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexanderBrese/go-server-browser-reload/pkg/utils"
	"github.com/fsnotify/fsnotify"
)

func (w *FileChangesDetection) watch() error {
	for {
		select {
		case evs, ok := <-w.watcher.Events:
			if !ok {
				return nil
			}
			if err := w.onChange(evs); err != nil {
				return err
			}
		case errs, ok := <-w.watcher.Errors:
			if !ok {
				return nil
			}
			return errs[len(errs)-1]
		}
	}
}

func (w *FileChangesDetection) crawlWatchedDirs(root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			if !w.isExcludedFile(path) {
				newChecksum, err := utils.FileChecksum(path)
				if err != nil {
					return err
				}
				w.watchedFileChecksums.UpdateFileChecksum(path, newChecksum)
			}

			return nil
		}

		return w.watchDirIfValid(path)
	})
}

func (w *FileChangesDetection) watchDirIfValid(path string) error {
	isExcluded, err := w.isExcludedDir(path)
	if err != nil {
		return err
	}
	if isExcluded {
		return filepath.SkipDir
	}

	if w.isIncludedDir(path) {
		if err := w.addWatchedDir(path); err != nil {
			return err
		}
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
	if len(incDirs) == 0 {
		return true
	}

	for _, d := range incDirs {
		incDir, err := utils.AbsolutePath(d)
		if err != nil {
			log.Printf("error: failed to get absolute path for %s: %s", d, err)
			return false
		}
		if dir == incDir || strings.HasPrefix(dir, incDir) {
			return true
		}
	}

	return false
}

func (w *FileChangesDetection) addWatchedDir(path string) error {
	return w.watcher.Add(path)
}

func (w *FileChangesDetection) onChange(evs []fsnotify.Event) error {
	hasChanged := false
	hasFiles := false

	for _, ev := range evs {
		path := ev.Name

		isDir, err := utils.IsDir(path)
		if err != nil {
			return err
		}
		if isDir {
			if err := w.onDirChange(ev, path); err != nil {
				return err
			}
		} else {
			hasFiles = true
			if !hasChanged {
				changeDetected, err := w.changed(ev, path)
				if err != nil {
					return err
				}
				if changeDetected {
					hasChanged = true
				}
			}
		}
	}

	if hasFiles {
		if hasChanged {
			w.notifyChange()
		} else {
			w.notifyNoChange()
		}
	}

	return nil
}
func (w *FileChangesDetection) changed(ev fsnotify.Event, path string) (bool, error) {
	if !utils.IsWrite(ev) || w.isExcludedFile(path) {
		return false, nil
	}
	newChecksum, err := utils.FileChecksum(path)
	if err != nil {
		return false, err
	}
	if w.watchedFileChecksums.HasChanged(path, newChecksum) {
		w.watchedFileChecksums.UpdateFileChecksum(path, newChecksum)
		return true, nil
	}
	return false, nil
}

func (w *FileChangesDetection) onDirChange(ev fsnotify.Event, path string) error {
	if utils.IsRemove(ev) {
		if err := w.removeWatchedDir(path); err != nil {
			return err
		}
	} else if utils.IsCreate(ev) {
		if err := w.watchDirIfValid(path); err != nil {
			return err
		}
	}
	return nil
}

func (w *FileChangesDetection) notifyNoChange() {
	if w.changeDetectedSubscription != nil {
		w.changeDetectedSubscription <- false
	}
}

func (w *FileChangesDetection) notifyChange() {
	w.changeDetected <- true
	if w.changeDetectedSubscription != nil {
		w.changeDetectedSubscription <- true
	}
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

func (w *FileChangesDetection) stopWatcher() {
	w.watcher.Close()
}
