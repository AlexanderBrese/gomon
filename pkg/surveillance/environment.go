package surveillance

import (
	"github.com/AlexanderBrese/Gomon/pkg/browsersync"
	"github.com/AlexanderBrese/Gomon/pkg/configuration"
	"github.com/AlexanderBrese/Gomon/pkg/reload"
	"github.com/AlexanderBrese/Gomon/pkg/utils"
)

type Environment struct {
	config   *configuration.Configuration
	detector *utils.Batcher
	reloader *reload.Reload
	sync     *browsersync.Server

	stopDetecting  chan bool
	stopRefreshing chan bool
}

func NewEnvironment(cfg *configuration.Configuration) (*Environment, error) {
	batcher, err := utils.NewBatcher(cfg.BufferTime())
	if err != nil {
		return nil, err
	}
	e := &Environment{
		config:         cfg,
		detector:       batcher,
		stopDetecting:  make(chan bool, 1),
		stopRefreshing: make(chan bool, 1),
	}

	if cfg.Reload {
		e.reloader = reload.NewReload(cfg)
		if err := e.checkRunEnvironment(); err != nil {
			return nil, err
		}
	}

	if cfg.Sync {
		e.sync = browsersync.NewServer(cfg.Port)
		e.sync.Start()
	}

	return e, nil
}

func (e *Environment) Teardown() {
	if e.config.Reload {
		e.reloader.Cleanup()
		<-e.reloader.FinishedKilling
	}

	if e.config.Sync {
		e.sync.Stop()
	}

	e.detector.Close()
	e.stopDetecting <- true
	e.stopRefreshing <- true
}

func (e *Environment) checkRunEnvironment() error {
	buildDir, err := e.config.BuildDir()
	if err != nil {
		return err
	}
	return utils.CreateBuildDirIfNotExist(buildDir)
}
