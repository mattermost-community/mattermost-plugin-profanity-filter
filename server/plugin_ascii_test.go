package main

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

func TestMessageWillBePosted(t *testing.T) {
	p := Plugin{
		configuration: &configuration{
			CensorCharacter: "*",
			RejectPosts:     false,
			BadWordsList:    "def ghi,abc",
			ExcludeBots:     true,
		},
	}
	p.badWordsRegex = regexp.MustCompile(wordListToRegex(p.getConfiguration().BadWordsList))

	t.Run("basic word replacement", func(t *testing.T) {
		in := &model.Post{
			Message: "123 abc 456",
		}
		expected := &model.Post{
			Message: "123 *** 456",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Word 'abc' should be replaced with '***'")
	})

	t.Run("case-insensitive matching", func(t *testing.T) {
		in := &model.Post{
			Message: "123 ABC AbC 456",
		}
		expected := &model.Post{
			Message: "123 *** *** 456",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Words 'ABC' and 'AbC' should be replaced case-insensitively")
	})

	t.Run("multi-word phrase replacement", func(t *testing.T) {
		in := &model.Post{
			Message: "123 def ghi 456",
		}
		expected := &model.Post{
			Message: "123 ******* 456",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Multi-word phrase 'def ghi' should be replaced with '*******'")
	})

	t.Run("word with punctuation", func(t *testing.T) {
		in := &model.Post{
			Message: "123 abc, 456",
		}
		expected := &model.Post{
			Message: "123 ***, 456",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Word 'abc' followed by punctuation should be replaced correctly")
	})

	t.Run("word boundary detection", func(t *testing.T) {
		in := &model.Post{
			Message: "helloabcworld helloabc abchello",
		}
		expected := &model.Post{
			Message: "helloabcworld helloabc abchello",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Words containing 'abc' as substring should not be replaced")
	})

	t.Run("bot message exclusion", func(t *testing.T) {
		in := &model.Post{
			Message: "abc",
		}
		in.AddProp("from_bot", "true")
		expected := &model.Post{
			Message: "abc",
		}
		expected.AddProp("from_bot", "true")

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s (from_bot: true)", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected, rpost, "Bot messages should not be filtered when ExcludeBots is true")
	})
}
