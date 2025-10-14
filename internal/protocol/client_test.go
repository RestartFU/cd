package protocol

import (
	"encoding/json"
	"net"
	"strings"
	"testing"
	"time"
)

// MockConn implements net.Conn for testing
type MockConn struct {
	readData   []byte
	writeData  []byte
	readIndex  int
	writeIndex int
	closed     bool
	readError  error
	writeError error
	closeError error
}

func NewMockConn(readData string) *MockConn {
	return &MockConn{
		readData:  []byte(readData),
		writeData: make([]byte, 1024),
	}
}

func (m *MockConn) Read(b []byte) (n int, err error) {
	if m.readError != nil {
		return 0, m.readError
	}
	if m.readIndex >= len(m.readData) {
		return 0, nil
	}

	n = copy(b, m.readData[m.readIndex:])
	m.readIndex += n
	return n, nil
}

func (m *MockConn) Write(b []byte) (n int, err error) {
	if m.writeError != nil {
		return 0, m.writeError
	}
	if m.writeIndex+len(b) > len(m.writeData) {
		// Expand buffer if needed
		newData := make([]byte, m.writeIndex+len(b))
		copy(newData, m.writeData[:m.writeIndex])
		m.writeData = newData
	}

	n = copy(m.writeData[m.writeIndex:], b)
	m.writeIndex += n
	return n, nil
}

func (m *MockConn) Close() error {
	m.closed = true
	return m.closeError
}

func (m *MockConn) LocalAddr() net.Addr                { return nil }
func (m *MockConn) RemoteAddr() net.Addr               { return nil }
func (m *MockConn) SetDeadline(t time.Time) error      { return nil }
func (m *MockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *MockConn) SetWriteDeadline(t time.Time) error { return nil }

func (m *MockConn) GetWrittenData() string {
	return string(m.writeData[:m.writeIndex])
}

func TestNewClient(t *testing.T) {
	client := NewClient()

	if client == nil {
		t.Error("NewClient should return a non-nil client")
	}
}

func TestClient_Connect(t *testing.T) {
	client := NewClient()

	// Test connection to invalid address
	err := client.Connect("invalid-address:99999")
	if err == nil {
		t.Error("Connect should fail with invalid address")
	}
}

func TestClient_Close(t *testing.T) {
	client := NewClient()

	// Test closing without connection
	err := client.Close()
	if err != nil {
		t.Errorf("Close should not error when no connection exists, got %v", err)
	}

	// Test closing with mock connection
	mockConn := NewMockConn("")
	client.conn = mockConn

	err = client.Close()
	if err != nil {
		t.Errorf("Close should not error with valid connection, got %v", err)
	}

	if !mockConn.closed {
		t.Error("Connection should be closed")
	}
}

func TestClient_Authenticate(t *testing.T) {
	client := NewClient()
	mockConn := NewMockConn("")
	client.conn = mockConn
	client.encoder = json.NewEncoder(mockConn)

	err := client.Authenticate("test-api-key")
	if err != nil {
		t.Errorf("Authenticate should not error, got %v", err)
	}

	// Verify the written data
	writtenData := mockConn.GetWrittenData()
	if !strings.Contains(writtenData, "test-api-key") {
		t.Error("Written data should contain the API key")
	}

	if !strings.Contains(writtenData, string(MessageTypeAuth)) {
		t.Error("Written data should contain auth message type")
	}
}

func TestClient_Authenticate_NoConnection(t *testing.T) {
	client := NewClient()

	err := client.Authenticate("test-key")
	if err == nil {
		t.Error("Authenticate should fail when not connected")
	}
}

func TestClient_Deploy(t *testing.T) {
	client := NewClient()

	// Test without connection
	_, err := client.Deploy("https://github.com/test/repo.git", "production")
	if err == nil {
		t.Error("Deploy should fail when not connected")
	}

	// Test with mock connection
	mockConn := NewMockConn("")
	client.conn = mockConn
	client.encoder = json.NewEncoder(mockConn)

	updateChan, err := client.Deploy("https://github.com/test/repo.git", "production")
	if err != nil {
		t.Errorf("Deploy should not error with valid connection, got %v", err)
	}

	if updateChan == nil {
		t.Error("Deploy should return a non-nil update channel")
	}

	// Verify the written data
	writtenData := mockConn.GetWrittenData()
	if !strings.Contains(writtenData, "https://github.com/test/repo.git") {
		t.Error("Written data should contain the Git URL")
	}

	if !strings.Contains(writtenData, "production") {
		t.Error("Written data should contain the environment")
	}
}

func TestClient_WaitForAuth_Success(t *testing.T) {
	// Create auth success response
	statusMsg := StatusMessage{
		Stage:    "authenticated",
		Message:  "Authentication successful",
		Progress: 100,
	}

	msg, err := NewMessage(MessageTypeStatus, statusMsg)
	if err != nil {
		t.Fatalf("Failed to create test message: %v", err)
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal test message: %v", err)
	}

	client := NewClient()
	mockConn := NewMockConn(string(msgBytes) + "\n")
	client.conn = mockConn
	client.decoder = json.NewDecoder(mockConn)

	err = client.WaitForAuth()
	if err != nil {
		t.Errorf("WaitForAuth should succeed with valid auth response, got %v", err)
	}
}

