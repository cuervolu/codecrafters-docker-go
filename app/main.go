package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {
	command := os.Args[3]
	args := os.Args[4:len(os.Args)]

	cmd := exec.Command(command, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		//Check the status of the error
		var e *exec.ExitError
		ok := errors.As(err, &e) // Check if the error is an ExitError
		if ok {
			os.Exit(e.ExitCode()) // Exit with the exit code of the command
		}
		fmt.Printf("Error: %v\n", err)
	}
}
