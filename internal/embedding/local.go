package embedding

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	ort "github.com/yalue/onnxruntime_go"
)

// LocalEmbedder generates embeddings using a local ONNX model.
type LocalEmbedder struct {
	session   *ort.DynamicAdvancedSession
	tokenizer *Tokenizer
	dimension int
	maxLength int
	mu        sync.Mutex
}

// LocalConfig configures the local embedder.
type LocalConfig struct {
	// ModelPath is the path to the ONNX model file.
	ModelPath string

	// TokenizerPath is the path to the tokenizer vocabulary file.
	TokenizerPath string

	// Dimension is the embedding dimension (default: 384 for MiniLM).
	Dimension int

	// MaxLength is the maximum token sequence length (default: 256).
	MaxLength int
}

// DefaultModelDir returns the default directory for embedding models.
func DefaultModelDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "ayo", "models")
}

// DefaultModelPath returns the path to the default embedding model.
func DefaultModelPath() string {
	return filepath.Join(DefaultModelDir(), "all-MiniLM-L6-v2.onnx")
}

// DefaultTokenizerPath returns the path to the default tokenizer vocabulary.
func DefaultTokenizerPath() string {
	return filepath.Join(DefaultModelDir(), "tokenizer.json")
}

// NewLocalEmbedder creates a new local embedder with the given configuration.
func NewLocalEmbedder(cfg LocalConfig) (*LocalEmbedder, error) {
	if cfg.Dimension == 0 {
		cfg.Dimension = Dimension
	}
	if cfg.MaxLength == 0 {
		cfg.MaxLength = 256
	}
	if cfg.ModelPath == "" {
		cfg.ModelPath = DefaultModelPath()
	}
	if cfg.TokenizerPath == "" {
		cfg.TokenizerPath = DefaultTokenizerPath()
	}

	// Check if model exists
	if _, err := os.Stat(cfg.ModelPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("embedding model not found at %s: run 'ayo setup' to download", cfg.ModelPath)
	}

	// Check if tokenizer exists
	if _, err := os.Stat(cfg.TokenizerPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("tokenizer not found at %s: run 'ayo setup' to download", cfg.TokenizerPath)
	}

	// Initialize ONNX runtime if not already done
	if err := initONNXRuntime(); err != nil {
		return nil, fmt.Errorf("failed to initialize ONNX runtime: %w", err)
	}

	// Load tokenizer
	tokenizer, err := LoadTokenizer(cfg.TokenizerPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load tokenizer: %w", err)
	}

	// Create ONNX session
	session, err := ort.NewDynamicAdvancedSession(
		cfg.ModelPath,
		[]string{"input_ids", "attention_mask", "token_type_ids"},
		[]string{"last_hidden_state"},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ONNX session: %w", err)
	}

	return &LocalEmbedder{
		session:   session,
		tokenizer: tokenizer,
		dimension: cfg.Dimension,
		maxLength: cfg.MaxLength,
	}, nil
}

var (
	onnxInitOnce sync.Once
	onnxInitErr  error
)

func initONNXRuntime() error {
	onnxInitOnce.Do(func() {
		// Set library path - check env var first, then auto-detect
		if libPath := os.Getenv("ONNX_LIBRARY_PATH"); libPath != "" {
			ort.SetSharedLibraryPath(libPath)
		} else if libPath := findONNXLibrary(); libPath != "" {
			ort.SetSharedLibraryPath(libPath)
		}
		onnxInitErr = ort.InitializeEnvironment()
	})
	return onnxInitErr
}

// findONNXLibrary searches common locations for the ONNX Runtime library.
func findONNXLibrary() string {
	// Common Homebrew locations (macOS)
	homebrewPaths := []string{
		"/opt/homebrew/lib/libonnxruntime.dylib",     // Apple Silicon
		"/usr/local/lib/libonnxruntime.dylib",        // Intel Mac
		"/opt/homebrew/opt/onnxruntime/lib/libonnxruntime.dylib",
		"/usr/local/opt/onnxruntime/lib/libonnxruntime.dylib",
	}
	
	// Linux paths (apt, dnf, pacman)
	linuxPaths := []string{
		"/usr/lib/libonnxruntime.so",
		"/usr/local/lib/libonnxruntime.so",
		"/usr/lib/x86_64-linux-gnu/libonnxruntime.so",   // Debian/Ubuntu x86_64
		"/usr/lib/aarch64-linux-gnu/libonnxruntime.so",  // Debian/Ubuntu ARM64
	}
	
	// User-local paths (legacy install location)
	home, _ := os.UserHomeDir()
	userPaths := []string{
		filepath.Join(home, ".local", "share", "ayo", "lib", "libonnxruntime.dylib"),
		filepath.Join(home, ".local", "share", "ayo", "lib", "libonnxruntime.so"),
	}
	
	allPaths := append(homebrewPaths, linuxPaths...)
	allPaths = append(allPaths, userPaths...)
	
	for _, p := range allPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	
	return ""
}

