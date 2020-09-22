package main

import (
	"fmt"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	badWords map[string]bool
}

func (p *Plugin) WordIsBad(word string) bool {
	_, ok := p.badWords[strings.ToLower(removeAccents(word))]
	return ok
}

func (p *Plugin) FilterPost(post *model.Post) (*model.Post, string) {
	configuration := p.getConfiguration()

	message := post.Message
	words := strings.Split(message, " ")
	for i, word := range words {
		if p.WordIsBad(word) {
			if configuration.RejectPosts {
				p.API.SendEphemeralPost(post.UserId, &model.Post{
					ChannelId: post.ChannelId,
					Message:   fmt.Sprintf(configuration.WarningMessage, word),
				})
				return nil, "Profane word not allowed: " + word
			}
			words[i] = strings.Repeat(configuration.CensorCharacter, len(word))
		}
	}

	post.Message = strings.Join(words, " ")
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
