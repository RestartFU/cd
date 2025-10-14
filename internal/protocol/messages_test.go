package protocol

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewMessage(t *testing.T) {
	// Test creating a message with valid data
	authData := AuthMessage{APIKey: "test-key"}
	msg, err := NewMessage(MessageTypeAuth, authData)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if msg.Type != MessageTypeAuth {
		t.Errorf("Expected type %s, got %s", MessageTypeAuth, msg.Type)
	}

	if msg.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}

	// Verify data can be parsed back
	var parsedAuth AuthMessage
	err = msg.ParseData(&parsedAuth)
	if err != nil {
		t.Fatalf("Failed to parse data: %v", err)
	}

	if parsedAuth.APIKey != "test-key" {
		t.Errorf("Expected API key 'test-key', got '%s'", parsedAuth.APIKey)
	}
}

func TestMessage_ParseData(t *testing.T) {
	tests := []struct {
		name        string
		messageType MessageType
		data        interface{}
		expectError bool
	}{
		{
			name:        "AuthMessage",
			messageType: MessageTypeAuth,
			data:        AuthMessage{APIKey: "secret"},
			expectError: false,
		},
		{
			name:        "DeployMessage",
			messageType: MessageTypeDeploy,
			data:        DeployMessage{GitURL: "https://github.com/test/repo.git", Environment: "prod"},
			expectError: false,
		},
		{
			name:        "LogMessage",
			messageType: MessageTypeLog,
			data:        LogMessage{Level: "info", Message: "test log", Source: "system"},
			expectError: false,
		},
		{
			name:        "StatusMessage",
			messageType: MessageTypeStatus,
			data:        StatusMessage{Stage: "building", Message: "Building image", Progress: 50},
			expectError: false,
		},
		{
			name:        "ResultMessage",
			messageType: MessageTypeResult,
			data:        ResultMessage{Success: true, Message: "SUCCESS", Duration: "30s"},
			expectError: false,
		},
		{
			name:        "ErrorMessage",
			messageType: MessageTypeError,
			data:        ErrorMessage{Code: "TEST_ERROR", Message: "Test error message"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create message
			msg, err := NewMessage(tt.messageType, tt.data)
			if err != nil {
				t.Fatalf("Failed to create message: %v", err)
			}

			// Parse data back based on type
			switch tt.messageType {
			case MessageTypeAuth:
				var parsed AuthMessage
				err = msg.ParseData(&parsed)
				if err != nil && !tt.expectError {
					t.Errorf("Unexpected error parsing auth message: %v", err)
				}
				if !tt.expectError {
					original := tt.data.(AuthMessage)
					if parsed.APIKey != original.APIKey {
						t.Errorf("Expected API key '%s', got '%s'", original.APIKey, parsed.APIKey)
					}
				}

			case MessageTypeDeploy:
				var parsed DeployMessage
				err = msg.ParseData(&parsed)
				if err != nil && !tt.expectError {
					t.Errorf("Unexpected error parsing deploy message: %v", err)
				}
				if !tt.expectError {
					original := tt.data.(DeployMessage)
					if parsed.GitURL != original.GitURL {
						t.Errorf("Expected Git URL '%s', got '%s'", original.GitURL, parsed.GitURL)
					}
					if parsed.Environment != original.Environment {
						t.Errorf("Expected environment '%s', got '%s'", original.Environment, parsed.Environment)
					}
				}

			case MessageTypeLog:
				var parsed LogMessage
				err = msg.ParseData(&parsed)
				if err != nil && !tt.expectError {
					t.Errorf("Unexpected error parsing log message: %v", err)
				}
				if !tt.expectError {
					original := tt.data.(LogMessage)
					if parsed.Level != original.Level {
						t.Errorf("Expected level '%s', got '%s'", original.Level, parsed.Level)
					}
					if parsed.Message != original.Message {
						t.Errorf("Expected message '%s', got '%s'", original.Message, parsed.Message)
					}
					if parsed.Source != original.Source {
						t.Errorf("Expected source '%s', got '%s'", original.Source, parsed.Source)
					}
				}

			case MessageTypeStatus:
				var parsed StatusMessage
				err = msg.ParseData(&parsed)
				if err != nil && !tt.expectError {
					t.Errorf("Unexpected error parsing status message: %v", err)
				}
				if !tt.expectError {
					original := tt.data.(StatusMessage)
					if parsed.Stage != original.Stage {
						t.Errorf("Expected stage '%s', got '%s'", original.Stage, parsed.Stage)
					}
					if parsed.Progress != original.Progress {
						t.Errorf("Expected progress %d, got %d", original.Progress, parsed.Progress)
					}
				}

			case MessageTypeResult:
				var parsed ResultMessage
				err = msg.ParseData(&parsed)
				if err != nil && !tt.expectError {
					t.Errorf("Unexpected error parsing result message: %v", err)
				}
				if !tt.expectError {
					original := tt.data.(ResultMessage)
					if parsed.Success != original.Success {
						t.Errorf("Expected success %v, got %v", original.Success, parsed.Success)
					}
					if parsed.Message != original.Message {
						t.Errorf("Expected message '%s', got '%s'", original.Message, parsed.Message)
					}
				}

			case MessageTypeError:
				var parsed ErrorMessage
				err = msg.ParseData(&parsed)
				if err != nil && !tt.expectError {
					t.Errorf("Unexpected error parsing error message: %v", err)
				}
				if !tt.expectError {
					original := tt.data.(ErrorMessage)
					if parsed.Code != original.Code {
						t.Errorf("Expected code '%s', got '%s'", original.Code, parsed.Code)
					}
					if parsed.Message != original.Message {
						t.Errorf("Expected message '%s', got '%s'", original.Message, parsed.Message)
					}
				}
			}
		})
	}
}

