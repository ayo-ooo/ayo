package jsonl

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/alexcabrera/ayo/internal/session"
)

// Reader reads session data from a JSONL file.
type Reader struct {
	file    *os.File
	scanner *bufio.Scanner
	header  *SessionHeader
	path    string
}

// NewReader opens a session file for reading.
func NewReader(path string) (*Reader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	scanner := bufio.NewScanner(file)
	// Increase buffer size for large lines
	const maxLineSize = 10 * 1024 * 1024 // 10MB
	scanner.Buffer(make([]byte, 64*1024), maxLineSize)

	// Read header
	if !scanner.Scan() {
		file.Close()
		if scanner.Err() != nil {
			return nil, fmt.Errorf("read header: %w", scanner.Err())
		}
		return nil, ErrEmptyFile
	}

	header, err := ParseSessionHeader(scanner.Bytes())
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("parse header: %w", err)
	}

	return &Reader{
		file:    file,
		scanner: scanner,
		header:  header,
		path:    path,
	}, nil
}

// Header returns the session header.
func (r *Reader) Header() *SessionHeader {
	return r.header
}

// Session returns the session metadata from the header.
func (r *Reader) Session() session.Session {
	return headerToSession(r.header)
}

// NextMessage reads the next message from the file.
// Returns nil, nil when no more messages are available.
func (r *Reader) NextMessage() (*session.Message, error) {
	if !r.scanner.Scan() {
		if err := r.scanner.Err(); err != nil {
			return nil, fmt.Errorf("read message: %w", err)
		}
		return nil, nil // EOF
	}

	line, err := ParseMessageLine(r.scanner.Bytes())
	if err != nil {
		return nil, fmt.Errorf("parse message: %w", err)
	}

	msg := lineToMessage(line, r.header.ID)
	return &msg, nil
}

// ReadAllMessages reads all remaining messages from the file.
func (r *Reader) ReadAllMessages() ([]session.Message, error) {
	var messages []session.Message
	for {
		msg, err := r.NextMessage()
		if err != nil {
			return nil, err
		}
		if msg == nil {
			break
		}
		messages = append(messages, *msg)
	}
	return messages, nil
}

// Close closes the reader.
func (r *Reader) Close() error {
	return r.file.Close()
}

// ReadSession reads a complete session from a file.
func ReadSession(path string) (session.Session, []session.Message, error) {
	reader, err := NewReader(path)
	if err != nil {
		return session.Session{}, nil, err
	}
	defer reader.Close()

	sess := reader.Session()
	messages, err := reader.ReadAllMessages()
	if err != nil {
		return session.Session{}, nil, err
	}

	return sess, messages, nil
}

// ReadSessionHeader reads only the header from a file (for indexing).
func ReadSessionHeader(path string) (*SessionHeader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		if scanner.Err() != nil {
			return nil, fmt.Errorf("read header: %w", scanner.Err())
		}
		return nil, ErrEmptyFile
	}

	return ParseSessionHeader(scanner.Bytes())
}

func headerToSession(h *SessionHeader) session.Session {
	sess := session.Session{
		ID:           h.ID,
		AgentHandle:  h.AgentHandle,
		Title:        h.Title,
		Source:       h.Source,
		ChainDepth:   int64(h.ChainDepth),
		ChainSource:  h.ChainSource,
		MessageCount: int64(h.MessageCount),
		CreatedAt:    h.CreatedAt.Unix(),
		UpdatedAt:    h.UpdatedAt.Unix(),
	}

	if h.FinishedAt != nil {
		sess.FinishedAt = h.FinishedAt.Unix()
	}
	if h.InputSchema != nil {
		sess.InputSchema = *h.InputSchema
	}
	if h.OutputSchema != nil {
		sess.OutputSchema = *h.OutputSchema
	}
	if h.StructuredInput != nil {
		sess.StructuredInput = *h.StructuredInput
	}
	if h.StructuredOutput != nil {
		sess.StructuredOutput = *h.StructuredOutput
	}

	return sess
}

func lineToMessage(line *MessageLine, sessionID string) session.Message {
	msg := session.Message{
		ID:        line.ID,
		SessionID: sessionID,
		Role:      session.MessageRole(line.Role),
		Model:     line.Model,
		Provider:  line.Provider,
		CreatedAt: line.CreatedAt.Unix(),
		UpdatedAt: line.UpdatedAt.Unix(),
	}

	if line.FinishedAt != nil {
		msg.FinishedAt = line.FinishedAt.Unix()
	}

	// Parse parts
	msg.Parts = partsFromJSON(line.Parts)

	return msg
}

