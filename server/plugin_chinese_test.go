package main

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

func TestChineseProfanityFilter(t *testing.T) {
	p := Plugin{
		configuration: &configuration{
			CensorCharacter: "*",
			RejectPosts:     false,
			BadWordsList:    "笨蛋,白痴,白癡,傻瓜",
			ExcludeBots:     false,
		},
	}
	p.badWordsRegex = regexp.MustCompile(wordListToRegex(p.getConfiguration().BadWordsList))

	t.Run("simplified chinese profanity word matches", func(t *testing.T) {
		in := &model.Post{
			Message: "你是个笨蛋！",
		}
		expected := &model.Post{
			Message: "你是个**！",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Simplified Chinese profanity word '笨蛋' should be replaced with '**'")
	})

	t.Run("traditional chinese profanity word matches", func(t *testing.T) {
		in := &model.Post{
			Message: "你是白癡嗎？",
		}
		expected := &model.Post{
			Message: "你是**嗎？",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Traditional Chinese profanity word '白癡' should be replaced with '**'")
	})

	t.Run("chinese words with english mixed content", func(t *testing.T) {
		in := &model.Post{
			Message: "Hello 笨蛋 world and goodbye 白痴 person",
		}
		expected := &model.Post{
			Message: "Hello ** world and goodbye ** person",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Chinese words '笨蛋' and '白痴' in mixed content should be replaced with '**'")
	})

	t.Run("chinese character context sensitivity", func(t *testing.T) {
		in := &model.Post{
			Message: "这个傻瓜不明白。",
		}
		expected := &model.Post{
			Message: "这个**不明白。",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Chinese profanity word '傻瓜' should be replaced with '**' in sentence context")
	})

	t.Run("utf-8 character length handling", func(t *testing.T) {
		// Test that replacement length matches visual character count
		testCases := []struct {
			name           string
			input          string
			word           string
			expectedOutput string
		}{
			{"simplified 2 chars", "笨蛋", "笨蛋", "**"},
			{"traditional 2 chars", "白癡", "白癡", "**"},
			{"simplified 2 chars alt", "白痴", "白痴", "**"},
			{"simplified 2 chars fool", "傻瓜", "傻瓜", "**"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				in := &model.Post{Message: tc.input}

				rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
				assert.Empty(t, s)

				t.Logf("Input: %s (word: %s)", tc.input, tc.word)
				t.Logf("Output: %s", rpost.Message)
				t.Logf("Expected: %s", tc.expectedOutput)

				// Count visual characters vs byte length
				runeCount := len([]rune(tc.word))
				byteCount := len(tc.word)

				t.Logf("Word '%s': Visual chars=%d, Byte count=%d", tc.word, runeCount, byteCount)

				// This test verifies that replacement uses visual character count, not byte count
				assert.Equal(t, tc.expectedOutput, rpost.Message,
					"Word '%s' should be replaced with %d asterisks (visual character count), not %d (byte count)",
					tc.word, runeCount, byteCount)
			})
		}
	})
}
