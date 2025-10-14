package protocol

import (
	"encoding/json"
	"time"
)

// MessageType represents the type of message being sent
type MessageType string

const (
	// Client to Server messages
	MessageTypeDeploy MessageType = "deploy"
	MessageTypeAuth   MessageType = "auth"

	// Server to Client messages
	MessageTypeLog    MessageType = "log"
	MessageTypeStatus MessageType = "status"
	MessageTypeResult MessageType = "result"
	MessageTypeError  MessageType = "error"
)

// Message is the base structure for all TCP communications
type Message struct {
	Type      MessageType     `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// AuthMessage is sent by client to authenticate
type AuthMessage struct {
	APIKey string `json:"api_key"`
}

// DeployMessage is sent by client to request deployment
type DeployMessage struct {
	GitURL      string            `json:"git_url"`
	Environment string            `json:"environment"`
	EnvVars     map[string]string `json:"env_vars,omitempty"`
	Secrets     map[string]string `json:"secrets,omitempty"`
}

// LogMessage is sent by server to stream logs in real-time
type LogMessage struct {
	Level   string `json:"level"` // info, warn, error
	Message string `json:"message"`
	Source  string `json:"source"` // git, docker, system
}

// StatusMessage is sent by server to indicate deployment progress
type StatusMessage struct {
	Stage    string `json:"stage"` // cloning, building, starting, complete
	Message  string `json:"message"`
	Progress int    `json:"progress"` // 0-100
}

// ResultMessage is sent by server at the end of deployment
type ResultMessage struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	ContainerID string `json:"container_id,omitempty"`
	Duration    string `json:"duration"`
}

// ErrorMessage is sent by server when an error occurs
type ErrorMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewMessage creates a new message with the current timestamp
func NewMessage(msgType MessageType, data interface{}) (*Message, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &Message{
		Type:      msgType,
		Timestamp: time.Now(),
		Data:      dataBytes,
	}, nil
}

// ParseData unmarshals the message data into the provided interface
func (m *Message) ParseData(v interface{}) error {
	return json.Unmarshal(m.Data, v)
}
