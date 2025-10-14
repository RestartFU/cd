package protocol

import (
	"fmt"
	"strings"

	"github.com/restartfu/cd/internal/adapters/handler"
)

// LoggingHandler wraps the original handler to provide real-time logging
type LoggingHandler struct {
	originalHandler *handler.Adapter
	client          *ClientConnection
	server          *Server
}

// Deploy executes deployment with real-time logging
func (h *LoggingHandler) Deploy(gitURL, environment string) *handler.DeployResult {
	return h.DeployWithEnv(gitURL, environment, nil, nil)
}

// DeployWithEnv executes deployment with environment variables and real-time logging
func (h *LoggingHandler) DeployWithEnv(gitURL, environment string, envVars, secrets map[string]string) *handler.DeployResult {
	h.server.sendLog(h.client, "info", fmt.Sprintf("Starting deployment for repository: %s", gitURL), "system")
	h.server.sendStatus(h.client, "preparing", "Preparing deployment", 10)

	// Log environment variables (without sensitive values)
	if len(envVars) > 0 {
		h.server.sendLog(h.client, "info", fmt.Sprintf("Environment variables: %d provided", len(envVars)), "system")
	}
	if len(secrets) > 0 {
		h.server.sendLog(h.client, "info", fmt.Sprintf("Secrets: %d provided", len(secrets)), "system")
	}

	// Destroy existing container is handled by the original handler
	h.server.sendLog(h.client, "info", "Checking for existing container...", "docker")

	h.server.sendStatus(h.client, "cloning", "Cloning repository", 25)

	// Execute deployment using the original handler with logging
	h.server.sendLog(h.client, "info", "Executing deployment...", "system")
	result := h.originalHandler.DeployWithEnv(gitURL, environment, envVars, secrets)

	if result.Error != nil {
		h.server.sendLog(h.client, "error", fmt.Sprintf("Deployment failed: %v", result.Error), "system")
	} else {
		h.server.sendLog(h.client, "info", result.Message, "system")
	}

	h.server.sendStatus(h.client, "complete", "Deployment completed", 100)
	return result
}

// LoggingProgress implements git.Progress to capture clone progress
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
