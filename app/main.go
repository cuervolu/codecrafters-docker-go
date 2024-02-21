package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"syscall"
)

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {
	command := os.Args[3]
	args := os.Args[4:len(os.Args)]

	cmd, err := tempEnv(command, args)
	if err != nil {
		fmt.Printf("Error creating temporary environment: %v\n", err)
		os.Exit(1)
	}

	if err := cmd.Run(); err != nil {
		handleRunError(err)
	}
}

func tempEnv(command string, args []string) (*exec.Cmd, error) {
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	//Create temporary directory
	dir, err := os.MkdirTemp("", "temp")
	if err != nil {
		return nil, err
	}

	//Copy command to temporary directory
	if err := CopyFile(command, path.Join(dir, command)); err != nil {
		return nil, err
	}

	// Create /dev/null
	if err := os.MkdirAll("/dev", 0755); err != nil {
		return nil, err
	}
	if devNull, err := os.Create("/dev/null"); err != nil {
		return nil, err
	} else {
		err := devNull.Close()
		if err != nil {
			return nil, err
		}
	}

	// Chroot to temporary directory
	if err := syscall.Chroot(dir); err != nil {
		return nil, err
	}
	// Change working directory to /
	if err := syscall.Chdir("/"); err != nil {
		return nil, err
	}
	return cmd, nil
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(srcFile *os.File) {
		err := srcFile.Close()
		if err != nil {
			fmt.Printf("Error closing file: %v\n", err)
		}
	}(srcFile)

	// Create the destination file
	if err := os.MkdirAll(path.Dir(dst), 0755); err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(dstFile *os.File) {
		err := dstFile.Close()
		if err != nil {
			fmt.Printf("Error closing file: %v\n", err)
		}
	}(dstFile)
	if err := os.Chmod(dst, 0755); err != nil {
		return err
	}
	// Copy the file
	_, err = io.Copy(dstFile, srcFile)
	return err
}

// handleRunError handles errors that occur during command execution
func handleRunError(err error) {
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		os.Exit(exitError.ExitCode())
	}
	fmt.Printf("Error executing command: %v\n", err)
	os.Exit(1)
}
