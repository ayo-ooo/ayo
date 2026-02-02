/**
 * LLM RPC Protocol for Console-based communication
 * 
 * Uses OSC (Operating System Command) escape sequences to frame JSON messages
 * within the console I/O stream.
 * 
 * Format: ESC ] AYO ; <json> BEL
 *         \x1B ] AYO ; {...} \x07
 * 
 * This allows LLM requests/responses to be multiplexed with regular terminal I/O.
 */

// Protocol constants
const ESC = '\x1B';
const BEL = '\x07';
const OSC_START = `${ESC}]AYO;`;
const OSC_END = BEL;

// Message types
const MessageType = {
    // Requests (guest -> host)
    LLM_REQUEST: 'llm:request',
    LLM_CANCEL: 'llm:cancel',
    
    // Responses (host -> guest)
    LLM_RESPONSE: 'llm:response',
    LLM_CHUNK: 'llm:chunk',
    LLM_ERROR: 'llm:error',
    LLM_DONE: 'llm:done',
    
    // Filesystem (both directions)
    FS_READ: 'fs:read',
    FS_WRITE: 'fs:write',
    FS_LIST: 'fs:list',
    FS_RESPONSE: 'fs:response',
    
    // System
    PING: 'ping',
    PONG: 'pong',
};

/**
 * Encode a message as an OSC escape sequence
 * @param {string} type Message type
 * @param {object} payload Message payload
 * @returns {string} Encoded message
 */
function encodeMessage(type, payload = {}) {
    const message = JSON.stringify({ type, ...payload, ts: Date.now() });
    return `${OSC_START}${message}${OSC_END}`;
}

/**
 * Parse messages from a console output stream
 * Returns both the regular output and any parsed messages
 */
class ProtocolParser {
    constructor() {
        this.buffer = '';
        this.pendingRequests = new Map(); // id -> { resolve, reject }
        this.nextId = 1;
    }
    
    /**
     * Feed data into the parser
     * @param {string} data Raw console output
     * @returns {{ output: string, messages: object[] }} Parsed result
     */
    parse(data) {
        this.buffer += data;
        const messages = [];
        let output = '';
        
        while (true) {
            // Find start of OSC sequence
            const startIdx = this.buffer.indexOf(OSC_START);
            
            if (startIdx === -1) {
                // No more messages, everything is regular output
                output += this.buffer;
                this.buffer = '';
                break;
            }
            
            // Output everything before the message
            output += this.buffer.substring(0, startIdx);
            
            // Find end of message
            const endIdx = this.buffer.indexOf(OSC_END, startIdx);
            if (endIdx === -1) {
                // Incomplete message, keep in buffer
                this.buffer = this.buffer.substring(startIdx);
                break;
            }
            
            // Extract and parse message
            const jsonStr = this.buffer.substring(startIdx + OSC_START.length, endIdx);
            try {
                const message = JSON.parse(jsonStr);
                messages.push(message);
                
                // Handle response to pending request
                if (message.id && this.pendingRequests.has(message.id)) {
                    const { resolve, reject } = this.pendingRequests.get(message.id);
                    if (message.type === MessageType.LLM_ERROR) {
                        reject(new Error(message.error || 'Unknown error'));
                    } else if (message.type === MessageType.LLM_DONE) {
                        resolve(message);
                        this.pendingRequests.delete(message.id);
                    }
                    // LLM_CHUNK messages are handled via callbacks, not promises
                }
            } catch (e) {
                console.error('Failed to parse OSC message:', jsonStr, e);
            }
            
            // Remove processed message from buffer
            this.buffer = this.buffer.substring(endIdx + OSC_END.length);
        }
        
        return { output, messages };
    }
    
    /**
     * Generate a unique request ID
     * @returns {number} Unique ID
     */
    generateId() {
        return this.nextId++;
    }
    
    /**
     * Register a pending request
     * @param {number} id Request ID
     * @returns {Promise} Promise that resolves when response is received
     */
    registerRequest(id) {
        return new Promise((resolve, reject) => {
            this.pendingRequests.set(id, { resolve, reject });
        });
    }
    
    /**
     * Cancel a pending request
     * @param {number} id Request ID
     */
    cancelRequest(id) {
        this.pendingRequests.delete(id);
    }
}

/**
 * LLM RPC Client for use in the guest VM
 * Sends requests via console escape sequences
 */
