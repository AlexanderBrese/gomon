package reload

import (
	"io"
	"os"
	"os/exec"

	"github.com/AlexanderBrese/Gomon/pkg/utils"
)

func (r *Reload) kill(cmd *exec.Cmd, stdout io.ReadCloser, stderr io.ReadCloser) error {
	<-r.stopRunning
	defer func() {
		stdout.Close()
		stderr.Close()
		r.FinishedKilling <- true
		r.removeBinary()
	}()

	var err error
	_, err = r.killCmd(cmd)
	if err != nil {
		if cmd.ProcessState != nil && !cmd.ProcessState.Exited() {
			os.Exit(1)
		}
		return err
	}
	utils.WithLock(&r.mu, func() {
		r.running = false
	})

	return nil
}

func (r *Reload) removeBinary() error {
	binary, err := r.config.Binary()
	if err != nil {
		return err
	}
	return utils.RemoveFileIfExist(binary)
}