// Embed generates an embedding for the given text.
func (e *LocalEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	results, err := e.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	return results[0], nil
}

// EmbedBatch generates embeddings for multiple texts.
func (e *LocalEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if e == nil || e.session == nil {
		return nil, ErrModelNotLoaded
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	batchSize := len(texts)
	if batchSize == 0 {
		return nil, nil
	}

	// Tokenize all texts
	inputIDs := make([]int64, 0, batchSize*e.maxLength)
	attentionMask := make([]int64, 0, batchSize*e.maxLength)
	tokenTypeIDs := make([]int64, 0, batchSize*e.maxLength)

	for _, text := range texts {
		tokens := e.tokenizer.Encode(text, e.maxLength)

		// Pad or truncate to maxLength
		for i := 0; i < e.maxLength; i++ {
			if i < len(tokens.InputIDs) {
				inputIDs = append(inputIDs, tokens.InputIDs[i])
				attentionMask = append(attentionMask, tokens.AttentionMask[i])
				tokenTypeIDs = append(tokenTypeIDs, tokens.TokenTypeIDs[i])
			} else {
				inputIDs = append(inputIDs, 0)
				attentionMask = append(attentionMask, 0)
				tokenTypeIDs = append(tokenTypeIDs, 0)
			}
		}
	}

	// Create input tensors
	shape := ort.Shape{int64(batchSize), int64(e.maxLength)}

	inputIDsTensor, err := ort.NewTensor(shape, inputIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to create input_ids tensor: %w", err)
	}
	defer inputIDsTensor.Destroy()

	attentionMaskTensor, err := ort.NewTensor(shape, attentionMask)
	if err != nil {
		return nil, fmt.Errorf("failed to create attention_mask tensor: %w", err)
	}
	defer attentionMaskTensor.Destroy()

	tokenTypeIDsTensor, err := ort.NewTensor(shape, tokenTypeIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to create token_type_ids tensor: %w", err)
	}
	defer tokenTypeIDsTensor.Destroy()

	// Run inference - outputs will be auto-allocated
	outputs := []ort.Value{nil}
	err = e.session.Run(
		[]ort.Value{inputIDsTensor, attentionMaskTensor, tokenTypeIDsTensor},
		outputs,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to run inference: %w", err)
	}
	defer func() {
		for _, o := range outputs {
			if o != nil {
				o.Destroy()
			}
		}
	}()

	// Extract embeddings from output
	// Output shape: [batch_size, sequence_length, hidden_size]
	outputTensor, ok := outputs[0].(*ort.Tensor[float32])
	if !ok {
		return nil, fmt.Errorf("unexpected output tensor type")
	}

	outputData := outputTensor.GetData()
	outputShape := outputTensor.GetShape()

	// Mean pooling over sequence dimension
	results := make([][]float32, batchSize)
	seqLen := int(outputShape[1])
	hiddenSize := int(outputShape[2])

	for b := 0; b < batchSize; b++ {
		embedding := make([]float32, hiddenSize)
		validTokens := 0

		for s := 0; s < seqLen; s++ {
			// Only pool over valid tokens (attention_mask = 1)
			maskIdx := b*e.maxLength + s
			if maskIdx < len(attentionMask) && attentionMask[maskIdx] == 1 {
				validTokens++
				for h := 0; h < hiddenSize; h++ {
					idx := b*seqLen*hiddenSize + s*hiddenSize + h
					embedding[h] += outputData[idx]
				}
			}
		}

		// Average pooling
		if validTokens > 0 {
			for h := 0; h < hiddenSize; h++ {
				embedding[h] /= float32(validTokens)
			}
		}

		// Normalize the embedding
		results[b] = Normalize(embedding)
	}

	return results, nil
}

// Dimension returns the embedding dimension.
func (e *LocalEmbedder) Dimension() int {
	return e.dimension
}

// Close releases resources.
func (e *LocalEmbedder) Close() error {
	if e == nil {
		return nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.session != nil {
		if err := e.session.Destroy(); err != nil {
			return err
		}
		e.session = nil
	}
	return nil
}

// IsModelAvailable checks if the embedding model is available locally.
func IsModelAvailable() bool {
	_, err := os.Stat(DefaultModelPath())
	return err == nil
}
