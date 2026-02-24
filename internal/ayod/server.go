package ayod

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

// Server is the ayod RPC server.
type Server struct {
	users    *UserManager
	executor *Executor
	files    *FileHandler
	start    time.Time

	mu       sync.Mutex
	listener net.Listener
	done     chan struct{}
	wg       sync.WaitGroup
}

// NewServer creates a new ayod server.
func NewServer() *Server {
	users := NewUserManager()
	return &Server{
		users:    users,
		executor: NewExecutor(users),
		files:    NewFileHandler(),
		start:    time.Now(),
		done:     make(chan struct{}),
	}
}

// Serve starts accepting connections on the listener.
func (s *Server) Serve(listener net.Listener) error {
	s.mu.Lock()
	s.listener = listener
	s.mu.Unlock()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.done:
				return nil
			default:
				log.Printf("accept error: %v", err)
				continue
			}
		}

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handleConn(conn)
		}()
	}
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown() {
	close(s.done)
	s.mu.Lock()
	if s.listener != nil {
		s.listener.Close()
	}
	s.mu.Unlock()
	s.wg.Wait()
}

// handleConn handles a single client connection.
func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	encoder := json.NewEncoder(conn)

	for {
		// Read line-delimited JSON
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("read error: %v", err)
			}
			return
		}

		var req RPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.sendError(encoder, fmt.Sprintf("invalid request: %v", err))
			continue
		}

		resp := s.handleRequest(req)
		if err := encoder.Encode(resp); err != nil {
			log.Printf("encode error: %v", err)
			return
		}
	}
}

// handleRequest dispatches an RPC request to the appropriate handler.
func (s *Server) handleRequest(req RPCRequest) RPCResponse {
	switch req.Method {
	case "UserAdd":
		return s.handleUserAdd(req.Params)
	case "Exec":
		return s.handleExec(req.Params)
	case "ReadFile":
		return s.handleReadFile(req.Params)
	case "WriteFile":
		return s.handleWriteFile(req.Params)
	case "Health":
		return s.handleHealth()
	default:
		return RPCResponse{Error: fmt.Sprintf("unknown method: %s", req.Method)}
	}
}

func (s *Server) handleUserAdd(params []byte) RPCResponse {
	var req UserAddRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return RPCResponse{Error: fmt.Sprintf("invalid params: %v", err)}
	}

	resp, err := s.users.AddUser(req)
	if err != nil {
		return RPCResponse{Error: err.Error()}
	}

	result, _ := json.Marshal(resp)
	return RPCResponse{Result: result}
}

func (s *Server) handleExec(params []byte) RPCResponse {
	var req ExecRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return RPCResponse{Error: fmt.Sprintf("invalid params: %v", err)}
	}

	resp, err := s.executor.Exec(req)
	if err != nil {
		return RPCResponse{Error: err.Error()}
	}

	result, _ := json.Marshal(resp)
	return RPCResponse{Result: result}
}

func (s *Server) handleReadFile(params []byte) RPCResponse {
	var req ReadFileRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return RPCResponse{Error: fmt.Sprintf("invalid params: %v", err)}
	}

	resp, err := s.files.ReadFile(req)
	if err != nil {
		return RPCResponse{Error: err.Error()}
	}

	result, _ := json.Marshal(resp)
	return RPCResponse{Result: result}
}

func (s *Server) handleWriteFile(params []byte) RPCResponse {
	var req WriteFileRequest
	if err := json.Unmarshal(params, &req); err != nil {
		return RPCResponse{Error: fmt.Sprintf("invalid params: %v", err)}
	}

	if err := s.files.WriteFile(req); err != nil {
		return RPCResponse{Error: err.Error()}
	}

	return RPCResponse{Result: []byte(`{}`)}
}

func (s *Server) handleHealth() RPCResponse {
	resp := HealthResponse{
		Status: "ok",
		Users:  s.users.ListUsers(),
		Uptime: time.Since(s.start).Seconds(),
	}

	result, _ := json.Marshal(resp)
	return RPCResponse{Result: result}
}

func (s *Server) sendError(encoder *json.Encoder, msg string) {
	encoder.Encode(RPCResponse{Error: msg})
}

// Run is the main entry point for ayod when run as PID 1.
// It sets up the socket, starts the server, and handles signals.
func Run() error {
	// Create socket directory
	if err := os.MkdirAll("/run", 0755); err != nil {
		return fmt.Errorf("create /run: %w", err)
	}

	// Remove stale socket
	os.Remove(SocketPath)

	// Create socket
	listener, err := net.Listen("unix", SocketPath)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer listener.Close()

	// Make socket world-accessible
	if err := os.Chmod(SocketPath, 0666); err != nil {
		return fmt.Errorf("chmod socket: %w", err)
	}

	log.Printf("ayod listening on %s", SocketPath)

	// Start server
	server := NewServer()
	return server.Serve(listener)
}
