/**
 * IndexedDB Storage Tests
 * 
 * Tests for the OfflineStorage class that manages IndexedDB operations
 * for API keys, models, filesystem, sessions, and assets.
 */

// Simple test framework for browser
const TestRunner = {
    tests: [],
    passed: 0,
    failed: 0,
    
    test(name, fn) {
        this.tests.push({ name, fn });
    },
    
    async run() {
        console.log('Starting Storage Tests...\n');
        
        for (const { name, fn } of this.tests) {
            try {
                await fn();
                this.passed++;
                console.log(`  OK: ${name}`);
            } catch (error) {
                this.failed++;
                console.error(`  FAIL: ${name}`);
                console.error(`        ${error.message}`);
            }
        }
        
        console.log(`\nResults: ${this.passed} passed, ${this.failed} failed`);
        return this.failed === 0;
    }
};

// Assertions
function assert(condition, message) {
    if (!condition) {
        throw new Error(message || 'Assertion failed');
    }
}

function assertEqual(actual, expected, message) {
    if (actual !== expected) {
        throw new Error(message || `Expected ${expected}, got ${actual}`);
    }
}

function assertNotNull(value, message) {
    if (value === null || value === undefined) {
        throw new Error(message || 'Expected non-null value');
    }
}

// Test setup
let storage;

async function setup() {
    // Delete existing test database
    await new Promise((resolve, reject) => {
        const req = indexedDB.deleteDatabase('ayo-offline');
        req.onsuccess = resolve;
        req.onerror = () => reject(req.error);
    });
    
    // Create fresh storage
    storage = await new OfflineStorage().init();
    storage.setEncryptionKey('test-key-12345');
}

// Tests for Config Operations
TestRunner.test('Config: set and get value', async () => {
    await setup();
    
    await storage.setConfig('testKey', 'testValue');
    const value = await storage.getConfig('testKey');
    
    assertEqual(value, 'testValue');
});

TestRunner.test('Config: update existing value', async () => {
    await setup();
    
    await storage.setConfig('updateKey', 'original');
    await storage.setConfig('updateKey', 'updated');
    const value = await storage.getConfig('updateKey');
    
    assertEqual(value, 'updated');
});

TestRunner.test('Config: delete value', async () => {
    await setup();
    
    await storage.setConfig('deleteKey', 'value');
    await storage.deleteConfig('deleteKey');
    const value = await storage.getConfig('deleteKey');
    
    assertEqual(value, undefined);
});

TestRunner.test('Config: get non-existent value returns undefined', async () => {
    await setup();
    
    const value = await storage.getConfig('nonexistent');
    
    assertEqual(value, undefined);
});

// Tests for API Key Operations (Encrypted)
TestRunner.test('API Key: store and retrieve encrypted key', async () => {
    await setup();
    
    await storage.setApiKey('openai', 'sk-test123456');
    const key = await storage.getApiKey('openai');
    
    assertEqual(key, 'sk-test123456');
});

TestRunner.test('API Key: different providers stored separately', async () => {
    await setup();
    
    await storage.setApiKey('openai', 'sk-openai-key');
    await storage.setApiKey('anthropic', 'sk-anthropic-key');
    
    const openaiKey = await storage.getApiKey('openai');
    const anthropicKey = await storage.getApiKey('anthropic');
    
    assertEqual(openaiKey, 'sk-openai-key');
    assertEqual(anthropicKey, 'sk-anthropic-key');
});

TestRunner.test('API Key: delete key', async () => {
    await setup();
    
    await storage.setApiKey('toDelete', 'key');
    await storage.deleteApiKey('toDelete');
    const key = await storage.getApiKey('toDelete');
    
    assertEqual(key, null);
});

TestRunner.test('API Key: list providers', async () => {
    await setup();
    
    await storage.setApiKey('provider1', 'key1');
    await storage.setApiKey('provider2', 'key2');
    
    const providers = await storage.listApiKeyProviders();
    
    assert(providers.includes('provider1'), 'Should include provider1');
    assert(providers.includes('provider2'), 'Should include provider2');
    assertEqual(providers.length, 2);
});

