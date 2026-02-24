# TinyEMU Web Build System
# Builds WASM emulator and RISC-V ayo binary for offline web client

.PHONY: all wasm riscv ayod rootfs clean help serve test test-coverage test-coverage-html

# Paths
TINYEMU_DIR := .read-only/tinyemu-go
WEB_DIR := web
WASM_DIR := $(WEB_DIR)/wasm
ASSETS_DIR := $(WEB_DIR)/assets
BUILD_DIR := build

# Go settings
GO := go
GOFLAGS := -ldflags="-s -w"

# Outputs
WASM_OUTPUT := $(ASSETS_DIR)/tinyemu.wasm
RISCV_OUTPUT := $(BUILD_DIR)/ayo-riscv64
AYOD_OUTPUT := $(BUILD_DIR)/ayod-linux-amd64
WASM_EXEC_JS := $(ASSETS_DIR)/wasm_exec.js

# Default target
all: wasm riscv ayod

# Help
help:
	@echo "TinyEMU Web Build System"
	@echo ""
	@echo "Targets:"
	@echo "  make wasm     - Build TinyEMU WASM binary"
	@echo "  make riscv    - Build ayo RISC-V binary"
	@echo "  make ayod     - Build ayod Linux binary (in-sandbox daemon)"
	@echo "  make rootfs   - Build minimal rootfs (requires Linux)"
	@echo "  make serve    - Start local dev server"
	@echo "  make clean    - Remove build artifacts"
	@echo ""
	@echo "Testing:"
	@echo "  make test              - Run all tests"
	@echo "  make test-coverage     - Run tests with coverage (summary)"
	@echo "  make test-coverage-html - Generate HTML coverage report"
	@echo ""
	@echo "Outputs:"
	@echo "  $(WASM_OUTPUT)"
	@echo "  $(WASM_EXEC_JS)"
	@echo "  $(RISCV_OUTPUT)"
	@echo "  $(AYOD_OUTPUT)"

# Build WASM emulator
wasm: $(WASM_OUTPUT) $(WASM_EXEC_JS)

$(WASM_OUTPUT): $(WASM_DIR)/main.go
	@echo "Building TinyEMU WASM..."
	@mkdir -p $(ASSETS_DIR)
	GOOS=js GOARCH=wasm $(GO) build $(GOFLAGS) -o $@ ./$(WASM_DIR)
	@echo "WASM size: $$(du -h $@ | cut -f1)"

$(WASM_EXEC_JS):
	@echo "Copying wasm_exec.js..."
	@mkdir -p $(ASSETS_DIR)
	cp "$$($(GO) env GOROOT)/lib/wasm/wasm_exec.js" $@

# Build RISC-V ayo binary
riscv: $(RISCV_OUTPUT)

$(RISCV_OUTPUT): $(shell find cmd internal -name '*.go')
	@echo "Building ayo for RISC-V..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=riscv64 $(GO) build $(GOFLAGS) -o $@ ./cmd/ayo
	@echo "RISC-V binary size: $$(du -h $@ | cut -f1)"

# Build ayod (in-sandbox daemon) for Linux amd64
ayod: $(AYOD_OUTPUT)

$(AYOD_OUTPUT): $(shell find cmd/ayod internal/ayod -name '*.go')
	@echo "Building ayod for Linux amd64..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $@ ./cmd/ayod
	@echo "ayod binary size: $$(du -h $@ | cut -f1)"

# Build rootfs (requires Linux with buildroot)
rootfs:
	@echo "Building rootfs requires Linux host with buildroot..."
	@echo "See $(TINYEMU_DIR)/images/buildroot/README.md"
	@if [ "$$(uname -s)" = "Linux" ]; then \
		cd $(TINYEMU_DIR)/images/buildroot && ./build.sh; \
	else \
		echo "Skipping rootfs build on non-Linux host"; \
	fi

# Development server
serve: wasm
	@echo "Starting development server on http://localhost:8080"
	cd $(WEB_DIR) && python3 -m http.server 8080

# Clean build artifacts
clean: clean-coverage
	rm -rf $(BUILD_DIR)
	rm -f $(WASM_OUTPUT)
	rm -f $(WASM_EXEC_JS)
	rm -f $(ASSETS_DIR)/*.wasm

# Version info
version:
	@echo "Go version: $$($(GO) version)"
	@echo "GOROOT: $$($(GO) env GOROOT)"
	@echo "TinyEMU: $(TINYEMU_DIR)"

# =============================================================================
# Testing
# =============================================================================

# Coverage output paths
COVERAGE_DIR := $(BUILD_DIR)/coverage
COVERAGE_OUT := $(COVERAGE_DIR)/coverage.out
COVERAGE_HTML := $(COVERAGE_DIR)/coverage.html

# Run tests
test:
	@echo "Running tests..."
	$(GO) test ./...

# Run tests with coverage
test-coverage: $(COVERAGE_OUT)
	@echo "Coverage summary:"
	@$(GO) tool cover -func=$(COVERAGE_OUT) | tail -1
	@echo ""
	@echo "Per-package coverage:"
	@$(GO) tool cover -func=$(COVERAGE_OUT) | grep -E '^[^t]' | sort -t'%' -k3 -rn | head -20
	@echo ""
	@echo "Coverage profile: $(COVERAGE_OUT)"

$(COVERAGE_OUT):
	@echo "Running tests with coverage..."
	@mkdir -p $(COVERAGE_DIR)
	$(GO) test -coverprofile=$(COVERAGE_OUT) -covermode=atomic ./internal/...
	@echo ""

# Generate HTML coverage report
test-coverage-html: $(COVERAGE_OUT)
	@echo "Generating HTML coverage report..."
	$(GO) tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@echo "Coverage report: $(COVERAGE_HTML)"
	@if command -v open >/dev/null 2>&1; then \
		echo "Opening in browser..."; \
		open $(COVERAGE_HTML); \
	elif command -v xdg-open >/dev/null 2>&1; then \
		echo "Opening in browser..."; \
		xdg-open $(COVERAGE_HTML); \
	fi

# Clean coverage files
clean-coverage:
	rm -rf $(COVERAGE_DIR)
