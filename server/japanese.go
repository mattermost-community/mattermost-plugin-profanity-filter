package main

import (
	"fmt"
	"strings"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

func (p *Plugin) initializeJapaneseTokenizer() error {
	// Initialize Japanese tokenizer
	t, err := tokenizer.New(ipa.Dict())
	if err != nil {
		return fmt.Errorf("failed to initialize Japanese tokenizer: %w", err)
	}
	p.japaneseTokenizer = t

	return nil
}

// isJapaneseRune checks if a rune is a Japanese character (Hiragana, Katakana, or Kanji)
func isJapaneseRune(r rune) bool {
	// Hiragana: U+3040-U+309F
	// Katakana: U+30A0-U+30FF
	// CJK Unified Ideographs (Kanji): U+4E00-U+9FFF
	return (r >= 0x3040 && r <= 0x309F) || // Hiragana
		(r >= 0x30A0 && r <= 0x30FF) || // Katakana
		(r >= 0x4E00 && r <= 0x9FFF) // Kanji
}

// isJapaneseWord checks if a word contains Japanese characters
func isJapaneseWord(word string) bool {
	for _, r := range word {
		if isJapaneseRune(r) {
			return true
		}
	}
	return false
}

// isJapaneseText checks if text contains Japanese characters (alias for isJapaneseWord)
func isJapaneseText(text string) bool {
	return isJapaneseWord(text)
}

// tokenizeJapanese tokenizes Japanese text using Kagome morphological analyzer
func tokenizeJapanese(text string, tokenizer *tokenizer.Tokenizer) []string {
	tokens := tokenizer.Tokenize(text)
	var words []string

	for _, token := range tokens {
		surface := token.Surface
		if surface != "" && surface != " " {
			words = append(words, strings.ToLower(surface))
		}
	}

	return words
}

// detectJapaneseWords uses tokenization for Japanese words to ensure proper word boundaries
func (p *Plugin) detectJapaneseWords(text string, japaneseWords []string) []string {
	var detected []string

	// Only tokenize if text contains Japanese characters
	if !isJapaneseText(text) {
		return detected
	}

	// Tokenize the Japanese text
	tokens := tokenizeJapanese(text, p.getJapaneseTokenizer())

	// Check each Japanese bad word against the tokenized text
	for _, badWord := range japaneseWords {
		if badWord == "" {
			continue
		}

		badWordLower := strings.ToLower(strings.TrimSpace(badWord))

		// First try exact token matching (for proper morphological words)
		tokenMatched := false
		for _, token := range tokens {
			if token == badWordLower {
				detected = append(detected, badWord)
				tokenMatched = true
				break
			}
		}

		// If no token match, fall back to substring matching for compound words
		// This handles cases where compounds like "クソ野郎" might be tokenized as separate parts
		if !tokenMatched && strings.Contains(strings.ToLower(text), badWordLower) {
			detected = append(detected, badWord)
		}
	}

	return detected
}

// detectJapaneseWordsWithTokenization uses tokenization + regex approach for Japanese text
func (p *Plugin) detectJapaneseWordsWithTokenization(text string, japaneseWords []string) []string {
	var detected []string

	// Only process if text contains Japanese characters
	if !isJapaneseText(text) {
		return detected
	}

	// Get the pre-compiled regex
	regex := p.getJapaneseWordsRegex()
	if regex == nil {
		return p.detectJapaneseWords(text, japaneseWords)
	}

	// Tokenize the Japanese text to create word boundaries
	tokens := tokenizeJapanese(text, p.getJapaneseTokenizer())
	tokenizedText := strings.Join(tokens, " ") // Create spaces between tokens

	// Find matches in tokenized text
	matches := regex.FindAllString(strings.ToLower(tokenizedText), -1)

	// Return the original words that match
	for _, match := range matches {
		for _, word := range japaneseWords {
			word = strings.TrimSpace(word)
			if word != "" && strings.ToLower(word) == match {
				detected = append(detected, word)
				break
			}
		}
	}

	// If no matches found with tokenization, fall back to the old approach
	if len(detected) == 0 {
		return p.detectJapaneseWords(text, japaneseWords)
	}

	return detected
}
