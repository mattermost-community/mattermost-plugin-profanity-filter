package main

import (
	"strings"
)

// splitWordList splits a word list by ASCII commas and cleans up whitespace
// Note: Japanese commas are normalized to ASCII commas at configuration load time
func splitWordList(wordList string) []string {
	// Split by ASCII comma
	words := strings.Split(wordList, ",")

	// Clean up each word
	var cleanWords []string
	for _, word := range words {
		word = strings.TrimSpace(word)
		if word != "" {
			cleanWords = append(cleanWords, word)
		}
	}

	return cleanWords
}

// normalizeWordListCommas converts Japanese commas to ASCII commas in a word list string
func normalizeWordListCommas(wordList string) string {
	// Replace Japanese commas with ASCII commas
	normalized := strings.ReplaceAll(wordList, "、", ",")  // Ideographic comma (U+3001)
	normalized = strings.ReplaceAll(normalized, "，", ",") // Full-width comma (U+FF0C)
	return normalized
}
