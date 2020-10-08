package main

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

func TestMessageWillBePosted(t *testing.T) {
	p := Plugin{
		configuration: &configuration{
			CensorCharacter: "*",
			RejectPosts:     false,
			BadWordsList:    "def ghi,abc",
		},
	}
	p.badWordsRegex = regexp.MustCompile(wordListToRegex(p.getConfiguration().BadWordsList))

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
			Message: "123 ABC AbC 456",
		}
		out := &model.Post{
			Message: "123 *** *** 456",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)
		assert.Equal(t, out, rpost)
	})

	t.Run("word with spaces matches", func(t *testing.T) {
		in := &model.Post{
			Message: "123 def ghi 456",
		}
		out := &model.Post{
			Message: "123 ******* 456",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)
		assert.Equal(t, out, rpost)
	})

	t.Run("word matches with punctuation", func(t *testing.T) {
		in := &model.Post{
			Message: "123 abc, 456",
		}
		out := &model.Post{
			Message: "123 ***, 456",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)
		assert.Equal(t, out, rpost)
	})

	t.Run("word shouldn't match because it in another word", func(t *testing.T) {
		in := &model.Post{
			Message: "helloabcworld helloabc abchello",
		}
		out := &model.Post{
			Message: "helloabcworld helloabc abchello",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)
		assert.Equal(t, out, rpost)
	})
}
