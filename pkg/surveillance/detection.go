package surveillance

import (
	"os"
	"path/filepath"

	"github.com/AlexanderBrese/GOATmon/pkg/utils"
	"github.com/fsnotify/fsnotify"
)

type Detection struct {
	environment  *Environment
	notification *Notification
	checksums    *utils.FileChecksums
	filter       *Filter
}

func NewDetection(env *Environment, n *Notification) (*Detection, error) {
	d := &Detection{
		environment:  env,
		notification: n,
		filter:       NewFilter(env.config),
		checksums:    utils.NewFileChecksums(),
	}

	if err := d.observe(env.config.Root); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Detection) Run() error {
	for {
		select {
		case <-d.environment.stopDetecting:
			close(d.environment.stopDetecting)
			return nil
		case evs, ok := <-d.environment.detector.Events:
			if !ok {
				return nil
			}
			if err := d.on(evs); err != nil {
				return err
			}
		case errs, ok := <-d.environment.detector.Errors:
			if !ok {
				return nil
			}
			return errs[len(errs)-1]
		}
	}
}

func (d *Detection) observe(root string) error {
	return filepath.WalkDir(root, func(path string, e os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		isFile := !e.IsDir()
		if isFile {
			return d.cacheFile(path)
		}

		return d.addIfIncluded(path)
	})
}

func (d *Detection) cacheFile(path string) error {
	isExcluded, err := d.filter.IsExcludedFile(path)
	if err != nil {
		return err
	}
	if !isExcluded {
		newChecksum, err := utils.FileChecksum(path)
		if err != nil {
			return err
		}
		d.checksums.UpdateFileChecksum(path, newChecksum)
	}

	return nil
}

func (d *Detection) add(path string) error {
	return d.environment.detector.Add(path)
}

func (d *Detection) remove(path string) error {
	return d.environment.detector.Remove(path)
}

func (d *Detection) addIfIncluded(path string) error {
	isExcluded, err := d.filter.IsExcludedDir(path)
	if err != nil {
		return err
	}
	if isExcluded {
		return filepath.SkipDir
	}
	isIncluded, err := d.filter.IsIncludedDir(path)
	if err != nil {
		return err
	}
	if isIncluded {
		if err := d.add(path); err != nil {
			return err
		}
	}
	return nil
}

func (d *Detection) on(evs []fsnotify.Event) error {
	hasChanged := false
	hasFiles := false

	for _, ev := range evs {
		path := ev.Name

		isDir, err := utils.IsDir(path)
		if err != nil {
			return err
		}
		if isDir {
			if err := d.dirChange(ev, path); err != nil {
				return err
			}
		} else {
			hasFiles = true
			if !hasChanged {
				changeDetected, err := d.fileChange(ev, path)
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
			d.notification.NotfiyChange()
		} else {
			d.notification.NotifyNoChange()
		}
	}

	return nil
}
func (d *Detection) fileChange(ev fsnotify.Event, path string) (bool, error) {
	isExcluded, err := d.filter.IsExcludedFile(path)
	if err != nil {
		return false, err
	}

	if !utils.IsWrite(ev) || isExcluded {
		return false, nil
	}
	newChecksum, err := utils.FileChecksum(path)
	if err != nil {
		return false, err
	}
	if d.checksums.HasChanged(path, newChecksum) {
		d.checksums.UpdateFileChecksum(path, newChecksum)
		return true, nil
	}
	return false, nil
}

func (d *Detection) dirChange(ev fsnotify.Event, path string) error {
	if utils.IsRemove(ev) {
		if err := d.remove(path); err != nil {
			return err
		}
	} else if utils.IsCreate(ev) {
		if err := d.addIfIncluded(path); err != nil {
			return err
		}
	}
	return nil
}
