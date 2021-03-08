package surveillance

import (
	"github.com/AlexanderBrese/GOATmon/pkg/browsersync"
	"github.com/AlexanderBrese/GOATmon/pkg/configuration"
	"github.com/AlexanderBrese/GOATmon/pkg/reload"
	"github.com/AlexanderBrese/GOATmon/pkg/utils"
)

type Environment struct {
	config   *configuration.Configuration
	detector *utils.Batcher
	reloader *reload.Reload
	sync     *browsersync.Server
}

func NewEnvironment(cfg *configuration.Configuration) (*Environment, error) {
	batcher, err := utils.NewBatcher(cfg.BufferTime())
	if err != nil {
		return nil, err
	}
	s := &Environment{
		config:   cfg,
		detector: batcher,
	}

	if cfg.Reload {
		s.reloader = reload.NewReload(cfg)
	}

	if cfg.Sync {
		s.sync = browsersync.NewServer(cfg.Port)
	}

	return s, nil
}

func (s *Environment) Teardown() {
	s.detector.Close()
	if s.config.Reload {
		s.reloader.Cleanup()
	}

	if s.config.Sync {
		s.sync.Stop()
	}
}

func (s *Environment) Run() error {
	if s.config.Reload {
		if err := s.checkRunEnvironment(); err != nil {
			return err
		}
	}

	if s.config.Sync {
		s.sync.Start()
	}
	return nil
}

func (s *Environment) checkRunEnvironment() error {
	buildDir, err := s.config.BuildDir()
	if err != nil {
		return err
	}
	return utils.CreateBuildDirIfNotExist(buildDir)
}
