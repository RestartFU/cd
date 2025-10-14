package protocol

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/restartfu/cd/internal/adapters/handler"
	"github.com/restartfu/cd/internal/config"
)

// Server handles TCP connections and manages deployment requests
type Server struct {
	listener      net.Listener
	handler       *handler.Adapter
	config        config.Config
	connections   map[string]*ClientConnection
	connectionsMu sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

// ClientConnection represents an active client connection
type ClientConnection struct {
	conn          net.Conn
	encoder       *json.Encoder
	authenticated bool
	mu            sync.Mutex
}

// NewServer creates a new TCP server
func NewServer(handler *handler.Adapter, cfg config.Config) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		handler:     handler,
		config:      cfg,
		connections: make(map[string]*ClientConnection),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start begins listening for TCP connections
func (s *Server) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener

	log.Printf("TCP server listening on %s", addr)

	for {
		select {
		case <-s.ctx.Done():
			log.Println("Server context cancelled, stopping accept loop")
			return nil
		default:
		}

		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				// Server is shutting down, this is expected
				return nil
			default:
				log.Printf("Failed to accept connection: %v", err)
				continue
			}
		}

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

// Stop gracefully shuts down the server
func (s *Server) Stop() error {
	log.Println("Initiating server shutdown...")

	// Cancel context to signal shutdown
	s.cancel()

	// Close listener to stop accepting new connections
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			log.Printf("Error closing listener: %v", err)
		}
	}

	// Close all existing connections
	s.connectionsMu.Lock()
	for addr, client := range s.connections {
		log.Printf("Closing connection: %s", addr)
		client.conn.Close()
	}
	s.connectionsMu.Unlock()

	// Wait for all goroutines to finish
	log.Println("Waiting for connections to close...")
	s.wg.Wait()

	log.Println("Server shutdown complete")
	return nil
}

// handleConnection manages a single client connection
func (s *Server) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	log.Printf("New connection from %s", clientAddr)

	client := &ClientConnection{
		conn:    conn,
		encoder: json.NewEncoder(conn),
	}

	s.connectionsMu.Lock()
	s.connections[clientAddr] = client
	s.connectionsMu.Unlock()

	defer func() {
		s.connectionsMu.Lock()
		delete(s.connections, clientAddr)
		s.connectionsMu.Unlock()
		log.Printf("Connection closed: %s", clientAddr)
	}()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		select {
		case <-s.ctx.Done():
			log.Printf("Server shutting down, closing connection: %s", clientAddr)
			return
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var msg Message
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			s.sendError(client, "INVALID_JSON", "Invalid JSON message")
			continue
		}

		if err := s.processMessage(client, &msg); err != nil {
			log.Printf("Error processing message from %s: %v", clientAddr, err)
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		log.Printf("Error reading from connection %s: %v", clientAddr, err)
	}
}

// processMessage handles different types of messages from clients
func (s *Server) processMessage(client *ClientConnection, msg *Message) error {
	switch msg.Type {
	case MessageTypeAuth:
		return s.handleAuth(client, msg)
	case MessageTypeDeploy:
		return s.handleDeploy(client, msg)
	default:
		return s.sendError(client, "UNKNOWN_MESSAGE_TYPE", "Unknown message type")
	}
}

// handleAuth processes authentication messages
func (s *Server) handleAuth(client *ClientConnection, msg *Message) error {
	var authMsg AuthMessage
	if err := msg.ParseData(&authMsg); err != nil {
		return s.sendError(client, "INVALID_AUTH", "Invalid authentication data")
	}

	// Check API key
	for _, key := range s.config.APIKeys {
		if key == authMsg.APIKey {
			client.mu.Lock()
			client.authenticated = true
			client.mu.Unlock()

			return s.sendStatus(client, "authenticated", "Authentication successful", 100)
		}
	}

	return s.sendError(client, "AUTH_FAILED", "Invalid API key")
}

// handleDeploy processes deployment requests
func (s *Server) handleDeploy(client *ClientConnection, msg *Message) error {
	client.mu.Lock()
	authenticated := client.authenticated
	client.mu.Unlock()

	if !authenticated {
		return s.sendError(client, "NOT_AUTHENTICATED", "Authentication required")
	}

	var deployMsg DeployMessage
	if err := msg.ParseData(&deployMsg); err != nil {
		return s.sendError(client, "INVALID_DEPLOY", "Invalid deployment data")
	}

	// Start deployment in a goroutine to allow real-time logging
	go s.performDeployment(client, deployMsg)
	return nil
}

// performDeployment executes the deployment with real-time logging
func (s *Server) performDeployment(client *ClientConnection, deployMsg DeployMessage) {
	startTime := time.Now()

	// Send initial status
	s.sendStatus(client, "starting", "Deployment started", 0)
	s.sendLog(client, "info", "Starting deployment process", "system")

	// Create a custom handler that captures logs
	logHandler := &LoggingHandler{
		originalHandler: s.handler,
		client:          client,
		server:          s,
	}

	// Execute deployment
	response := logHandler.DeployWithEnv(deployMsg.GitURL, deployMsg.Environment, deployMsg.EnvVars, deployMsg.Secrets)
	duration := time.Since(startTime)

	// Send final result
	result := ResultMessage{
		Success:  response.Success,
		Message:  response.Message,
		Duration: duration.String(),
	}

	if response.Success {
		s.sendResult(client, result)
	} else {
		s.sendError(client, "DEPLOY_FAILED", response.Message)
	}
}

// sendMessage sends a message to a client
func (s *Server) sendMessage(client *ClientConnection, msg *Message) error {
	client.mu.Lock()
	defer client.mu.Unlock()

	return client.encoder.Encode(msg)
}

// sendLog sends a log message to a client
func (s *Server) sendLog(client *ClientConnection, level, message, source string) error {
	logMsg := LogMessage{
		Level:   level,
		Message: message,
		Source:  source,
	}

	msg, err := NewMessage(MessageTypeLog, logMsg)
	if err != nil {
		return err
	}

	return s.sendMessage(client, msg)
}

// sendStatus sends a status update to a client
func (s *Server) sendStatus(client *ClientConnection, stage, message string, progress int) error {
	statusMsg := StatusMessage{
		Stage:    stage,
		Message:  message,
		Progress: progress,
	}

	msg, err := NewMessage(MessageTypeStatus, statusMsg)
	if err != nil {
		return err
	}

	return s.sendMessage(client, msg)
}

// sendResult sends the final deployment result to a client
func (s *Server) sendResult(client *ClientConnection, result ResultMessage) error {
	msg, err := NewMessage(MessageTypeResult, result)
	if err != nil {
		return err
	}

	return s.sendMessage(client, msg)
}

// sendError sends an error message to a client
func (s *Server) sendError(client *ClientConnection, code, message string) error {
	errorMsg := ErrorMessage{
		Code:    code,
		Message: message,
	}

	msg, err := NewMessage(MessageTypeError, errorMsg)
	if err != nil {
		return err
	}

	return s.sendMessage(client, msg)
}
