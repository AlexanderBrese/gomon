package reload

import (
	"fmt"
	"io"
	"os"
)

func (r *Reload) BuildCleanup() {
	select {
	case <-r.startBuilding:
		r.stop <- true
	default:
	}
}

func (r *Reload) build() error {
	binary, err := r.config.Binary()
	if err != nil {
		return err
	}
	srcDir, err := r.config.SrcDir()
	if err != nil {
		return err
	}
	buildCmd := fmt.Sprintf("%s %s %s", r.config.BuildCommand, binary, srcDir)
	cmd, stdout, stderr, err := r.startCmd(buildCmd)
	if err != nil {
		return err
	}
	defer func() {
		stdout.Close()
		stderr.Close()
	}()
	_, _ = io.Copy(os.Stdout, stdout)
	_, _ = io.Copy(os.Stderr, stderr)

	return cmd.Wait()
}
