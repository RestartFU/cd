package protocol

import (
	"testing"

	"github.com/restartfu/cd/internal/adapters/handler"
	"github.com/restartfu/cd/internal/config"
)

// MockDockerAdapterIntegration implements DockerAdapter for integration testing
type MockDockerAdapterIntegration struct {
	buildAndStartResult string
	buildAndStartError  error
	destroyError        error
	calls               []string
}

func (m *MockDockerAdapterIntegration) BuildAndStartContainer(imageName, name string, envVars map[string]string) (string, error) {
	m.calls = append(m.calls, "BuildAndStartContainer")
	return m.buildAndStartResult, m.buildAndStartError
}

func (m *MockDockerAdapterIntegration) DestroyContainer(containerID string) error {
	m.calls = append(m.calls, "DestroyContainer")
	return m.destroyError
}

func (m *MockDockerAdapterIntegration) GetCalls() []string {
	result := make([]string, len(m.calls))
	copy(result, m.calls)
	return result
}

func TestHandlerIntegration(t *testing.T) {
	// Test that handler works with mock docker
	mockDocker := &MockDockerAdapterIntegration{
		buildAndStartResult: "Container abc123 started successfully",
		buildAndStartError:  nil,
		destroyError:        nil,
	}

	// Create handler
	handlerAdapter := handler.NewAdapter(mockDocker)

	// Test deployment (will fail at git clone but tests the flow)
	result := handlerAdapter.Deploy("invalid-git-url", "test")

	if result == nil {
		t.Fatal("Expected deployment result")
	}

	// Should fail due to invalid git URL
	if result.Success {
		t.Error("Expected deployment to fail with invalid git URL")
	}

	if result.Error == nil {
		t.Error("Expected error to be set")
	}

	// Verify docker methods were called
	calls := mockDocker.GetCalls()
	if len(calls) == 0 {
		t.Error("Expected docker methods to be called")
	}

	// Should have destroy called
	foundDestroy := false
	for _, call := range calls {
		if call == "DestroyContainer" {
			foundDestroy = true
			break
		}
	}

	if !foundDestroy {
		t.Error("Expected DestroyContainer to be called")
	}
}

func TestServerConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		config    config.Config
		expectErr bool
	}{
		{
			name: "Valid config",
			config: config.Config{
				ListenAddr: ":8080",
				APIKeys:    []string{"test-key"},
			},
			expectErr: false,
		},
		{
			name: "Multiple API keys",
			config: config.Config{
				ListenAddr: ":9090",
				APIKeys:    []string{"key1", "key2", "key3"},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDocker := &MockDockerAdapterIntegration{}
			handlerAdapter := handler.NewAdapter(mockDocker)

			server := NewServer(handlerAdapter, tt.config)

			if server == nil && !tt.expectErr {
				t.Error("Expected server to be created")
			}

			if server != nil && tt.expectErr {
				t.Error("Expected server creation to fail")
			}
		})
	}
}

func TestLoggingHandlerIntegration(t *testing.T) {
	// Test that logging handler works with basic deployment
	mockDocker := &MockDockerAdapterIntegration{
		buildAndStartResult: "Container started",
		buildAndStartError:  nil,
	}

	handlerAdapter := handler.NewAdapter(mockDocker)

	// Create a simple logging handler test
	logHandler := &LoggingHandler{
		originalHandler: handlerAdapter,
		client:          nil, // No client for unit test
		server:          nil, // No server for unit test
	}

	result := logHandler.Deploy("invalid-url", "test")

	if result == nil {
		t.Fatal("Expected deployment result")
	}

	// Should fail due to invalid URL
	if result.Success {
		t.Error("Expected deployment to fail")
	}

	// Verify docker was called
	calls := mockDocker.GetCalls()
	if len(calls) == 0 {
		t.Error("Expected docker methods to be called")
	}
}

func TestServerCreation(t *testing.T) {
	// Test basic server creation without network operations
	mockDocker := &MockDockerAdapterIntegration{}
	handlerAdapter := handler.NewAdapter(mockDocker)

	cfg := config.Config{
		ListenAddr: ":8080",
		APIKeys:    []string{"test-key"},
	}

	server := NewServer(handlerAdapter, cfg)

	if server == nil {
		t.Fatal("Expected server to be created")
	}

	// Test Stop without Start (should not panic)
	err := server.Stop()
	if err != nil {
		t.Errorf("Stop should not error when server not started: %v", err)
	}
}

func TestMessageCreation(t *testing.T) {
	// Test message creation without network operations
	authMsg, err := NewMessage(MessageTypeAuth, AuthMessage{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("Failed to create auth message: %v", err)
	}

	if authMsg.Type != MessageTypeAuth {
		t.Errorf("Expected auth message type, got %s", authMsg.Type)
	}

	deployMsg, err := NewMessage(MessageTypeDeploy, DeployMessage{
		GitURL:      "https://github.com/test/repo.git",
		Environment: "production",
	})
	if err != nil {
		t.Fatalf("Failed to create deploy message: %v", err)
	}

	if deployMsg.Type != MessageTypeDeploy {
		t.Errorf("Expected deploy message type, got %s", deployMsg.Type)
	}
}

func TestConfigIntegration(t *testing.T) {
	// Test that config works with server creation
	cfg := config.DefaultConfig()

	mockDocker := &MockDockerAdapterIntegration{}
	handlerAdapter := handler.NewAdapter(mockDocker)

	server := NewServer(handlerAdapter, cfg)

	if server == nil {
		t.Fatal("Expected server to be created with default config")
	}
}
