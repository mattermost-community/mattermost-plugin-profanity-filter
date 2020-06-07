package main

import (
	"net/http"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
)

type Plugin struct {
	plugin.MattermostPlugin

	configurationLock sync.RWMutex
	configuration     *configuration

	badWords map[string]bool
}

func main() {
	plugin.ClientMain(&Plugin{})
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	switch path := r.URL.Path; path {
	default:
		http.NotFound(w, r)
	}
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
				return nil, "Profane word not allowed: " + word
			}
			words[i] = strings.Repeat(configuration.CensorCharacter, len(word))
		}
	}

	post.Message = strings.Join(words, " ")
	return post, ""
}

func (p *Plugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
	return p.FilterPost(post)
}

func (p *Plugin) MessageWillBeUpdated(c *plugin.Context, newPost *model.Post, _ *model.Post) (*model.Post, string) {
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
