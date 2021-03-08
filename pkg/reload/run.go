package reload

import (
	"io"
	"os"

	"github.com/AlexanderBrese/Gomon/pkg/utils"
)

// RunCleanup stops the run
func (r *Reload) RunCleanup() {
	utils.WithLock(&r.mu, func() {
		if r.running {
			r.stopRunning <- true
		}
	})
}

func (r *Reload) run() error {
	cmd, stdout, stderr, err := r.startCmd(r.config.ExecutionCommand)
	if err != nil {
		return err
	}
	r.FinishedRunning <- true
	utils.WithLock(&r.mu, func() {
		r.running = true

	})

	go func() {
		_, _ = io.Copy(os.Stdout, stdout)
		_, _ = io.Copy(os.Stderr, stderr)
	}()

	go r.kill(cmd, stdout, stderr)
	return nil
}
