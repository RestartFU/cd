package protocol

import (
	"fmt"
	"strings"

	"github.com/restartfu/cd/internal/ports"
)

type LoggingHandler struct {
	handler ports.Handler
	client  *ClientConnection
	server  *Server
}

func (h *LoggingHandler) Deploy(gitURL, environment string, envVars, secrets map[string]string) *ports.DeployResult {
	h.server.sendLog(h.client, "info", fmt.Sprintf("Starting deployment for repository: %s", gitURL), "system")
	h.server.sendStatus(h.client, "preparing", "Preparing deployment", 10)

	if len(envVars) > 0 {
		h.server.sendLog(h.client, "info", fmt.Sprintf("Environment variables: %d provided", len(envVars)), "system")
	}
	if len(secrets) > 0 {
		h.server.sendLog(h.client, "info", fmt.Sprintf("Secrets: %d provided", len(secrets)), "system")
	}

	h.server.sendLog(h.client, "info", "Checking for existing container...", "docker")

	h.server.sendStatus(h.client, "cloning", "Cloning repository", 25)

	h.server.sendLog(h.client, "info", "Executing deployment...", "system")
	result := h.handler.Deploy(gitURL, environment, envVars, secrets)

	if result.Error != nil {
		h.server.sendLog(h.client, "error", fmt.Sprintf("Deployment failed: %v", result.Error), "system")
	} else {
		h.server.sendLog(h.client, "info", result.Message, "system")
	}

	h.server.sendStatus(h.client, "complete", "Deployment completed", 100)
	return result
}

type LoggingProgress struct {
	handler *LoggingHandler
}

func (p *LoggingProgress) Write(data []byte) (int, error) {
	message := strings.TrimSpace(string(data))
	if message != "" {
		p.handler.server.sendLog(p.handler.client, "info", message, "git")
	}
	return len(data), nil
}
