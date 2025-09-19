package main

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

func TestKoreanProfanityFilter(t *testing.T) {
	p := Plugin{
		configuration: &configuration{
			CensorCharacter: "*",
			RejectPosts:     false,
			BadWordsList:    "바보,멍청이,똥개,병신",
			ExcludeBots:     false,
		},
	}
	p.badWordsRegex = regexp.MustCompile(wordListToRegex(p.getConfiguration().BadWordsList))

	t.Run("korean hangul profanity word matches", func(t *testing.T) {
		in := &model.Post{
			Message: "너는 바보야!",
		}
		expected := &model.Post{
			Message: "너는 **야!",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Korean profanity word '바보' should be replaced with '**'")
	})

	t.Run("korean longer word profanity matches", func(t *testing.T) {
		in := &model.Post{
			Message: "정말 멍청이구나.",
		}
		expected := &model.Post{
			Message: "정말 ***구나.",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Korean profanity word '멍청이' should be replaced with '***'")
	})

	t.Run("korean words with english mixed content", func(t *testing.T) {
		in := &model.Post{
			Message: "Hello 바보 world and goodbye 멍청이 person",
		}
		expected := &model.Post{
			Message: "Hello ** world and goodbye *** person",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Korean words '바보' and '멍청이' in mixed content should be replaced correctly")
	})

	t.Run("korean character context sensitivity", func(t *testing.T) {
		in := &model.Post{
			Message: "이 똥개가 뭔가요?",
		}
		expected := &model.Post{
			Message: "이 **가 뭔가요?",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Korean profanity word '똥개' should be replaced with '**' in sentence context")
	})

	t.Run("utf-8 character length handling", func(t *testing.T) {
		// Test that replacement length matches visual character count
		testCases := []struct {
			name           string
			input          string
			word           string
			expectedOutput string
		}{
			{"hangul 2 chars", "바보", "바보", "**"},
			{"hangul 3 chars", "멍청이", "멍청이", "***"},
			{"hangul 2 chars alt", "똥개", "똥개", "**"},
			{"hangul 2 chars offensive", "병신", "병신", "**"},
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
