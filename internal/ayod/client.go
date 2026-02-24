package ayod

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// Client is a client for connecting to ayod.
type Client struct {
	conn   net.Conn
	reader *bufio.Reader
}

// Connect establishes a connection to the ayod socket.
func Connect(socketPath string) (*Client, error) {
	conn, err := net.DialTimeout("unix", socketPath, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("connect to ayod: %w", err)
	}

	return &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}, nil
}

// Close closes the connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// call sends an RPC request and returns the response.
func (c *Client) call(method string, params any) ([]byte, error) {
	// Marshal params
	paramsData, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("marshal params: %w", err)
	}

	// Build request
	req := RPCRequest{
		Method: method,
		Params: paramsData,
	}

	// Send request (line-delimited JSON)
	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	reqData = append(reqData, '\n')

	if _, err := c.conn.Write(reqData); err != nil {
		return nil, fmt.Errorf("write request: %w", err)
	}

	// Read response
	respData, err := c.reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var resp RPCResponse
	if err := json.Unmarshal(respData, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("rpc error: %s", resp.Error)
	}

	return resp.Result, nil
}

// UserAdd creates a new user in the sandbox.
func (c *Client) UserAdd(req UserAddRequest) (*UserAddResponse, error) {
	result, err := c.call("UserAdd", req)
	if err != nil {
		return nil, err
	}

	var resp UserAddResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &resp, nil
}

// Exec executes a command as the specified user.
func (c *Client) Exec(req ExecRequest) (*ExecResponse, error) {
	result, err := c.call("Exec", req)
	if err != nil {
		return nil, err
	}

	var resp ExecResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &resp, nil
}

// ReadFile reads a file from the sandbox.
func (c *Client) ReadFile(path string) (*ReadFileResponse, error) {
	result, err := c.call("ReadFile", ReadFileRequest{Path: path})
	if err != nil {
		return nil, err
	}

	var resp ReadFileResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &resp, nil
}

// WriteFile writes a file to the sandbox.
func (c *Client) WriteFile(req WriteFileRequest) error {
	_, err := c.call("WriteFile", req)
	return err
}

// Health checks if ayod is healthy.
func (c *Client) Health() (*HealthResponse, error) {
	result, err := c.call("Health", struct{}{})
	if err != nil {
		return nil, err
	}

	var resp HealthResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return &resp, nil
}
