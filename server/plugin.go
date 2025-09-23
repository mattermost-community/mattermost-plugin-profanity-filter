package main

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

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

	// Use hybrid detection system that separates ASCII and non-ASCII word detection for better multilingual support
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

// runeLength returns the number of visual characters (runes) in a string
func runeLength(s string) int {
	return len([]rune(s))
}

// detectAllProfanityWords uses detection for ASCII and Japanese words
func (p *Plugin) detectAllProfanityWords(text, wordList string) []string {
	words := strings.Split(wordList, ",")
	asciiWords, japaneseWords := separateASCIIAndJapanese(words)

	var detected []string

	// ASCII words: Use existing regex (fast & precise)
	if len(asciiWords) > 0 {
		detected = append(detected, p.detectASCIIWords(text, asciiWords)...)
	}

	// Japanese words: Use tokenization + regex approach
	if len(japaneseWords) > 0 {
		detected = append(detected, p.detectJapaneseWordsWithTokenization(text, japaneseWords)...)
	}

	return detected
}
