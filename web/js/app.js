/**
 * Ayo Offline Web Client - Main Application
 * 
 * Handles mode switching between Connected and Offline modes,
 * and integrates all components.
 */

// Application modes
const AppMode = {
    CONNECTED: 'connected',
    OFFLINE: 'offline',
    DETECTING: 'detecting'
};

// Tab IDs
const Tab = {
    CHAT: 'chat',
    TERMINAL: 'terminal',
    FILES: 'files',
    SETTINGS: 'settings'
};

/**
 * Main Application Class
 */
class AyoApp {
    constructor() {
        this.mode = AppMode.DETECTING;
        this.activeTab = Tab.CHAT;
        this.storage = null;
        this.llmRouter = null;
        this.emulator = null;
        this.terminal = null;
        this.serverUrl = null;
        this.isServerAvailable = false;
        this.emulatorInitPromise = null;
        
        // UI callbacks
        this.onModeChange = () => {};
        this.onTabChange = () => {};
        this.onStatusChange = () => {};
    }
    
    /**
     * Initialize the application
     */
    async init() {
        // Initialize storage
        this.storage = await new OfflineStorage().init();
        
        // Get saved encryption key or generate one
        let encKey = await this.storage.getConfig('encryptionKey');
        if (!encKey) {
            // Generate a device-specific key
            encKey = crypto.randomUUID();
            await this.storage.setConfig('encryptionKey', encKey);
        }
        this.storage.setEncryptionKey(encKey);
        
        // Check for server URL in query params or localStorage
        const urlParams = new URLSearchParams(window.location.search);
        this.serverUrl = urlParams.get('server') || localStorage.getItem('ayoServerUrl');
        
        // Detect mode
        await this.detectMode();
        
        // Initialize LLM router for offline mode
        if (this.mode === AppMode.OFFLINE) {
            await this.initOfflineMode();
        }
        
        return this;
    }
    
    /**
     * Detect whether we're in connected or offline mode
     */
    async detectMode() {
        this.mode = AppMode.DETECTING;
        this.onModeChange(this.mode);
        
        // Try to connect to server if URL is set
        if (this.serverUrl) {
            try {
                const response = await fetch(`${this.serverUrl}/health`, {
                    method: 'GET',
                    mode: 'cors',
                    signal: AbortSignal.timeout(5000)
                });
                
                if (response.ok) {
                    this.isServerAvailable = true;
                    this.mode = AppMode.CONNECTED;
                    localStorage.setItem('ayoServerUrl', this.serverUrl);
                    this.onModeChange(this.mode);
                    return;
                }
            } catch (e) {
                console.log('Server not available:', e.message);
            }
        }
        
        // Fall back to offline mode
        this.isServerAvailable = false;
        this.mode = AppMode.OFFLINE;
        this.onModeChange(this.mode);
    }
    
    /**
     * Initialize offline mode components
     */
    async initOfflineMode() {
        // Initialize LLM router
        this.llmRouter = await new LLMRouter(this.storage).init();
        
        // Check if we have any LLM backend available
        const backends = await this.llmRouter.getBackends();
        const hasBackend = backends.some(b => b.available);
        
        if (!hasBackend) {
            this.onStatusChange({
                type: 'warning',
                message: 'No LLM backend available. Configure API keys or download a WebLLM model in Settings.'
            });
        }
    }
    
    /**
     * Initialize emulator (lazy, only when needed)
     */
    async initEmulator() {
        if (this.emulator) {
            return this.emulator;
        }
        
        // Prevent concurrent initialization
        if (this.emulatorInitPromise) {
            return this.emulatorInitPromise;
        }
        
        this.emulatorInitPromise = (async () => {
            this.emulator = await new EmulatorController({
                onOutput: (text) => this.handleEmulatorOutput(text),
                onStatus: (status) => this.handleEmulatorStatus(status),
                onError: (error) => this.handleEmulatorError(error)
            }).init();
            
            // Set up LLM host for handling requests from VM
            if (this.llmRouter) {
                const llmHost = new LLMProtocol.LLMRPCHost(
                    this.llmRouter,
                    (message) => this.emulator.sendInput(message)
                );
                this.emulator.setLLMHost(llmHost);
            }
            
            return this.emulator;
        })();
        
        return this.emulatorInitPromise;
    }
    
    /**
     * Initialize terminal with xterm.js
     */
    async initTerminal(container) {
        if (this.terminal) {
            // Just reattach to new container
            this.terminal.fit();
            return this.terminal;
        }
        
        // Initialize emulator first if needed
        const emulator = await this.initEmulator();
        
        // Create terminal manager
        this.terminal = new TerminalManager({
            onReady: () => {
                this.onStatusChange({ type: 'success', message: 'Terminal ready' });
            },
            onError: (error) => {
                this.onStatusChange({ type: 'error', message: error });
            }
        });
        
        // Initialize with container and emulator
        await this.terminal.init(container, emulator);
        
        // Show boot message
        this.terminal.showBootMessage();
        
        return this.terminal;
    }
    
    /**
     * Start the VM and connect terminal
     */
    async startVM() {
        if (!this.emulator) {
            throw new Error('Emulator not initialized');
        }
        
        this.onStatusChange({ type: 'info', message: 'Starting VM...' });
        
        try {
            await this.emulator.start();
            this.onStatusChange({ type: 'success', message: 'VM running' });
            
            if (this.terminal) {
                this.terminal.showReadyMessage();
            }
        } catch (error) {
            this.onStatusChange({ type: 'error', message: `VM start failed: ${error.message}` });
            throw error;
        }
    }
    
