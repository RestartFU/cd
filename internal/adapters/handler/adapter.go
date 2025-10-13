package handler

import (
	"fmt"
	"path/filepath"

	"github.com/restartfu/cd/internal/core/domain"
	"github.com/restartfu/cd/internal/ports"
	"github.com/restartfu/cd/pkg/rmux"
)

const dockerImageName = "docker.io/library/ubuntu:latest"

type Adapter struct {
	dockerAdapter ports.Docker
}

func NewAdapter(dockerAdapter ports.Docker) *Adapter {
	return &Adapter{dockerAdapter: dockerAdapter}
}

func (a *Adapter) Deploy(q domain.QueryDeploy) *rmux.Response {
	containerName := filepath.Base(q.GitURL)
	_ = a.dockerAdapter.DestroyContainer(containerName) // ignore if not found

	// Simulate a realistic "deployment" inside Ubuntu
	// This chain of commands will:
	// 1. Update packages
	// 2. Install git
	// 3. Clone a repository
	// 4. Attempt to run a non-existent build command (forces failure)

	cmds := []string{
		"apt-get update",
		"apt-get install -y git",
	}

	cmds = append(cmds, []string{
		fmt.Sprintf("git clone %s", q.GitURL),
		fmt.Sprintf("cd %s", containerName),
		"apt-get install -y $(jq -r '.packages[]' deps.json)",
		`vars=$(jq -r '.variables | to_entries | map("\(.key)=\(.value)") | .[]' deps.json)
cmd=$(jq -r '.command' deps.json)

for var in $vars; do
  export $var
done

eval $cmd`,
	}...)

	msg, err := a.dockerAdapter.CreateAndStartContainer(dockerImageName, containerName, cmds)
	if err != nil {
		fmt.Println(err)
		return rmux.NewResponse(err.Error(), 200)
	}
	fmt.Println(msg)
	return rmux.NewResponse("SUCCESS", 200)
}