class LLMRPCClient {
    constructor(consoleWrite) {
        this.write = consoleWrite; // Function to write to console
        this.parser = new ProtocolParser();
    }
    
    /**
     * Send an LLM generation request
     * @param {object} params Request parameters
     * @param {function} onChunk Callback for streaming chunks
     * @returns {Promise} Resolves when generation is complete
     */
    async generate(params, onChunk = null) {
        const id = this.parser.generateId();
        const message = encodeMessage(MessageType.LLM_REQUEST, {
            id,
            method: 'generate',
            params
        });
        
        this.write(message);
        
        // The response handling is done by the parser
        // which is fed console input from the host
        return this.parser.registerRequest(id);
    }
    
    /**
     * Cancel an ongoing generation
     * @param {number} id Request ID to cancel
     */
    cancel(id) {
        const message = encodeMessage(MessageType.LLM_CANCEL, { id });
        this.write(message);
        this.parser.cancelRequest(id);
    }
    
    /**
     * Send a ping and wait for pong
     * @returns {Promise<number>} Round-trip time in ms
     */
    async ping() {
        const id = this.parser.generateId();
        const start = Date.now();
        const message = encodeMessage(MessageType.PING, { id });
        
        this.write(message);
        await this.parser.registerRequest(id);
        
        return Date.now() - start;
    }
    
    /**
     * Process incoming console data
     * @param {string} data Raw console input
     * @returns {{ output: string, messages: object[] }} Parsed result
     */
    processInput(data) {
        return this.parser.parse(data);
    }
}

/**
 * LLM RPC Host for use in the browser
 * Handles requests from the guest VM
 */
class LLMRPCHost {
    constructor(llmRouter, consoleWrite) {
        this.router = llmRouter; // LLMRouter instance
        this.write = consoleWrite; // Function to write to VM console
        this.parser = new ProtocolParser();
        this.activeRequests = new Map(); // id -> AbortController
    }
    
    /**
     * Process console output from VM
     * @param {string} data Raw console output
     * @returns {string} Regular output (with messages removed)
     */
    processOutput(data) {
        const { output, messages } = this.parser.parse(data);
        
        for (const message of messages) {
            this.handleMessage(message);
        }
        
        return output;
    }
    
    /**
     * Handle a message from the guest
     * @param {object} message Parsed message
     */
    async handleMessage(message) {
        switch (message.type) {
            case MessageType.LLM_REQUEST:
                await this.handleLLMRequest(message);
                break;
            case MessageType.LLM_CANCEL:
                this.handleLLMCancel(message);
                break;
            case MessageType.PING:
                this.handlePing(message);
                break;
            default:
                console.warn('Unknown message type:', message.type);
        }
    }
    
    /**
     * Handle LLM generation request
     */
    async handleLLMRequest(message) {
        const { id, params } = message;
        const controller = new AbortController();
        this.activeRequests.set(id, controller);
        
        try {
            await this.router.generate(
                params,
                // onChunk callback
                (chunk) => {
                    const response = encodeMessage(MessageType.LLM_CHUNK, {
                        id,
                        chunk: chunk.content,
                        done: false
                    });
                    this.write(response);
                },
                controller.signal
            );
            
            // Send completion message
            const response = encodeMessage(MessageType.LLM_DONE, { id });
            this.write(response);
            
        } catch (error) {
            if (error.name === 'AbortError') {
                return; // Cancelled, no response needed
            }
            const response = encodeMessage(MessageType.LLM_ERROR, {
                id,
                error: error.message
            });
            this.write(response);
        } finally {
            this.activeRequests.delete(id);
        }
    }
    
    /**
     * Handle cancellation request
     */
    handleLLMCancel(message) {
        const { id } = message;
        const controller = this.activeRequests.get(id);
        if (controller) {
            controller.abort();
            this.activeRequests.delete(id);
        }
    }
    
    /**
     * Handle ping request
     */
    handlePing(message) {
        const response = encodeMessage(MessageType.PONG, { id: message.id });
        this.write(response);
    }
}

// Export
if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        MessageType,
        encodeMessage,
        ProtocolParser,
        LLMRPCClient,
        LLMRPCHost
    };
} else if (typeof window !== 'undefined') {
    window.LLMProtocol = {
        MessageType,
        encodeMessage,
        ProtocolParser,
        LLMRPCClient,
        LLMRPCHost
    };
}
