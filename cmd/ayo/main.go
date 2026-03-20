package main

import (
	"os"

	"github.com/ayo-ooo/ayo/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
