package internal

import (
	"github.com/restartfu/cd/internal/adapters/docker"
	"github.com/restartfu/cd/internal/adapters/handler"
)

// CreateHandler creates a complete handler with docker adapter
func CreateHandler() (*handler.Adapter, error) {
	dockerAdapter, err := docker.NewAdapter()
	if err != nil {
		return nil, err
	}
	return handler.NewAdapter(dockerAdapter), nil
}
