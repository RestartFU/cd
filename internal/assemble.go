package internal

import (
	"github.com/restartfu/cd/internal/adapters/docker"
	"github.com/restartfu/cd/internal/adapters/handler"
	"github.com/restartfu/cd/internal/config"
	"github.com/restartfu/cd/internal/protocol"
)

func Assemble(cfg config.Config) (*protocol.Server, error) {
	dockerAdapter, err := docker.NewAdapter()
	if err != nil {
		return nil, err
	}

	h := handler.NewAdapter(dockerAdapter, cfg)
	return protocol.NewServer(h, cfg), nil
}
