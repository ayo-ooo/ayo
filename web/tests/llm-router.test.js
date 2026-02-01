/**
 * Browser LLM Router Tests
 * 
 * Tests for the LLMRouter class that handles routing between
 * WebLLM and cloud API providers.
 */

const LLMTestRunner = {
    tests: [],
    passed: 0,
    failed: 0,
    
    test(name, fn) {
        this.tests.push({ name, fn });
    },
    
    async run() {
        console.log('Starting LLM Router Tests...\n');
        
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
    if (!condition) throw new Error(message || 'Assertion failed');
}

function assertEqual(actual, expected, message) {
    if (actual !== expected) throw new Error(message || `Expected ${expected}, got ${actual}`);
}

function assertNotNull(value, message) {
    if (value === null || value === undefined) throw new Error(message || 'Expected non-null');
}

// Mock storage for tests
class MockStorage {
    constructor() {
        this.apiKeys = new Map();
    }
    
    async listApiKeyProviders() {
        return Array.from(this.apiKeys.keys());
    }
    
    async getApiKey(provider) {
        return this.apiKeys.get(provider) || null;
    }
    
    async setApiKey(provider, key) {
        this.apiKeys.set(provider, key);
    }
    
    async deleteApiKey(provider) {
        this.apiKeys.delete(provider);
    }
}

// Test setup
let storage;
let router;

async function setup() {
    storage = new MockStorage();
    router = new LLMRouter(storage);
    await router.init();
}

// Tests for WebGPU Detection
LLMTestRunner.test('WebGPU: checkWebGPU returns object with available property', async () => {
    const result = await checkWebGPU();
    
    assertNotNull(result);
    assert('available' in result, 'Result should have available property');
    assert(typeof result.available === 'boolean', 'available should be boolean');
});

LLMTestRunner.test('WebGPU: checkWebGPU returns reason when not available', async () => {
    const result = await checkWebGPU();
    
    if (!result.available) {
        assertNotNull(result.reason, 'Should have reason when not available');
    }
});

// Tests for Provider Configuration
LLMTestRunner.test('Provider: PROVIDERS contains expected providers', async () => {
    assert('openai' in LLM_PROVIDERS, 'Should have openai');
    assert('anthropic' in LLM_PROVIDERS, 'Should have anthropic');
    assert('openrouter' in LLM_PROVIDERS, 'Should have openrouter');
});

LLMTestRunner.test('Provider: providers have required fields', async () => {
    for (const [id, config] of Object.entries(LLM_PROVIDERS)) {
        assertNotNull(config.name, `${id} should have name`);
        assertNotNull(config.baseUrl, `${id} should have baseUrl`);
        assertNotNull(config.models, `${id} should have models`);
        assert(Array.isArray(config.models), `${id} models should be array`);
    }
});

// Tests for WebLLM Models
LLMTestRunner.test('WebLLM: WEBLLM_MODELS is defined', async () => {
    assertNotNull(WEBLLM_MODELS);
    assert(Array.isArray(WEBLLM_MODELS), 'Should be array');
    assert(WEBLLM_MODELS.length > 0, 'Should have models');
});

LLMTestRunner.test('WebLLM: models have required fields', async () => {
    for (const model of WEBLLM_MODELS) {
        assertNotNull(model.id, 'Model should have id');
        assertNotNull(model.size, 'Model should have size');
        assertNotNull(model.minVRAM, 'Model should have minVRAM');
    }
});

// Tests for Router Initialization
LLMTestRunner.test('Router: init sets webgpuInfo', async () => {
    await setup();
    
    assertNotNull(router.webgpuInfo);
    assert('available' in router.webgpuInfo);
});

LLMTestRunner.test('Router: init detects configured providers', async () => {
    storage = new MockStorage();
    await storage.setApiKey('openai', 'sk-test');
    
    router = new LLMRouter(storage);
    await router.init();
    
    assert(router.providers.includes('openai'), 'Should detect openai');
});

// Tests for Backend Detection
LLMTestRunner.test('Router: getBackends returns array', async () => {
    await setup();
    
    const backends = await router.getBackends();
    
    assert(Array.isArray(backends), 'Should return array');
    assert(backends.length >= 4, 'Should have WebLLM + 3 cloud providers');
});

LLMTestRunner.test('Router: getBackends includes WebLLM', async () => {
    await setup();
    
    const backends = await router.getBackends();
    const webllm = backends.find(b => b.type === 'webllm');
    
    assertNotNull(webllm, 'Should include webllm backend');
    assertEqual(webllm.name, 'Local (WebLLM)');
});

LLMTestRunner.test('Router: getBackends shows cloud providers as unavailable without keys', async () => {
    await setup();
    
    const backends = await router.getBackends();
    const cloudBackends = backends.filter(b => b.type === 'cloud');
    
    for (const backend of cloudBackends) {
        assertEqual(backend.available, false, `${backend.id} should be unavailable`);
    }
});

LLMTestRunner.test('Router: getBackends shows cloud provider as available with key', async () => {
    storage = new MockStorage();
    await storage.setApiKey('openai', 'sk-test');
    
    router = new LLMRouter(storage);
    await router.init();
    
    const backends = await router.getBackends();
    const openai = backends.find(b => b.id === 'openai');
    
    assertNotNull(openai);
    assertEqual(openai.available, true);
});

// Tests for Availability Check
LLMTestRunner.test('Router: isAvailable returns false with no backends', async () => {
    await setup();
    router.webgpuInfo = { available: false };
    router.providers = [];
    
    assertEqual(router.isAvailable(), false);
});

LLMTestRunner.test('Router: isAvailable returns true with WebGPU', async () => {
    await setup();
    router.webgpuInfo = { available: true };
    router.providers = [];
    
    assertEqual(router.isAvailable(), true);
});

LLMTestRunner.test('Router: isAvailable returns true with API key', async () => {
    await setup();
    router.webgpuInfo = { available: false };
    router.providers = ['openai'];
    
    assertEqual(router.isAvailable(), true);
});

// Tests for Generation (Mock)
LLMTestRunner.test('Router: generate throws without any backend', async () => {
    await setup();
    router.webgpuInfo = { available: false };
    router.providers = [];
    
    let threw = false;
    try {
        await router.generate({ messages: [{ role: 'user', content: 'Hi' }] });
    } catch (e) {
        threw = true;
        assert(e.message.includes('No LLM backend'), 'Should throw no backend error');
    }
    
    assert(threw, 'Should throw error');
});

// Tests for Model Loading
LLMTestRunner.test('Router: loadWebLLMModel throws without WebGPU', async () => {
    await setup();
    router.webgpuInfo = { available: false };
    
    let threw = false;
    try {
        await router.loadWebLLMModel('test-model');
    } catch (e) {
        threw = true;
        assert(e.message.includes('WebGPU not available'));
    }
    
    assert(threw, 'Should throw error');
});

LLMTestRunner.test('Router: unloadWebLLMModel clears current model', async () => {
    await setup();
    router.currentModel = 'test';
    router.webllmEngine = { some: 'engine' };
    
    await router.unloadWebLLMModel();
    
    assertEqual(router.currentModel, null);
    assertEqual(router.webllmEngine, null);
});

// Run tests if in browser
if (typeof window !== 'undefined') {
    window.runLLMRouterTests = () => LLMTestRunner.run();
}

// Export for Node.js testing
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { LLMTestRunner, runTests: () => LLMTestRunner.run() };
}
