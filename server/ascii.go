package main

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func removeAccents(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	output, _, e := transform.String(t, s)
	if e != nil {
		return s
	}

	return output
}

// detectASCIIWords uses regex with word boundaries for ASCII words
func (p *Plugin) detectASCIIWords(text string, asciiWords []string) []string {
	regex := p.getASCIIWordsRegex()
	if regex == nil {
		return []string{}
	}

	return regex.FindAllString(removeAccents(text), -1)
}

// separateASCIIAndJapanese separates a word list into ASCII words and Japanese words
func separateASCIIAndJapanese(wordList []string) (asciiWords, japaneseWords []string) {
	for _, word := range wordList {
		word = strings.TrimSpace(word)
		if word == "" {
			continue
		}
		if isJapaneseWord(word) {
			japaneseWords = append(japaneseWords, word)
		} else {
			// Treat everything non-Japanese as ASCII (including other non-ASCII languages)
			asciiWords = append(asciiWords, word)
		}
	}
	return asciiWords, japaneseWords
}
