package main

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

// configuration captures the plugin's external configuration as exposed in the Mattermost server
// configuration, as well as values computed from the configuration. Any public fields will be
// deserialized from the Mattermost server configuration in OnConfigurationChange.
//
// As plugins are inherently concurrent (hooks being called asynchronously), and the plugin
// configuration can change at any time, access to the configuration must be synchronized. The
// strategy used in this plugin is to guard a pointer to the configuration, and clone the entire
// struct whenever it changes. You may replace this with whatever strategy you choose.
//
// If you add non-reference types to your configuration struct, be sure to rewrite Clone as a deep
// copy appropriate for your types.
type configuration struct {
	ExcludeBots     bool
	RejectPosts     bool
	CensorCharacter string
	BadWordsList    string
	WarningMessage  string `json:"WarningMessage"`
}

// Clone shallow copies the configuration. Your implementation may require a deep copy if
// your configuration has reference types.
func (c *configuration) Clone() *configuration {
	var clone = *c
	return &clone
}

// getConfiguration retrieves the active configuration under lock, making it safe to use
// concurrently. The active configuration may change underneath the client of this method, but
// the struct returned by this API call is considered immutable.
func (p *Plugin) getConfiguration() *configuration {
	p.configurationLock.RLock()
	defer p.configurationLock.RUnlock()

	if p.configuration == nil {
		return &configuration{}
	}

	return p.configuration
}

// setConfiguration replaces the active configuration under lock.
//
// Do not call setConfiguration while holding the configurationLock, as sync.Mutex is not
// reentrant. In particular, avoid using the plugin API entirely, as this may in turn trigger a
// hook back into the plugin. If that hook attempts to acquire this lock, a deadlock may occur.
//
// This method panics if setConfiguration is called with the existing configuration. This almost
// certainly means that the configuration was modified without being cloned and may result in
// an unsafe access.
func (p *Plugin) setConfiguration(configuration *configuration) {
	p.configurationLock.Lock()
	defer p.configurationLock.Unlock()

	if configuration != nil && p.configuration == configuration {
		// Ignore assignment if the configuration struct is empty. Go will optimize the
		// allocation for same to point at the same memory address, breaking the check
		// above.
		if reflect.ValueOf(*configuration).NumField() == 0 {
			return
		}

		panic("setConfiguration called with the existing configuration")
	}

	p.configuration = configuration
}

// OnConfigurationChange is invoked when configuration changes may have been made.
func (p *Plugin) OnConfigurationChange() error {
	var configuration = new(configuration)

	// Load the public configuration fields from the Mattermost server configuration.
	if err := p.API.LoadPluginConfiguration(configuration); err != nil {
		return errors.Wrap(err, "failed to load plugin configuration")
	}

	// Normalize Japanese commas to ASCII commas in BadWordsList for consistent processing
	configuration.BadWordsList = normalizeWordListCommas(configuration.BadWordsList)

	p.setConfiguration(configuration)

	// Compile regex patterns for both ASCII and Japanese words
	if err := p.compileWordRegexes(configuration.BadWordsList); err != nil {
		return err
	}

	// Initialize Japanese tokenizer
	if err := p.initializeJapaneseTokenizer(); err != nil {
		return err
	}

	return nil
}

// compileWordRegexes compiles regex patterns for both ASCII and Japanese words
func (p *Plugin) compileWordRegexes(wordList string) error {
	words := splitWordList(wordList)
	asciiWords, japaneseWords := separateASCIIAndJapanese(words)

	// Compile ASCII words regex
	if len(asciiWords) > 0 {
		// Sort by length (longest first) to match longer words first
		sort.Slice(asciiWords, func(i, j int) bool { return len(asciiWords[i]) > len(asciiWords[j]) })
		asciiRegexStr := fmt.Sprintf(`(?mi)\b(%s)\b`, strings.Join(asciiWords, "|"))
		asciiRegex, err := regexp.Compile(asciiRegexStr)
		if err != nil {
			return fmt.Errorf("failed to compile ASCII words regex: %w", err)
		}
		p.asciiWordsRegex = asciiRegex
	} else {
		p.asciiWordsRegex = nil
	}

	// Compile Japanese words regex
	if len(japaneseWords) > 0 {
		var escapedWords []string
		for _, word := range japaneseWords {
			escapedWords = append(escapedWords, regexp.QuoteMeta(strings.ToLower(strings.TrimSpace(word))))
		}
		japaneseRegexStr := fmt.Sprintf(`(?i)\b(%s)\b`, strings.Join(escapedWords, "|"))
		japaneseRegex, err := regexp.Compile(japaneseRegexStr)
		if err != nil {
			return fmt.Errorf("failed to compile Japanese words regex: %w", err)
		}
		p.japaneseWordsRegex = japaneseRegex
	} else {
		p.japaneseWordsRegex = nil
	}

	return nil
}
