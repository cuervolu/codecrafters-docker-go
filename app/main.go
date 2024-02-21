package main

import (
	"fmt"
	"github.com/codecrafters-io/docker-starter-go/app/helpers"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/codecrafters-io/docker-starter-go/app/docker"
)

func main() {
	image := os.Args[2]
	command := os.Args[3]
	args := os.Args[4:]

	err := runContainer(image, command, args)
	if err != nil {
		log.Fatalf("Error running container: %v", err)
	}
}

func runContainer(image, command string, args []string) error {
	imageRetriever, err := docker.NewOCIImageRetriever(docker.ParseImageString(image))
	if err != nil {
		return helpers.HandleRunError(fmt.Errorf("failed to create image retriever: %w", err))
	}

	imagesDir, err := imageRetriever.Pull()
	if err != nil {
		return helpers.HandleRunError(fmt.Errorf("failed to pull image: %w", err))
	}
	defer removeDirectory(imagesDir)

	containerDir, err := os.MkdirTemp("", "containers-root")
	if err != nil {
		return helpers.HandleRunError(fmt.Errorf("failed to create temporary directory: %w", err))
	}
	defer removeDirectory(containerDir)

	err = extractTarFiles(imagesDir, containerDir)
	if err != nil {
		return helpers.HandleRunError(fmt.Errorf("failed to extract tar files: %w", err))
	}

	err = setupContainer(containerDir)
	if err != nil {
		return helpers.HandleRunError(fmt.Errorf("failed to setup container: %w", err))
	}

	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return helpers.HandleRunError(fmt.Errorf("failed to run command: %w", err))
	}

	return nil
}

func extractTarFiles(sourceDir, targetDir string) error {
	files, err := os.ReadDir(sourceDir)
	if err != nil {
		return helpers.HandleRunError(fmt.Errorf("failed to read directory: %w", err))
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".tar") {
			tarPath := filepath.Join(sourceDir, file.Name())
			cmd := exec.Command("tar", "-xvf", tarPath, "-C", targetDir)
			err := cmd.Run()
			if err != nil {
				return helpers.HandleRunError(fmt.Errorf("failed to extract tar file: %w", err))
			}
		}
	}
	return nil
}

func setupContainer(containerDir string) error {
	if err := syscall.Chroot(containerDir); err != nil {
		return helpers.HandleRunError(fmt.Errorf("failed to chroot: %w", err))
	}

	if err := os.Chdir("/"); err != nil {
		return helpers.HandleRunError(fmt.Errorf("failed to chdir: %w", err))
	}

	devNull, err := os.Create("/dev/null")
	if err != nil {
		return helpers.HandleRunError(fmt.Errorf("failed to create /dev/null: %w", err))
	}
	if cerr := devNull.Close(); cerr != nil {
		return helpers.HandleRunError(fmt.Errorf("failed to close /dev/null: %w", cerr))
	}

	return nil
}

func removeDirectory(dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		log.Printf("Failed to remove directory: %v", err)
	}
}
