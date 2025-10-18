package ports

type Docker interface {
	BuildAndStartContainer(dockerfileDir, containerName string, envVars map[string]string) (string, error)
	DestroyContainer(nameOrID string) error
}

type Handler interface {
	Deploy(gitURL, environment string, envVars, secrets map[string]string) *DeployResult
}

type DeployResult struct {
	Message string
	Success bool
	Error   error
}
