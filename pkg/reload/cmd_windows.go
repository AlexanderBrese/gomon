package reload

import (
	"io"
	"os/exec"
	"strconv"
	"strings"
)

func (r *Reload) StartCmd(cmd string) (*exec.Cmd, io.ReadCloser, io.ReadCloser, error) {
	var err error

	c := exec.Command("cmd", "/c", cmd)
	if !strings.Contains(cmd, ".exe") {
		r.logger.Run("CMD will not recognize non .exe file for execution, path: %s", cmd)
	}
	stderr, err := c.StderrPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	stdout, err := c.StdoutPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	err = c.Start()
	if err != nil {
		return nil, nil, nil, err
	}

	return c, stdout, stderr, err
}

func (r *Reload) KillCmd(cmd *exec.Cmd) (pid int, err error) {
	pid = cmd.Process.Pid
	// https://stackoverflow.com/a/44551450
	kill := exec.Command("TASKKILL", "/T", "/F", "/PID", strconv.Itoa(pid))
	return pid, kill.Run()
}
