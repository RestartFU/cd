package handler

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/restartfu/cd/internal/config"
	"github.com/restartfu/cd/internal/ports"
)

type Adapter struct {
	dockerAdapter ports.Docker
	config        config.Config
}

func NewAdapter(dockerAdapter ports.Docker, cfg config.Config) *Adapter {
	return &Adapter{
		dockerAdapter: dockerAdapter,
		config:        cfg,
	}
}

func (a *Adapter) Deploy(gitURL, environment string, envVars, secrets map[string]string) *ports.DeployResult {
	var auth *ssh.PublicKeys
	var err error

	// Use provided SSH key path, then fall back to config, then to default
	keyPath := a.config.SSHKeyPath
	if keyPath == "" && a.config.SSHKeyPath != "" {
		keyPath = a.config.SSHKeyPath
	}
	if keyPath == "" {
		keyPath = "/home/restart/.ssh/id_ed25519"
	}

	auth, err = ssh.NewPublicKeysFromFile("git", keyPath, "")
	if err != nil {
		fmt.Println("Failed to load SSH key from", keyPath, ":", err)
		return &ports.DeployResult{Message: fmt.Sprintf("Failed to load SSH key from %s: %v", keyPath, err), Success: false, Error: err}
	}
	containerName := filepath.Base(gitURL)

	_ = a.dockerAdapter.DestroyContainer(containerName)

	tmpDir, err2 := os.MkdirTemp("", "repo-*")
	if err2 != nil {
		fmt.Println("Failed to create temp dir:", err2)
		return &ports.DeployResult{Message: err2.Error(), Success: false, Error: err2}
	}
	defer os.RemoveAll(tmpDir)

	fmt.Println("Cloning repository:", gitURL)
	_, err = git.PlainClone(tmpDir, false, &git.CloneOptions{
		Auth:     auth,
		URL:      gitURL,
		Progress: os.Stdout,
	})
	if err != nil {
		fmt.Println("Git clone failed:", err)
		return &ports.DeployResult{Message: err.Error(), Success: false, Error: err}
	}

	allEnvVars := make(map[string]string)

	for k, v := range envVars {
		allEnvVars[k] = v
	}

	for k, v := range secrets {
		allEnvVars[k] = v
	}

	allEnvVars["DEPLOYMENT_ENV"] = environment

	msg, err := a.dockerAdapter.BuildAndStartContainer(tmpDir, containerName, allEnvVars)
	if err != nil {
		fmt.Println("Docker build/start failed:", err)
		return &ports.DeployResult{Message: err.Error(), Success: false, Error: err}
	}

	fmt.Println(msg)
	return &ports.DeployResult{Message: "SUCCESS", Success: true, Error: nil}
}
