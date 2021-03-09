package reload

import (
	"fmt"
	"io"

	"github.com/AlexanderBrese/gomon/pkg/utils"
)

// BuildCleanup stops the build
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
	buildCmd := fmt.Sprintf("%s %s %s", r.config.Build.Command, binary, srcDir)
	cmd, stdout, stderr, err := r.StartCmd(buildCmd)
	if err != nil {
		return err
	}
	defer func() {
		stdout.Close()
		stderr.Close()
	}()
	buildLog, err := r.logger.BuildLog()
	if err != nil {
		return err
	}
	defer utils.CloseFile(buildLog)
	_, _ = io.Copy(buildLog, stdout)
	_, _ = io.Copy(buildLog, stderr)

	return cmd.Wait()
}
