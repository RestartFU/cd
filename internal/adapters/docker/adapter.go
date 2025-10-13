package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type Adapter struct {
	cli *client.Client
}

func NewAdapter() (*Adapter, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to init docker client: %w", err)
	}
	return &Adapter{cli: cli}, nil
}

func (a *Adapter) BuildAndStartContainer(dockerfileDir, containerName string) (string, error) {
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
				HostPort: port, // port.Port() returns string, fine here
			},
		}
	}

	resp, err := a.cli.ContainerCreate(ctx, &container.Config{
		Image: containerName + ":latest",
		Tty:   false,
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
		if client.IsErrNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to remove container: %w", err)
	}
	return nil
}
