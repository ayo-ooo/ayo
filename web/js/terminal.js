/**
 * Terminal Manager
 * 
 * Wraps xterm.js and connects it to the EmulatorController.
 * Handles terminal initialization, input/output, and resizing.
 */

/**
 * TerminalManager class
 */
class TerminalManager {
    constructor(options = {}) {
        this.terminal = null;
        this.fitAddon = null;
        this.emulator = null;
        this.container = null;
        this.initialized = false;
        this.onReady = options.onReady || (() => {});
        this.onError = options.onError || (() => {});
        this.resizeObserver = null;
        this.inputBuffer = '';
    }
    
    /**
     * Initialize the terminal
     * @param {HTMLElement} container - DOM element to attach terminal to
     * @param {EmulatorController} emulator - Emulator controller instance
     */
    async init(container, emulator) {
        if (this.initialized) {
            return this;
        }
        
        this.container = container;
        this.emulator = emulator;
        
        // Wait for xterm.js to load
        await this.loadXterm();
        
        // Create terminal instance
        this.terminal = new Terminal({
            cursorBlink: true,
            cursorStyle: 'block',
            fontFamily: '"Fira Code", "Cascadia Code", "SF Mono", Menlo, Monaco, monospace',
            fontSize: 14,
            lineHeight: 1.2,
            theme: {
                background: '#000000',
                foreground: '#eee',
                cursor: '#9d4edd',
                cursorAccent: '#000000',
                selectionBackground: 'rgba(157, 78, 221, 0.3)',
                black: '#000000',
                red: '#f87171',
                green: '#4ade80',
                yellow: '#fbbf24',
                blue: '#60a5fa',
                magenta: '#9d4edd',
                cyan: '#22d3ee',
                white: '#eee',
                brightBlack: '#666',
                brightRed: '#f87171',
                brightGreen: '#4ade80',
                brightYellow: '#fbbf24',
                brightBlue: '#60a5fa',
                brightMagenta: '#c084fc',
                brightCyan: '#22d3ee',
                brightWhite: '#fff'
            },
            allowProposedApi: true,
            scrollback: 10000,
            tabStopWidth: 4
        });
        
        // Create fit addon
        this.fitAddon = new FitAddon.FitAddon();
        this.terminal.loadAddon(this.fitAddon);
        
        // Open terminal in container
        this.terminal.open(container);
        
        // Fit to container
        this.fitAddon.fit();
        
        // Set up resize observer
        this.resizeObserver = new ResizeObserver(() => {
            this.fit();
        });
        this.resizeObserver.observe(container);
        
        // Connect terminal input to emulator
        this.terminal.onData((data) => {
            if (this.emulator && this.emulator.isRunning()) {
                this.emulator.sendInput(data);
            } else {
                // Buffer input until emulator is ready
                this.inputBuffer += data;
            }
        });
        
        // Connect emulator output to terminal
        if (this.emulator) {
            const originalOnOutput = this.emulator.onOutput;
            this.emulator.onOutput = (data) => {
                this.write(data);
                originalOnOutput(data);
            };
        }
        
        // Set up paste handling
        this.terminal.attachCustomKeyEventHandler((event) => {
            // Allow Ctrl+V / Cmd+V for paste
            if ((event.ctrlKey || event.metaKey) && event.key === 'v') {
                return false; // Let browser handle paste
            }
            // Allow Ctrl+C / Cmd+C for copy
            if ((event.ctrlKey || event.metaKey) && event.key === 'c') {
                if (this.terminal.hasSelection()) {
                    return false; // Let browser handle copy
                }
            }
            return true;
        });
        
        // Handle paste event
        container.addEventListener('paste', async (event) => {
            event.preventDefault();
            const text = event.clipboardData.getData('text');
            if (text && this.emulator && this.emulator.isRunning()) {
                this.emulator.sendInput(text);
            }
        });
        
        this.initialized = true;
        this.onReady();
        
        return this;
    }
    
