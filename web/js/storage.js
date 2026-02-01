/**
 * IndexedDB Storage Layer for Ayo Offline Web Client
 * 
 * Provides persistent storage for:
 * - API keys and configuration
 * - WebLLM model cache
 * - VM filesystem overlay
 * - Offline sessions
 * - Cached assets (WASM, rootfs)
 */

const DB_NAME = 'ayo-offline';
const DB_VERSION = 1;

/**
 * Store definitions
 */
const STORES = {
    config: 'config',       // API keys, preferences
    models: 'models',       // WebLLM model files
    filesystem: 'filesystem', // VM filesystem overlay
    sessions: 'sessions',   // Chat sessions
    assets: 'assets'        // Cached WASM/rootfs
};

/**
 * Open the database, creating stores if needed
 */
function openDB() {
    return new Promise((resolve, reject) => {
        const request = indexedDB.open(DB_NAME, DB_VERSION);
        
        request.onerror = () => reject(request.error);
        request.onsuccess = () => resolve(request.result);
        
        request.onupgradeneeded = (event) => {
            const db = event.target.result;
            
            // Config store
            if (!db.objectStoreNames.contains(STORES.config)) {
                db.createObjectStore(STORES.config, { keyPath: 'key' });
            }
            
            // Models store
            if (!db.objectStoreNames.contains(STORES.models)) {
                const modelStore = db.createObjectStore(STORES.models, { keyPath: 'id' });
                modelStore.createIndex('downloadedAt', 'downloadedAt');
            }
            
            // Filesystem store
            if (!db.objectStoreNames.contains(STORES.filesystem)) {
                const fsStore = db.createObjectStore(STORES.filesystem, { keyPath: 'path' });
                fsStore.createIndex('mtime', 'mtime');
            }
            
            // Sessions store
            if (!db.objectStoreNames.contains(STORES.sessions)) {
                const sessionStore = db.createObjectStore(STORES.sessions, { keyPath: 'id' });
                sessionStore.createIndex('agentHandle', 'agentHandle');
                sessionStore.createIndex('createdAt', 'createdAt');
            }
            
            // Assets store
            if (!db.objectStoreNames.contains(STORES.assets)) {
                db.createObjectStore(STORES.assets, { keyPath: 'url' });
            }
        };
    });
}

/**
 * Encrypt a value using Web Crypto API
 * Uses PBKDF2 for key derivation and AES-GCM for encryption
 */
async function encrypt(plaintext, password) {
    const enc = new TextEncoder();
    const salt = crypto.getRandomValues(new Uint8Array(16));
    const iv = crypto.getRandomValues(new Uint8Array(12));
    
    const keyMaterial = await crypto.subtle.importKey(
        'raw',
        enc.encode(password),
        'PBKDF2',
        false,
        ['deriveKey']
    );
    
    const key = await crypto.subtle.deriveKey(
        { name: 'PBKDF2', salt, iterations: 100000, hash: 'SHA-256' },
        keyMaterial,
        { name: 'AES-GCM', length: 256 },
        false,
        ['encrypt']
    );
    
    const ciphertext = await crypto.subtle.encrypt(
        { name: 'AES-GCM', iv },
        key,
        enc.encode(plaintext)
    );
    
    return {
        salt: Array.from(salt),
        iv: Array.from(iv),
        ciphertext: Array.from(new Uint8Array(ciphertext))
    };
}

/**
 * Decrypt a value using Web Crypto API
 */
async function decrypt(encrypted, password) {
    const enc = new TextEncoder();
    const dec = new TextDecoder();
    
    const salt = new Uint8Array(encrypted.salt);
    const iv = new Uint8Array(encrypted.iv);
    const ciphertext = new Uint8Array(encrypted.ciphertext);
    
    const keyMaterial = await crypto.subtle.importKey(
        'raw',
        enc.encode(password),
        'PBKDF2',
        false,
        ['deriveKey']
    );
    
    const key = await crypto.subtle.deriveKey(
        { name: 'PBKDF2', salt, iterations: 100000, hash: 'SHA-256' },
        keyMaterial,
        { name: 'AES-GCM', length: 256 },
        false,
        ['decrypt']
    );
    
    const plaintext = await crypto.subtle.decrypt(
        { name: 'AES-GCM', iv },
        key,
        ciphertext
    );
    
    return dec.decode(plaintext);
}

