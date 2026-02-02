/**
 * Browser LLM Router
 * 
 * Routes LLM requests to either:
 * 1. WebLLM (local GPU inference via WebGPU) if available and model is downloaded
 * 2. Wllama (local CPU inference via WebAssembly) as fallback for browsers without WebGPU
 * 3. Cloud API (OpenAI, Anthropic, etc.) if API key is configured
 */

/**
 * LLM Provider configurations
 */
const PROVIDERS = {
    openai: {
        name: 'OpenAI',
        baseUrl: 'https://api.openai.com/v1',
        models: ['gpt-4o', 'gpt-4o-mini', 'gpt-4-turbo', 'gpt-3.5-turbo'],
        keyPrefix: 'sk-'
    },
    anthropic: {
        name: 'Anthropic',
        baseUrl: 'https://api.anthropic.com/v1',
        models: ['claude-3-5-sonnet-20241022', 'claude-3-5-haiku-20241022', 'claude-3-opus-20240229'],
        keyPrefix: 'sk-ant-'
    },
    openrouter: {
        name: 'OpenRouter',
        baseUrl: 'https://openrouter.ai/api/v1',
        models: [], // Supports many models
        keyPrefix: 'sk-or-'
    }
};

/**
 * WebLLM model configurations (WebGPU - fast)
 * Model IDs must match exactly what's in WebLLM's prebuiltAppConfig
 */
const WEBLLM_MODELS = [
    { id: 'Llama-3.2-1B-Instruct-q4f16_1-MLC', name: 'Llama 3.2 1B', size: '880MB', minVRAM: 1 },
    { id: 'Llama-3.2-3B-Instruct-q4f16_1-MLC', name: 'Llama 3.2 3B', size: '2.3GB', minVRAM: 3 },
    { id: 'SmolLM2-1.7B-Instruct-q4f16_1-MLC', name: 'SmolLM2 1.7B', size: '1.8GB', minVRAM: 2 },
    { id: 'SmolLM2-360M-Instruct-q4f32_1-MLC', name: 'SmolLM2 360M', size: '580MB', minVRAM: 1 },
    { id: 'Qwen2.5-1.5B-Instruct-q4f16_1-MLC', name: 'Qwen2.5 1.5B', size: '1.6GB', minVRAM: 2 },
    { id: 'Phi-3.5-mini-instruct-q4f16_1-MLC', name: 'Phi 3.5 Mini', size: '3.7GB', minVRAM: 4 },
    { id: 'TinyLlama-1.1B-Chat-v1.0-q4f32_1-MLC', name: 'TinyLlama 1.1B', size: '840MB', minVRAM: 1 }
];

/**
 * Wllama model configurations (WebAssembly - slower but universal)
 * These are smaller models suitable for CPU inference
 */
const WLLAMA_MODELS = [
    { 
        id: 'smollm2-360m-instruct-q8_0',
        name: 'SmolLM2 360M',
        size: '390MB',
        hfRepo: 'HuggingFaceTB/SmolLM2-360M-Instruct-GGUF',
        hfFile: 'smollm2-360m-instruct-q8_0.gguf',
        description: 'Tiny but capable. Good for simple tasks.'
    },
    { 
        id: 'smollm2-1.7b-instruct-q4_k_m',
        name: 'SmolLM2 1.7B',
        size: '1GB',
        hfRepo: 'HuggingFaceTB/SmolLM2-1.7B-Instruct-GGUF',
        hfFile: 'smollm2-1.7b-instruct-q4_k_m.gguf',
        description: 'Best balance of size and quality.'
    },
    { 
        id: 'qwen2.5-0.5b-instruct-q8_0',
        name: 'Qwen2.5 0.5B',
        size: '530MB',
        hfRepo: 'Qwen/Qwen2.5-0.5B-Instruct-GGUF',
        hfFile: 'qwen2.5-0.5b-instruct-q8_0.gguf',
        description: 'Fast responses, good for chat.'
    },
    { 
        id: 'tinyllama-1.1b-chat-v1.0-q4_k_m',
        name: 'TinyLlama 1.1B',
        size: '670MB',
        hfRepo: 'TheBloke/TinyLlama-1.1B-Chat-v1.0-GGUF',
        hfFile: 'tinyllama-1.1b-chat-v1.0.Q4_K_M.gguf',
        description: 'Classic small model, well-tested.'
    }
];

