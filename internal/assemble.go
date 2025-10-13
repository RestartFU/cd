package internal

import (
	"github.com/restartfu/cd/internal/adapters/docker"
	"github.com/restartfu/cd/internal/adapters/handler"
	"github.com/restartfu/cd/internal/config"
	"github.com/restartfu/cd/internal/core/service"
	"github.com/restartfu/cd/pkg/rmux"
)

func Assemble(cfg config.Config, allowFunc rmux.AllowFunc) (*service.Service, error) {
	dockerAdapter, err := docker.NewAdapter()
	if err != nil {
		return nil, err
	}
	handler := handler.NewAdapter(dockerAdapter)
	s := service.NewService(handler, allowFunc)
	return s, nil
}
