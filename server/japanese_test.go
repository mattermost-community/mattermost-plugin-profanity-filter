package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJapaneseTextNormalizer(t *testing.T) {
	normalizer := NewJapaneseTextNormalizer()

	t.Run("ToHiragana", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"バカ", "ばか"},             // katakana to hiragana
			{"カタカナ", "かたかな"},         // katakana to hiragana
			{"ひらがな", "ひらがな"},         // hiragana stays hiragana
			{"Romaji", "Romaji"},     // latin stays latin
			{"混合バカtext", "混合ばかtext"}, // mixed text
		}

		for _, test := range tests {
			result := normalizer.ToHiragana(test.input)
			assert.Equal(t, test.expected, result, "ToHiragana(%s) should equal %s", test.input, test.expected)
		}
	})

	t.Run("ToKatakana", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"ばか", "バカ"},             // hiragana to katakana
			{"ひらがな", "ヒラガナ"},         // hiragana to katakana
			{"バカ", "バカ"},             // katakana stays katakana
			{"Romaji", "Romaji"},     // latin stays latin
			{"混合ばかtext", "混合バカtext"}, // mixed text
		}

		for _, test := range tests {
			result := normalizer.ToKatakana(test.input)
			assert.Equal(t, test.expected, result, "ToKatakana(%s) should equal %s", test.input, test.expected)
		}
	})

	t.Run("IsHiragana", func(t *testing.T) {
		tests := []struct {
			input    rune
			expected bool
		}{
			{'あ', true},
			{'か', true},
			{'ん', true},
			{'ア', false},
			{'カ', false},
			{'漢', false},
			{'a', false},
		}

		for _, test := range tests {
			result := normalizer.IsHiragana(test.input)
			assert.Equal(t, test.expected, result, "IsHiragana(%c) should be %v", test.input, test.expected)
		}
	})

	t.Run("IsKatakana", func(t *testing.T) {
		tests := []struct {
			input    rune
			expected bool
		}{
			{'ア', true},
			{'カ', true},
			{'ン', true},
			{'あ', false},
			{'か', false},
			{'漢', false},
			{'a', false},
		}

		for _, test := range tests {
			result := normalizer.IsKatakana(test.input)
			assert.Equal(t, test.expected, result, "IsKatakana(%c) should be %v", test.input, test.expected)
		}
	})

	t.Run("IsRomaji", func(t *testing.T) {
		tests := []struct {
			input    string
			expected bool
		}{
			{"baka", true},
			{"konnichiwa", true},
			{"test123", true},
			{"ばか", false},
			{"バカ", false},
			{"漢字", false},
			{"mixed ばか text", false},
			{"", false},
		}

		for _, test := range tests {
			result := normalizer.IsRomaji(test.input)
			assert.Equal(t, test.expected, result, "IsRomaji(%s) should be %v", test.input, test.expected)
		}
	})

	t.Run("ContainsJapanese", func(t *testing.T) {
		tests := []struct {
			input    string
			expected bool
		}{
			{"ばか", true},
			{"バカ", true},
			{"漢字", true},
			{"mixed ばか text", true},
			{"mixed バカ text", true},
			{"baka", false},
			{"english only", false},
			{"123", false},
			{"", false},
		}

		for _, test := range tests {
			result := normalizer.ContainsJapanese(test.input)
			assert.Equal(t, test.expected, result, "ContainsJapanese(%s) should be %v", test.input, test.expected)
		}
	})

	t.Run("ToHalfWidth", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"ＴＥＳＴ", "TEST"},
			{"１２３", "123"},
			{"test", "test"},
			{"", ""},
		}

		for _, test := range tests {
			result := normalizer.ToHalfWidth(test.input)
			assert.Equal(t, test.expected, result, "ToHalfWidth(%s) should equal %s", test.input, test.expected)
		}
	})

	t.Run("ToFullWidth", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"TEST", "ＴＥＳＴ"},
			{"123", "１２３"},
			{"ＴＥＳＴ", "ＴＥＳＴ"},
			{"", ""},
		}

		for _, test := range tests {
			result := normalizer.ToFullWidth(test.input)
			assert.Equal(t, test.expected, result, "ToFullWidth(%s) should equal %s", test.input, test.expected)
		}
	})

	t.Run("NormalizeJapaneseText", func(t *testing.T) {
		tests := []struct {
			input    string
			contains []string // variations that should be included
		}{
			{
				"baka",
				[]string{"baka", "ばか", "バカ"},
			},
			{
				"ばか",
				[]string{"ばか", "バカ", "baka"},
			},
			{
				"バカ",
				[]string{"バカ", "ばか", "baka"},
			},
			{
				"TEST",
				[]string{"TEST", "ＴＥＳＴ"},
			},
		}

		for _, test := range tests {
			variations := normalizer.NormalizeJapaneseText(test.input)

			// Check that all expected variations are present
			for _, expected := range test.contains {
				found := false
				for _, variation := range variations {
					if variation == expected {
						found = true
						break
					}
				}
				assert.True(t, found, "NormalizeJapaneseText(%s) should contain %s, got %v", test.input, expected, variations)
			}
		}
	})

	t.Run("GenerateMatchingVariations", func(t *testing.T) {
		tests := []struct {
			input       string
			shouldMatch []string
		}{
			{
				"baka",
				[]string{"baka", "ばか", "バカ"},
			},
			{
				"ばか",
				[]string{"ばか", "バカ", "baka"},
			},
		}

		for _, test := range tests {
			variations := normalizer.GenerateMatchingVariations(test.input)

			for _, expected := range test.shouldMatch {
				found := false
				for _, variation := range variations {
					if variation == expected {
						found = true
						break
					}
				}
				assert.True(t, found, "GenerateMatchingVariations(%s) should contain %s", test.input, expected)
			}
		}
	})
}

func TestPluginJapaneseFiltering(t *testing.T) {
	plugin := &Plugin{
		japaneseNormalizer: NewJapaneseTextNormalizer(),
	}

	t.Run("replaceWordInText", func(t *testing.T) {
		config := &configuration{
			EnableJapaneseSupport: true,
			CensorCharacter:       "*",
		}

		plugin.configuration = config

		tests := []struct {
			text        string
			badWord     string
			replacement string
			expected    string
		}{
			{
				"This is baka text",
				"baka",
				"****",
				"This is **** text",
			},
			{
				"これはばかです",
				"baka",
				"****",
				"これは****です",
			},
			{
				"これはバカです",
				"baka",
				"****",
				"これは****です",
			},
			{
				"mixed ばか content",
				"baka",
				"****",
				"mixed **** content",
			},
		}

		for _, test := range tests {
			result := plugin.replaceWordInText(test.text, test.badWord, test.replacement)
			assert.Equal(t, test.expected, result, "replaceWordInText should properly replace %s in %s", test.badWord, test.text)
		}
	})

	t.Run("replaceWordInText without Japanese support", func(t *testing.T) {
		config := &configuration{
			EnableJapaneseSupport: false,
			CensorCharacter:       "*",
		}

		plugin.configuration = config

		// Without Japanese support, only exact matches should be replaced
		result := plugin.replaceWordInText("これはばかです", "baka", "****")
		assert.Equal(t, "これはばかです", result, "Without Japanese support, Japanese text should not be modified")

		result = plugin.replaceWordInText("This is baka text", "baka", "****")
		assert.Equal(t, "This is **** text", result, "Without Japanese support, exact Latin matches should still work")
	})
}
