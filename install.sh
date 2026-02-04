#!/bin/bash
set -e

# Colors for output (fallback when gum not available)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Required Ollama models
REQUIRED_MODELS=("ministral-3:3b" "nomic-embed-text")

# Docker images for sandbox
SANDBOX_BASE_IMAGE="busybox:stable"
SANDBOX_EXTENDED_IMAGE="ayo-sandbox:latest"

# Apple Container (macOS 15+)
APPLE_CONTAINER_MIN_VERSION="15.0"
APPLE_CONTAINER_PKG_URL="https://github.com/apple/container/releases"

# Flags
CLEAN=false
FORCE=false

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --clean)
            CLEAN=true
            shift
            ;;
        --force|-f)
            FORCE=true
            shift
            ;;
        --help|-h)
            echo "Usage: ./install.sh [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --clean     Remove all ayo data before installing"
            echo "  --force,-f  Skip confirmation prompts"
            echo "  --help,-h   Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Run './install.sh --help' for usage"
            exit 1
            ;;
    esac
done

# Check if gum is available and we have a TTY (interactive mode)
has_gum() {
    command -v gum &> /dev/null && [[ -t 0 ]] && [[ -t 1 ]]
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
    if $FORCE; then
        return 0
    fi
    if has_gum; then
        gum confirm "$prompt"
    else
        read -p "$prompt [y/N] " -n 1 -r
        echo
        [[ "$REPLY" =~ ^[Yy]$ ]]
    fi
}

# Determine if this is a dev install (in git repo working directory)
is_dev_install() {
    # Check if we're in a git repo with install.sh
    [[ -d ".git" && -f "install.sh" && -f "go.mod" ]]
}

