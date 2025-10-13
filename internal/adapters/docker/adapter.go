package docker

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type Adapter struct {
	cli *client.Client
}

// NewAdapter initializes a Docker client
func NewAdapter() (*Adapter, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to init docker client: %w", err)
	}
	return &Adapter{cli: cli}, nil
}

// CreateAndStartContainer pulls, starts, and executes commands one by one
func (a *Adapter) CreateAndStartContainer(imageName, containerName string, cmds []string) (string, error) {
	ctx := context.Background()

	// Pull image
	out, err := a.cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to pull image: %w", err)
	}
	io.Copy(io.Discard, out)
	out.Close()

	// Create container with a shell that keeps it alive
	resp, err := a.cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Cmd:   []string{"/bin/sh", "-c", "tail -f /dev/null"}, // Keep container running
		Tty:   false,
	}, nil, nil, nil, containerName)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	if err := a.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	// Execute commands one by one
	for i, cmd := range cmds {
		isLastCommand := i == len(cmds)-1

		// Create exec instance
		execConfig := container.ExecOptions{
			Cmd:          []string{"/bin/sh", "-c", cmd},
			AttachStdout: true,
			AttachStderr: true,
		}

		execID, err := a.cli.ContainerExecCreate(ctx, resp.ID, execConfig)
		if err != nil {
			return "", fmt.Errorf("failed to create exec for command %d: %w", i+1, err)
		}

		// Attach to exec
		execAttach, err := a.cli.ContainerExecAttach(ctx, execID.ID, container.ExecStartOptions{})
		if err != nil {
			return "", fmt.Errorf("failed to attach to exec for command %d: %w", i+1, err)
		}

		// Read output using stdcopy to properly demultiplex Docker streams
		var stdout, stderr strings.Builder
		go func() {
			stdcopy.StdCopy(&stdout, &stderr, execAttach.Reader)
		}()

		// For non-last commands, wait indefinitely
		if !isLastCommand {
			done := make(chan error, 1)
			go func() {
				for {
					inspect, err := a.cli.ContainerExecInspect(ctx, execID.ID)
					if err != nil {
						done <- err
						return
					}
					if !inspect.Running {
						if inspect.ExitCode != 0 {
							done <- fmt.Errorf("command %d failed with exit code %d", i+1, inspect.ExitCode)
						} else {
							done <- nil
						}
						return
					}
					time.Sleep(100 * time.Millisecond)
				}
			}()

			err := <-done
			execAttach.Close()
			time.Sleep(100 * time.Millisecond) // Give time for output capture
			if err != nil {
				stdoutStr := strings.TrimSpace(stdout.String())
				stderrStr := strings.TrimSpace(stderr.String())
				return "", fmt.Errorf("%w\nCommand: %s\nStdout: %s\nStderr: %s",
					err, cmd, stdoutStr, stderrStr)
			}
		} else {
			// Last command - give it 10 seconds to start, then assume it's a server
			time.Sleep(10 * time.Second)

			inspect, err := a.cli.ContainerExecInspect(ctx, execID.ID)
			if err != nil {
				execAttach.Close()
				return "", fmt.Errorf("failed to inspect last command: %w", err)
			}

			// If it exited quickly with error, report it
			if !inspect.Running && inspect.ExitCode != 0 {
				execAttach.Close()
				time.Sleep(100 * time.Millisecond) // Give time for output capture
				stdoutStr := strings.TrimSpace(stdout.String())
				stderrStr := strings.TrimSpace(stderr.String())
				return "", fmt.Errorf("last command failed with exit code %d\nCommand: %s\nStdout: %s\nStderr: %s",
					inspect.ExitCode, cmd, stdoutStr, stderrStr)
			}

			execAttach.Close()
			// If still running or exited successfully, consider it good
		}
	}

	return fmt.Sprintf("✅ Deployment container started: %s", resp.ID), nil
}

// DestroyContainer forcibly removes container by name or ID
func (a *Adapter) DestroyContainer(nameOrID string) error {
	ctx := context.Background()
	err := a.cli.ContainerRemove(ctx, nameOrID, container.RemoveOptions{Force: true})
	if err != nil {
		if client.IsErrNotFound(err) {
			return nil // no-op if already gone
		}
		return fmt.Errorf("failed to remove container: %w", err)
	}
	return nil
}
