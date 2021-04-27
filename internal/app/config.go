package app

import "github.com/b2r2/character-counter-bot/internal/scrape"

type Config struct {
	Token       string            `toml:"token"`
	ChatID      int64             `toml:"chat_id"`
	LogLevel    string            `toml:"log_level"`
	BotLogLevel bool              `toml:"bot_log_level"`
	AccessUsers []string          `toml:"access_users"`
	Text        map[string]string `toml:"messages"`
	Webhook     struct {
		IsWebhook bool   `toml:"is_webhook"`
		Cert      string `toml:"cert"`
		Addr      string `toml:"addr"`
	} `toml:"webhook"`
	Scraper *scrape.Config
}

func NewConfig() *Config {
	return &Config{
		LogLevel:    "debug",
		BotLogLevel: true,
		Scraper:     scrape.NewConfig(),
		Text:        make(map[string]string),
	}
}