func TestMessageSerialization(t *testing.T) {
	// Test that messages can be serialized and deserialized properly
	authData := AuthMessage{APIKey: "test-api-key"}
	msg, err := NewMessage(MessageTypeAuth, authData)
	if err != nil {
		t.Fatalf("Failed to create message: %v", err)
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	// Deserialize from JSON
	var deserializedMsg Message
	err = json.Unmarshal(jsonData, &deserializedMsg)
	if err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	// Verify message type
	if deserializedMsg.Type != MessageTypeAuth {
		t.Errorf("Expected type %s, got %s", MessageTypeAuth, deserializedMsg.Type)
	}

	// Verify data
	var parsedAuth AuthMessage
	err = deserializedMsg.ParseData(&parsedAuth)
	if err != nil {
		t.Fatalf("Failed to parse deserialized data: %v", err)
	}

	if parsedAuth.APIKey != "test-api-key" {
		t.Errorf("Expected API key 'test-api-key', got '%s'", parsedAuth.APIKey)
	}
}

func TestMessageTypes(t *testing.T) {
	// Test that all message type constants are defined correctly
	expectedTypes := []MessageType{
		MessageTypeDeploy,
		MessageTypeAuth,
		MessageTypeLog,
		MessageTypeStatus,
		MessageTypeResult,
		MessageTypeError,
	}

	for _, msgType := range expectedTypes {
		if string(msgType) == "" {
			t.Errorf("Message type %v should not be empty", msgType)
		}
	}
}

func TestInvalidMessageData(t *testing.T) {
	// Test parsing invalid JSON data
	msg := &Message{
		Type:      MessageTypeAuth,
		Timestamp: time.Now(),
		Data:      json.RawMessage(`{"invalid": json}`),
	}

	var authMsg AuthMessage
	err := msg.ParseData(&authMsg)
	if err == nil {
		t.Error("Expected error when parsing invalid JSON, got nil")
	}
}

func TestEmptyMessageData(t *testing.T) {
	// Test creating message with nil data
	_, err := NewMessage(MessageTypeAuth, nil)
	if err != nil {
		t.Errorf("Expected no error with nil data, got %v", err)
	}
}
