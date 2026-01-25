#!/bin/bash
set -e

# Colors for output (fallback when gum not available)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check for Crush and offer to install it
check_crush() {
    if command -v crush &> /dev/null; then
        echo "Crush detected: $(command -v crush)"
        return 0
    fi

    echo ""
    echo "Crush is not installed."
    echo ""
    echo "Crush is an AI-powered coding agent that ayo uses for all source code"
    echo "creation and modification tasks. Without it, coding features will be"
    echo "unavailable."
    echo ""
    read -p "Would you like to install Crush now? [Y/n] " -n 1 -r
    echo ""

    if [[ $REPLY =~ ^[Nn]$ ]]; then
        echo "Skipping Crush installation. You can install it later with:"
        echo "  go install github.com/charmbracelet/crush@latest"
        return 0
    fi

    echo "Installing Crush..."
    go install github.com/charmbracelet/crush@latest

    if command -v crush &> /dev/null; then
        echo "Crush installed successfully: $(command -v crush)"
    else
        echo ""
        echo "Warning: Crush was installed but is not in your PATH."
        echo "Make sure your Go bin directory is in your PATH:"
        echo "  export PATH=\"\$PATH:\$(go env GOPATH)/bin\""
    fi
}

# Check for Crush first
check_crush

echo ""

# Required Ollama models
REQUIRED_MODELS=("ministral-3:3b" "nomic-embed-text")

# Check if gum is available
has_gum() {
    command -v gum &> /dev/null
}

# Styled output functions
info() {
    if has_gum; then
        gum style --foreground 4 "$1"
    else
        echo -e "${BLUE}$1${NC}"
    fi
}

success() {
    if has_gum; then
        gum style --foreground 2 "✓ $1"
    else
        echo -e "${GREEN}✓ $1${NC}"
    fi
}

warn() {
    if has_gum; then
        gum style --foreground 3 "! $1"
    else
        echo -e "${YELLOW}! $1${NC}"
    fi
}

error() {
    if has_gum; then
        gum style --foreground 1 "✗ $1"
    else
        echo -e "${RED}✗ $1${NC}"
    fi
}

header() {
    echo ""
    if has_gum; then
        gum style --bold --border double --padding "0 2" "$1"
    else
        echo "========================================"
        echo "  $1"
        echo "========================================"
    fi
    echo ""
}

# Spinner for long-running commands
spin() {
    local title="$1"
    shift
    if has_gum; then
        gum spin --spinner dot --title "$title" -- "$@"
    else
        echo "$title"
        "$@"
    fi
}

# Confirm prompt
confirm() {
    local prompt="$1"
    if has_gum; then
        gum confirm "$prompt"
    else
        read -p "$prompt [y/N] " -n 1 -r
        echo
        [[ "$REPLY" =~ ^[Yy]$ ]]
    fi
}

# Check if Ollama is installed
check_ollama_installed() {
    command -v ollama &> /dev/null
}

# Check if Ollama is running
check_ollama_running() {
    curl -s http://localhost:11434/api/tags &> /dev/null
}

# Install Ollama
install_ollama() {
    header "Ollama Setup"
    
    if has_gum; then
        gum format << EOF
Ayo uses **Ollama** for local AI features:

* Memory formation (remembering your preferences)
* Session title generation
* Semantic search across memories

All processing happens **locally** on your machine.
No data is sent to external servers.
EOF
    else
        echo "Ayo uses Ollama for local AI features:"
        echo ""
        echo "  - Memory formation (remembering your preferences)"
        echo "  - Session title generation"
        echo "  - Semantic search across memories"
        echo ""
        echo "All processing happens locally on your machine."
    fi
    echo ""
    
    if confirm "Install Ollama?"; then
        info "Installing Ollama..."
        if curl -fsSL https://ollama.ai/install.sh | sh; then
            success "Ollama installed"
        else
            error "Failed to install Ollama"
            echo ""
            echo "You can install manually from: https://ollama.ai"
            return 1
        fi
    else
        warn "Skipping Ollama installation"
        echo "Memory features will be disabled without Ollama."
        echo "Install later with: curl -fsSL https://ollama.ai/install.sh | sh"
        return 1
    fi
}

# Start Ollama service
start_ollama() {
    if check_ollama_running; then
        return 0
    fi
    
    info "Starting Ollama service..."
    
    # Try to start Ollama
    if [[ "$(uname -s)" == "Darwin" ]]; then
        # macOS: start the app or service
        open -a Ollama 2>/dev/null || ollama serve &>/dev/null &
    else
        # Linux: start in background
        ollama serve &>/dev/null &
    fi
    
    # Wait for Ollama to be ready
    local max_attempts=30
    local attempt=0
    while ! check_ollama_running; do
        sleep 1
        ((attempt++))
        if [[ $attempt -ge $max_attempts ]]; then
            error "Ollama failed to start"
            return 1
        fi
    done
    
    success "Ollama service started"
}

# Check if a model is installed
has_model() {
    local model="$1"
    ollama list 2>/dev/null | grep -q "^${model}"
}

# Pull required models
pull_models() {
    for model in "${REQUIRED_MODELS[@]}"; do
        if has_model "$model"; then
            success "$model already installed"
        else
            info "Pulling $model..."
            if spin "Pulling $model..." ollama pull "$model"; then
                success "$model installed"
            else
                error "Failed to pull $model"
            fi
        fi
    done
}

# Setup Ollama and models
setup_ollama() {
    if ! check_ollama_installed; then
        install_ollama || return 1
    else
        success "Ollama already installed"
    fi
    
    if ! check_ollama_running; then
        start_ollama || return 1
    else
        success "Ollama service running"
    fi
    
    echo ""
    info "Checking required models..."
    pull_models
}

# Main installation flow
main() {
    header "Ayo Installation"
    
    # Determine install location based on git state
    local branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
    local dirty=$(git status --porcelain 2>/dev/null)
    local behind_ahead=$(git rev-list --left-right --count origin/main...HEAD 2>/dev/null || echo "0 0")
    
    if [[ "$branch" == "main" && -z "$dirty" && "$behind_ahead" == "0	0" ]]; then
        # Clean main branch in sync with origin - install to standard location
        info "Installing to standard GOBIN location..."
        spin "Building ayo..." go install ./cmd/ayo
        success "ayo installed to GOBIN"
        echo ""
        ayo setup
    else
        # Any other state - install to local .local/bin
        info "Installing to .local/bin/ (branch: $branch, dirty: ${dirty:+yes}${dirty:-no})..."
        mkdir -p .local/bin
        spin "Building ayo..." env GOBIN="$(pwd)/.local/bin" go install ./cmd/ayo
        success "ayo installed to .local/bin/"
        echo ""
        .local/bin/ayo setup
    fi
    
    # Setup Ollama for local AI features
    setup_ollama
    
    # Final summary
    header "Installation Complete"
    
    if check_ollama_running; then
        success "Ollama is running"
        for model in "${REQUIRED_MODELS[@]}"; do
            if has_model "$model"; then
                success "$model ready"
            else
                warn "$model not installed"
            fi
        done
    else
        warn "Ollama not running - memory features disabled"
    fi
    
    echo ""
    info "Run 'ayo --help' to get started"
}

main "$@"