TestRunner.test('API Key: throws error without encryption key', async () => {
    await setup();
    storage.encryptionKey = null;
    
    let threw = false;
    try {
        await storage.setApiKey('test', 'key');
    } catch (e) {
        threw = true;
        assert(e.message.includes('Encryption key not set'));
    }
    
    assert(threw, 'Should throw error');
});

// Tests for Model Operations
TestRunner.test('Model: save and check downloaded', async () => {
    await setup();
    
    const data = new Uint8Array([1, 2, 3, 4, 5]);
    await storage.saveModel('llama-3.2-1b', data, { name: 'Llama 3.2 1B' });
    
    const isDownloaded = await storage.isModelDownloaded('llama-3.2-1b');
    assert(isDownloaded, 'Model should be marked as downloaded');
});

TestRunner.test('Model: get model info', async () => {
    await setup();
    
    const data = new Uint8Array([1, 2, 3, 4, 5]);
    await storage.saveModel('test-model', data, { name: 'Test Model' });
    
    const info = await storage.getModelInfo('test-model');
    
    assertEqual(info.id, 'test-model');
    assertEqual(info.size, 5);
    assertEqual(info.name, 'Test Model');
    assertNotNull(info.downloadedAt);
});

TestRunner.test('Model: list models', async () => {
    await setup();
    
    await storage.saveModel('model1', new Uint8Array([1]), { name: 'M1' });
    await storage.saveModel('model2', new Uint8Array([1, 2]), { name: 'M2' });
    
    const models = await storage.listModels();
    
    assertEqual(models.length, 2);
    assert(models.some(m => m.id === 'model1'));
    assert(models.some(m => m.id === 'model2'));
});

TestRunner.test('Model: delete model', async () => {
    await setup();
    
    await storage.saveModel('toDelete', new Uint8Array([1]));
    await storage.deleteModel('toDelete');
    
    const isDownloaded = await storage.isModelDownloaded('toDelete');
    assert(!isDownloaded, 'Model should be deleted');
});

// Tests for Filesystem Operations
TestRunner.test('Filesystem: save and get file', async () => {
    await setup();
    
    await storage.saveFile('/test/file.txt', 'Hello, World!');
    const file = await storage.getFile('/test/file.txt');
    
    assertEqual(file.path, '/test/file.txt');
    assertEqual(file.content, 'Hello, World!');
});

TestRunner.test('Filesystem: list files', async () => {
    await setup();
    
    await storage.saveFile('/dir/a.txt', 'A');
    await storage.saveFile('/dir/b.txt', 'B');
    await storage.saveFile('/other/c.txt', 'C');
    
    const allFiles = await storage.listFiles();
    assertEqual(allFiles.length, 3);
    
    const dirFiles = await storage.listFiles('/dir');
    assertEqual(dirFiles.length, 2);
});

TestRunner.test('Filesystem: delete file', async () => {
    await setup();
    
    await storage.saveFile('/toDelete.txt', 'content');
    await storage.deleteFile('/toDelete.txt');
    
    const file = await storage.getFile('/toDelete.txt');
    assertEqual(file, undefined);
});

// Tests for Session Operations
TestRunner.test('Session: save and get session', async () => {
    await setup();
    
    const session = {
        id: 'session-123',
        title: 'Test Session',
        agentHandle: '@ayo',
        createdAt: Date.now(),
        messages: [{ role: 'user', content: 'Hello' }]
    };
    
    await storage.saveSession(session);
    const retrieved = await storage.getSession('session-123');
    
    assertEqual(retrieved.id, 'session-123');
    assertEqual(retrieved.title, 'Test Session');
    assertEqual(retrieved.messages.length, 1);
});

