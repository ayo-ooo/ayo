# ayo Offline Mode User Guide

This guide explains how to use ayo in offline mode, which runs entirely in your browser without requiring a server connection.

## Overview

ayo's offline mode provides a complete AI assistant experience that works without an internet connection after initial setup:

- **Chat with AI**: Use WebLLM for local inference or configure cloud API keys
- **Terminal Access**: Full Linux shell running in a TinyEMU virtual machine
- **File Management**: Create, edit, and manage files in the browser-based filesystem
- **Persistent Storage**: Your sessions, files, and settings are stored locally

## Getting Started

### 1. Open the Web App

Navigate to the ayo offline web client in your browser. The app will automatically detect whether you're in connected or offline mode.

### 2. Configure an LLM Provider

You need at least one LLM provider to chat with ayo:

**Option A: Local Models (WebLLM)**

If your browser supports WebGPU (Chrome 113+, Edge 113+):

1. Go to the **Settings** tab
2. Under "Local Models (WebLLM)", you'll see available models
3. Click **Download** on your preferred model
4. Wait for the download to complete (models are 700MB-4GB)
5. The model is now cached locally for offline use

Recommended models:
- **Llama-3.2-1B-Instruct**: Fastest, good for simple tasks (700MB, 2GB VRAM)
- **Phi-3.5-mini-instruct**: Best quality/size ratio (2GB, 4GB VRAM)
- **Llama-3.2-3B-Instruct**: Higher quality (1.8GB, 4GB VRAM)

**Option B: Cloud API Keys**

For better performance or if WebGPU isn't available:

1. Go to the **Settings** tab
2. Under "API Keys", enter your key for one of:
   - OpenAI: Get key at [platform.openai.com](https://platform.openai.com/api-keys)
   - Anthropic: Get key at [console.anthropic.com](https://console.anthropic.com/settings/keys)
   - OpenRouter: Get key at [openrouter.ai](https://openrouter.ai/keys) (access to 100+ models)
3. Click **Save** and optionally **Test** to verify the key works

## Using the Chat Tab

The Chat tab is your main interface for interacting with ayo:

1. Type your message in the input box
2. Press **Enter** or click **Send**
3. The response will stream in real-time
4. Press **Escape** to cancel a running request

### Session Persistence

Your chat sessions are automatically saved to IndexedDB:
- Sessions persist across browser refreshes
- Chat history is stored locally

## Using the Terminal Tab

The Terminal tab provides access to a full Linux shell:

1. Click the **Terminal** tab (or press `Ctrl+2`)
2. Wait for the VM to boot (first time takes 5-15 seconds)
3. Once ready, you have a full BusyBox Linux shell

### What You Can Do

- Run shell commands (`ls`, `cat`, `echo`, etc.)
- Create and edit files
- Use text-based tools
- Execute scripts

### Limitations

- No network access from within the VM
- Limited to the tools included in BusyBox
- VM state is reset when you refresh the page

## Using the Files Tab

The Files tab lets you browse and edit files in the overlay filesystem:

1. Click **Refresh** to load the current file list
2. Click a file to view/edit its contents
3. Make changes in the editor
4. Click **Save** to persist changes

### File Operations

- **+ File**: Create a new file
- **Download**: Download the selected file to your computer
- **Delete**: Remove the file from the overlay

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+1` | Switch to Chat tab |
| `Ctrl+2` | Switch to Terminal tab |
| `Ctrl+3` | Switch to Files tab |
| `Ctrl+4` | Switch to Settings tab |
| `Escape` | Cancel current chat request |

## Storage Management

All data is stored in your browser's IndexedDB:

### View Storage Usage

1. Go to **Settings** tab
2. See "Storage Usage" under the Storage section

### Clear All Data

1. Go to **Settings** tab
2. Click **Clear All Data** under Storage
3. Confirm the action

This will delete:
- Downloaded models
- Saved API keys
- Chat sessions
- Files in the overlay
- Cached assets

## Troubleshooting

### "WebGPU Not Available"

WebGPU is required for local model inference. To fix:
- Use Chrome 113+ or Edge 113+
- Enable WebGPU in browser flags if needed
- Some GPUs may not be supported

Workaround: Use cloud API keys instead.

### "No LLM backend available"

You need to either:
- Download a WebLLM model (if WebGPU is available)
- Configure an API key in Settings

### VM Won't Start

If the terminal shows an error:
- Refresh the page and try again
- Clear browser data and reload
- Check if you have enough memory available

### Slow Performance

- WebLLM models run on your GPU; performance varies by hardware
- Consider using a smaller model (1B instead of 3B)
- Cloud APIs are generally faster than local inference

## Mode Badge

The mode badge in the top right shows your current status:

- **Connected** (green): Server connection active
- **Offline** (yellow): Running locally in browser
- **Detecting...** (blue): Checking connection status

## Privacy & Security

### API Key Storage

Your API keys are encrypted using AES-256-GCM before being stored in IndexedDB. The encryption key is derived from a randomly generated device-specific key using PBKDF2 with 100,000 iterations.

### Data Locality

All data stays in your browser:
- No data is sent to ayo servers in offline mode
- WebLLM models run entirely on your GPU
- Chat history is stored in IndexedDB

When using cloud API keys, your messages are sent to the respective provider (OpenAI, Anthropic, etc.).

## Browser Requirements

### Minimum Requirements
- Chrome 90+, Firefox 90+, Safari 16+, or Edge 90+
- 4GB available RAM
- IndexedDB support

### Recommended for WebLLM
- Chrome 113+ or Edge 113+
- WebGPU-capable GPU with 4GB+ VRAM
- 8GB+ system RAM

## FAQ

**Q: Can I use ayo offline without downloading a model?**
A: Yes, if you configure a cloud API key. However, you'll need internet access when chatting.

**Q: How much storage do models use?**
A: Models range from 700MB to 4GB. The model cache is stored in IndexedDB.

**Q: Is my data secure?**
A: API keys are encrypted at rest. All other data is stored in plain text in IndexedDB. Clear your browser data to remove it completely.

**Q: Why is the first boot slow?**
A: The VM needs to load the WASM binary and boot Linux. Subsequent boots may be faster due to caching.
