package main

import (
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
)

type Plugin struct {
	api           plugin.API
	configuration atomic.Value
	badWords      map[string]bool
}

func (p *Plugin) OnActivate(api plugin.API) error {
	p.api = api
	if err := p.OnConfigurationChange(); err != nil {
		return err
	}

	config := p.config()
	config.SetDefaults()
	if err := config.IsValid(); err != nil {
		return err
	}

	p.badWords = make(map[string]bool, len(badWords))
	for _, word := range badWords {
		p.badWords[word] = true
	}

	return nil
}

func (p *Plugin) config() *Configuration {
	return p.configuration.Load().(*Configuration)
}

func (p *Plugin) OnConfigurationChange() error {
	var configuration Configuration
	err := p.api.LoadPluginConfiguration(&configuration)
	p.configuration.Store(&configuration)
	return err
}

func (p *Plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	config := p.config()
	if err := config.IsValid(); err != nil {
		http.Error(w, "This plugin is not configured.", http.StatusNotImplemented)
		return
	}

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
			if p.config().RejectPosts {
				return nil, "Profane word not allowed: " + word
			}
			words[i] = strings.Repeat(p.config().CensorCharacter, len(word))
		}
	}

	post.Message = strings.Join(words, " ")
	return post, ""
}

func (p *Plugin) MessageWillBePosted(post *model.Post) (*model.Post, string) {
	return p.FilterPost(post)
}

func (p *Plugin) MessageWillBeUpdated(newPost *model.Post, _ *model.Post) (*model.Post, string) {
	return p.FilterPost(newPost)
}
