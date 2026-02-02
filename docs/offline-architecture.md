# TinyEMU Offline Web Client Architecture

This document describes the architecture of the ayo offline web client, which enables ayo to run entirely in the browser without requiring a server connection.

## Overview

The offline web client uses a RISC-V emulator (TinyEMU) compiled to WebAssembly to run a full Linux system with ayo inside the browser. LLM requests are handled either by WebLLM (local GPU inference) or cloud APIs with user-provided keys.

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              Browser                                     │
├─────────────────────────────────────────────────────────────────────────┤
│  Main Thread                                                             │
│  ├── AyoApp (app.js)           # Application controller                 │
│  ├── LLMRouter (llm-router.js) # Route requests to WebLLM/API           │
│  ├── OfflineStorage (storage.js) # IndexedDB wrapper                    │
│  └── EmulatorController (emulator.js) # Web Worker interface            │
├─────────────────────────────────────────────────────────────────────────┤
│  Web Worker (worker.js)                                                  │
│  └── TinyEMU WASM (tinyemu.wasm)                                        │
│      └── Linux VM                                                        │
│          └── ayo (RISC-V binary)                                        │
├─────────────────────────────────────────────────────────────────────────┤
│  IndexedDB                                                               │
│  ├── config      # API keys, preferences                                │
│  ├── models      # WebLLM model cache                                   │
│  ├── filesystem  # VM filesystem overlay                                │
│  ├── sessions    # Chat history                                         │
│  └── assets      # Cached WASM, rootfs                                  │
└─────────────────────────────────────────────────────────────────────────┘
```

## Operating Modes

### Connected Mode
- Server handles LLM requests and tool execution
- Uses SSE for streaming responses
- Falls back to offline mode if server unavailable

### Offline Mode
- LLM via WebLLM (local GPU) or cloud API (user keys)
- Tool execution via TinyEMU Linux VM
- All data persisted in IndexedDB

## Components

### 1. WASM Emulator (`web/wasm/main.go`)

Go program compiled to WebAssembly that provides the TinyEMU interface:

```go
// Exposed to JavaScript
tinyemuInit(callback)   // Initialize with console output callback
tinyemuStart()          // Start the VM
tinyemuStop()           // Stop the VM
tinyemuSendInput(text)  // Send input to VM console
tinyemuVersion()        // Get version string
```

**Build**: `make wasm`

### 2. Web Worker (`web/worker.js`)

Runs the WASM emulator in a separate thread to avoid blocking the UI:

```javascript
// Message protocol
Main → Worker:
  { type: 'init' }              // Initialize emulator
  { type: 'start' }             // Start VM
  { type: 'stop' }              // Stop VM
  { type: 'input', text: '...'} // Console input

Worker → Main:
  { type: 'output', data: '...' }  // Console output
  { type: 'status', status: '...'} // Status update
  { type: 'error', error: '...' }  // Error
```

### 3. LLM Protocol (`web/js/protocol.js`)

Console escape sequence-based RPC for LLM requests from the VM:

```
Format: ESC ] AYO ; <json> BEL
        \x1B ] AYO ; {...} \x07

Request:  {"type":"llm:request","id":1,"params":{...}}
Response: {"type":"llm:chunk","id":1,"chunk":"...","done":false}
Complete: {"type":"llm:done","id":1}
Error:    {"type":"llm:error","id":1,"error":"..."}
```

### 4. LLM Router (`web/js/llm-router.js`)

Routes LLM requests to appropriate backend:

| Priority | Backend | Requirements |
|----------|---------|--------------|
| 1 | WebLLM | WebGPU + downloaded model |
| 2 | OpenAI | API key configured |
| 3 | Anthropic | API key configured |
| 4 | OpenRouter | API key configured |

### 5. IndexedDB Storage (`web/js/storage.js`)

Persistent storage with encrypted API keys:

```javascript
const storage = await new OfflineStorage().init();
storage.setEncryptionKey(deviceKey);
await storage.setApiKey('openai', 'sk-...');
await storage.saveFile('/home/user/file.txt', content);
```

### 6. Application Controller (`web/js/app.js`)

Main application logic:

```javascript
const app = await new AyoApp().init();
app.onModeChange = (mode) => { /* update UI */ };
await app.sendChatMessage('Hello!');
```

## Data Flow

### Chat Message (Offline Mode)

```
User types message
    ↓
AyoApp.sendChatMessage()
    ↓
LLMRouter.generate()
    ↓
WebLLM / Cloud API
    ↓
Streaming response chunks
    ↓
UI updates
```

### Tool Execution (via VM)

```
LLM returns tool call
    ↓
Send to VM via console escape sequence
    ↓
ayo in VM executes tool
    ↓
Tool output captured
    ↓
Send back to LLM
```

## Build System

```bash
make wasm      # Build WASM binary (2.4MB)
make riscv     # Build RISC-V ayo (52MB)
make rootfs    # Build Linux rootfs (Linux only)
make serve     # Start dev server
```

## File Structure

```
web/
├── index.html              # Main page with UI
├── worker.js               # Web Worker entry
├── assets/
│   ├── tinyemu.wasm        # WASM binary
│   └── wasm_exec.js        # Go WASM runtime
└── js/
    ├── app.js              # Application controller
    ├── emulator.js         # Emulator controller
    ├── llm-router.js       # LLM routing
    ├── protocol.js         # LLM RPC protocol
    └── storage.js          # IndexedDB storage
```

## Security

### API Key Storage
- Keys encrypted with AES-256-GCM
- Key derivation via PBKDF2 (100k iterations)
- Device-specific encryption key

### VM Isolation
- WASM sandbox isolates VM from browser
- No direct filesystem access
- Network via browser fetch only

## Performance

| Metric | Value |
|--------|-------|
| WASM size | 2.4 MB |
| ayo binary | 52 MB (compressed ~12 MB) |
| Rootfs | ~5.5 MB compressed |
| Boot time | ~5-10s (estimate) |
| LLM latency | Depends on backend |

## Browser Support

| Browser | WebGPU | Status |
|---------|--------|--------|
| Chrome 113+ | Yes | Full support |
| Safari 17+ | Yes | Full support |
| Firefox | Experimental | Cloud API only |
| Edge | Yes | Full support |
| Mobile | Limited | Cloud API recommended |
