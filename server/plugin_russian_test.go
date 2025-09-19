package main

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

func TestRussianProfanityFilter(t *testing.T) {
	p := Plugin{
		configuration: &configuration{
			CensorCharacter: "*",
			RejectPosts:     false,
			BadWordsList:    "дурак,идиот,тупой,глупец",
			ExcludeBots:     false,
		},
	}
	p.badWordsRegex = regexp.MustCompile(wordListToRegex(p.getConfiguration().BadWordsList))

	t.Run("russian cyrillic profanity word matches", func(t *testing.T) {
		in := &model.Post{
			Message: "Ты дурак!",
		}
		expected := &model.Post{
			Message: "Ты *****!",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Russian profanity word 'дурак' should be replaced with '*****'")
	})

	t.Run("russian longer word profanity matches", func(t *testing.T) {
		in := &model.Post{
			Message: "Какой идиот это сделал?",
		}
		expected := &model.Post{
			Message: "Какой ***** это сделал?",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Russian profanity word 'идиот' should be replaced with '*****'")
	})

	t.Run("russian words with english mixed content", func(t *testing.T) {
		in := &model.Post{
			Message: "Hello дурак world and goodbye идиот person",
		}
		expected := &model.Post{
			Message: "Hello ***** world and goodbye ***** person",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Russian words 'дурак' and 'идиот' in mixed content should be replaced correctly")
	})

	t.Run("russian adjective profanity matches", func(t *testing.T) {
		in := &model.Post{
			Message: "Это очень тупой вопрос.",
		}
		expected := &model.Post{
			Message: "Это очень ***** вопрос.",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Russian profanity word 'тупой' should be replaced with '*****' in sentence context")
	})

	t.Run("utf-8 character length handling", func(t *testing.T) {
		// Test that replacement length matches visual character count
		testCases := []struct {
			name           string
			input          string
			word           string
			expectedOutput string
		}{
			{"cyrillic 5 chars", "дурак", "дурак", "*****"},
			{"cyrillic 5 chars idiot", "идиот", "идиот", "*****"},
			{"cyrillic 5 chars stupid", "тупой", "тупой", "*****"},
			{"cyrillic 6 chars fool", "глупец", "глупец", "******"},
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