/**
 * Check if WebGPU is available
 */
async function checkWebGPU() {
    if (!navigator.gpu) {
        return { available: false, reason: 'WebGPU not supported in this browser' };
    }
    
    try {
        const adapter = await navigator.gpu.requestAdapter();
        if (!adapter) {
            return { available: false, reason: 'No GPU adapter found' };
        }
        
        // Try to get adapter info - may not be available in all browsers
        let info = {};
        if (adapter.requestAdapterInfo) {
            try {
                info = await adapter.requestAdapterInfo();
            } catch (e) {
                // requestAdapterInfo not supported (e.g., older Firefox)
                info = { vendor: 'Unknown', device: 'GPU detected' };
            }
        } else {
            info = { vendor: 'Unknown', device: 'GPU detected' };
        }
        
        return {
            available: true,
            vendor: info.vendor || 'Unknown',
            architecture: info.architecture,
            device: info.device
        };
    } catch (e) {
        return { available: false, reason: e.message };
    }
}

/**
 * Check if Wllama (WebAssembly) is available
 * This works in all modern browsers
 */
function checkWllama() {
    // WebAssembly is supported in all modern browsers
    const wasmSupported = typeof WebAssembly !== 'undefined';
    
    // SharedArrayBuffer is needed for multi-threading (optional but faster)
    // Requires COOP/COEP headers
    const hasSharedArrayBuffer = typeof SharedArrayBuffer !== 'undefined';
    
    return {
        available: wasmSupported,
        multiThreaded: hasSharedArrayBuffer,
        reason: wasmSupported ? null : 'WebAssembly not supported'
    };
}

/**
 * LLM Router class
 */
class LLMRouter {
    constructor(storage) {
        this.storage = storage; // OfflineStorage instance
        this.webllmEngine = null;
        this.wllamaInstance = null;
        this.webgpuInfo = null;
        this.wllamaInfo = null;
        this.currentModel = null;
        this.currentBackend = null; // 'webllm' or 'wllama'
        this.onProgress = null; // Callback for model loading progress
    }
    
    /**
     * Initialize the router
     */
    async init() {
        // Check WebGPU availability
        this.webgpuInfo = await checkWebGPU();
        
        // Check Wllama availability
        this.wllamaInfo = checkWllama();
        
        // Get configured providers
        this.providers = await this.storage.listApiKeyProviders();
        
        return this;
    }
    
    /**
     * Check if any LLM backend is available
     */
    isAvailable() {
        return this.webgpuInfo?.available || this.wllamaInfo?.available || this.providers.length > 0;
    }
    
    /**
     * Get the preferred local backend
     */
    getPreferredBackend() {
        if (this.webgpuInfo?.available) {
            return 'webllm';
        } else if (this.wllamaInfo?.available) {
            return 'wllama';
        }
        return null;
    }
    
    /**
     * Get available backends and their status
     */
    async getBackends() {
        const backends = [];
        
        // WebLLM backend (WebGPU)
        backends.push({
            type: 'webllm',
            name: 'Local GPU (WebLLM)',
            available: this.webgpuInfo?.available || false,
            reason: this.webgpuInfo?.reason,
            models: this.webgpuInfo?.available ? WEBLLM_MODELS : [],
            performance: 'fast'
        });
        
        // Wllama backend (WebAssembly)
        backends.push({
            type: 'wllama',
            name: 'Local CPU (Wllama)',
            available: this.wllamaInfo?.available || false,
            reason: this.wllamaInfo?.reason,
            models: this.wllamaInfo?.available ? WLLAMA_MODELS : [],
            performance: 'slower',
            multiThreaded: this.wllamaInfo?.multiThreaded || false
        });
        
        // Cloud backends
        for (const [id, config] of Object.entries(PROVIDERS)) {
            const hasKey = this.providers.includes(id);
            backends.push({
                type: 'cloud',
                id,
                name: config.name,
                available: hasKey,
                reason: hasKey ? null : 'API key not configured',
                models: config.models
            });
        }
        
        return backends;
    }
    
