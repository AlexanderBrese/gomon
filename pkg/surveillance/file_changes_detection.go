package surveillance

import (
	"fmt"

	"github.com/AlexanderBrese/GOATmon/pkg/browsersync"
	"github.com/AlexanderBrese/GOATmon/pkg/configuration"
	"github.com/AlexanderBrese/GOATmon/pkg/reload"
	"github.com/AlexanderBrese/GOATmon/pkg/utils"
)

const MAX_WATCHED_FILES = 1000

type FileChangesDetection struct {
	config   *configuration.Configuration
	watcher  *utils.Batcher
	reloader *reload.Reload

	stopWatching chan bool
	syncServer   *browsersync.Server

	changeDetectedSubscription chan bool
	changeDetected             chan bool
	watchedFileChecksums       *utils.FileChecksums
}

func NewFileChangesDetection(cfg *configuration.Configuration) (*FileChangesDetection, error) {
	batcher, err := utils.NewBatcher(cfg.BufferTime())
	if err != nil {
		return nil, err
	}

	w := &FileChangesDetection{
		config:                     cfg,
		watcher:                    batcher,
		changeDetectedSubscription: nil,
		changeDetected:             make(chan bool, MAX_WATCHED_FILES),
		stopWatching:               make(chan bool),
		watchedFileChecksums:       utils.NewFileChecksums(),
	}

	if cfg.Reload {
		w.reloader = reload.NewReload(cfg)
	}

	if cfg.Sync {
		w.syncServer = browsersync.NewServer(cfg.Port)
	}

	return w, nil
}

func (w *FileChangesDetection) Subscribe(watchedFilesSubscription chan bool) {
	w.changeDetectedSubscription = watchedFilesSubscription
}

func (w *FileChangesDetection) Init() error {
	if w.config.Reload {
		if err := w.checkRunEnvironment(); err != nil {
			return err
		}
	}

	if err := w.crawlWatchedDirs(w.config.Root); err != nil {
		return err
	}

	if w.config.Sync {
		w.syncServer.Start()
	}

	return nil
}

func (w *FileChangesDetection) Surveil() error {
	go w.watch()

	if err := w.control(); err != nil {
		return err
	}

	w.cleanup()
	return nil
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
	firstRun := make(chan bool, 1)
	firstRun <- true

	for {
		select {
		case <-w.stopWatching:
			return nil
		case <-w.changeDetected:
			fmt.Println("change detected")
		case <-firstRun:
			break
		}

		if w.config.Reload {
			w.reloader.Reload()
			<-w.reloader.FinishedRunning
		}
		if w.config.Sync {
			w.syncServer.Sync()
		}
	}
}

func (w *FileChangesDetection) cleanup() {
	w.stopWatcher()

	if w.config.Reload {
		w.reloader.Cleanup()
	}

	if w.config.Sync {
		w.syncServer.Stop()
	}
}
