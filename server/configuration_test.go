package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWordListToRegex(t *testing.T) {
	p := Plugin{
		configuration: &configuration{
			BadWordsList: "abc,def ghi",
		},
	}

	t.Run("Build Regex", func(t *testing.T) {
		regexStr := wordListToRegex(p.getConfiguration().BadWordsList)

		assert.Equal(t, regexStr, "(?mi)(def ghi|abc)")
	})

	p2 := Plugin{
		configuration: &configuration{
			BadWordsList: "abc,abc def",
		},
	}

	t.Run("Build In double Regex", func(t *testing.T) {
		regexStr := wordListToRegex(p2.getConfiguration().BadWordsList)

		assert.Equal(t, regexStr, "(?mi)(abc def|abc)")
	})
}
