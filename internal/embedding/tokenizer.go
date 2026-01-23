package embedding

import (
	"encoding/json"
	"os"
	"strings"
	"unicode"
)

// Tokenizer handles text tokenization for BERT-based models.
type Tokenizer struct {
	vocab    map[string]int64
	unkToken string
	clsToken string
	sepToken string
	padToken string
	unkID    int64
	clsID    int64
	sepID    int64
	padID    int64
}

// TokenizerOutput contains the tokenized output.
type TokenizerOutput struct {
	InputIDs      []int64
	AttentionMask []int64
	TokenTypeIDs  []int64
}

// tokenizerJSON is the structure of HuggingFace tokenizer.json files.
type tokenizerJSON struct {
	Model struct {
		Vocab map[string]int64 `json:"vocab"`
	} `json:"model"`
	AddedTokens []struct {
		Content string `json:"content"`
		ID      int64  `json:"id"`
	} `json:"added_tokens"`
}

// LoadTokenizer loads a tokenizer from a HuggingFace tokenizer.json file.
func LoadTokenizer(path string) (*Tokenizer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var tj tokenizerJSON
	if err := json.Unmarshal(data, &tj); err != nil {
		return nil, err
	}

	vocab := tj.Model.Vocab

	// Add special tokens from added_tokens
	for _, t := range tj.AddedTokens {
		vocab[t.Content] = t.ID
	}

	t := &Tokenizer{
		vocab:    vocab,
		unkToken: "[UNK]",
		clsToken: "[CLS]",
		sepToken: "[SEP]",
		padToken: "[PAD]",
	}

	// Get special token IDs
	if id, ok := vocab["[UNK]"]; ok {
		t.unkID = id
	}
	if id, ok := vocab["[CLS]"]; ok {
		t.clsID = id
	}
	if id, ok := vocab["[SEP]"]; ok {
		t.sepID = id
	}
	if id, ok := vocab["[PAD]"]; ok {
		t.padID = id
	}

	return t, nil
}

// Encode tokenizes text and returns token IDs with attention mask.
func (t *Tokenizer) Encode(text string, maxLength int) TokenizerOutput {
	// Basic preprocessing
	text = strings.ToLower(text)
	text = t.cleanText(text)

	// Tokenize
	tokens := t.tokenize(text)

	// Truncate if needed (leave room for [CLS] and [SEP])
	if len(tokens) > maxLength-2 {
		tokens = tokens[:maxLength-2]
	}

	// Build input IDs with special tokens
	inputIDs := make([]int64, 0, len(tokens)+2)
	inputIDs = append(inputIDs, t.clsID)

	for _, token := range tokens {
		if id, ok := t.vocab[token]; ok {
			inputIDs = append(inputIDs, id)
		} else {
			// Try WordPiece tokenization for unknown tokens
			subwords := t.wordPieceTokenize(token)
			for _, sw := range subwords {
				if id, ok := t.vocab[sw]; ok {
					inputIDs = append(inputIDs, id)
				} else {
					inputIDs = append(inputIDs, t.unkID)
				}
			}
		}
	}

	inputIDs = append(inputIDs, t.sepID)

	// Create attention mask
	attentionMask := make([]int64, len(inputIDs))
	for i := range attentionMask {
		attentionMask[i] = 1
	}

	// Create token type IDs (all zeros for single sentence)
	tokenTypeIDs := make([]int64, len(inputIDs))

	return TokenizerOutput{
		InputIDs:      inputIDs,
		AttentionMask: attentionMask,
		TokenTypeIDs:  tokenTypeIDs,
	}
}

// cleanText performs basic text cleaning.
func (t *Tokenizer) cleanText(text string) string {
	var result strings.Builder
	for _, r := range text {
		if r == 0 || r == 0xfffd || unicode.Is(unicode.Cc, r) {
			continue
		}
		if unicode.IsSpace(r) {
			result.WriteRune(' ')
		} else {
			result.WriteRune(r)
		}
	}
	return strings.TrimSpace(result.String())
}

// tokenize performs basic whitespace and punctuation tokenization.
func (t *Tokenizer) tokenize(text string) []string {
	var tokens []string
	var current strings.Builder

	for _, r := range text {
		if unicode.IsSpace(r) {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		} else if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			tokens = append(tokens, string(r))
		} else {
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

// wordPieceTokenize breaks a word into WordPiece subwords.
func (t *Tokenizer) wordPieceTokenize(word string) []string {
	if len(word) == 0 {
		return nil
	}

	var subwords []string
	start := 0

	for start < len(word) {
		end := len(word)
		found := false

		for end > start {
			substr := word[start:end]
			if start > 0 {
				substr = "##" + substr
			}

			if _, ok := t.vocab[substr]; ok {
				subwords = append(subwords, substr)
				found = true
				break
			}
			end--
		}

		if !found {
			// Single character not in vocab, use [UNK]
			subwords = append(subwords, t.unkToken)
			start++
		} else {
			start = end
		}
	}

	return subwords
}
