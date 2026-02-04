// Package jsonl provides JSONL-based session file storage.
package jsonl

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/alexcabrera/ayo/internal/session"
)

// Writer writes session data to a JSONL file.
type Writer struct {
	mu       sync.Mutex
	file     *os.File
	writer   *bufio.Writer
	header   SessionHeader
	path     string
	msgCount int
}

// NewWriter creates a new JSONL session writer.
// It creates the file and writes the initial session header.
func NewWriter(structure *Structure, sess session.Session) (*Writer, error) {
	// Ensure directory exists
	createdAt := time.Unix(sess.CreatedAt, 0)
	if err := structure.EnsureDir(sess.AgentHandle, createdAt); err != nil {
		return nil, fmt.Errorf("ensure dir: %w", err)
	}

	path := structure.SessionPath(sess.AgentHandle, sess.ID, createdAt)

	// Create file
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}

	header := sessionToHeader(sess)

	w := &Writer{
		file:   file,
		writer: bufio.NewWriter(file),
		header: header,
		path:   path,
	}

	// Write initial header
	if err := w.writeHeader(); err != nil {
		file.Close()
		return nil, fmt.Errorf("write header: %w", err)
	}

	return w, nil
}

// OpenWriter opens an existing session file for appending.
func OpenWriter(path string) (*Writer, error) {
	// Read existing header first
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	scanner := bufio.NewScanner(file)
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

	// Count existing messages
	msgCount := 0
	for scanner.Scan() {
		msgCount++
	}
	if scanner.Err() != nil {
		file.Close()
		return nil, fmt.Errorf("count messages: %w", scanner.Err())
	}

	// Seek to end for appending
	if _, err := file.Seek(0, 2); err != nil {
		file.Close()
		return nil, fmt.Errorf("seek to end: %w", err)
	}

	return &Writer{
		file:     file,
		writer:   bufio.NewWriter(file),
		header:   *header,
		path:     path,
		msgCount: msgCount,
	}, nil
}

// WriteMessage writes a message to the session file.
func (w *Writer) WriteMessage(msg session.Message) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	line := messageToLine(msg)
	data, err := MarshalLine(line)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	if _, err := w.writer.Write(data); err != nil {
		return fmt.Errorf("write message: %w", err)
	}
	if err := w.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("write newline: %w", err)
	}
	if err := w.writer.Flush(); err != nil {
		return fmt.Errorf("flush: %w", err)
	}

	w.msgCount++
	w.header.MessageCount = w.msgCount
	w.header.UpdatedAt = time.Now().UTC()

	return nil
}

// Finish marks the session as finished and updates the header.
func (w *Writer) Finish(structuredOutput string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now().UTC()
	w.header.FinishedAt = &now
	w.header.UpdatedAt = now
	if structuredOutput != "" {
		w.header.StructuredOutput = &structuredOutput
	}

	return w.rewriteHeader()
}

// Close flushes and closes the writer.
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Update header with final counts
	w.header.UpdatedAt = time.Now().UTC()

	// Rewrite header with final state
	if err := w.rewriteHeader(); err != nil {
		w.file.Close()
		return fmt.Errorf("update header: %w", err)
	}

	if err := w.writer.Flush(); err != nil {
		w.file.Close()
		return fmt.Errorf("flush: %w", err)
	}

	return w.file.Close()
}

// Path returns the file path.
func (w *Writer) Path() string {
	return w.path
}

// MessageCount returns the number of messages written.
func (w *Writer) MessageCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.msgCount
}

func (w *Writer) writeHeader() error {
	data, err := MarshalLine(w.header)
	if err != nil {
		return err
	}
	if _, err := w.writer.Write(data); err != nil {
		return err
	}
	if err := w.writer.WriteByte('\n'); err != nil {
		return err
	}
	return w.writer.Flush()
}

// rewriteHeader rewrites the header in place by rewriting the entire file.
// This is expensive but necessary for updating the header.
func (w *Writer) rewriteHeader() error {
	// Read all content after header
	if err := w.writer.Flush(); err != nil {
		return err
	}

	// Seek to start, skip header line
	if _, err := w.file.Seek(0, 0); err != nil {
		return err
	}

	scanner := bufio.NewScanner(w.file)
	if !scanner.Scan() {
		return ErrEmptyFile
	}

	// Collect remaining lines
	var lines [][]byte
	for scanner.Scan() {
		line := make([]byte, len(scanner.Bytes()))
		copy(line, scanner.Bytes())
		lines = append(lines, line)
	}
	if scanner.Err() != nil {
		return scanner.Err()
	}

	// Truncate and rewrite
	if err := w.file.Truncate(0); err != nil {
		return err
	}
	if _, err := w.file.Seek(0, 0); err != nil {
		return err
	}

	w.writer.Reset(w.file)

	// Write updated header
	if err := w.writeHeader(); err != nil {
		return err
	}

	// Write all messages
	for _, line := range lines {
		if _, err := w.writer.Write(line); err != nil {
			return err
		}
		if err := w.writer.WriteByte('\n'); err != nil {
			return err
		}
	}

	return w.writer.Flush()
}

