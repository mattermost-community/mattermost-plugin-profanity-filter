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

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

// Plugin represents the profanity filter plugin instance.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	badWordsRegex      *regexp.Regexp
	japaneseNormalizer *JapaneseTextNormalizer
}

// FilterPost processes a post to filter profanity based on configured settings.
func (p *Plugin) FilterPost(post *model.Post) (*model.Post, string) {
	config := p.getConfiguration()
	_, fromBot := post.GetProps()["from_bot"]

	if config.ExcludeBots && fromBot {
		return post, ""
	}

	// Get text variations for matching
	textVariations := []string{post.Message}

	// Add Japanese text normalization if enabled
	if config.EnableJapaneseSupport && p.japaneseNormalizer != nil {
		japaneseVariations := p.japaneseNormalizer.NormalizeJapaneseText(post.Message)
		textVariations = append(textVariations, japaneseVariations...)
	}

	// Also add accent-removed version for backward compatibility
	postMessageWithoutAccents := removeAccents(post.Message)
	textVariations = append(textVariations, postMessageWithoutAccents)

	// Check all text variations against regex
	var detectedBadWords []string
	detectedWordsMap := make(map[string]bool)

	for _, textVariation := range textVariations {
		if p.badWordsRegex.MatchString(textVariation) {
			matches := p.badWordsRegex.FindAllString(textVariation, -1)
			for _, match := range matches {
				if !detectedWordsMap[match] {
					detectedWordsMap[match] = true
					detectedBadWords = append(detectedBadWords, match)
				}
			}
		}
	}

	if len(detectedBadWords) == 0 {
		return post, ""
	}

	if config.RejectPosts {
		p.API.SendEphemeralPost(post.UserId, &model.Post{
			ChannelId: post.ChannelId,
			Message:   fmt.Sprintf(config.WarningMessage, strings.Join(detectedBadWords, ", ")),
			RootId:    post.RootId,
		})

		return nil, fmt.Sprintf("Profane word not allowed: %s", strings.Join(detectedBadWords, ", "))
	}

	// Censor detected words in the original message
	for _, word := range detectedBadWords {
		// Try to replace in original message and its variations
		for _, variation := range textVariations {
			if strings.Contains(variation, word) {
				// Find the word in the original message and replace it
				post.Message = p.replaceWordInText(post.Message, word, strings.Repeat(config.CensorCharacter, len(word)))
				break
			}
		}
	}

	return post, ""
}

// replaceWordInText intelligently replaces profane words considering Japanese text boundaries
func (p *Plugin) replaceWordInText(text, badWord, replacement string) string {
	config := p.getConfiguration()

	// Simple replacement for non-Japanese mode or if Japanese support is disabled
	if !config.EnableJapaneseSupport || p.japaneseNormalizer == nil {
		return strings.ReplaceAll(text, badWord, replacement)
	}

	// For Japanese text, we need to be more careful about word boundaries
	// Since Japanese doesn't use spaces, we'll do a direct replacement
	// but consider different script variations

	if p.japaneseNormalizer.ContainsJapanese(text) || p.japaneseNormalizer.ContainsJapanese(badWord) {
		// Generate all variations of the bad word for comprehensive replacement
		variations := p.japaneseNormalizer.GenerateMatchingVariations(badWord)

		result := text
		for _, variation := range variations {
			if strings.Contains(result, variation) {
				result = strings.ReplaceAll(result, variation, replacement)
			}
		}
		return result
	}

	// Fallback to simple replacement
	return strings.ReplaceAll(text, badWord, replacement)
}

// MessageWillBePosted is called when a message is about to be posted.
func (p *Plugin) MessageWillBePosted(_ *plugin.Context, post *model.Post) (*model.Post, string) {
	return p.FilterPost(post)
}

// MessageWillBeUpdated is called when a message is about to be updated.
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
