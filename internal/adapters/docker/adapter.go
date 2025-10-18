package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type Adapter struct {
	cli *client.Client
}

func NewAdapter() (*Adapter, error) {
	dockerHost := getDockerHost()

	var opts []client.Opt
	if dockerHost != "" {
		log.Printf("Using Docker host: %s", dockerHost)
		opts = append(opts, client.WithHost(dockerHost))
	} else {
		log.Printf("Using Docker client auto-detection (FromEnv)")
		opts = append(opts, client.FromEnv)
	}
	opts = append(opts, client.WithAPIVersionNegotiation())

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to init docker client: %w", err)
	}

	ctx := context.Background()
	_, err = cli.Ping(ctx)
	if err != nil {
		log.Printf("Docker ping failed: %v", err)
		return nil, fmt.Errorf("failed to connect to Docker daemon: %w", err)
	}

	log.Printf("Successfully connected to Docker daemon")
	return &Adapter{cli: cli}, nil
}

func DiagnoseDocker() error {
	log.Printf("=== Docker Diagnostics ===")
	log.Printf("OS: %s", runtime.GOOS)
	log.Printf("Architecture: %s", runtime.GOARCH)

	if host := os.Getenv("DOCKER_HOST"); host != "" {
		log.Printf("DOCKER_HOST: %s", host)
	} else {
		log.Printf("DOCKER_HOST: not set")
	}

	switch runtime.GOOS {
	case "darwin":
		paths := []string{
			os.ExpandEnv("$HOME/.docker/run/docker.sock"),
			"/var/run/docker.sock",
		}
		for _, path := range paths {
			if stat, err := os.Stat(path); err == nil {
				log.Printf("Socket found: %s (size: %d bytes)", path, stat.Size())
			} else {
				log.Printf("Socket missing: %s (%v)", path, err)
			}
		}
	case "linux":
		path := "/var/run/docker.sock"
		if stat, err := os.Stat(path); err == nil {
			log.Printf("Socket found: %s (size: %d bytes)", path, stat.Size())
		} else {
			log.Printf("Socket missing: %s (%v)", path, err)
		}
	case "windows":
		log.Printf("Windows: using named pipe ////./pipe/docker_engine")
	}

	adapter, err := NewAdapter()
	if err != nil {
		log.Printf("Failed to create Docker adapter: %v", err)
		return err
	}

	ctx := context.Background()
	info, err := adapter.cli.Info(ctx)
	if err != nil {
		log.Printf("Failed to get Docker info: %v", err)
		return err
	}

	log.Printf("Docker info: %s %s", info.ServerVersion, info.OSType)
	log.Printf("=== Diagnostics Complete ===")
	return nil
}

func (a *Adapter) BuildAndStartContainer(dockerfileDir, containerName string, envVars map[string]string) (string, error) {
	ctx := context.Background()

	tarBuf := new(bytes.Buffer)
	tw := tar.NewWriter(tarBuf)

	err := filepath.Walk(dockerfileDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(dockerfileDir, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		hdr := &tar.Header{
			Name: relPath,
			Mode: 0644,
			Size: int64(len(data)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		_, err = tw.Write(data)
		return err
	})
	if err != nil {
		return "", fmt.Errorf("failed to create tar of Dockerfile dir: %w", err)
	}
	tw.Close()

	imageBuildResp, err := a.cli.ImageBuild(ctx, bytes.NewReader(tarBuf.Bytes()), build.ImageBuildOptions{
		Tags:       []string{containerName + ":latest"},
		Remove:     true,
		Dockerfile: "Dockerfile",
	})
	if err != nil {
		return "", fmt.Errorf("failed to build image: %w", err)
	}
	defer imageBuildResp.Body.Close()
	io.Copy(os.Stdout, imageBuildResp.Body)

	inspect, _, err := a.cli.ImageInspectWithRaw(ctx, containerName+":latest")
	if err != nil {
		return "", fmt.Errorf("failed to inspect image: %w", err)
	}

	portBindings := nat.PortMap{}
	for port := range inspect.Config.ExposedPorts {
		portBindings[nat.Port(port)] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: port,
			},
		}
	}

	var envSlice []string
	for key, value := range envVars {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", key, value))
	}

	resp, err := a.cli.ContainerCreate(ctx, &container.Config{
		Image: containerName + ":latest",
		Tty:   false,
		Env:   envSlice,
	}, &container.HostConfig{
		PortBindings: portBindings,
	}, nil, nil, containerName)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	if err := a.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	return fmt.Sprintf("✅ Container started: %s", resp.ID), nil
}

func (a *Adapter) DestroyContainer(nameOrID string) error {
	ctx := context.Background()
	err := a.cli.ContainerRemove(ctx, nameOrID, container.RemoveOptions{Force: true})
	if err != nil {
		if errdefs.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to remove container: %w", err)
	}
	return nil
}
