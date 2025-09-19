package main

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

func TestMultilingualProfanityFilter(t *testing.T) {
	p := Plugin{
		configuration: &configuration{
			CensorCharacter: "*",
			RejectPosts:     false,
			// Mixed word list with multiple languages
			BadWordsList: "bad,stupid,ばか,バカ,馬鹿,笨蛋,白痴,바보,멍청이,дурак,идиот,أحمق,غبي",
			ExcludeBots:  false,
		},
	}
	p.badWordsRegex = regexp.MustCompile(wordListToRegex(p.getConfiguration().BadWordsList))

	t.Run("multilingual mixed content", func(t *testing.T) {
		in := &model.Post{
			Message: "Hello bad ばか and 笨蛋 바보 дурак أحمق world",
		}
		expected := &model.Post{
			Message: "Hello *** ** and ** ** ***** **** world",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Multiple languages should be filtered correctly in one message")
	})

	t.Run("complex multilingual sentence", func(t *testing.T) {
		in := &model.Post{
			Message: "Don't be stupid like that ばか person who acts like 笨蛋.",
		}
		expected := &model.Post{
			Message: "Don't be ****** like that ** person who acts like **.",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Complex multilingual sentences should be filtered correctly")
	})

	t.Run("ascii and rune word separation", func(t *testing.T) {
		// Test that ASCII and rune words are processed differently
		in := &model.Post{
			Message: "This bad guy is a real 바보 and total идиот!",
		}
		expected := &model.Post{
			Message: "This *** guy is a real ** and total *****!",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "ASCII and rune words should both be filtered correctly")
	})

	t.Run("different script families together", func(t *testing.T) {
		in := &model.Post{
			Message: "Japanese バカ, Chinese 白痴, Korean 멍청이, Russian идиот, Arabic غبي",
		}
		expected := &model.Post{
			Message: "Japanese **, Chinese **, Korean ***, Russian *****, Arabic ***",
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Different script families should all be handled correctly")
	})

	t.Run("performance with large multilingual word list", func(t *testing.T) {
		// Test performance doesn't degrade significantly with mixed language lists
		largeWordList := "bad,stupid,idiot,fool,damn,hell,ばか,バカ,馬鹿,あほ,間抜け,笨蛋,白痴,傻瓜,蠢貨,愚蠢,바보,멍청이,똥개,병신,어리석은,дурак,идиот,тупой,глупец,болван,أحمق,غبي,حمار,جاهل,سفيه"

		p2 := Plugin{
			configuration: &configuration{
				CensorCharacter: "*",
				RejectPosts:     false,
				BadWordsList:    largeWordList,
				ExcludeBots:     false,
			},
		}
		p2.badWordsRegex = regexp.MustCompile(wordListToRegex(p2.getConfiguration().BadWordsList))

		in := &model.Post{
			Message: "This bad ばか is 笨蛋 and 바보 like дурак or أحمق person",
		}
		expected := &model.Post{
			Message: "This *** ** is ** and ** like ***** or **** person",
		}

		rpost, s := p2.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Large multilingual word list should still filter correctly")
	})

	t.Run("utf-8 character count accuracy across languages", func(t *testing.T) {
		// Test that character count replacement is accurate across different Unicode scripts
		testCases := []struct {
			name     string
			input    string
			expected string
			wordList string
		}{
			{
				"mixed lengths",
				"bad ばか 笨蛋 바보 дурак أحمق",
				"*** ** ** ** ***** ****",
				"bad,ばか,笨蛋,바보,дурак,أحمق",
			},
			{
				"varying byte lengths",
				"go バカ 白痴 멍청이 идиот",
				"** ** ** *** *****",
				"go,バカ,白痴,멍청이,идиот",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				testPlugin := Plugin{
					configuration: &configuration{
						CensorCharacter: "*",
						RejectPosts:     false,
						BadWordsList:    tc.wordList,
						ExcludeBots:     false,
					},
				}
				testPlugin.badWordsRegex = regexp.MustCompile(wordListToRegex(testPlugin.getConfiguration().BadWordsList))

				in := &model.Post{Message: tc.input}
				rpost, s := testPlugin.MessageWillBePosted(&plugin.Context{}, in)
				assert.Empty(t, s)

				t.Logf("Input: %s", tc.input)
				t.Logf("Output: %s", rpost.Message)
				t.Logf("Expected: %s", tc.expected)

				assert.Equal(t, tc.expected, rpost.Message, "Character count should be accurate across all Unicode scripts")
			})
		}
	})

	t.Run("edge cases with mixed scripts", func(t *testing.T) {
		in := &model.Post{
			Message: "ばかstupid笨蛋", // No spaces between words
		}
		expected := &model.Post{
			Message: "**********", // All detected words should be replaced: ばか(2) + stupid(6) + 笨蛋(2) = 10 chars
		}

		rpost, s := p.MessageWillBePosted(&plugin.Context{}, in)
		assert.Empty(t, s)

		t.Logf("Input: %s", in.Message)
		t.Logf("Output: %s", rpost.Message)
		t.Logf("Expected: %s", expected.Message)

		assert.Equal(t, expected.Message, rpost.Message, "Mixed scripts without spaces should detect and replace all profanity words")
	})
}
