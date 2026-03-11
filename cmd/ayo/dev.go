package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var verbose bool

func newDevCmd() *cobra.Command {
	var runAfterBuild bool

	cmd := &cobra.Command{
		Use:   "dev <directory>",
		Short: "Development mode with hot reload",
		Long: `Run in development mode with automatic rebuilds on file changes.

The dev command:
1. Watches for changes in config.toml, prompts/, skills/, tools/
2. Automatically rebuilds the agent on changes
3. Optionally runs the agent after each build
4. Provides detailed build logs

Usage:
  ayo dev <directory>              Watch and rebuild on changes
  ayo dev <directory> --run        Watch, rebuild, and run
  ayo dev <directory> --verbose    Enable verbose logging

This is ideal for rapid iteration during development.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := args[0]
			return runDev(dir, runAfterBuild)
		},
	}

	cmd.Flags().BoolVar(&runAfterBuild, "run", false, "Run the agent after building")
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	return cmd
}

func runDev(dir string, runAfterBuild bool) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("error resolving directory: %w", err)
	}

	configPath, err := findConfigPath(absDir)
	if err != nil {
		return fmt.Errorf("error finding config: %w", err)
	}

	if verbose {
		log.Printf("Starting dev mode for %s", absDir)
		log.Printf("Watching: %s", configPath)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating watcher: %w", err)
	}
	defer watcher.Close()

	// Add directories to watch
	promptsDir := filepath.Join(absDir, "prompts")
	skillsDir := filepath.Join(absDir, "skills")
	toolsDir := filepath.Join(absDir, "tools")

	if _, err := os.Stat(configPath); err == nil {
		if err := watcher.Add(configPath); err != nil {
			return fmt.Errorf("error watching config: %w", err)
		}
	}

	if _, err := os.Stat(promptsDir); err == nil {
		if err := watcher.Add(promptsDir); err != nil {
			return fmt.Errorf("error watching prompts: %w", err)
		}
	}

	if _, err := os.Stat(skillsDir); err == nil {
		if err := watcher.Add(skillsDir); err != nil {
			return fmt.Errorf("error watching skills: %w", err)
		}
	}

	if _, err := os.Stat(toolsDir); err == nil {
		if err := watcher.Add(toolsDir); err != nil {
			return fmt.Errorf("error watching tools: %w", err)
		}
	}

	log.Println("Watching for changes... (Press Ctrl+C to stop)")

	// Initial build
	if err := triggerBuild(absDir, runAfterBuild); err != nil {
		log.Printf("Initial build failed: %v", err)
	}

	// Debounce timer to avoid rapid successive builds
	debounceTimer := time.NewTimer(0)
	<-debounceTimer.C

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}

			if verbose {
				log.Printf("File changed: %s", event.Name)
			}

			// Debounce: reset timer and wait for 500ms of no changes
			if !debounceTimer.Stop() {
				<-debounceTimer.C
			}
			debounceTimer.Reset(500 * time.Millisecond)

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("Watch error: %v", err)

		case <-debounceTimer.C:
			log.Println("Files changed, rebuilding...")
			if err := triggerBuild(absDir, runAfterBuild); err != nil {
				log.Printf("Build failed: %v", err)
			}
		}
	}
}

func triggerBuild(dir string, runAfterBuild bool) error {
	log.Println("Building...")

	outputPath := filepath.Join(dir, ".build", "bin", "dev-agent")

	if err := runBuild(dir, outputPath, "", ""); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	log.Printf("✓ Build successful: %s", outputPath)

	if runAfterBuild {
		log.Println("Running agent...")
		cmd := exec.Command(outputPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			log.Printf("Agent exited with error: %v", err)
		}
	}

	return nil
}

func findConfigPath(dir string) (string, error) {
	configPath := filepath.Join(dir, "config.toml")
	if _, err := os.Stat(configPath); err == nil {
		return configPath, nil
	}

	teamPath := filepath.Join(dir, "team.toml")
	if _, err := os.Stat(teamPath); err == nil {
		return teamPath, nil
	}

	return "", fmt.Errorf("no config.toml or team.toml found")
}