    /**
     * Handle console output from emulator
     */
    handleEmulatorOutput(text) {
        // Dispatch to appropriate handler based on active tab
        if (this.activeTab === Tab.TERMINAL && this.onTerminalOutput) {
            this.onTerminalOutput(text);
        }
        // Chat tab may also need to see output during tool execution
        if (this.onChatOutput) {
            this.onChatOutput(text);
        }
    }
    
    /**
     * Handle emulator status changes
     */
    handleEmulatorStatus(status) {
        this.onStatusChange({
            type: 'info',
            message: `Emulator: ${status}`
        });
    }
    
    /**
     * Handle emulator errors
     */
    handleEmulatorError(error) {
        this.onStatusChange({
            type: 'error',
            message: `Emulator error: ${error}`
        });
    }
    
    /**
     * Switch to a different mode
     */
    async switchMode(newMode) {
        if (newMode === this.mode) return;
        
        // Clean up current mode
        if (this.mode === AppMode.OFFLINE && this.emulator) {
            await this.emulator.stop();
        }
        
        this.mode = newMode;
        this.onModeChange(newMode);
        
        // Initialize new mode
        if (newMode === AppMode.OFFLINE) {
            await this.initOfflineMode();
        }
    }
    
    /**
     * Switch active tab
     */
    switchTab(tabId) {
        if (tabId === this.activeTab) return;
        
        this.activeTab = tabId;
        this.onTabChange(tabId);
        
        // Initialize emulator if switching to terminal and in offline mode
        if (tabId === Tab.TERMINAL && this.mode === AppMode.OFFLINE) {
            this.initEmulator().catch(e => {
                this.onStatusChange({
                    type: 'error',
                    message: `Failed to initialize emulator: ${e.message}`
                });
            });
        }
    }
    
    /**
     * Send a chat message
     * @param {string} message - User message
     * @param {string} agentHandle - Agent handle
     * @param {function} onChunk - Streaming callback
     * @param {AbortSignal} signal - Abort signal
     */
    async sendChatMessage(message, agentHandle = '@ayo', onChunk = null, signal = null) {
        if (this.mode === AppMode.CONNECTED) {
            return this.sendConnectedMessage(message, agentHandle, onChunk, signal);
        } else {
            return this.sendOfflineMessage(message, agentHandle, onChunk, signal);
        }
    }
    
    /**
     * Send message in connected mode (via server)
     */
    async sendConnectedMessage(message, agentHandle, onChunk = null, signal = null) {
        // Use SSE for streaming response
        const response = await fetch(`${this.serverUrl}/chat`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                message,
                agent: agentHandle,
                session_id: await this.getOrCreateSessionId()
            }),
            signal
        });
        
        if (!response.ok) {
            throw new Error(`Server error: ${response.status}`);
        }
        
        // Handle SSE streaming
        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let buffer = '';
        let fullContent = '';
        
        while (true) {
            const { done, value } = await reader.read();
            if (done) break;
            
            buffer += decoder.decode(value, { stream: true });
            const lines = buffer.split('\n');
            buffer = lines.pop() || '';
            
            for (const line of lines) {
                if (line.startsWith('data: ')) {
                    try {
                        const data = JSON.parse(line.slice(6));
                        if (data.content) {
                            fullContent += data.content;
                            if (onChunk) {
                                onChunk({ content: data.content, done: false });
                            }
                        }
                    } catch (e) {
                        // Ignore parse errors
                    }
                }
            }
        }
        
        if (onChunk) {
            onChunk({ content: '', done: true });
        }
        
        return { content: fullContent };
    }
    
    /**
     * Send message in offline mode (via LLM router)
     */
    async sendOfflineMessage(message, agentHandle, onChunk = null, signal = null) {
        // Ensure LLM router is initialized
        if (!this.llmRouter) {
            await this.initOfflineMode();
        }
        
        // Build messages array
        const messages = [
            { role: 'system', content: this.getSystemPrompt(agentHandle) },
            { role: 'user', content: message }
        ];
        
        // Generate response
        const result = await this.llmRouter.generate(
            { messages, temperature: 0.7, maxTokens: 2000 },
            onChunk,
            signal
        );
        
        return result;
    }
    
    /**
     * Get system prompt for an agent
     */
    getSystemPrompt(agentHandle) {
        // Simplified system prompt for offline mode
        return `You are ${agentHandle}, a helpful AI assistant running locally in the browser.
You can help with coding, answering questions, and general tasks.
When you need to execute commands, describe what you would do.`;
    }
    
    /**
     * Get or create a session ID
     */
    async getOrCreateSessionId() {
        let sessionId = sessionStorage.getItem('ayoSessionId');
        if (!sessionId) {
            sessionId = crypto.randomUUID();
            sessionStorage.setItem('ayoSessionId', sessionId);
        }
        return sessionId;
    }
    
    /**
     * Get storage statistics
     */
    async getStorageStats() {
        return this.storage.getStorageUsage();
    }
    
    /**
     * Clear all offline data
     */
    async clearOfflineData() {
        await this.storage.clearAll();
        this.onStatusChange({
            type: 'success',
            message: 'Offline data cleared'
        });
    }
}

// Export
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { AyoApp, AppMode, Tab };
} else if (typeof window !== 'undefined') {
    window.AyoApp = AyoApp;
    window.AppMode = AppMode;
    window.Tab = Tab;
}
