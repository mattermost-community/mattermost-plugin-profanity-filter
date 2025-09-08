package main

import (
	"strings"
	"unicode"

	"github.com/gojp/kana"
)

// JapaneseTextNormalizer provides utilities for Japanese text processing
type JapaneseTextNormalizer struct{}

// NewJapaneseTextNormalizer creates a new instance of Japanese text normalizer
func NewJapaneseTextNormalizer() *JapaneseTextNormalizer {
	return &JapaneseTextNormalizer{}
}

// NormalizeJapaneseText converts text to multiple normalized forms for comprehensive matching
func (j *JapaneseTextNormalizer) NormalizeJapaneseText(text string) []string {
	if text == "" {
		return []string{text}
	}

	variations := make(map[string]bool)
	variations[text] = true // Always include original

	// Convert to hiragana if possible
	if hiragana := j.ToHiragana(text); hiragana != text {
		variations[hiragana] = true
	}

	// Convert to katakana if possible
	if katakana := j.ToKatakana(text); katakana != text {
		variations[katakana] = true
	}

	// Convert from romaji if it appears to be romaji
	if j.IsRomaji(text) {
		if hiragana := kana.RomajiToHiragana(text); hiragana != text {
			variations[hiragana] = true
		}
		if katakana := kana.RomajiToKatakana(text); katakana != text {
			variations[katakana] = true
		}
	}

	// Convert to romaji if it's Japanese text
	if j.ContainsJapanese(text) {
		if romaji := kana.KanaToRomaji(text); romaji != text {
			variations[romaji] = true
		}
	}

	// Convert full-width to half-width and vice versa
	if halfWidth := j.ToHalfWidth(text); halfWidth != text {
		variations[halfWidth] = true
	}
	if fullWidth := j.ToFullWidth(text); fullWidth != text {
		variations[fullWidth] = true
	}

	// Convert to slice
	result := make([]string, 0, len(variations))
	for variation := range variations {
		result = append(result, variation)
	}

	return result
}

// ToHiragana converts katakana characters to hiragana
func (j *JapaneseTextNormalizer) ToHiragana(text string) string {
	runes := []rune(text)
	for i, r := range runes {
		if j.IsKatakana(r) {
			// Convert katakana to hiragana by subtracting the offset
			if r >= 0x30A1 && r <= 0x30F6 {
				runes[i] = r - 0x60
			}
		}
	}
	return string(runes)
}

// ToKatakana converts hiragana characters to katakana
func (j *JapaneseTextNormalizer) ToKatakana(text string) string {
	runes := []rune(text)
	for i, r := range runes {
		if j.IsHiragana(r) {
			// Convert hiragana to katakana by adding the offset
			if r >= 0x3041 && r <= 0x3096 {
				runes[i] = r + 0x60
			}
		}
	}
	return string(runes)
}

// IsHiragana checks if a rune is hiragana
func (j *JapaneseTextNormalizer) IsHiragana(r rune) bool {
	return r >= 0x3041 && r <= 0x3096
}

// IsKatakana checks if a rune is katakana
func (j *JapaneseTextNormalizer) IsKatakana(r rune) bool {
	return (r >= 0x30A1 && r <= 0x30F6) || (r >= 0xFF66 && r <= 0xFF9F)
}

// IsKanji checks if a rune is kanji
func (j *JapaneseTextNormalizer) IsKanji(r rune) bool {
	return kana.IsKanji(string(r))
}

// IsRomaji checks if text appears to be romaji (Latin characters)
func (j *JapaneseTextNormalizer) IsRomaji(text string) bool {
	if text == "" {
		return false
	}

	// If text contains any Japanese characters, it's not pure romaji
	if j.ContainsJapanese(text) {
		return false
	}

	latinCount := 0
	totalCount := 0

	for _, r := range text {
		if unicode.IsLetter(r) {
			totalCount++
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				latinCount++
			}
		}
	}

	// Consider it romaji if more than 80% of letters are Latin
	return totalCount > 0 && float64(latinCount)/float64(totalCount) > 0.8
}

// ContainsJapanese checks if text contains Japanese characters
func (j *JapaneseTextNormalizer) ContainsJapanese(text string) bool {
	for _, r := range text {
		if j.IsHiragana(r) || j.IsKatakana(r) || j.IsKanji(r) {
			return true
		}
	}
	return false
}

// ToHalfWidth converts full-width characters to half-width
func (j *JapaneseTextNormalizer) ToHalfWidth(text string) string {
	var result strings.Builder

	for _, r := range text {
		// Convert full-width ASCII to half-width
		switch {
		case r >= 0xFF01 && r <= 0xFF5E:
			result.WriteRune(r - 0xFEE0)
		case r == 0xFF5F:
			result.WriteRune(0x2985) // ⦅
		case r == 0xFF60:
			result.WriteRune(0x2986) // ⦆
		default:
			result.WriteRune(r)
		}
	}

	return result.String()
}

// ToFullWidth converts half-width characters to full-width
func (j *JapaneseTextNormalizer) ToFullWidth(text string) string {
	var result strings.Builder

	for _, r := range text {
		// Convert half-width ASCII to full-width
		if r >= 0x21 && r <= 0x7E {
			result.WriteRune(r + 0xFEE0)
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// GenerateMatchingVariations creates all possible text variations for matching
func (j *JapaneseTextNormalizer) GenerateMatchingVariations(badWord string) []string {
	variations := j.NormalizeJapaneseText(badWord)

	// Also generate accent-removed versions
	var allVariations []string
	for _, variation := range variations {
		allVariations = append(allVariations, variation)
		if normalized := removeAccents(variation); normalized != variation {
			allVariations = append(allVariations, normalized)
		}
	}

	// Remove duplicates
	seen := make(map[string]bool)
	var unique []string
	for _, v := range allVariations {
		if !seen[v] && v != "" {
			seen[v] = true
			unique = append(unique, v)
		}
	}

	return unique
}