/**
 * OfflineStorage class - main API
 */
class OfflineStorage {
    constructor() {
        this.db = null;
        this.encryptionKey = null; // Set via setEncryptionKey()
    }
    
    /**
     * Initialize the storage
     */
    async init() {
        this.db = await openDB();
        return this;
    }
    
    /**
     * Set the encryption key for API key storage
     * Should be derived from user password or device-specific key
     */
    setEncryptionKey(key) {
        this.encryptionKey = key;
    }
    
    // ========== Config Operations ==========
    
    /**
     * Get a config value
     */
    async getConfig(key) {
        const tx = this.db.transaction(STORES.config, 'readonly');
        const store = tx.objectStore(STORES.config);
        return new Promise((resolve, reject) => {
            const request = store.get(key);
            request.onsuccess = () => resolve(request.result?.value);
            request.onerror = () => reject(request.error);
        });
    }
    
    /**
     * Set a config value
     */
    async setConfig(key, value) {
        const tx = this.db.transaction(STORES.config, 'readwrite');
        const store = tx.objectStore(STORES.config);
        return new Promise((resolve, reject) => {
            const request = store.put({ key, value, updatedAt: Date.now() });
            request.onsuccess = () => resolve();
            request.onerror = () => reject(request.error);
        });
    }
    
    /**
     * Delete a config value
     */
    async deleteConfig(key) {
        const tx = this.db.transaction(STORES.config, 'readwrite');
        const store = tx.objectStore(STORES.config);
        return new Promise((resolve, reject) => {
            const request = store.delete(key);
            request.onsuccess = () => resolve();
            request.onerror = () => reject(request.error);
        });
    }
    
    // ========== API Key Operations (Encrypted) ==========
    
    /**
     * Store an API key (encrypted)
     */
    async setApiKey(provider, apiKey) {
        if (!this.encryptionKey) {
            throw new Error('Encryption key not set');
        }
        const encrypted = await encrypt(apiKey, this.encryptionKey);
        await this.setConfig(`apikey:${provider}`, encrypted);
    }
    
    /**
     * Get an API key (decrypted)
     */
    async getApiKey(provider) {
        if (!this.encryptionKey) {
            throw new Error('Encryption key not set');
        }
        const encrypted = await this.getConfig(`apikey:${provider}`);
        if (!encrypted) return null;
        return decrypt(encrypted, this.encryptionKey);
    }
    
    /**
     * Delete an API key
     */
    async deleteApiKey(provider) {
        await this.deleteConfig(`apikey:${provider}`);
    }
    
    /**
     * List all configured API key providers
     */
    async listApiKeyProviders() {
        const tx = this.db.transaction(STORES.config, 'readonly');
        const store = tx.objectStore(STORES.config);
        return new Promise((resolve, reject) => {
            const request = store.getAllKeys();
            request.onsuccess = () => {
                const providers = request.result
                    .filter(k => k.startsWith('apikey:'))
                    .map(k => k.replace('apikey:', ''));
                resolve(providers);
            };
            request.onerror = () => reject(request.error);
        });
    }
    
    // ========== Model Operations ==========
    
    /**
     * Check if a model is downloaded
     */
    async isModelDownloaded(modelId) {
        const tx = this.db.transaction(STORES.models, 'readonly');
        const store = tx.objectStore(STORES.models);
        return new Promise((resolve, reject) => {
            const request = store.get(modelId);
            request.onsuccess = () => resolve(!!request.result);
            request.onerror = () => reject(request.error);
        });
    }
    
    /**
     * Get model info (without loading full data)
     */
    async getModelInfo(modelId) {
        const tx = this.db.transaction(STORES.models, 'readonly');
        const store = tx.objectStore(STORES.models);
        return new Promise((resolve, reject) => {
            const request = store.get(modelId);
            request.onsuccess = () => {
                if (!request.result) return resolve(null);
                const { data, ...info } = request.result;
                resolve(info);
            };
            request.onerror = () => reject(request.error);
        });
    }
    
