package main

import (
	"net/http"
	"strings"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
)

type Plugin struct {
	plugin.MattermostPlugin
	badWords map[string]bool

	RejectPosts     bool
	CensorCharacter string
}

func main() {
	plugin.ClientMain(&Plugin{})
}

func (p *Plugin) OnActivate() error {
	p.badWords = make(map[string]bool, len(badWords))
	for _, word := range badWords {
		p.badWords[word] = true
	}

	return nil
}

func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	switch path := r.URL.Path; path {
	default:
		http.NotFound(w, r)
	}
}

func (p *Plugin) WordIsBad(word string) bool {
	_, ok := p.badWords[word]
	return ok
}

func (p *Plugin) FilterPost(post *model.Post) (*model.Post, string) {
	message := post.Message
	words := strings.Split(message, " ")
	for i, word := range words {
		if p.WordIsBad(word) {
			if p.RejectPosts {
				return nil, "Profane word not allowed: " + word
			}
			words[i] = strings.Repeat(p.CensorCharacter, len(word))
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
