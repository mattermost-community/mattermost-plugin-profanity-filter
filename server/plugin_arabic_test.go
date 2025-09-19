package main

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

func TestArabicProfanityFilter(t *testing.T) {
	p := Plugin{
		configuration: &configuration{
			CensorCharacter: "*",
			RejectPosts:     false,
			BadWordsList:    "أحمق,غبي,حمار,جاهل",
			ExcludeBots:     false,
		},
	}
	p.badWordsRegex = regexp.MustCompile(wordListToRegex(p.getConfiguration().BadWordsList))

	t.Run("arabic profanity word matches", func(t *testing.T) {
		in := &model.Post{
			Message: "أنت أحمق!",
		}
		expected := &model.Post{
			Message: "أنت ****!",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Arabic profanity word 'أحمق' should be replaced with '****'")
	})

	t.Run("arabic second word profanity matches", func(t *testing.T) {
		in := &model.Post{
			Message: "هذا غبي جداً.",
		}
		expected := &model.Post{
			Message: "هذا *** جداً.",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Arabic profanity word 'غبي' should be replaced with '***'")
	})

	t.Run("arabic words with english mixed content", func(t *testing.T) {
		in := &model.Post{
			Message: "Hello أحمق world and goodbye غبي person",
		}
		expected := &model.Post{
			Message: "Hello **** world and goodbye *** person",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Arabic words 'أحمق' and 'غبي' in mixed content should be replaced correctly")
	})

	t.Run("arabic rtl text profanity matches", func(t *testing.T) {
		in := &model.Post{
			Message: "لا تكن حمار في هذا الموضوع.",
		}
		expected := &model.Post{
			Message: "لا تكن **** في هذا الموضوع.",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Arabic profanity word 'حمار' should be replaced with '****' in RTL text")
	})

	t.Run("utf-8 character length handling", func(t *testing.T) {
		// Test that replacement length matches visual character count
		testCases := []struct {
			name           string
			input          string
			word           string
			expectedOutput string
		}{
			{"arabic 4 chars fool", "أحمق", "أحمق", "****"},
			{"arabic 3 chars stupid", "غبي", "غبي", "***"},
			{"arabic 4 chars donkey", "حمار", "حمار", "****"},
			{"arabic 4 chars ignorant", "جاهل", "جاهل", "****"},
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