    /**
     * Save model data
     */
    async saveModel(modelId, data, metadata = {}) {
        const tx = this.db.transaction(STORES.models, 'readwrite');
        const store = tx.objectStore(STORES.models);
        return new Promise((resolve, reject) => {
            const request = store.put({
                id: modelId,
                data,
                size: data.byteLength || data.length,
                downloadedAt: Date.now(),
                ...metadata
            });
            request.onsuccess = () => resolve();
            request.onerror = () => reject(request.error);
        });
    }
    
    /**
     * Get model data
     */
    async getModel(modelId) {
        const tx = this.db.transaction(STORES.models, 'readonly');
        const store = tx.objectStore(STORES.models);
        return new Promise((resolve, reject) => {
            const request = store.get(modelId);
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }
    
    /**
     * Delete model
     */
    async deleteModel(modelId) {
        const tx = this.db.transaction(STORES.models, 'readwrite');
        const store = tx.objectStore(STORES.models);
        return new Promise((resolve, reject) => {
            const request = store.delete(modelId);
            request.onsuccess = () => resolve();
            request.onerror = () => reject(request.error);
        });
    }
    
    /**
     * List all downloaded models
     */
    async listModels() {
        const tx = this.db.transaction(STORES.models, 'readonly');
        const store = tx.objectStore(STORES.models);
        return new Promise((resolve, reject) => {
            const request = store.getAll();
            request.onsuccess = () => {
                resolve(request.result.map(m => ({
                    id: m.id,
                    size: m.size,
                    downloadedAt: m.downloadedAt
                })));
            };
            request.onerror = () => reject(request.error);
        });
    }
    
    // ========== Filesystem Operations ==========
    
    /**
     * Get a file from the overlay
     */
    async getFile(path) {
        const tx = this.db.transaction(STORES.filesystem, 'readonly');
        const store = tx.objectStore(STORES.filesystem);
        return new Promise((resolve, reject) => {
            const request = store.get(path);
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }
    
    /**
     * Save a file to the overlay
     */
    async saveFile(path, content) {
        const tx = this.db.transaction(STORES.filesystem, 'readwrite');
        const store = tx.objectStore(STORES.filesystem);
        return new Promise((resolve, reject) => {
            const request = store.put({
                path,
                content,
                mtime: Date.now(),
                size: content.byteLength || content.length
            });
            request.onsuccess = () => resolve();
            request.onerror = () => reject(request.error);
        });
    }
    
    /**
     * Delete a file from the overlay
     */
    async deleteFile(path) {
        const tx = this.db.transaction(STORES.filesystem, 'readwrite');
        const store = tx.objectStore(STORES.filesystem);
        return new Promise((resolve, reject) => {
            const request = store.delete(path);
            request.onsuccess = () => resolve();
            request.onerror = () => reject(request.error);
        });
    }
    
    /**
     * List files in the overlay (optionally filtered by prefix)
     */
    async listFiles(prefix = '') {
        const tx = this.db.transaction(STORES.filesystem, 'readonly');
        const store = tx.objectStore(STORES.filesystem);
        return new Promise((resolve, reject) => {
            const request = store.getAll();
            request.onsuccess = () => {
                let files = request.result.map(f => ({
                    path: f.path,
                    mtime: f.mtime,
                    size: f.size
                }));
                if (prefix) {
                    files = files.filter(f => f.path.startsWith(prefix));
                }
                resolve(files);
            };
            request.onerror = () => reject(request.error);
        });
    }
    
    // ========== Session Operations ==========
    
    /**
     * Save a session
     */
    async saveSession(session) {
        const tx = this.db.transaction(STORES.sessions, 'readwrite');
        const store = tx.objectStore(STORES.sessions);
        return new Promise((resolve, reject) => {
            const request = store.put({
                ...session,
                updatedAt: Date.now()
            });
            request.onsuccess = () => resolve();
            request.onerror = () => reject(request.error);
        });
    }
    
    /**
     * Get a session
     */
    async getSession(id) {
        const tx = this.db.transaction(STORES.sessions, 'readonly');
        const store = tx.objectStore(STORES.sessions);
        return new Promise((resolve, reject) => {
            const request = store.get(id);
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }
    
    /**
     * Delete a session
     */
    async deleteSession(id) {
        const tx = this.db.transaction(STORES.sessions, 'readwrite');
        const store = tx.objectStore(STORES.sessions);
        return new Promise((resolve, reject) => {
            const request = store.delete(id);
            request.onsuccess = () => resolve();
            request.onerror = () => reject(request.error);
        });
    }
    
    /**
     * List sessions (optionally filtered by agent)
     */
    async listSessions(agentHandle = null, limit = 50) {
        const tx = this.db.transaction(STORES.sessions, 'readonly');
        const store = tx.objectStore(STORES.sessions);
        
        return new Promise((resolve, reject) => {
            let request;
            if (agentHandle) {
                const index = store.index('agentHandle');
                request = index.getAll(agentHandle);
            } else {
                request = store.getAll();
            }
            
            request.onsuccess = () => {
                let sessions = request.result
                    .map(s => ({
                        id: s.id,
                        title: s.title,
                        agentHandle: s.agentHandle,
                        createdAt: s.createdAt,
                        updatedAt: s.updatedAt,
                        messageCount: s.messages?.length || 0
                    }))
                    .sort((a, b) => b.updatedAt - a.updatedAt)
                    .slice(0, limit);
                resolve(sessions);
            };
            request.onerror = () => reject(request.error);
        });
    }
    
    // ========== Asset Operations ==========
    
    /**
     * Cache an asset
     */
    async cacheAsset(url, data, metadata = {}) {
        const tx = this.db.transaction(STORES.assets, 'readwrite');
        const store = tx.objectStore(STORES.assets);
        return new Promise((resolve, reject) => {
            const request = store.put({
                url,
                data,
                size: data.byteLength || data.length,
                cachedAt: Date.now(),
                ...metadata
            });
            request.onsuccess = () => resolve();
            request.onerror = () => reject(request.error);
        });
    }
    
    /**
     * Get a cached asset
     */
    async getAsset(url) {
        const tx = this.db.transaction(STORES.assets, 'readonly');
        const store = tx.objectStore(STORES.assets);
        return new Promise((resolve, reject) => {
            const request = store.get(url);
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }
    
    /**
     * Delete a cached asset
     */
    async deleteAsset(url) {
        const tx = this.db.transaction(STORES.assets, 'readwrite');
        const store = tx.objectStore(STORES.assets);
        return new Promise((resolve, reject) => {
            const request = store.delete(url);
            request.onsuccess = () => resolve();
            request.onerror = () => reject(request.error);
        });
    }
    
    // ========== Utility Operations ==========
    
    /**
     * Get total storage usage
     */
    async getStorageUsage() {
        const stores = [STORES.config, STORES.models, STORES.filesystem, STORES.sessions, STORES.assets];
        const usage = {};
        
        for (const storeName of stores) {
            const tx = this.db.transaction(storeName, 'readonly');
            const store = tx.objectStore(storeName);
            const result = await new Promise((resolve, reject) => {
                const request = store.getAll();
                request.onsuccess = () => resolve(request.result);
                request.onerror = () => reject(request.error);
            });
            
            let size = 0;
            for (const item of result) {
                if (item.data) {
                    size += item.data.byteLength || item.data.length || 0;
                }
                if (item.content) {
                    size += item.content.byteLength || item.content.length || 0;
                }
                size += JSON.stringify(item).length; // Rough estimate of metadata
            }
            usage[storeName] = { count: result.length, size };
        }
        
        usage.total = Object.values(usage).reduce((sum, s) => sum + s.size, 0);
        return usage;
    }
    
    /**
     * Clear all data
     */
    async clearAll() {
        const stores = [STORES.config, STORES.models, STORES.filesystem, STORES.sessions, STORES.assets];
        
        for (const storeName of stores) {
            const tx = this.db.transaction(storeName, 'readwrite');
            const store = tx.objectStore(storeName);
            await new Promise((resolve, reject) => {
                const request = store.clear();
                request.onsuccess = () => resolve();
                request.onerror = () => reject(request.error);
            });
        }
    }
    
    /**
     * Close the database
     */
    close() {
        if (this.db) {
            this.db.close();
            this.db = null;
        }
    }
}

// Export for use in modules or global scope
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { OfflineStorage, encrypt, decrypt };
} else if (typeof window !== 'undefined') {
    window.OfflineStorage = OfflineStorage;
}