    /**
     * Load xterm.js from CDN if not already loaded
     */
    async loadXterm() {
        // Check if already loaded
        if (typeof Terminal !== 'undefined' && typeof FitAddon !== 'undefined') {
            return;
        }
        
        // Load CSS
        if (!document.querySelector('link[href*="xterm.css"]')) {
            const css = document.createElement('link');
            css.rel = 'stylesheet';
            css.href = 'https://cdn.jsdelivr.net/npm/xterm@5.3.0/css/xterm.css';
            document.head.appendChild(css);
        }
        
        // Load xterm.js
        if (typeof Terminal === 'undefined') {
            await this.loadScript('https://cdn.jsdelivr.net/npm/xterm@5.3.0/lib/xterm.min.js');
        }
        
        // Load fit addon
        if (typeof FitAddon === 'undefined') {
            await this.loadScript('https://cdn.jsdelivr.net/npm/xterm-addon-fit@0.8.0/lib/xterm-addon-fit.min.js');
        }
    }
    
    /**
     * Load a script dynamically
     */
    loadScript(src) {
        return new Promise((resolve, reject) => {
            const script = document.createElement('script');
            script.src = src;
            script.onload = resolve;
            script.onerror = () => reject(new Error(`Failed to load ${src}`));
            document.head.appendChild(script);
        });
    }
    
    /**
     * Write data to the terminal
     */
    write(data) {
        if (this.terminal) {
            this.terminal.write(data);
        }
    }
    
    /**
     * Write a line to the terminal
     */
    writeln(data) {
        if (this.terminal) {
            this.terminal.writeln(data);
        }
    }
    
    /**
     * Clear the terminal
     */
    clear() {
        if (this.terminal) {
            this.terminal.clear();
        }
    }
    
    /**
     * Reset the terminal
     */
    reset() {
        if (this.terminal) {
            this.terminal.reset();
        }
    }
    
    /**
     * Focus the terminal
     */
    focus() {
        if (this.terminal) {
            this.terminal.focus();
        }
    }
    
    /**
     * Blur the terminal
     */
    blur() {
        if (this.terminal) {
            this.terminal.blur();
        }
    }
    
    /**
     * Fit terminal to container
     */
    fit() {
        if (this.fitAddon && this.container) {
            try {
                this.fitAddon.fit();
                
                // Notify emulator of new dimensions
                if (this.emulator && this.terminal) {
                    const dims = {
                        cols: this.terminal.cols,
                        rows: this.terminal.rows
                    };
                    // Could send resize event to VM if needed
                }
            } catch (e) {
                // Ignore fit errors (can happen during transitions)
            }
        }
    }
    
    /**
     * Scroll to bottom
     */
    scrollToBottom() {
        if (this.terminal) {
            this.terminal.scrollToBottom();
        }
    }
    
    /**
     * Get current selection
     */
    getSelection() {
        return this.terminal ? this.terminal.getSelection() : '';
    }
    
    /**
     * Check if terminal has selection
     */
    hasSelection() {
        return this.terminal ? this.terminal.hasSelection() : false;
    }
    
    /**
     * Flush buffered input to emulator
     */
    flushInputBuffer() {
        if (this.inputBuffer && this.emulator && this.emulator.isRunning()) {
            this.emulator.sendInput(this.inputBuffer);
            this.inputBuffer = '';
        }
    }
    
    /**
     * Show boot message
     */
    showBootMessage() {
        this.writeln('\x1b[1;35mayo\x1b[0m Offline Terminal');
        this.writeln('');
        this.writeln('Starting Linux VM...');
        this.writeln('');
    }
    
    /**
     * Show ready message
     */
    showReadyMessage() {
        // Flush any buffered input now that we're ready
        this.flushInputBuffer();
    }
    
    /**
     * Dispose of the terminal
     */
    dispose() {
        if (this.resizeObserver) {
            this.resizeObserver.disconnect();
            this.resizeObserver = null;
        }
        
        if (this.terminal) {
            this.terminal.dispose();
            this.terminal = null;
        }
        
        this.fitAddon = null;
        this.emulator = null;
        this.container = null;
        this.initialized = false;
    }
}

// Export
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { TerminalManager };
} else if (typeof window !== 'undefined') {
    window.TerminalManager = TerminalManager;
}
