/**
 * Browser LLM Router
 * 
 * Routes LLM requests to either:
 * 1. WebLLM (local GPU inference) if model is downloaded
 * 2. Cloud API (OpenAI, Anthropic, etc.) if API key is configured
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
 * WebLLM model configurations
 */
const WEBLLM_MODELS = [
    { id: 'Llama-3.2-1B-Instruct-q4f16_1', size: '700MB', minVRAM: 2 },
    { id: 'Llama-3.2-3B-Instruct-q4f16_1', size: '1.8GB', minVRAM: 4 },
    { id: 'Phi-3.5-mini-instruct-q4f16_1', size: '2GB', minVRAM: 4 },
    { id: 'Qwen2.5-1.5B-Instruct-q4f16_1', size: '1GB', minVRAM: 3 },
    { id: 'Mistral-7B-Instruct-v0.3-q4f16_1', size: '4GB', minVRAM: 8 }
];

/**
 * Check if WebGPU is available
 */
async function checkWebGPU() {
    if (!navigator.gpu) {
        return { available: false, reason: 'WebGPU not supported' };
    }
    
    try {
        const adapter = await navigator.gpu.requestAdapter();
        if (!adapter) {
            return { available: false, reason: 'No GPU adapter found' };
        }
        
        const info = await adapter.requestAdapterInfo();
        return {
            available: true,
            vendor: info.vendor,
            architecture: info.architecture,
            device: info.device
        };
    } catch (e) {
        return { available: false, reason: e.message };
    }
}

/**
 * LLM Router class
 */
class LLMRouter {
    constructor(storage) {
        this.storage = storage; // OfflineStorage instance
        this.webllmEngine = null;
        this.webgpuInfo = null;
        this.currentModel = null;
        this.onProgress = null; // Callback for model loading progress
    }
    
    /**
     * Initialize the router
     */
    async init() {
        // Check WebGPU availability
        this.webgpuInfo = await checkWebGPU();
        
        // Get configured providers
        this.providers = await this.storage.listApiKeyProviders();
        
        return this;
    }
    
    /**
     * Check if any LLM backend is available
     */
    isAvailable() {
        return this.webgpuInfo?.available || this.providers.length > 0;
    }
    
    /**
     * Get available backends and their status
     */
    async getBackends() {
        const backends = [];
        
        // WebLLM backend
        backends.push({
            type: 'webllm',
            name: 'Local (WebLLM)',
            available: this.webgpuInfo?.available || false,
            reason: this.webgpuInfo?.reason,
            models: this.webgpuInfo?.available ? WEBLLM_MODELS : []
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
     * Load a WebLLM model
     * @param {string} modelId Model ID to load
     * @param {function} onProgress Progress callback
     */
    async loadWebLLMModel(modelId, onProgress = null) {
        if (!this.webgpuInfo?.available) {
            throw new Error('WebGPU not available');
        }
        
        // Dynamically import WebLLM
        const { CreateMLCEngine } = await import('@mlc-ai/web-llm');
        
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
    }
    
    /**
     * Unload the current WebLLM model
     */
    async unloadWebLLMModel() {
        if (this.webllmEngine) {
            // WebLLM doesn't have an explicit unload, but we can null it
            this.webllmEngine = null;
            this.currentModel = null;
        }
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
        if (this.webllmEngine && this.currentModel) {
            return this.generateWithWebLLM(messages, temperature, maxTokens, onChunk, signal);
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
        
        throw new Error('No LLM backend available');
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
    module.exports = { LLMRouter, PROVIDERS, WEBLLM_MODELS, checkWebGPU };
} else if (typeof window !== 'undefined') {
    window.LLMRouter = LLMRouter;
    window.LLM_PROVIDERS = PROVIDERS;
    window.WEBLLM_MODELS = WEBLLM_MODELS;
    window.checkWebGPU = checkWebGPU;
}
