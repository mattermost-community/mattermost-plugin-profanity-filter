package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

func TestJapaneseCommaSupport(t *testing.T) {
	// Test with Japanese ideographic commas (、)
	t.Run("Japanese ideographic comma support", func(t *testing.T) {
		config := &configuration{
			CensorCharacter: "*",
			RejectPosts:     false,
			BadWordsList:    "ばか、バカ、馬鹿、クソ野郎",
			ExcludeBots:     false,
		}

		p := createMockPlugin(t, config)
		err := p.OnConfigurationChange()
		assert.NoError(t, err)

		in := &model.Post{
			Message: "あなたはばかです。バカな人だ。",
		}
		expected := &model.Post{
			Message: "あなたは**です。**な人だ。",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Japanese words separated by ideographic commas should be filtered")
	})

	// Test with Japanese full-width commas (，)
	t.Run("Japanese full-width comma support", func(t *testing.T) {
		config := &configuration{
			CensorCharacter: "*",
			RejectPosts:     false,
			BadWordsList:    "ばか，バカ，馬鹿，クソ野郎",
			ExcludeBots:     false,
		}

		p := createMockPlugin(t, config)
		err := p.OnConfigurationChange()
		assert.NoError(t, err)

		in := &model.Post{
			Message: "あなたはばかです。バカな人だ。",
		}
		expected := &model.Post{
			Message: "あなたは**です。**な人だ。",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Japanese words separated by full-width commas should be filtered")
	})

	// Test with mixed comma types
	t.Run("Mixed comma types support", func(t *testing.T) {
		config := &configuration{
			CensorCharacter: "*",
			RejectPosts:     false,
			BadWordsList:    "bad,ばか、バカ，stupid",
			ExcludeBots:     false,
		}

		p := createMockPlugin(t, config)
		err := p.OnConfigurationChange()
		assert.NoError(t, err)

		in := &model.Post{
			Message: "This is bad, あなたはばかです。You are stupid and バカな人だ。",
		}
		expected := &model.Post{
			Message: "This is ***, あなたは**です。You are ****** and **な人だ。",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Mixed ASCII and Japanese words with different comma types should be filtered")
	})
}
