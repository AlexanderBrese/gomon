package reload

import (
	"sync"

	"github.com/AlexanderBrese/GOATmon/pkg/configuration"
)

// Reload recompiles the build and restarts the binary
type Reload struct {
	config *configuration.Configuration
	mu     sync.Mutex

	running         bool
	startBuilding   chan bool
	stop            chan bool
	stopRunning     chan bool
	FinishedRunning chan bool
	FinishedKilling chan bool
}

// NewReload creates a new Reload with the config provided
func NewReload(cfg *configuration.Configuration) *Reload {
	return &Reload{
		config:          cfg,
		running:         false,
		startBuilding:   make(chan bool, 1),
		stop:            make(chan bool, 1),
		stopRunning:     make(chan bool),
		FinishedRunning: make(chan bool, 1),
		FinishedKilling: make(chan bool, 1),
	}
}

// Cleanup stops the current build and the run
func (r *Reload) Cleanup() {
	r.BuildCleanup()
	r.RunCleanup()
}

// Run cleans up and starts the new build
func (r *Reload) Run() {
	r.Cleanup()
	go r.start()
}

func (r *Reload) start() error {
	r.startBuilding <- true
	defer func() {
		<-r.startBuilding
	}()

	select {
	case <-r.stop:
		return nil
	default:
	}
	if err := r.build(); err != nil {
		return err
	}

	select {
	case <-r.stop:
		return nil
	default:
	}

	return r.run()
}
