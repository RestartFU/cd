package ports

import (
	"github.com/restartfu/cd/internal/core/domain"
	"github.com/restartfu/cd/pkg/rmux"
)

type Handler interface {
	Deploy(domain.QueryDeploy) *rmux.Response
}

type Docker interface {
	CreateAndStartContainer(imageName, name string, cmd []string) (string, error)
	DestroyContainer(containerID string) error
}