func sessionToHeader(sess session.Session) SessionHeader {
	createdAt := time.Unix(sess.CreatedAt, 0).UTC()
	updatedAt := time.Unix(sess.UpdatedAt, 0).UTC()

	h := SessionHeader{
		Type:         LineTypeSession,
		ID:           sess.ID,
		AgentHandle:  sess.AgentHandle,
		Title:        sess.Title,
		Source:       sess.Source,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
		ChainDepth:   int(sess.ChainDepth),
		ChainSource:  sess.ChainSource,
		MessageCount: int(sess.MessageCount),
	}

	if sess.FinishedAt > 0 {
		finishedAt := time.Unix(sess.FinishedAt, 0).UTC()
		h.FinishedAt = &finishedAt
	}

	if sess.InputSchema != "" {
		h.InputSchema = &sess.InputSchema
	}
	if sess.OutputSchema != "" {
		h.OutputSchema = &sess.OutputSchema
	}
	if sess.StructuredInput != "" {
		h.StructuredInput = &sess.StructuredInput
	}
	if sess.StructuredOutput != "" {
		h.StructuredOutput = &sess.StructuredOutput
	}

	return h
}

func messageToLine(msg session.Message) MessageLine {
	createdAt := time.Unix(msg.CreatedAt, 0).UTC()
	updatedAt := time.Unix(msg.UpdatedAt, 0).UTC()

	line := MessageLine{
		Type:      LineTypeMessage,
		ID:        msg.ID,
		Role:      string(msg.Role),
		Model:     msg.Model,
		Provider:  msg.Provider,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	if msg.FinishedAt > 0 {
		finishedAt := time.Unix(msg.FinishedAt, 0).UTC()
		line.FinishedAt = &finishedAt
	}

	// Convert parts to JSON
	parts := partsToJSON(msg.Parts)
	partsData, _ := json.Marshal(parts)
	line.Parts = partsData

	return line
}

func partsToJSON(parts []session.ContentPart) []ContentPart {
	result := make([]ContentPart, 0, len(parts))
	for _, part := range parts {
		cp := ContentPart{}
		switch p := part.(type) {
		case session.TextContent:
			cp.Type = "text"
			data, _ := json.Marshal(TextData{Text: p.Text})
			cp.Data = data
		case session.ReasoningContent:
			cp.Type = "reasoning"
			rd := ReasoningData{
				Text:      p.Text,
				Signature: p.Signature,
			}
			if p.StartedAt > 0 {
				t := time.Unix(p.StartedAt, 0).UTC()
				rd.StartedAt = &t
			}
			if p.FinishedAt > 0 {
				t := time.Unix(p.FinishedAt, 0).UTC()
				rd.FinishedAt = &t
			}
			data, _ := json.Marshal(rd)
			cp.Data = data
		case session.ToolCall:
			cp.Type = "tool_call"
			td := ToolCallData{
				ID:               p.ID,
				Name:             p.Name,
				Input:            json.RawMessage(p.Input),
				ProviderExecuted: p.ProviderExecuted,
				Finished:         p.Finished,
			}
			data, _ := json.Marshal(td)
			cp.Data = data
		case session.ToolResult:
			cp.Type = "tool_result"
			// Properly marshal content as JSON string
			contentJSON, _ := json.Marshal(p.Content)
			td := ToolResultData{
				ToolCallID: p.ToolCallID,
				Name:       p.Name,
				Content:    contentJSON,
				IsError:    p.IsError,
			}
			data, _ := json.Marshal(td)
			cp.Data = data
		case session.FileContent:
			cp.Type = "file"
			fd := FileData{
				Filename:  p.Filename,
				Data:      base64.StdEncoding.EncodeToString(p.Data),
				MediaType: p.MediaType,
			}
			data, _ := json.Marshal(fd)
			cp.Data = data
		case session.Finish:
			cp.Type = "finish"
			fd := FinishData{
				Reason: string(p.Reason),
			}
			if p.Time > 0 {
				t := time.Unix(p.Time, 0).UTC()
				fd.Time = &t
			}
			if p.Message != "" {
				fd.Message = &p.Message
			}
			data, _ := json.Marshal(fd)
			cp.Data = data
		}
		result = append(result, cp)
	}
	return result
}