func partsFromJSON(data json.RawMessage) []session.ContentPart {
	if len(data) == 0 {
		return nil
	}

	var parts []ContentPart
	if err := json.Unmarshal(data, &parts); err != nil {
		return nil
	}

	result := make([]session.ContentPart, 0, len(parts))
	for _, cp := range parts {
		switch cp.Type {
		case "text":
			var td TextData
			if err := json.Unmarshal(cp.Data, &td); err != nil {
				continue
			}
			result = append(result, session.TextContent{Text: td.Text})

		case "reasoning":
			var rd ReasoningData
			if err := json.Unmarshal(cp.Data, &rd); err != nil {
				continue
			}
			rc := session.ReasoningContent{
				Text:      rd.Text,
				Signature: rd.Signature,
			}
			if rd.StartedAt != nil {
				rc.StartedAt = rd.StartedAt.Unix()
			}
			if rd.FinishedAt != nil {
				rc.FinishedAt = rd.FinishedAt.Unix()
			}
			result = append(result, rc)

		case "tool_call":
			var td ToolCallData
			if err := json.Unmarshal(cp.Data, &td); err != nil {
				continue
			}
			result = append(result, session.ToolCall{
				ID:               td.ID,
				Name:             td.Name,
				Input:            string(td.Input),
				ProviderExecuted: td.ProviderExecuted,
				Finished:         td.Finished,
			})

		case "tool_result":
			var td ToolResultData
			if err := json.Unmarshal(cp.Data, &td); err != nil {
				continue
			}
			var content string
			if len(td.Content) > 0 {
				_ = json.Unmarshal(td.Content, &content)
			}
			result = append(result, session.ToolResult{
				ToolCallID: td.ToolCallID,
				Name:       td.Name,
				Content:    content,
				IsError:    td.IsError,
			})

		case "file":
			var fd FileData
			if err := json.Unmarshal(cp.Data, &fd); err != nil {
				continue
			}
			data, _ := base64.StdEncoding.DecodeString(fd.Data)
			result = append(result, session.FileContent{
				Filename:  fd.Filename,
				Data:      data,
				MediaType: fd.MediaType,
			})

		case "finish":
			var fd FinishData
			if err := json.Unmarshal(cp.Data, &fd); err != nil {
				continue
			}
			f := session.Finish{
				Reason: session.FinishReason(fd.Reason),
			}
			if fd.Time != nil {
				f.Time = fd.Time.Unix()
			}
			if fd.Message != nil {
				f.Message = *fd.Message
			}
			result = append(result, f)
		}
	}

	return result
}

// ListSessionHeaders reads headers from all session files in a directory structure.
func ListSessionHeaders(structure *Structure) ([]SessionHeader, error) {
	paths, err := structure.ListAllSessions()
	if err != nil {
		return nil, err
	}

	headers := make([]SessionHeader, 0, len(paths))
	for _, path := range paths {
		h, err := ReadSessionHeader(path)
		if err != nil {
			continue // Skip invalid files
		}
		headers = append(headers, *h)
	}

	// Sort by created_at desc
	sortHeadersByCreatedDesc(headers)

	return headers, nil
}

func sortHeadersByCreatedDesc(headers []SessionHeader) {
	for i := 0; i < len(headers)-1; i++ {
		for j := i + 1; j < len(headers); j++ {
			if headers[j].CreatedAt.After(headers[i].CreatedAt) {
				headers[i], headers[j] = headers[j], headers[i]
			}
		}
	}
}

// FindSessionByID searches for a session by ID.
func FindSessionByID(structure *Structure, sessionID string) (*SessionHeader, string, error) {
	path, _, err := structure.SessionPathByID(sessionID)
	if err != nil {
		return nil, "", err
	}

	header, err := ReadSessionHeader(path)
	if err != nil {
		return nil, "", err
	}

	return header, path, nil
}

// FindSessionsByAgent returns all session headers for an agent.
func FindSessionsByAgent(structure *Structure, agentHandle string) ([]SessionHeader, error) {
	paths, err := structure.ListSessions(agentHandle)
	if err != nil {
		return nil, err
	}

	headers := make([]SessionHeader, 0, len(paths))
	for _, path := range paths {
		h, err := ReadSessionHeader(path)
		if err != nil {
			continue
		}
		headers = append(headers, *h)
	}

	sortHeadersByCreatedDesc(headers)
	return headers, nil
}

// SessionCount returns the total number of sessions.
func SessionCount(structure *Structure) (int, error) {
	paths, err := structure.ListAllSessions()
	if err != nil {
		return 0, err
	}
	return len(paths), nil
}

// DeleteSession removes a session file.
func DeleteSession(structure *Structure, sessionID string) error {
	path, _, err := structure.SessionPathByID(sessionID)
	if err != nil {
		return err
	}
	return os.Remove(path)
}
