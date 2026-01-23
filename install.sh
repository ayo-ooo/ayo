#!/bin/bash
set -e

# Model download configuration
MODEL_DIR="${HOME}/.local/share/ayo/models"
ONNX_MODEL_URL="https://huggingface.co/sentence-transformers/all-MiniLM-L6-v2/resolve/main/onnx/model.onnx"
TOKENIZER_URL="https://huggingface.co/sentence-transformers/all-MiniLM-L6-v2/resolve/main/tokenizer.json"
ONNX_MODEL_PATH="${MODEL_DIR}/all-MiniLM-L6-v2.onnx"
TOKENIZER_PATH="${MODEL_DIR}/tokenizer.json"

download_file() {
    local url="$1"
    local dest="$2"
    local name="$3"
    
    if [[ -f "$dest" ]]; then
        echo "  [skip] $name already exists"
        return 0
    fi
    
    echo "  [download] $name..."
    mkdir -p "$(dirname "$dest")"
    
    if command -v curl &> /dev/null; then
        curl -fsSL "$url" -o "$dest" || {
            echo "  [error] Failed to download $name"
            rm -f "$dest"
            return 1
        }
    elif command -v wget &> /dev/null; then
        wget -q "$url" -O "$dest" || {
            echo "  [error] Failed to download $name"
            rm -f "$dest"
            return 1
        }
    else
        echo "  [error] Neither curl nor wget available"
        return 1
    fi
    
    echo "  [done] $name downloaded"
}

setup_onnx_runtime() {
    # Check if library already exists in common locations
    if [[ -f "/opt/homebrew/lib/libonnxruntime.dylib" ]] || \
       [[ -f "/usr/local/lib/libonnxruntime.dylib" ]] || \
       [[ -f "/usr/lib/libonnxruntime.so" ]] || \
       [[ -f "/usr/lib/x86_64-linux-gnu/libonnxruntime.so" ]] || \
       [[ -f "/usr/lib/aarch64-linux-gnu/libonnxruntime.so" ]]; then
        echo "  [skip] ONNX Runtime already installed"
        return 0
    fi
    
    # Check if Homebrew is available (macOS/Linux)
    if command -v brew &> /dev/null; then
        # Check if onnxruntime is already installed
        if brew list onnxruntime &> /dev/null; then
            echo "  [skip] ONNX Runtime already installed via Homebrew"
            return 0
        fi
        
        echo "  [install] ONNX Runtime via Homebrew..."
        brew install onnxruntime
        echo "  [done] ONNX Runtime installed"
        return 0
    fi
    
    # Try Linux package managers
    if [[ "$(uname -s)" == "Linux" ]]; then
        # Debian/Ubuntu (apt)
        if command -v apt-get &> /dev/null; then
            echo "  [install] ONNX Runtime via apt..."
            echo "  Note: You may need to add the Microsoft repository first."
            echo "  See: https://onnxruntime.ai/docs/install/"
            echo ""
            read -p "  Attempt to install libonnxruntime-dev? [y/N] " -n 1 -r
            echo
            if [[ "$REPLY" =~ ^[Yy]$ ]]; then
                sudo apt-get update && sudo apt-get install -y libonnxruntime-dev && {
                    echo "  [done] ONNX Runtime installed"
                    return 0
                }
            fi
        fi
        
        # Fedora/RHEL (dnf)
        if command -v dnf &> /dev/null; then
            echo "  [install] ONNX Runtime via dnf..."
            read -p "  Attempt to install onnxruntime-devel? [y/N] " -n 1 -r
            echo
            if [[ "$REPLY" =~ ^[Yy]$ ]]; then
                sudo dnf install -y onnxruntime-devel && {
                    echo "  [done] ONNX Runtime installed"
                    return 0
                }
            fi
        fi
        
        # Arch Linux (pacman)
        if command -v pacman &> /dev/null; then
            echo "  [install] ONNX Runtime via pacman..."
            read -p "  Attempt to install onnxruntime? [y/N] " -n 1 -r
            echo
            if [[ "$REPLY" =~ ^[Yy]$ ]]; then
                sudo pacman -S --noconfirm onnxruntime && {
                    echo "  [done] ONNX Runtime installed"
                    return 0
                }
            fi
        fi
    fi
    
    # Fallback message
    echo ""
    echo "  [warn] Could not auto-install ONNX Runtime."
    echo "         For local embeddings, install manually:"
    echo "           macOS:       brew install onnxruntime"
    echo "           Debian/Ubuntu: apt install libonnxruntime-dev"
    echo "           Fedora/RHEL:   dnf install onnxruntime-devel"
    echo "           Arch Linux:    pacman -S onnxruntime"
    echo ""
    echo "         Or use cloud embeddings (OpenAI, Ollama) instead."
    return 0
}

get_onnx_library_path() {
    # Check Homebrew first
    if command -v brew &> /dev/null; then
        local prefix=$(brew --prefix onnxruntime 2>/dev/null)
        if [[ -n "$prefix" && -d "$prefix/lib" ]]; then
            echo "$prefix/lib"
            return 0
        fi
    fi
    
    # Check common system locations
    if [[ -f "/usr/local/lib/libonnxruntime.dylib" ]]; then
        echo "/usr/local/lib"
        return 0
    fi
    if [[ -f "/opt/homebrew/lib/libonnxruntime.dylib" ]]; then
        echo "/opt/homebrew/lib"
        return 0
    fi
    if [[ -f "/usr/lib/libonnxruntime.so" ]]; then
        echo "/usr/lib"
        return 0
    fi
    
    # Linux arch-specific paths
    if [[ -f "/usr/lib/x86_64-linux-gnu/libonnxruntime.so" ]]; then
        echo "/usr/lib/x86_64-linux-gnu"
        return 0
    fi
    if [[ -f "/usr/lib/aarch64-linux-gnu/libonnxruntime.so" ]]; then
        echo "/usr/lib/aarch64-linux-gnu"
        return 0
    fi
    
    return 1
}

