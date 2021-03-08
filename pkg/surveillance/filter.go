package surveillance

import (
	"path/filepath"
	"strings"

	"github.com/AlexanderBrese/GOATmon/pkg/configuration"
	"github.com/AlexanderBrese/GOATmon/pkg/utils"
)

type Filter struct {
	config *configuration.Configuration
}

func NewFilter(cfg *configuration.Configuration) *Filter {
	return &Filter{
		config: cfg,
	}
}

func (f *Filter) IsExcludedDir(path string) (bool, error) {
	isIgnored, err := f.IsIgnoredDir(path)
	if err != nil {
		return false, err
	}
	isBuildDir, err := f.IsBuildDir(path)
	if err != nil {
		return false, err
	}
	isLogDir, err := f.IsLogDir(path)
	if err != nil {
		return false, err
	}

	isExcluded := isBuildDir || isLogDir || f.IsHiddenDir(path) || isIgnored
	return isExcluded, nil
}

func (f *Filter) IsIncludedDir(dir string) (bool, error) {
	incDirs := f.config.IncludeDirs
	if len(incDirs) == 0 {
		return true, nil
	}

	for _, d := range incDirs {
		incDir, err := utils.CurrentAbsolutePath(d)
		if err != nil {
			return false, err
		}
		if dir == incDir || strings.HasPrefix(dir, incDir) {
			return true, nil
		}
	}

	return false, nil
}

func (f *Filter) IsBuildDir(path string) (bool, error) {
	buildDir, err := f.config.BuildDir()
	if err != nil {
		return false, err
	}
	return path == buildDir, nil
}

func (f *Filter) IsLogDir(path string) (bool, error) {
	logDir, err := f.config.LogDir()
	if err != nil {
		return false, err
	}
	return path == logDir, nil
}

func (f *Filter) IsHiddenDir(path string) bool {
	return len(path) > 1 && strings.HasPrefix(filepath.Base(path), ".")
}

func (f *Filter) IsIgnoredDir(path string) (bool, error) {
	relPath, err := utils.RelPath(f.config.Root, path)
	if err != nil {
		return false, err
	}

	rootParent := strings.Split(relPath, "/")[0]
	for _, ignoredDir := range f.config.ExcludeDirs {
		if rootParent == ignoredDir {
			return true, nil
		}
	}

	return false, nil
}

func (f *Filter) IsExcludedFile(path string) (bool, error) {
	isIgnored, err := f.IsIgnoredFile(path)
	if err != nil {
		return false, err
	}
	return isIgnored || f.IsIgnoredExt(path), nil
}

func (f *Filter) IsIgnoredFile(path string) (bool, error) {
	for _, f := range f.config.IgnoreFiles {
		absIgnoredFile, err := utils.CurrentAbsolutePath(f)
		if err != nil {
			return false, err
		}

		if path == absIgnoredFile {
			return true, nil
		}
	}
	return false, nil
}

func (f *Filter) IsIgnoredExt(path string) bool {
	ext := filepath.Ext(path)

	for _, e := range f.config.IncludeExts {
		if ext == "."+strings.TrimSpace(e) {
			return false
		}
	}

	return true
}
