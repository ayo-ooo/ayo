package main

import (
	"context"
	_ "embed"
	"io"
	"os"

	"github.com/charmbracelet/fang"

	"github.com/alexcabrera/ayo/internal/builtin"
	"github.com/alexcabrera/ayo/internal/version"
)

//go:embed ayo-config-schema.json
var configSchema []byte

func init() {
	builtin.ConfigSchema = configSchema
}

func main() {
	ctx := context.Background()
	cmd := newRootCmd()

	// Custom error handler that suppresses "input validation failed" since we already printed it
	errorHandler := func(w io.Writer, styles fang.Styles, err error) {
		if err.Error() == "input validation failed" {
			return // Already printed custom error
		}
		fang.DefaultErrorHandler(w, styles, err)
	}

	if err := fang.Execute(ctx, cmd,
		fang.WithVersion(version.Version),
		fang.WithErrorHandler(errorHandler),
	); err != nil {
		os.Exit(1)
	}
}