TestRunner.test('Session: list sessions', async () => {
    await setup();
    
    await storage.saveSession({
        id: 's1',
        title: 'Session 1',
        agentHandle: '@ayo',
        createdAt: Date.now() - 1000,
        messages: []
    });
    
    await storage.saveSession({
        id: 's2',
        title: 'Session 2',
        agentHandle: '@other',
        createdAt: Date.now(),
        messages: []
    });
    
    const allSessions = await storage.listSessions();
    assertEqual(allSessions.length, 2);
    
    const ayoSessions = await storage.listSessions('@ayo');
    assertEqual(ayoSessions.length, 1);
    assertEqual(ayoSessions[0].id, 's1');
});

TestRunner.test('Session: delete session', async () => {
    await setup();
    
    await storage.saveSession({
        id: 'toDelete',
        title: 'Delete Me',
        agentHandle: '@ayo',
        createdAt: Date.now(),
        messages: []
    });
    
    await storage.deleteSession('toDelete');
    const session = await storage.getSession('toDelete');
    
    assertEqual(session, undefined);
});

// Tests for Asset Operations
TestRunner.test('Asset: cache and get asset', async () => {
    await setup();
    
    const data = new Uint8Array([1, 2, 3, 4, 5]);
    await storage.cacheAsset('https://example.com/asset.wasm', data, { version: '1.0' });
    
    const asset = await storage.getAsset('https://example.com/asset.wasm');
    
    assertEqual(asset.url, 'https://example.com/asset.wasm');
    assertEqual(asset.size, 5);
    assertEqual(asset.version, '1.0');
});

TestRunner.test('Asset: delete asset', async () => {
    await setup();
    
    await storage.cacheAsset('https://example.com/toDelete', new Uint8Array([1]));
    await storage.deleteAsset('https://example.com/toDelete');
    
    const asset = await storage.getAsset('https://example.com/toDelete');
    assertEqual(asset, undefined);
});

// Tests for Utility Operations
TestRunner.test('Utility: get storage usage', async () => {
    await setup();
    
    await storage.setConfig('test', 'value');
    await storage.saveFile('/test.txt', 'Some content here');
    
    const usage = await storage.getStorageUsage();
    
    assertNotNull(usage.config);
    assertNotNull(usage.filesystem);
    assertNotNull(usage.total);
    assert(usage.total > 0, 'Total should be greater than 0');
});

TestRunner.test('Utility: clear all data', async () => {
    await setup();
    
    await storage.setConfig('test', 'value');
    await storage.saveFile('/test.txt', 'content');
    
    await storage.clearAll();
    
    const config = await storage.getConfig('test');
    const files = await storage.listFiles();
    
    assertEqual(config, undefined);
    assertEqual(files.length, 0);
});

// Encryption Tests
TestRunner.test('Encryption: encrypted data is not plaintext', async () => {
    await setup();
    
    const secretKey = 'super-secret-api-key-12345';
    await storage.setApiKey('test', secretKey);
    
    // Read raw encrypted data from IndexedDB
    const encrypted = await storage.getConfig('apikey:test');
    
    // The encrypted data should not contain the plaintext
    assert(
        JSON.stringify(encrypted).indexOf(secretKey) === -1,
        'Encrypted data should not contain plaintext'
    );
    
    // But we should be able to decrypt it
    const decrypted = await storage.getApiKey('test');
    assertEqual(decrypted, secretKey);
});

TestRunner.test('Encryption: different encryption keys produce different ciphertext', async () => {
    await setup();
    
    storage.setEncryptionKey('key1');
    await storage.setApiKey('test1', 'secret');
    const encrypted1 = await storage.getConfig('apikey:test1');
    
    storage.setEncryptionKey('key2');
    await storage.setApiKey('test2', 'secret');
    const encrypted2 = await storage.getConfig('apikey:test2');
    
    // Ciphertexts should be different
    assert(
        JSON.stringify(encrypted1.ciphertext) !== JSON.stringify(encrypted2.ciphertext),
        'Different keys should produce different ciphertext'
    );
});

// Run tests if in browser
if (typeof window !== 'undefined') {
    window.runStorageTests = () => TestRunner.run();
}

// Export for Node.js testing
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { TestRunner, runTests: () => TestRunner.run() };
}
