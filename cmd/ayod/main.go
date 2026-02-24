// ayod is the in-sandbox daemon for ayo.
// It runs as PID 1 inside sandboxes and provides a consistent interface
// for all sandbox operations via a Unix socket.
//
// Usage:
//
//	ayod
//
// ayod listens on /run/ayod.sock and handles RPC requests for:
//   - User management (create agent users)
//   - Command execution (run commands as users)
//   - File operations (read/write files)
//   - Health checks
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/alexcabrera/ayo/internal/ayod"
)

func main() {
	// Set up logging
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.SetPrefix("ayod: ")

	// Handle signals for clean shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	// Run server in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- ayod.Run()
	}()

	// Wait for signal or error
	select {
	case sig := <-sigCh:
		log.Printf("received signal %v, shutting down", sig)
	case err := <-errCh:
		if err != nil {
			log.Fatalf("server error: %v", err)
		}
	}
}
