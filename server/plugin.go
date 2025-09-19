package main

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	badWordsRegex *regexp.Regexp
}

func (p *Plugin) FilterPost(post *model.Post) (*model.Post, string) {
	configuration := p.getConfiguration()
	_, fromBot := post.GetProps()["from_bot"]

	if configuration.ExcludeBots && fromBot {
		return post, ""
	}

	// Use new hybrid detection system
	detectedBadWords := p.detectAllProfanityWords(post.Message, configuration.BadWordsList)

	if len(detectedBadWords) == 0 {
		return post, ""
	}

	if configuration.RejectPosts {
		p.API.SendEphemeralPost(post.UserId, &model.Post{
			ChannelId: post.ChannelId,
			Message:   fmt.Sprintf(configuration.WarningMessage, strings.Join(detectedBadWords, ", ")),
			RootId:    post.RootId,
		})

		return nil, fmt.Sprintf("Profane word not allowed: %s", strings.Join(detectedBadWords, ", "))
	}

	// Use rune-based replacement for correct character count
	for _, word := range detectedBadWords {
		post.Message = strings.ReplaceAll(
			post.Message,
			word,
			strings.Repeat(p.getConfiguration().CensorCharacter, runeLength(word)),
		)
	}

	return post, ""
}

func (p *Plugin) MessageWillBePosted(_ *plugin.Context, post *model.Post) (*model.Post, string) {
	return p.FilterPost(post)
}

func (p *Plugin) MessageWillBeUpdated(_ *plugin.Context, newPost *model.Post, _ *model.Post) (*model.Post, string) {
	return p.FilterPost(newPost)
}

func removeAccents(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	output, _, e := transform.String(t, s)
	if e != nil {
		return s
	}

	return output
}

// requiresRuneProcessing checks if text contains non-ASCII characters that need rune-based processing
func requiresRuneProcessing(text string) bool {
	for _, r := range text {
		if r > 127 { // Non-ASCII character
			return true
		}
	}
	return false
}

// runeLength returns the number of visual characters (runes) in a string
func runeLength(s string) int {
	return len([]rune(s))
}

// separateWordsByType separates a word list into ASCII words and rune words
func separateWordsByType(wordList []string) (asciiWords, runeWords []string) {
	for _, word := range wordList {
		word = strings.TrimSpace(word)
		if word == "" {
			continue
		}
		if requiresRuneProcessing(word) {
			runeWords = append(runeWords, word)
		} else {
			asciiWords = append(asciiWords, word)
		}
	}
	return asciiWords, runeWords
}

// detectAllProfanityWords uses hybrid detection for both ASCII and rune words
func (p *Plugin) detectAllProfanityWords(text, wordList string) []string {
	words := strings.Split(wordList, ",")
	asciiWords, runeWords := separateWordsByType(words)

	var detected []string

	// ASCII words: Use existing regex (fast & precise)
	if len(asciiWords) > 0 {
		detected = append(detected, p.detectASCIIWords(text, asciiWords)...)
	}

	// Rune words: Use substring matching (universal support)
	if len(runeWords) > 0 {
		detected = append(detected, p.detectRuneWords(text, runeWords)...)
	}

	return detected
}

// detectASCIIWords uses regex with word boundaries for ASCII words
func (p *Plugin) detectASCIIWords(text string, asciiWords []string) []string {
	if len(asciiWords) == 0 {
		return []string{}
	}

	// Use existing regex logic with \b boundaries
	regexStr := fmt.Sprintf(`(?mi)\b(%s)\b`, strings.Join(asciiWords, "|"))
	regex, err := regexp.Compile(regexStr)
	if err != nil {
		return []string{}
	}

	return regex.FindAllString(removeAccents(text), -1)
}

// detectRuneWords uses substring matching for rune words (Japanese, Russian, etc.)
func (p *Plugin) detectRuneWords(text string, runeWords []string) []string {
	var detected []string
	// Don't apply removeAccents to rune words as it can corrupt non-Latin scripts
	textLower := strings.ToLower(text)

	for _, word := range runeWords {
		wordLower := strings.ToLower(strings.TrimSpace(word))
		if wordLower != "" && strings.Contains(textLower, wordLower) {
			detected = append(detected, strings.TrimSpace(word))
		}
	}

	return detected
}
