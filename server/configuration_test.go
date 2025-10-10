package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompileWordRegexes(t *testing.T) {
	p := Plugin{
		configuration: &configuration{
			BadWordsList: "abc,def ghi",
		},
	}

	t.Run("Build ASCII Regex", func(t *testing.T) {
		err := p.compileWordRegexes(p.getConfiguration().BadWordsList)
		assert.NoError(t, err)

		asciiRegex := p.getASCIIWordsRegex()
		assert.NotNil(t, asciiRegex)
		assert.Equal(t, `(?mi)\b(def ghi|abc)\b`, asciiRegex.String())
	})

	p2 := Plugin{
		configuration: &configuration{
			BadWordsList: "abc,abc def",
		},
	}

	t.Run("Build In double Regex", func(t *testing.T) {
		err := p2.compileWordRegexes(p2.getConfiguration().BadWordsList)
		assert.NoError(t, err)

		asciiRegex := p2.getASCIIWordsRegex()
		assert.NotNil(t, asciiRegex)
		assert.Equal(t, `(?mi)\b(abc def|abc)\b`, asciiRegex.String())
	})
}
