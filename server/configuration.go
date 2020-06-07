package main

import (
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

type configuration struct {
	RejectPosts     bool
	CensorCharacter string
	BadWordsList    string
}

func (c *configuration) Clone() *configuration {
	var clone = *c
	return &clone
}

func (p *Plugin) getConfiguration() *configuration {
	p.configurationLock.RLock()
	defer p.configurationLock.RUnlock()

	if p.configuration == nil {
		return &configuration{}
	}

	return p.configuration
}

func (p *Plugin) setConfiguration(configuration *configuration) {
	p.configurationLock.Lock()
	defer p.configurationLock.Unlock()

	if configuration != nil && p.configuration == configuration {
		if reflect.ValueOf(*configuration).NumField() == 0 {
			return
		}

		panic("setConfiguration called with the existing configuration")
	}

	p.configuration = configuration
}

func (p *Plugin) OnConfigurationChange() error {
	configuration := p.getConfiguration().Clone()

	if err := p.API.LoadPluginConfiguration(configuration); err != nil {
		return errors.Wrap(err, "failed to load plugin configuration")
	}

	p.setConfiguration(configuration)

	badWordsFromSettings := strings.Split(configuration.BadWordsList, " ")
	p.badWords = make(map[string]bool, len(badWordsFromSettings))
	for _, word := range badWordsFromSettings {
		p.badWords[strings.ToLower(removeAccents(word))] = true
	}

	return nil
}
