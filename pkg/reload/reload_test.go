package reload

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/AlexanderBrese/go-server-browser-reload/pkg/configuration"
	"github.com/AlexanderBrese/go-server-browser-reload/pkg/utils"
)

const CHECK_RELOADING_DELAY = 600

func TestReload(t *testing.T) {
	cfg, err := configuration.TestConfiguration()
	if err != nil {
		t.Error(err)
	}
	reloader := NewReload(cfg)

	if err := buildPrepare(cfg); err != nil {
		t.Error(err)
	}

	defer runCleanup(reloader)

	reloadStart(reloader)
	time.Sleep(CHECK_RELOADING_DELAY * time.Millisecond)
	if err := reloadPassed(reloader); err != nil {
		t.Error(err)
	}
}

func reloadStart(reloader *Reload) {
	reloader.Reload()
}

func reloadPassed(reloader *Reload) error {
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
