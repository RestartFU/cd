package handler

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/restartfu/cd/internal/core/domain"
	"github.com/restartfu/cd/internal/ports"
	"github.com/restartfu/cd/pkg/rmux"
)

type Adapter struct {
	dockerAdapter ports.Docker
}

func NewAdapter(dockerAdapter ports.Docker) *Adapter {
	return &Adapter{dockerAdapter: dockerAdapter}
}

func (a *Adapter) Deploy(q domain.QueryDeploy) *rmux.Response {
	containerName := filepath.Base(q.GitURL)

	// Destroy existing container
	_ = a.dockerAdapter.DestroyContainer(containerName)

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "repo-*")
	if err != nil {
		fmt.Println("Failed to create temp dir:", err)
		return rmux.NewResponse(err.Error(), 200)
	}
	defer os.RemoveAll(tmpDir) // cleanup

	// Clone repository using go-git
	fmt.Println("Cloning repository:", q.GitURL)
	_, err = git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL:      q.GitURL,
		Progress: os.Stdout,
	})
	if err != nil {
		fmt.Println("Git clone failed:", err)
		return rmux.NewResponse(err.Error(), 200)
	}

	// Build and start Docker container
	msg, err := a.dockerAdapter.BuildAndStartContainer(tmpDir, containerName)
	if err != nil {
		fmt.Println("Docker build/start failed:", err)
		return rmux.NewResponse(err.Error(), 200)
	}

	fmt.Println(msg)
	return rmux.NewResponse("SUCCESS", 200)
}
