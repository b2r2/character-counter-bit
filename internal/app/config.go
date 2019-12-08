package app

import "github.com/b2r2/character-counter-bot/internal/scrape"

// Config ...
type Config struct {
	Token       string   `toml:"token"`
	ChatID      int64    `toml:"chat_id"`
	LogLevel    string   `toml:"log_level"`
	BotLogLevel bool     `toml:"bot_log_level"`
	AccessUsers []string `toml:"access_users"`
	Scraper     *scrape.Config
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{
		LogLevel:    "debug",
		BotLogLevel: true,
		Scraper:     scrape.NewConfig(),
	}
}
