package main

import (
	"os"

	"github.com/charmbracelet/ayo/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
