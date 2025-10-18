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

	"github.com/restartfu/cd/internal/config"
	"github.com/restartfu/cd/internal/ports"
)

type Server struct {
	listener      net.Listener
	handler       ports.Handler
	config        config.Config
	connections   map[string]*ClientConnection
	connectionsMu sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
}

type ClientConnection struct {
	conn          net.Conn
	encoder       *json.Encoder
	authenticated bool
	mu            sync.Mutex
}

func NewServer(handler ports.Handler, cfg config.Config) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		handler:     handler,
		config:      cfg,
		connections: make(map[string]*ClientConnection),
		ctx:         ctx,
		cancel:      cancel,
	}
}

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

func (s *Server) Stop() error {
	log.Println("Initiating server shutdown...")

	s.cancel()

	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			log.Printf("Error closing listener: %v", err)
		}
	}

	s.connectionsMu.Lock()
	for addr, client := range s.connections {
		log.Printf("Closing connection: %s", addr)
		client.conn.Close()
	}
	s.connectionsMu.Unlock()

	log.Println("Waiting for connections to close...")
	s.wg.Wait()

	log.Println("Server shutdown complete")
	return nil
}

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

func (s *Server) handleAuth(client *ClientConnection, msg *Message) error {
	var authMsg AuthMessage
	if err := msg.Unmarshal(&authMsg); err != nil {
		return s.sendError(client, "INVALID_AUTH", "Invalid authentication data")
	}

	if authMsg.APIKey == s.config.APIKey {
		client.mu.Lock()
		client.authenticated = true
		client.mu.Unlock()

		return s.sendStatus(client, "authenticated", "Authentication successful", 100)
	}

	return s.sendError(client, "AUTH_FAILED", "Invalid API key")
}

func (s *Server) handleDeploy(client *ClientConnection, msg *Message) error {
	client.mu.Lock()
	authenticated := client.authenticated
	client.mu.Unlock()

	if !authenticated {
		return s.sendError(client, "NOT_AUTHENTICATED", "Authentication required")
	}

	var deployMsg DeployMessage
	if err := msg.Unmarshal(&deployMsg); err != nil {
		return s.sendError(client, "INVALID_DEPLOY", "Invalid deployment data")
	}

	go s.performDeployment(client, deployMsg)
	return nil
}

func (s *Server) performDeployment(client *ClientConnection, deployMsg DeployMessage) {
	startTime := time.Now()

	s.sendStatus(client, "starting", "Deployment started", 0)
	s.sendLog(client, "info", "Starting deployment process", "system")

	logHandler := &LoggingHandler{
		handler: s.handler,
		client:  client,
		server:  s,
	}

	response := logHandler.Deploy(deployMsg.GitURL, deployMsg.Environment, deployMsg.EnvVars, deployMsg.Secrets)
	duration := time.Since(startTime)

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

func (s *Server) sendMessage(client *ClientConnection, msg *Message) error {
	client.mu.Lock()
	defer client.mu.Unlock()

	return client.encoder.Encode(msg)
}

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

func (s *Server) sendResult(client *ClientConnection, result ResultMessage) error {
	msg, err := NewMessage(MessageTypeResult, result)
	if err != nil {
		return err
	}

	return s.sendMessage(client, msg)
}

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