# Clean up ayo installation
clean_install() {
    header "Cleaning Ayo Installation"
    
    local is_dev=$(is_dev_install && echo "true" || echo "false")
    
    if [[ "$is_dev" == "true" ]]; then
        # Dev install - clean local directories without prompting
        info "Dev mode: cleaning local directories..."
        
        local dirs_to_clean=(".local" ".config/ayo")
        for dir in "${dirs_to_clean[@]}"; do
            if [[ -d "$dir" ]]; then
                rm -rf "$dir"
                success "Removed $dir"
            fi
        done
    else
        # Production install - prompt before cleaning
        warn "This will remove ALL ayo data including:"
        echo "  - Configuration (~/.config/ayo/)"
        echo "  - Database and sessions (~/.local/share/ayo/)"
        echo "  - Stored credentials"
        echo "  - Custom agents and skills"
        echo ""
        
        if ! confirm "Are you sure you want to remove all ayo data?"; then
            info "Clean cancelled."
            exit 0
        fi
        
        # Remove production directories
        local dirs_to_clean=(
            "$HOME/.config/ayo"
            "$HOME/.local/share/ayo"
        )
        
        for dir in "${dirs_to_clean[@]}"; do
            if [[ -d "$dir" ]]; then
                rm -rf "$dir"
                success "Removed $dir"
            fi
        done
        
        # Try to remove the binary from GOBIN
        local gobin="${GOBIN:-$HOME/go/bin}"
        if [[ -f "$gobin/ayo" ]]; then
            rm -f "$gobin/ayo"
            success "Removed $gobin/ayo"
        fi
    fi
    
    success "Clean complete"
    echo ""
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

# Check if Docker is available
check_docker() {
    command -v docker &> /dev/null && docker info &> /dev/null
}

# Check if a Docker image is available locally
has_docker_image() {
    local image="$1"
    docker images -q "$image" 2>/dev/null | grep -q .
}

# Setup Docker sandbox images
setup_docker() {
    header "Sandbox Setup"
    
    if ! check_docker; then
        warn "Docker not available - sandbox will run on host"
        echo "Install Docker Desktop to enable sandboxed execution:"
        echo "  https://docs.docker.com/get-docker/"
        return 0
    fi
    
    success "Docker available"
    
    # Pull base image
    if has_docker_image "$SANDBOX_BASE_IMAGE"; then
        success "$SANDBOX_BASE_IMAGE already available"
    else
        info "Pulling $SANDBOX_BASE_IMAGE..."
        if spin "Pulling $SANDBOX_BASE_IMAGE..." docker pull "$SANDBOX_BASE_IMAGE"; then
            success "$SANDBOX_BASE_IMAGE ready"
        else
            warn "Failed to pull $SANDBOX_BASE_IMAGE"
        fi
    fi
    
    # Check for extended image or build it
    if has_docker_image "$SANDBOX_EXTENDED_IMAGE"; then
        success "$SANDBOX_EXTENDED_IMAGE already available"
    else
        # Check if we have the Dockerfile to build it
        local dockerfile="internal/sandbox/images/Dockerfile"
        if [[ -f "$dockerfile" ]]; then
            info "Building $SANDBOX_EXTENDED_IMAGE..."
            if spin "Building $SANDBOX_EXTENDED_IMAGE..." docker build -t "$SANDBOX_EXTENDED_IMAGE" -f "$dockerfile" .; then
                success "$SANDBOX_EXTENDED_IMAGE built"
            else
                warn "Failed to build $SANDBOX_EXTENDED_IMAGE - using base image"
            fi
        else
            info "$SANDBOX_EXTENDED_IMAGE not found - using base image"
        fi
    fi
}

# Check if running macOS 15+ (Sequoia)
check_macos_version() {
    if [[ "$(uname -s)" != "Darwin" ]]; then
        return 1
    fi
    
    local version=$(sw_vers -productVersion 2>/dev/null | cut -d. -f1)
    [[ "$version" -ge 15 ]]
}

# Check if Apple Container is installed
check_apple_container() {
    command -v container &> /dev/null
}

# Check if Apple Container service is running
check_apple_container_running() {
    container system status &>/dev/null
}

# Setup Apple Container (macOS 15+ only)
setup_apple_container() {
    # Only run on macOS
    if [[ "$(uname -s)" != "Darwin" ]]; then
        return 0
    fi
    
    # Check macOS version
    if ! check_macos_version; then
        info "Apple Container requires macOS 15+ (current: $(sw_vers -productVersion 2>/dev/null || echo 'unknown'))"
        return 0
    fi
    
    # Check for Apple Silicon
    if [[ "$(uname -m)" != "arm64" ]]; then
        info "Apple Container requires Apple Silicon"
        return 0
    fi
    
    header "Apple Container Setup"
    
    if has_gum; then
        gum format << EOF
**Apple Container** provides native Linux container support on macOS 15+.

Benefits over Docker:
* Native virtualization (faster startup)
* Lower resource usage
* Optimized for Apple Silicon
* virtiofs for fast file sharing

This is optional - Docker works great too!
EOF
    else
        echo "Apple Container provides native Linux container support on macOS 15+."
        echo ""
        echo "Benefits over Docker:"
        echo "  - Native virtualization (faster startup)"
        echo "  - Lower resource usage"
        echo "  - Optimized for Apple Silicon"
        echo "  - virtiofs for fast file sharing"
        echo ""
        echo "This is optional - Docker works great too!"
    fi
    echo ""
    
    if check_apple_container; then
        success "Apple Container already installed"
        
        # Check if service is running
        if check_apple_container_running; then
            success "Apple Container service running"
        else
            if confirm "Start Apple Container service?"; then
                info "Starting Apple Container service..."
                if container system start; then
                    success "Apple Container service started"
                else
                    warn "Failed to start Apple Container service"
                fi
            fi
        fi
        return 0
    fi
    
    # Apple Container not installed - offer to install
    echo ""
    info "Apple Container is not installed."
    echo "Download the installer from: $APPLE_CONTAINER_PKG_URL"
    echo ""
    
    if confirm "Open the download page in your browser?"; then
        open "$APPLE_CONTAINER_PKG_URL"
        echo ""
        info "After installing, run: container system start"
    fi
}

# Main installation flow
main() {
    # Handle clean flag first
    if $CLEAN; then
        clean_install
    fi
    
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
        if $FORCE; then
            ayo setup --force
        else
            ayo setup
        fi
    else
        # Any other state - install to local .local/bin
        info "Installing to .local/bin/ (branch: $branch, dirty: ${dirty:+yes}${dirty:-no})..."
        mkdir -p .local/bin
        spin "Building ayo..." env GOBIN="$(pwd)/.local/bin" go install ./cmd/ayo
        success "ayo installed to .local/bin/"
        echo ""
        if $FORCE; then
            .local/bin/ayo setup --force
        else
            .local/bin/ayo setup
        fi
    fi
    
    # Setup Ollama for local AI features
    setup_ollama
    
    # Setup Docker sandbox images
    setup_docker
    
    # Setup Apple Container on macOS 15+
    setup_apple_container
    
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
    
    if check_docker; then
        if has_docker_image "$SANDBOX_BASE_IMAGE"; then
            success "Sandbox image ready"
        else
            warn "Sandbox image not pulled - sandbox will be slow on first use"
        fi
    fi
    
    # Check Apple Container on macOS
    if [[ "$(uname -s)" == "Darwin" ]] && check_apple_container; then
        success "Apple Container available (native virtualization)"
    fi
    
    echo ""
    info "Run 'ayo --help' to get started"
}

main "$@"
