package main

import (
	"flag"
	"log"

	"github.com/BurntSushi/toml"
	"github.com/b2r2/character-counter-bot/internal/app"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config-path", "./configs/bot.toml", "path to config file")
}

func main() {
	flag.Parse()

	config := app.NewConfig()

	_, err := toml.DecodeFile(configPath, config)
	if err != nil {
		log.Fatal(err)
	}

	b := app.New(config)
	if err := b.Run(); err != nil {
		log.Fatal(err)
	}

}
