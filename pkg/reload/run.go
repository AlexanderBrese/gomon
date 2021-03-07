package reload

import (
	"io"
	"os"

	"github.com/AlexanderBrese/GOATmon/pkg/utils"
)

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
	utils.WithLock(&r.mu, func() {
		r.running = true
		r.FinishedRunning <- true
	})

	go func() {
		_, _ = io.Copy(os.Stdout, stdout)
		_, _ = io.Copy(os.Stderr, stderr)
	}()

	go r.kill(cmd, stdout, stderr)
	return nil
}
