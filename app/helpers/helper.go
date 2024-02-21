package helpers

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// HandleRunError is a helper function to handle errors from running a command
func HandleRunError(err error) error {
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		os.Exit(exitError.ExitCode())
	}
	return fmt.Errorf("error executing command: %v", err)
}
