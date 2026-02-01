//go:build js && wasm

// Package main provides a WASM entry point for TinyEMU.
// This is a minimal test harness to verify WASM compilation works.
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"syscall/js"
	"time"
)

// ConsoleWriter writes to the JavaScript console and/or a callback function.
type ConsoleWriter struct {
	callback js.Value
}

func (c *ConsoleWriter) Write(p []byte) (n int, err error) {
	if !c.callback.IsUndefined() && !c.callback.IsNull() {
		c.callback.Invoke(string(p))
	}
	return len(p), nil
}

// ConsoleReader reads input from a JavaScript callback.
type ConsoleReader struct {
	buffer    *bytes.Buffer
	inputChan chan []byte
}

func NewConsoleReader() *ConsoleReader {
	return &ConsoleReader{
		buffer:    bytes.NewBuffer(nil),
		inputChan: make(chan []byte, 100),
	}
}

func (c *ConsoleReader) Read(p []byte) (n int, err error) {
	// Non-blocking read from buffer first
	if c.buffer.Len() > 0 {
		return c.buffer.Read(p)
	}

	// Try to get more input from channel (non-blocking)
	select {
	case data := <-c.inputChan:
		c.buffer.Write(data)
		return c.buffer.Read(p)
	default:
		return 0, nil
	}
}

func (c *ConsoleReader) Write(data []byte) {
	c.inputChan <- data
}

// Global state for the emulator
var (
	consoleWriter *ConsoleWriter
	consoleReader *ConsoleReader
	emulatorCtx   context.Context
	emulatorStop  context.CancelFunc
)

func main() {
	fmt.Println("TinyEMU WASM module loaded")

	// Register JavaScript functions
	js.Global().Set("tinyemuInit", js.FuncOf(initEmulator))
	js.Global().Set("tinyemuStart", js.FuncOf(startEmulator))
	js.Global().Set("tinyemuStop", js.FuncOf(stopEmulator))
	js.Global().Set("tinyemuSendInput", js.FuncOf(sendInput))
	js.Global().Set("tinyemuVersion", js.FuncOf(getVersion))

	// Keep the Go program running
	select {}
}

func getVersion(this js.Value, args []js.Value) interface{} {
	return "0.1.0"
}

func initEmulator(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return map[string]interface{}{"error": "missing callback argument"}
	}

	consoleWriter = &ConsoleWriter{callback: args[0]}
	consoleReader = NewConsoleReader()

	return map[string]interface{}{"status": "initialized"}
}

func startEmulator(this js.Value, args []js.Value) interface{} {
	if consoleWriter == nil {
		return map[string]interface{}{"error": "not initialized, call tinyemuInit first"}
	}

	// This is a placeholder - actual emulator start would go here
	// For now, just demonstrate the callback works
	go func() {
		consoleWriter.Write([]byte("TinyEMU starting...\n"))
		time.Sleep(100 * time.Millisecond)
		consoleWriter.Write([]byte("Boot sequence would start here\n"))
	}()

	return map[string]interface{}{"status": "starting"}
}

func stopEmulator(this js.Value, args []js.Value) interface{} {
	if emulatorStop != nil {
		emulatorStop()
	}
	return map[string]interface{}{"status": "stopped"}
}

func sendInput(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 || consoleReader == nil {
		return false
	}

	input := args[0].String()
	consoleReader.Write([]byte(input))
	return true
}

// Verify io.Writer and io.Reader interfaces are satisfied
var _ io.Writer = (*ConsoleWriter)(nil)
var _ io.Reader = (*ConsoleReader)(nil)
