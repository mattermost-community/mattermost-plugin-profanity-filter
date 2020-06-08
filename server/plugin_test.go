package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

func TestMessageWillBePosted(t *testing.T) {
	p := Plugin{
		badWords: map[string]bool{
			"abc": true,
		},
		configuration: &configuration{
			CensorCharacter: "*",
			RejectPosts:     false,
		},
	}

	t.Run("word matches", func(t *testing.T) {
		in := &model.Post{
			Message: "123 abc 456",
		}
		out := &model.Post{
			Message: "123 *** 456",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)
		assert.Equal(t, out, rpost)
	})

	t.Run("word matches case-insensitive", func(t *testing.T) {
		in := &model.Post{
			Message: "123 ABC 456",
		}
		out := &model.Post{
			Message: "123 *** 456",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)
		assert.Equal(t, out, rpost)
	})
}
