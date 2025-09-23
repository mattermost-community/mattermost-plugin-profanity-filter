package main

import (
	"testing"

	"github.com/mattermost/mattermost/server/public/plugin/plugintest"
	"github.com/stretchr/testify/mock"
)

// createMockPlugin creates a plugin with a mocked API for testing
func createMockPlugin(_ *testing.T, config *configuration) *Plugin {
	api := &plugintest.API{}

	// Mock the LoadPluginConfiguration method with proper parameter matching
	api.On("LoadPluginConfiguration", mock.AnythingOfType("*main.configuration")).Return(nil).Run(func(args mock.Arguments) {
		dest := args.Get(0).(*configuration)
		// Copy the mock config fields to the destination
		dest.CensorCharacter = config.CensorCharacter
		dest.RejectPosts = config.RejectPosts
		dest.BadWordsList = config.BadWordsList
		dest.ExcludeBots = config.ExcludeBots
		dest.WarningMessage = config.WarningMessage
	})

	plugin := &Plugin{}
	plugin.SetAPI(api)

	return plugin
}
