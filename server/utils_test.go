package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitWordList(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "ASCII commas only",
			input:    "word1,word2,word3",
			expected: []string{"word1", "word2", "word3"},
		},
		{
			name:     "Japanese words with ASCII commas (normalized at config level)",
			input:    "ばか,バカ,馬鹿",
			expected: []string{"ばか", "バカ", "馬鹿"},
		},
		{
			name:     "Spaces around words",
			input:    " word1 , ばか , バカ , word2 ",
			expected: []string{"word1", "ばか", "バカ", "word2"},
		},
		{
			name:     "Empty words filtered out",
			input:    "word1,,ばか,,,バカ,",
			expected: []string{"word1", "ばか", "バカ"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitWordList(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeWordListCommas(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "ASCII commas unchanged",
			input:    "word1,word2,word3",
			expected: "word1,word2,word3",
		},
		{
			name:     "Japanese ideographic commas to ASCII",
			input:    "ばか、バカ、馬鹿",
			expected: "ばか,バカ,馬鹿",
		},
		{
			name:     "Japanese full-width commas to ASCII",
			input:    "ばか，バカ，馬鹿",
			expected: "ばか,バカ,馬鹿",
		},
		{
			name:     "Mixed comma types normalized",
			input:    "word1,ばか、バカ，word2",
			expected: "word1,ばか,バカ,word2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeWordListCommas(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