    /**
     * Load a WebLLM model (WebGPU)
     * @param {string} modelId Model ID to load
     * @param {function} onProgress Progress callback
     */
    async loadWebLLMModel(modelId, onProgress = null) {
        if (!this.webgpuInfo?.available) {
            throw new Error('WebGPU not available');
        }
        
        // Dynamically import WebLLM
        const { CreateMLCEngine } = await import('https://esm.run/@mlc-ai/web-llm');
        
        this.webllmEngine = await CreateMLCEngine(modelId, {
            initProgressCallback: (progress) => {
                if (onProgress) {
                    onProgress({
                        stage: progress.text,
                        progress: progress.progress
                    });
                }
                if (this.onProgress) {
                    this.onProgress({
                        stage: progress.text,
                        progress: progress.progress
                    });
                }
            }
        });
        
        this.currentModel = modelId;
        this.currentBackend = 'webllm';
    }
    
    /**
     * Load a Wllama model (WebAssembly)
     * @param {string} modelId Model ID to load
     * @param {function} onProgress Progress callback
     */
    async loadWllamaModel(modelId, onProgress = null) {
        if (!this.wllamaInfo?.available) {
            throw new Error('WebAssembly not available');
        }
        
        const modelConfig = WLLAMA_MODELS.find(m => m.id === modelId);
        if (!modelConfig) {
            throw new Error(`Unknown model: ${modelId}`);
        }
        
        // Dynamically import Wllama from CDN
        const { Wllama } = await import('https://esm.run/@wllama/wllama');
        
        // Import the CDN wasm paths helper
        const WasmFromCDN = (await import('https://esm.run/@wllama/wllama/esm/wasm-from-cdn.js')).default;
        
        // Create Wllama instance
        this.wllamaInstance = new Wllama(WasmFromCDN, {
            // Suppress debug messages
            logger: {
                debug: () => {},
                log: (...args) => console.log('[Wllama]', ...args),
                warn: (...args) => console.warn('[Wllama]', ...args),
                error: (...args) => console.error('[Wllama]', ...args)
            }
        });
        
        // Load model from Hugging Face
        const hfUrl = `https://huggingface.co/${modelConfig.hfRepo}/resolve/main/${modelConfig.hfFile}`;
        
        await this.wllamaInstance.loadModelFromUrl(hfUrl, {
            n_threads: this.wllamaInfo.multiThreaded ? navigator.hardwareConcurrency || 4 : 1,
            progressCallback: ({ loaded, total }) => {
                const progress = total > 0 ? loaded / total : 0;
                if (onProgress) {
                    onProgress({
                        stage: `Downloading ${modelConfig.name}...`,
                        progress
                    });
                }
                if (this.onProgress) {
                    this.onProgress({
                        stage: `Downloading ${modelConfig.name}...`,
                        progress
                    });
                }
            }
        });
        
        this.currentModel = modelId;
        this.currentBackend = 'wllama';
    }
    
    /**
     * Unload the current WebLLM model
     */
    async unloadWebLLMModel() {
        if (this.webllmEngine) {
            this.webllmEngine = null;
        }
        if (this.currentBackend === 'webllm') {
            this.currentModel = null;
            this.currentBackend = null;
        }
    }
    
    /**
     * Unload the current Wllama model
     */
    async unloadWllamaModel() {
        if (this.wllamaInstance) {
            try {
                await this.wllamaInstance.exit();
            } catch (e) {
                // Ignore cleanup errors
            }
            this.wllamaInstance = null;
        }
        if (this.currentBackend === 'wllama') {
            this.currentModel = null;
            this.currentBackend = null;
        }
    }
    
    /**
     * Unload any loaded local model
     */
    async unloadModel() {
        await this.unloadWebLLMModel();
        await this.unloadWllamaModel();
    }
    
