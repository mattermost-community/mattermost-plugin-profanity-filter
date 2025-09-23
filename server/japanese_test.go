package main

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

func TestJapaneseProfanityFilter(t *testing.T) {
	p := Plugin{
		configuration: &configuration{
			CensorCharacter: "*",
			RejectPosts:     false,
			BadWordsList:    "ばか,バカ,馬鹿,クソ野郎,MySQL",
			ExcludeBots:     false,
		},
	}
	p.badWordsRegex = regexp.MustCompile(wordListToRegex(p.getConfiguration().BadWordsList))

	t.Run("hiragana profanity word matches", func(t *testing.T) {
		in := &model.Post{
			Message: "あなたはばかです。",
		}
		expected := &model.Post{
			Message: "あなたは**です。",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Hiragana profanity word 'ばか' should be replaced with '**'")
	})

	t.Run("katakana profanity word matches", func(t *testing.T) {
		in := &model.Post{
			Message: "あなたはバカです。",
		}
		expected := &model.Post{
			Message: "あなたは**です。",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Katakana profanity word 'バカ' should be replaced with '**'")
	})

	t.Run("kanji profanity word matches", func(t *testing.T) {
		in := &model.Post{
			Message: "あなたは馬鹿です。",
		}
		expected := &model.Post{
			Message: "あなたは**です。",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Kanji profanity word '馬鹿' should be replaced with '**'")
	})

	t.Run("mixed script profanity word matches", func(t *testing.T) {
		in := &model.Post{
			Message: "このクソ野郎が！",
		}
		expected := &model.Post{
			Message: "この****が！",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Mixed script profanity word 'クソ野郎' should be replaced with '****'")
	})

	t.Run("japanese words with english mixed content", func(t *testing.T) {
		in := &model.Post{
			Message: "Hello ばか world and goodbye バカ person",
		}
		expected := &model.Post{
			Message: "Hello ** world and goodbye ** person",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Japanese words 'ばか' and 'バカ' in mixed content should be replaced with '**'")
	})

	t.Run("utf-8 character length handling", func(t *testing.T) {
		// Test that replacement length matches visual character count
		testCases := []struct {
			name           string
			input          string
			word           string
			expectedOutput string
		}{
			{"hiragana 2 chars", "ばか", "ばか", "**"},
			{"katakana 2 chars", "バカ", "バカ", "**"},
			{"kanji 2 chars", "馬鹿", "馬鹿", "**"},
			{"mixed 4 chars", "クソ野郎", "クソ野郎", "****"},
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

	t.Run("replaces Japanese and English in the same sentence", func(t *testing.T) {
		in := "MySQLを使ってPostgresの代わりにするのはバカです。"
		rpost, s := p.MessageWillBePosted(&plugin.Context{}, &model.Post{Message: in})
		assert.Empty(t, s)
		expected := "*****を使ってPostgresの代わりにするのは**です。"
		assert.Equal(t, expected, rpost.Message, "Japanese word 'ばか' and English word 'MySQL' should be replaced with asterisks")
	})
}
