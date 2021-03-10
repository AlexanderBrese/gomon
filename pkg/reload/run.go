package reload

import (
	"io"

	"github.com/AlexanderBrese/gomon/pkg/logging"
	"github.com/AlexanderBrese/gomon/pkg/utils"
)

// RunCleanup stops the run
func (r *Reload) RunCleanup() {
	utils.WithLock(&r.mu, func() {
		if r.running {
			r.stopRunning <- true
		}
	})
}

func (r *Reload) run() {
	cmd, stdout, stderr, err := r.StartCmd(r.config.Build.ExecutionCommand)
	if err != nil {
		r.logger.Run("error: during run: %s", err)
		return
	}

	r.FinishedRunning <- true
	r.logger.Run("%s", "running")
	utils.WithLock(&r.mu, func() {
		r.running = true
	})

	go func() {
		_, _ = io.Copy(&logging.RunWriter{Logger: r.logger}, stdout)
		_, _ = io.Copy(&logging.ErrorWriter{Logger: r.logger}, stderr)
	}()

	go func() {
		if err := r.kill(cmd, stdout, stderr); err != nil {
			r.logger.Run("error: during kill: %s", err)
			return
		}
	}()
}
