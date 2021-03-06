package reload

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/AlexanderBrese/go-server-browser-reload/pkg/configuration"
	"github.com/AlexanderBrese/go-server-browser-reload/pkg/utils"
)

const CHECK_RUNNING_DELAY = 300

func TestRun(t *testing.T) {
	cfg, err := configuration.TestConfiguration()
	if err != nil {
		t.Error(err)
	}
	reloader := NewReload(cfg)

	if err := buildPrepare(cfg); err != nil {
		t.Error(err)
	}

	defer runCleanup(reloader)

	if err := runStart(reloader); err != nil {
		t.Error(err)
	}

	time.Sleep(CHECK_RUNNING_DELAY * time.Millisecond)

	if err := runPassed(reloader); err != nil {
		t.Error(err)
	}
}

func runStart(reloader *Reload) error {
	if err := reloader.build(); err != nil {
		return err
	}
	return reloader.run()
}

func runPassed(reloader *Reload) error {
	binary, err := reloader.Configuration().Binary()
	if err != nil {
		return err
	}
	if err := utils.CheckPath(binary); err != nil {
		return fmt.Errorf("error: there was no built binary found at %s", binary)
	}
	return utils.WithLockAndError(&reloader.mu, func() error {
		if !reloader.running {
			return errors.New("error: binary not running")
		}
		return nil
	})
}

func runCleanup(reloader *Reload) error {
	if err := buildCleanup(reloader); err != nil {
		return err
	}

	reloader.RunCleanup()

	return nil
}
