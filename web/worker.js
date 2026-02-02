/**
 * Web Worker for TinyEMU RISC-V Emulator
 * 
 * Runs the emulator in a separate thread and communicates with
 * the main thread via postMessage.
 */

// Import wasm_exec.js for Go WASM support
importScripts('assets/wasm_exec.js');

// State
let emulatorInitialized = false;
let emulatorRunning = false;

// Message handlers
const handlers = {
    /**
     * Initialize the emulator with images
     */
    async init(data) {
        try {
            postMessage({ type: 'status', status: 'loading_wasm' });
            
            // Load and instantiate WASM
            const go = new Go();
            const response = await fetch('assets/tinyemu.wasm');
            const result = await WebAssembly.instantiateStreaming(response, go.importObject);
            
            postMessage({ type: 'status', status: 'starting_go' });
            
            // Start Go runtime (non-blocking - runs in background)
            go.run(result.instance);
            
            // Wait for Go to initialize
            await new Promise(resolve => setTimeout(resolve, 100));
            
            // Initialize the emulator with console callback
            if (typeof tinyemuInit === 'function') {
                tinyemuInit((output) => {
                    postMessage({ type: 'output', data: output });
                });
            }
            
            emulatorInitialized = true;
            postMessage({ type: 'status', status: 'ready' });
            postMessage({ 
                type: 'init_complete',
                version: typeof tinyemuVersion === 'function' ? tinyemuVersion() : 'unknown'
            });
            
        } catch (error) {
            postMessage({ type: 'error', error: error.message });
        }
    },
    
    /**
     * Start the emulator
     */
    async start(data) {
        if (!emulatorInitialized) {
            postMessage({ type: 'error', error: 'Emulator not initialized' });
            return;
        }
        
        if (emulatorRunning) {
            postMessage({ type: 'error', error: 'Emulator already running' });
            return;
        }
        
        try {
            postMessage({ type: 'status', status: 'booting' });
            
            if (typeof tinyemuStart === 'function') {
                tinyemuStart();
            }
            
            emulatorRunning = true;
            postMessage({ type: 'status', status: 'running' });
            postMessage({ type: 'start_complete' });
            
        } catch (error) {
            postMessage({ type: 'error', error: error.message });
        }
    },
    
    /**
     * Stop the emulator
     */
    async stop(data) {
        if (!emulatorRunning) {
            return;
        }
        
        try {
            if (typeof tinyemuStop === 'function') {
                tinyemuStop();
            }
            
            emulatorRunning = false;
            postMessage({ type: 'status', status: 'stopped' });
            postMessage({ type: 'stop_complete' });
            
        } catch (error) {
            postMessage({ type: 'error', error: error.message });
        }
    },
    
    /**
     * Send input to the emulator console
     */
    input(data) {
        if (!emulatorRunning) {
            return;
        }
        
        if (typeof tinyemuSendInput === 'function') {
            tinyemuSendInput(data.text);
        }
    },
    
    /**
     * Send LLM response back to the emulator
     */
    llmResponse(data) {
        if (!emulatorRunning) {
            return;
        }
        
        // The LLM response is formatted as an OSC escape sequence
        // and written to the emulator's console input
        const { LLMProtocol } = self;
        if (LLMProtocol && typeof tinyemuSendInput === 'function') {
            const message = LLMProtocol.encodeMessage(data.messageType, data.payload);
            tinyemuSendInput(message);
        }
    },
    
    /**
     * Get emulator status
     */
    status(data) {
        postMessage({
            type: 'status_response',
            initialized: emulatorInitialized,
            running: emulatorRunning,
            version: typeof tinyemuVersion === 'function' ? tinyemuVersion() : 'unknown'
        });
    }
};

// Message listener
self.onmessage = async (event) => {
    const { type, ...data } = event.data;
    
    const handler = handlers[type];
    if (handler) {
        await handler(data);
    } else {
        postMessage({ type: 'error', error: `Unknown message type: ${type}` });
    }
};

// Error handler
self.onerror = (error) => {
    postMessage({ type: 'error', error: error.message || 'Unknown error' });
};

// Signal ready
postMessage({ type: 'worker_ready' });