    /**
     * Generate a response using the best available backend
     * @param {object} params Generation parameters
     * @param {function} onChunk Streaming callback
     * @param {AbortSignal} signal Abort signal
     */
    async generate(params, onChunk = null, signal = null) {
        const { model, messages, temperature = 0.7, maxTokens = 1000 } = params;
        
        // Try WebLLM first if model is loaded
        if (this.webllmEngine && this.currentBackend === 'webllm') {
            return this.generateWithWebLLM(messages, temperature, maxTokens, onChunk, signal);
        }
        
        // Try Wllama if model is loaded
        if (this.wllamaInstance && this.currentBackend === 'wllama') {
            return this.generateWithWllama(messages, temperature, maxTokens, onChunk, signal);
        }
        
        // Check if we have a configured local model that needs to be loaded
        const activeLocalModel = await this.storage.getConfig('activeLocalModel');
        if (activeLocalModel) {
            // Determine which backend to use based on model ID
            const isWebLLMModel = WEBLLM_MODELS.some(m => m.id === activeLocalModel);
            const isWllamaModel = WLLAMA_MODELS.some(m => m.id === activeLocalModel);
            
            if (isWebLLMModel && this.webgpuInfo?.available) {
                // Load WebLLM model
                await this.loadWebLLMModel(activeLocalModel, (progress) => {
                    // Show loading progress in status
                    console.log(`Loading model: ${progress.stage} (${Math.round(progress.progress * 100)}%)`);
                });
                return this.generateWithWebLLM(messages, temperature, maxTokens, onChunk, signal);
            } else if (isWllamaModel && this.wllamaInfo?.available) {
                // Load Wllama model
                await this.loadWllamaModel(activeLocalModel, (progress) => {
                    console.log(`Loading model: ${progress.stage} (${Math.round(progress.progress * 100)}%)`);
                });
                return this.generateWithWllama(messages, temperature, maxTokens, onChunk, signal);
            }
        }
        
        // Try cloud providers
        for (const provider of this.providers) {
            try {
                return await this.generateWithCloud(provider, model, messages, temperature, maxTokens, onChunk, signal);
            } catch (e) {
                console.warn(`Provider ${provider} failed:`, e);
                continue;
            }
        }
        
        throw new Error('No LLM backend available. Please download a local model or configure an API key in Settings.');
    }
    
    /**
     * Generate using WebLLM
     */
    async generateWithWebLLM(messages, temperature, maxTokens, onChunk, signal) {
        const response = await this.webllmEngine.chat.completions.create({
            messages,
            temperature,
            max_tokens: maxTokens,
            stream: true
        });
        
        let fullContent = '';
        
        for await (const chunk of response) {
            if (signal?.aborted) {
                throw new DOMException('Aborted', 'AbortError');
            }
            
            const delta = chunk.choices[0]?.delta?.content;
            if (delta) {
                fullContent += delta;
                if (onChunk) {
                    onChunk({ content: delta, done: false });
                }
            }
        }
        
        if (onChunk) {
            onChunk({ content: '', done: true });
        }
        
        return { content: fullContent };
    }
    
    /**
     * Generate using Wllama
     */
    async generateWithWllama(messages, temperature, maxTokens, onChunk, signal) {
        // Build prompt from messages
        // Most small models use ChatML or similar format
        let prompt = '';
        for (const msg of messages) {
            if (msg.role === 'system') {
                prompt += `<|im_start|>system\n${msg.content}<|im_end|>\n`;
            } else if (msg.role === 'user') {
                prompt += `<|im_start|>user\n${msg.content}<|im_end|>\n`;
            } else if (msg.role === 'assistant') {
                prompt += `<|im_start|>assistant\n${msg.content}<|im_end|>\n`;
            }
        }
        prompt += '<|im_start|>assistant\n';
        
        let fullContent = '';
        
        // Use createCompletion with streaming callback
        const result = await this.wllamaInstance.createCompletion(prompt, {
            nPredict: maxTokens,
            sampling: {
                temp: temperature,
                top_k: 40,
                top_p: 0.9
            },
            onNewToken: (token, piece, currentText, { abortSignal }) => {
                if (signal?.aborted) {
                    abortSignal();
                    return;
                }
                
                fullContent = currentText;
                if (onChunk) {
                    onChunk({ content: piece, done: false });
                }
            }
        });
        
        if (onChunk) {
            onChunk({ content: '', done: true });
        }
        
        // Clean up any trailing special tokens
        let cleanContent = result;
        const stopTokens = ['<|im_end|>', '<|endoftext|>', '</s>'];
        for (const token of stopTokens) {
            if (cleanContent.endsWith(token)) {
                cleanContent = cleanContent.slice(0, -token.length);
            }
        }
        
        return { content: cleanContent.trim() };
    }
    