func TestClient_WaitForAuth_Failure(t *testing.T) {
	// Create auth failure response
	errorMsg := ErrorMessage{
		Code:    "AUTH_FAILED",
		Message: "Invalid API key",
	}

	msg, err := NewMessage(MessageTypeError, errorMsg)
	if err != nil {
		t.Fatalf("Failed to create test message: %v", err)
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal test message: %v", err)
	}

	client := NewClient()
	mockConn := NewMockConn(string(msgBytes) + "\n")
	client.conn = mockConn
	client.decoder = json.NewDecoder(mockConn)

	err = client.WaitForAuth()
	if err == nil {
		t.Error("WaitForAuth should fail with error response")
	}

	if !strings.Contains(err.Error(), "Invalid API key") {
		t.Errorf("Error should contain the error message, got %v", err)
	}
}

func TestClient_ListenForUpdates(t *testing.T) {
	// Create test messages
	logMsg, _ := NewMessage(MessageTypeLog, LogMessage{
		Level:   "info",
		Message: "Test log message",
		Source:  "system",
	})

	statusMsg, _ := NewMessage(MessageTypeStatus, StatusMessage{
		Stage:    "building",
		Message:  "Building image",
		Progress: 50,
	})

	resultMsg, _ := NewMessage(MessageTypeResult, ResultMessage{
		Success:  true,
		Message:  "SUCCESS",
		Duration: "30s",
	})

	// Marshal messages
	logBytes, _ := json.Marshal(logMsg)
	statusBytes, _ := json.Marshal(statusMsg)
	resultBytes, _ := json.Marshal(resultMsg)

	// Combine messages with newlines
	testData := string(logBytes) + "\n" + string(statusBytes) + "\n" + string(resultBytes) + "\n"

	client := NewClient()
	mockConn := NewMockConn(testData)
	client.conn = mockConn

	updateChan := make(chan DeploymentUpdate, 10)
	go client.listenForUpdates(updateChan)

	// Collect updates
	var updates []DeploymentUpdate
	timeout := time.After(time.Second)

	for {
		select {
		case update, ok := <-updateChan:
			if !ok {
				// Channel closed, we're done
				goto done
			}
			updates = append(updates, update)

			// Stop after result message
			if update.Type == MessageTypeResult {
				goto done
			}

		case <-timeout:
			t.Fatal("Timeout waiting for updates")
		}
	}

done:
	if len(updates) < 3 {
		t.Errorf("Expected at least 3 updates, got %d", len(updates))
	}

	// Verify message types
	expectedTypes := []MessageType{MessageTypeLog, MessageTypeStatus, MessageTypeResult}
	for i, expectedType := range expectedTypes {
		if i >= len(updates) {
			t.Errorf("Missing update %d", i)
			continue
		}

		if updates[i].Type != expectedType {
			t.Errorf("Expected update %d to be type %s, got %s", i, expectedType, updates[i].Type)
		}
	}

	// Verify log message
	if updates[0].Log == nil {
		t.Error("First update should have log data")
	} else {
		if updates[0].Log.Message != "Test log message" {
			t.Errorf("Expected log message 'Test log message', got '%s'", updates[0].Log.Message)
		}
	}

	// Verify status message
	if updates[1].Status == nil {
		t.Error("Second update should have status data")
	} else {
		if updates[1].Status.Progress != 50 {
			t.Errorf("Expected progress 50, got %d", updates[1].Status.Progress)
		}
	}

	// Verify result message
	if updates[2].Result == nil {
		t.Error("Third update should have result data")
	} else {
		if !updates[2].Result.Success {
			t.Error("Expected result to be successful")
		}
	}
}

func TestClient_ListenForUpdates_InvalidJSON(t *testing.T) {
	client := NewClient()
	mockConn := NewMockConn("invalid json\n")
	client.conn = mockConn

	updateChan := make(chan DeploymentUpdate, 10)
	go client.listenForUpdates(updateChan)

	// Should receive an error update
	timeout := time.After(time.Second)
	select {
	case update := <-updateChan:
		if update.Type != MessageTypeError {
			t.Errorf("Expected error message type, got %s", update.Type)
		}

		if update.Error == nil {
			t.Error("Expected error data")
		} else {
			if update.Error.Code != "PARSE_ERROR" {
				t.Errorf("Expected error code 'PARSE_ERROR', got '%s'", update.Error.Code)
			}
		}

	case <-timeout:
		t.Fatal("Timeout waiting for error update")
	}
}

func TestDeploymentUpdate_Types(t *testing.T) {
	// Test DeploymentUpdate struct with different message types
	update := DeploymentUpdate{
		Type:      MessageTypeLog,
		Timestamp: time.Now(),
		Log: &LogMessage{
			Level:   "info",
			Message: "test",
			Source:  "system",
		},
	}

	if update.Type != MessageTypeLog {
		t.Errorf("Expected type %s, got %s", MessageTypeLog, update.Type)
	}

	if update.Log == nil {
		t.Error("Expected log data")
	}

	if update.Status != nil || update.Result != nil || update.Error != nil {
		t.Error("Other fields should be nil")
	}
}

func BenchmarkClient_Authenticate(b *testing.B) {
	client := NewClient()
	mockConn := NewMockConn("")
	client.conn = mockConn
	client.encoder = json.NewEncoder(mockConn)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Authenticate("test-key")
	}
}

func BenchmarkClient_Deploy(b *testing.B) {
	client := NewClient()
	mockConn := NewMockConn("")
	client.conn = mockConn
	client.encoder = json.NewEncoder(mockConn)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Deploy("https://github.com/test/repo.git", "production")
	}
}
