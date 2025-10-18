package protocol

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"
)

type Client struct {
	conn    net.Conn
	encoder *json.Encoder
	decoder *json.Decoder
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Connect(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	c.conn = conn
	c.encoder = json.NewEncoder(conn)
	c.decoder = json.NewDecoder(conn)

	return nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) Authenticate(apiKey string) error {
	authMsg := AuthMessage{
		APIKey: apiKey,
	}

	msg, err := NewMessage(MessageTypeAuth, authMsg)
	if err != nil {
		return fmt.Errorf("failed to create auth message: %w", err)
	}

	return c.sendMessage(msg)
}

func (c *Client) Deploy(gitURL, environment string, envVars, secrets map[string]string) (<-chan DeploymentUpdate, error) {
	deployMsg := DeployMessage{
		GitURL:      gitURL,
		Environment: environment,
		EnvVars:     envVars,
		Secrets:     secrets,
	}

	msg, err := NewMessage(MessageTypeDeploy, deployMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to create deploy message: %w", err)
	}

	if err := c.sendMessage(msg); err != nil {
		return nil, fmt.Errorf("failed to send deploy message: %w", err)
	}

	updateChan := make(chan DeploymentUpdate, 100)

	go c.listenForUpdates(updateChan)

	return updateChan, nil
}

type DeploymentUpdate struct {
	Type      MessageType
	Log       *LogMessage
	Status    *StatusMessage
	Result    *ResultMessage
	Error     *ErrorMessage
	Timestamp time.Time
}

func (c *Client) sendMessage(msg *Message) error {
	if c.encoder == nil {
		return fmt.Errorf("not connected to server")
	}
	return c.encoder.Encode(msg)
}

func (c *Client) listenForUpdates(updateChan chan<- DeploymentUpdate) {
	defer close(updateChan)

	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var msg Message
		if err := json.Unmarshal([]byte(line), &msg); err != nil {

			updateChan <- DeploymentUpdate{
				Type: MessageTypeError,
				Error: &ErrorMessage{
					Code:    "PARSE_ERROR",
					Message: fmt.Sprintf("Failed to parse message: %v", err),
				},
				Timestamp: time.Now(),
			}
			continue
		}

		update := DeploymentUpdate{
			Type:      msg.Type,
			Timestamp: msg.Timestamp,
		}

		switch msg.Type {
		case MessageTypeLog:
			var logMsg LogMessage
			if err := msg.Unmarshal(&logMsg); err == nil {
				update.Log = &logMsg
			}

		case MessageTypeStatus:
			var statusMsg StatusMessage
			if err := msg.Unmarshal(&statusMsg); err == nil {
				update.Status = &statusMsg
			}

		case MessageTypeResult:
			var resultMsg ResultMessage
			if err := msg.Unmarshal(&resultMsg); err == nil {
				update.Result = &resultMsg
			}

			updateChan <- update
			return

		case MessageTypeError:
			var errorMsg ErrorMessage
			if err := msg.Unmarshal(&errorMsg); err == nil {
				update.Error = &errorMsg
			}

			updateChan <- update
			return
		}

		updateChan <- update
	}

	if err := scanner.Err(); err != nil {
		updateChan <- DeploymentUpdate{
			Type: MessageTypeError,
			Error: &ErrorMessage{
				Code:    "CONNECTION_ERROR",
				Message: fmt.Sprintf("Connection error: %v", err),
			},
			Timestamp: time.Now(),
		}
	}
}

func (c *Client) WaitForAuth() error {
	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var msg Message
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			return fmt.Errorf("failed to parse auth response: %w", err)
		}

		switch msg.Type {
		case MessageTypeStatus:
			var statusMsg StatusMessage
			if err := msg.Unmarshal(&statusMsg); err != nil {
				return fmt.Errorf("failed to parse status message: %w", err)
			}
			if statusMsg.Stage == "authenticated" {
				return nil
			}

		case MessageTypeError:
			var errorMsg ErrorMessage
			if err := msg.Unmarshal(&errorMsg); err != nil {
				return fmt.Errorf("failed to parse error message: %w", err)
			}
			return fmt.Errorf("authentication failed: %s", errorMsg.Message)
		}
	}

	return fmt.Errorf("connection closed while waiting for auth response")
}