    /**
     * Generate using cloud API
     */
    async generateWithCloud(provider, model, messages, temperature, maxTokens, onChunk, signal) {
        const apiKey = await this.storage.getApiKey(provider);
        if (!apiKey) {
            throw new Error(`No API key for ${provider}`);
        }
        
        const config = PROVIDERS[provider];
        if (!config) {
            throw new Error(`Unknown provider: ${provider}`);
        }
        
        // Use the model parameter or default to first model
        const useModel = model || config.models[0];
        
        if (provider === 'anthropic') {
            return this.generateAnthropic(apiKey, useModel, messages, temperature, maxTokens, onChunk, signal);
        } else {
            // OpenAI-compatible API (OpenAI, OpenRouter, etc.)
            return this.generateOpenAI(config.baseUrl, apiKey, useModel, messages, temperature, maxTokens, onChunk, signal);
        }
    }
    
    /**
     * Generate using OpenAI-compatible API
     */
    async generateOpenAI(baseUrl, apiKey, model, messages, temperature, maxTokens, onChunk, signal) {
        const response = await fetch(`${baseUrl}/chat/completions`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${apiKey}`
            },
            body: JSON.stringify({
                model,
                messages,
                temperature,
                max_tokens: maxTokens,
                stream: !!onChunk
            }),
            signal
        });
        
        if (!response.ok) {
            const error = await response.json().catch(() => ({}));
            throw new Error(error.error?.message || `API error: ${response.status}`);
        }
        
        if (onChunk) {
            return this.streamOpenAIResponse(response, onChunk);
        } else {
            const data = await response.json();
            return { content: data.choices[0]?.message?.content || '' };
        }
    }
    
    /**
     * Stream OpenAI response
     */
    async streamOpenAIResponse(response, onChunk) {
        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let fullContent = '';
        let buffer = '';
        
        while (true) {
            const { done, value } = await reader.read();
            if (done) break;
            
            buffer += decoder.decode(value, { stream: true });
            const lines = buffer.split('\n');
            buffer = lines.pop() || '';
            
            for (const line of lines) {
                if (line.startsWith('data: ')) {
                    const data = line.slice(6);
                    if (data === '[DONE]') continue;
                    
                    try {
                        const parsed = JSON.parse(data);
                        const delta = parsed.choices[0]?.delta?.content;
                        if (delta) {
                            fullContent += delta;
                            onChunk({ content: delta, done: false });
                        }
                    } catch (e) {
                        // Ignore parse errors
                    }
                }
            }
        }
        
        onChunk({ content: '', done: true });
        return { content: fullContent };
    }
    
    /**
     * Generate using Anthropic API
     */
    async generateAnthropic(apiKey, model, messages, temperature, maxTokens, onChunk, signal) {
        // Convert messages to Anthropic format
        const systemMessage = messages.find(m => m.role === 'system');
        const chatMessages = messages.filter(m => m.role !== 'system');
        
        const response = await fetch('https://api.anthropic.com/v1/messages', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'x-api-key': apiKey,
                'anthropic-version': '2023-06-01',
                'anthropic-dangerous-direct-browser-access': 'true'
            },
            body: JSON.stringify({
                model,
                max_tokens: maxTokens,
                system: systemMessage?.content,
                messages: chatMessages,
                temperature,
                stream: !!onChunk
            }),
            signal
        });
        
        if (!response.ok) {
            const error = await response.json().catch(() => ({}));
            throw new Error(error.error?.message || `API error: ${response.status}`);
        }
        
        if (onChunk) {
            return this.streamAnthropicResponse(response, onChunk);
        } else {
            const data = await response.json();
            return { content: data.content[0]?.text || '' };
        }
    }
    
    /**
     * Stream Anthropic response
     */
    async streamAnthropicResponse(response, onChunk) {
        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let fullContent = '';
        let buffer = '';
        
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
                        if (data.type === 'content_block_delta') {
                            const delta = data.delta?.text;
                            if (delta) {
                                fullContent += delta;
                                onChunk({ content: delta, done: false });
                            }
                        }
                    } catch (e) {
                        // Ignore parse errors
                    }
                }
            }
        }
        
        onChunk({ content: '', done: true });
        return { content: fullContent };
    }
}

// Export
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { LLMRouter, PROVIDERS, WEBLLM_MODELS, WLLAMA_MODELS, checkWebGPU, checkWllama };
} else if (typeof window !== 'undefined') {
    window.LLMRouter = LLMRouter;
    window.LLM_PROVIDERS = PROVIDERS;
    window.WEBLLM_MODELS = WEBLLM_MODELS;
    window.WLLAMA_MODELS = WLLAMA_MODELS;
    window.checkWebGPU = checkWebGPU;
    window.checkWllama = checkWllama;
}