detect_shell_profile() {
    local shell_name=$(basename "$SHELL")
    
    case "$shell_name" in
        zsh)
            if [[ -f "$HOME/.zshrc" ]]; then
                echo "$HOME/.zshrc"
            elif [[ -f "$HOME/.zprofile" ]]; then
                echo "$HOME/.zprofile"
            else
                echo "$HOME/.zshrc"
            fi
            ;;
        bash)
            if [[ -f "$HOME/.bashrc" ]]; then
                echo "$HOME/.bashrc"
            elif [[ -f "$HOME/.bash_profile" ]]; then
                echo "$HOME/.bash_profile"
            elif [[ -f "$HOME/.profile" ]]; then
                echo "$HOME/.profile"
            else
                echo "$HOME/.bashrc"
            fi
            ;;
        fish)
            echo "$HOME/.config/fish/config.fish"
            ;;
        *)
            # Fallback to .profile for POSIX shells
            if [[ -f "$HOME/.profile" ]]; then
                echo "$HOME/.profile"
            else
                echo "$HOME/.bashrc"
            fi
            ;;
    esac
}

check_env_in_profile() {
    local profile="$1"
    local var_name="$2"
    
    if [[ -f "$profile" ]]; then
        grep -q "export ${var_name}=" "$profile" 2>/dev/null
        return $?
    fi
    return 1
}

add_env_to_profile() {
    local profile="$1"
    local var_name="$2"
    local var_value="$3"
    local shell_name=$(basename "$SHELL")
    
    # Create parent directory if needed (for fish)
    mkdir -p "$(dirname "$profile")"
    
    # Add newline if file doesn't end with one
    if [[ -f "$profile" ]] && [[ -s "$profile" ]]; then
        if [[ $(tail -c1 "$profile" | wc -l) -eq 0 ]]; then
            echo "" >> "$profile"
        fi
    fi
    
    # Add the export (fish uses different syntax)
    if [[ "$shell_name" == "fish" ]]; then
        echo "" >> "$profile"
        echo "# ONNX Runtime for ayo embeddings" >> "$profile"
        echo "set -gx ${var_name} ${var_value}" >> "$profile"
    else
        echo "" >> "$profile"
        echo "# ONNX Runtime for ayo embeddings" >> "$profile"
        echo "export ${var_name}=${var_value}" >> "$profile"
    fi
}

setup_embedding_models() {
    echo ""
    echo "Setting up embedding models..."
    
    # Install/check ONNX Runtime
    setup_onnx_runtime
    
    # Download embedding model
    download_file "$ONNX_MODEL_URL" "$ONNX_MODEL_PATH" "embedding model (all-MiniLM-L6-v2)"
    
    # Download tokenizer
    download_file "$TOKENIZER_URL" "$TOKENIZER_PATH" "tokenizer"
    
    # Configure library path if available
    local lib_path=$(get_onnx_library_path)
    if [[ -n "$lib_path" ]]; then
        local profile=$(detect_shell_profile)
        local profile_name=$(basename "$profile")
        
        # Check if already configured
        if check_env_in_profile "$profile" "ONNX_LIBRARY_PATH"; then
            echo "  [skip] ONNX_LIBRARY_PATH already in $profile_name"
        else
            echo ""
            echo "ONNX Runtime found at: $lib_path"
            echo ""
            read -p "Add ONNX_LIBRARY_PATH to ~/$profile_name? [Y/n] " -n 1 -r
            echo
            
            if [[ -z "$REPLY" || "$REPLY" =~ ^[Yy]$ ]]; then
                add_env_to_profile "$profile" "ONNX_LIBRARY_PATH" "$lib_path"
                echo "  [done] Added to ~/$profile_name"
                echo ""
                echo "Run this to use immediately:"
                echo "  source ~/$profile_name"
            else
                echo ""
                echo "To enable local embeddings, add to your shell profile:"
                echo "  export ONNX_LIBRARY_PATH=$lib_path"
            fi
        fi
    fi
}

# Determine if we're on an unmodified main branch in sync with origin
branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
dirty=$(git status --porcelain 2>/dev/null)
behind_ahead=$(git rev-list --left-right --count origin/main...HEAD 2>/dev/null || echo "0 0")

if [[ "$branch" == "main" && -z "$dirty" && "$behind_ahead" == "0	0" ]]; then
    # Clean main branch in sync with origin - install to standard location
    echo "Installing to standard GOBIN location..."
    go install ./cmd/ayo
    echo ""
    ayo setup
else
    # Any other state - install to local .local/bin
    echo "Installing to .local/bin/ (branch: $branch, dirty: ${dirty:+yes}${dirty:-no})..."
    mkdir -p .local/bin
    GOBIN="$(pwd)/.local/bin" go install ./cmd/ayo
    echo ""
    .local/bin/ayo setup
fi

# Download embedding models
setup_embedding_models
