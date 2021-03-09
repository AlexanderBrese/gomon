package reload

import (
	"io"
	"os/exec"
	"syscall"
	"time"

	"github.com/creack/pty"
)

const killDelay = 100

func (r *Reload) StartCmd(cmd string) (*exec.Cmd, io.ReadCloser, io.ReadCloser, error) {
	c := exec.Command("/bin/sh", "-c", cmd)

	f, err := pty.Start(c)
	return c, f, f, err
}

func (r *Reload) KillCmd(cmd *exec.Cmd) (pid int, err error) {
	pid = cmd.Process.Pid

	// Sending a signal to make it clear to the process that it is time to turn off
	if err = syscall.Kill(-pid, syscall.SIGINT); err != nil {
		return
	}
	time.Sleep(killDelay * time.Millisecond)

	// https://stackoverflow.com/questions/22470193/why-wont-go-kill-a-child-process-correctly
	err = syscall.Kill(-pid, syscall.SIGKILL)

	// Wait releases any resources associated with the Process.
	_, _ = cmd.Process.Wait()
	return
}
