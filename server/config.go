package main

type Configuration struct {
	RejectPosts     bool
	CensorCharacter string
}

func (c *Configuration) IsValid() error {
	return nil
}

func (c *Configuration) SetDefaults() {
	if c.CensorCharacter == "" {
		c.CensorCharacter = "\\*"
	}
}
