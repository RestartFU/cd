package handler

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

type DockerAdapter interface {
	BuildAndStartContainer(imageName, name string, envVars map[string]string) (string, error)
	DestroyContainer(containerID string) error
}

type DeployResult struct {
	Message string
	Success bool
	Error   error
}

type Adapter struct {
	dockerAdapter DockerAdapter
}

func NewAdapter(dockerAdapter DockerAdapter) *Adapter {
	return &Adapter{dockerAdapter: dockerAdapter}
}

func (a *Adapter) Deploy(gitURL, environment string) *DeployResult {
	return a.DeployWithEnv(gitURL, environment, nil, nil)
}

func (a *Adapter) DeployWithEnv(gitURL, environment string, envVars, secrets map[string]string) *DeployResult {
	containerName := filepath.Base(gitURL)

	// Destroy existing container
	_ = a.dockerAdapter.DestroyContainer(containerName)

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "repo-*")
	if err != nil {
		fmt.Println("Failed to create temp dir:", err)
		return &DeployResult{Message: err.Error(), Success: false, Error: err}
	}
	defer os.RemoveAll(tmpDir) // cleanup

	// Clone repository using go-git
	fmt.Println("Cloning repository:", gitURL)
	_, err = git.PlainClone(tmpDir, false, &git.CloneOptions{
		URL:      gitURL,
		Progress: os.Stdout,
	})
	if err != nil {
		fmt.Println("Git clone failed:", err)
		return &DeployResult{Message: err.Error(), Success: false, Error: err}
	}

	// Merge environment variables and secrets
	allEnvVars := make(map[string]string)

	// Add environment variables
	for k, v := range envVars {
		allEnvVars[k] = v
	}

	// Add secrets (they can override env vars)
	for k, v := range secrets {
		allEnvVars[k] = v
	}

	// Add deployment environment
	allEnvVars["DEPLOYMENT_ENV"] = environment

	// Build and start Docker container
	msg, err := a.dockerAdapter.BuildAndStartContainer(tmpDir, containerName, allEnvVars)
	if err != nil {
		fmt.Println("Docker build/start failed:", err)
		return &DeployResult{Message: err.Error(), Success: false, Error: err}
	}

	fmt.Println(msg)
	return &DeployResult{Message: "SUCCESS", Success: true, Error: nil}
}
