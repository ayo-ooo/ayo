/**
 * Emulator Controller
 * 
 * Main thread interface for controlling the TinyEMU Web Worker.
 * Handles message passing and provides a clean API for the UI.
 */

/**
 * EmulatorController class
 */
class EmulatorController {
    constructor(options = {}) {
        this.worker = null;
        this.llmHost = null;
        this.onOutput = options.onOutput || (() => {});
        this.onStatus = options.onStatus || (() => {});
        this.onError = options.onError || (() => {});
        this.onLLMRequest = options.onLLMRequest || (() => {});
        this.pendingCallbacks = new Map();
        this.nextCallbackId = 1;
        this.status = 'not_initialized';
        this.version = null;
    }
    
    /**
     * Initialize the emulator
     */
    async init() {
        return new Promise((resolve, reject) => {
            try {
                // Create Web Worker
                this.worker = new Worker('worker.js');
                
                // Set up message handler
                this.worker.onmessage = (event) => this.handleMessage(event.data);
                
                // Set up error handler
                this.worker.onerror = (error) => {
                    this.onError(error.message || 'Worker error');
                    reject(error);
                };
                
                // Wait for worker ready, then send init
                const checkReady = (data) => {
                    if (data.type === 'worker_ready') {
                        // Worker is ready, now initialize emulator
                        this.worker.postMessage({ type: 'init' });
                    } else if (data.type === 'init_complete') {
                        this.version = data.version;
                        this.status = 'ready';
                        resolve(this);
                    } else if (data.type === 'error') {
                        reject(new Error(data.error));
                    }
                };
                
                // Temporarily override message handler for init
                const originalHandler = this.handleMessage.bind(this);
                this.handleMessage = (data) => {
                    checkReady(data);
                    originalHandler(data);
                };
                
            } catch (error) {
                reject(error);
            }
        });
    }
    
    /**
     * Handle messages from the worker
     */
    handleMessage(data) {
        switch (data.type) {
            case 'output':
                this.handleOutput(data.data);
                break;
                
            case 'status':
                this.status = data.status;
                this.onStatus(data.status);
                break;
                
            case 'error':
                this.onError(data.error);
                break;
                
            case 'init_complete':
                this.version = data.version;
                break;
                
            case 'start_complete':
            case 'stop_complete':
                // These are handled by promises in start()/stop()
                const callback = this.pendingCallbacks.get(data.type);
                if (callback) {
                    callback.resolve();
                    this.pendingCallbacks.delete(data.type);
                }
                break;
                
            case 'status_response':
                const statusCallback = this.pendingCallbacks.get('status');
                if (statusCallback) {
                    statusCallback.resolve(data);
                    this.pendingCallbacks.delete('status');
                }
                break;
        }
    }
    
    /**
     * Handle console output from the emulator
     * Parses LLM protocol messages and passes regular output to callback
     */
    handleOutput(output) {
        if (this.llmHost) {
            // Parse for LLM protocol messages
            const regularOutput = this.llmHost.processOutput(output);
            if (regularOutput) {
                this.onOutput(regularOutput);
            }
        } else {
            this.onOutput(output);
        }
    }
    
    /**
     * Set the LLM host for handling LLM requests
     */
    setLLMHost(llmHost) {
        this.llmHost = llmHost;
    }
    
    /**
     * Start the emulator
     */
    async start() {
        return new Promise((resolve, reject) => {
            this.pendingCallbacks.set('start_complete', { resolve, reject });
            this.worker.postMessage({ type: 'start' });
            
            // Timeout after 60 seconds (boot can take a while)
            setTimeout(() => {
                if (this.pendingCallbacks.has('start_complete')) {
                    this.pendingCallbacks.delete('start_complete');
                    reject(new Error('Start timeout'));
                }
            }, 60000);
        });
    }
    
    /**
     * Stop the emulator
     */
    async stop() {
        return new Promise((resolve, reject) => {
            this.pendingCallbacks.set('stop_complete', { resolve, reject });
            this.worker.postMessage({ type: 'stop' });
            
            setTimeout(() => {
                if (this.pendingCallbacks.has('stop_complete')) {
                    this.pendingCallbacks.delete('stop_complete');
                    resolve(); // Don't reject on stop timeout
                }
            }, 5000);
        });
    }
    
    /**
     * Send input to the emulator console
     */
    sendInput(text) {
        if (this.worker) {
            this.worker.postMessage({ type: 'input', text });
        }
    }
    
    /**
     * Send LLM response to the emulator
     */
    sendLLMResponse(messageType, payload) {
        if (this.worker) {
            this.worker.postMessage({ 
                type: 'llmResponse',
                messageType,
                payload
            });
        }
    }
    
    /**
     * Get current status
     */
    async getStatus() {
        return new Promise((resolve) => {
            this.pendingCallbacks.set('status', { resolve });
            this.worker.postMessage({ type: 'status' });
            
            setTimeout(() => {
                if (this.pendingCallbacks.has('status')) {
                    this.pendingCallbacks.delete('status');
                    resolve({ 
                        initialized: false, 
                        running: false, 
                        version: this.version 
                    });
                }
            }, 1000);
        });
    }
    
    /**
     * Terminate the worker
     */
    terminate() {
        if (this.worker) {
            this.worker.terminate();
            this.worker = null;
            this.status = 'terminated';
        }
    }
    
    /**
     * Check if emulator is running
     */
    isRunning() {
        return this.status === 'running';
    }
    
    /**
     * Check if emulator is ready
     */
    isReady() {
        return this.status === 'ready' || this.status === 'running' || this.status === 'stopped';
    }
}

// Export
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { EmulatorController };
} else if (typeof window !== 'undefined') {
    window.EmulatorController = EmulatorController;
}
