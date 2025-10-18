package protocol

import (
	"encoding/json"
	"time"
)

type MessageType string

const (
	MessageTypeDeploy MessageType = "deploy"
	MessageTypeAuth   MessageType = "auth"

	MessageTypeLog    MessageType = "log"
	MessageTypeStatus MessageType = "status"
	MessageTypeResult MessageType = "result"
	MessageTypeError  MessageType = "error"
)

type Message struct {
	Type      MessageType     `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

type AuthMessage struct {
	APIKey string `json:"api_key"`
}

type DeployMessage struct {
	GitURL      string            `json:"git_url"`
	Environment string            `json:"environment"`
	EnvVars     map[string]string `json:"env_vars,omitempty"`
	Secrets     map[string]string `json:"secrets,omitempty"`
}

type LogMessage struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Source  string `json:"source"`
}

type StatusMessage struct {
	Stage    string `json:"stage"`
	Message  string `json:"message"`
	Progress int    `json:"progress"`
}

type ResultMessage struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	ContainerID string `json:"container_id,omitempty"`
	Duration    string `json:"duration"`
}

type ErrorMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewMessage(msgType MessageType, data any) (*Message, error) {
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

func (m *Message) Unmarshal(v any) error {
	return json.Unmarshal(m.Data, v)
}
