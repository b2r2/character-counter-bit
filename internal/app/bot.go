package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/b2r2/character-counter-bot/internal/scrape"

	api "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
)

// BotAPI ...
type BotAPI struct {
	config  *Config
	logger  *logrus.Logger
	bot     *api.BotAPI
	scraper *scrape.Scraper
}

// New ...
func New(config *Config) *BotAPI {
	return &BotAPI{
		config: config,
		logger: logrus.New(),
	}
}

// Run ...
func (b *BotAPI) Run() error {
	if err := b.configureLogger(); err != nil {
		logrus.Println("configure logger:", err)
		return err
	}

	if err := b.configureBot(); err != nil {
		b.logger.Println("configure bot:", err)
		return err
	}

	b.configureScraper()

	b.logger.Infof("Authorized on account %s, debuging mode: %t", b.bot.Self.UserName, b.config.BotLogLevel)

	if err := b.handleUpdates(); err != nil {
		b.logger.Println("handle updates:", err)
		return err
	}
	return nil
}

func (b *BotAPI) configureBot() error {
	bot, err := api.NewBotAPI(b.config.Token)
	if err != nil {
		b.logger.Println("configure bot:", err)
		return err
	}
	bot.Debug = b.config.BotLogLevel

	b.bot = bot
	return nil
}

func (b *BotAPI) configureScraper() {
	c := scrape.New(b.config.Scraper)
	b.scraper = c
}

func (b *BotAPI) handleUpdates() error {
	u := api.NewUpdate(0)
	u.Timeout = 60
	updates, err := b.bot.GetUpdatesChan(u)
	if err != nil {
		b.logger.Println("get updates:", err)
		return err
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}
		user := update.Message.From.UserName
		userID := int64(update.Message.From.ID)
		replyToUser := api.NewMessage(userID, "")
		if !b.verifyUser(user) {
			replyToUser.Text = "401 Unauthorized"
			if _, err := b.bot.Send(replyToUser); err != nil {
				b.logger.Println("send error:", err)
				return err
			}
			continue
		}
		if update.Message.IsCommand() {
			if update.Message.Command() == "start" {
				replyToUser.Text = b.config.Text["first_message"]
			} else {
				replyToUser.Text = b.config.Text["unknown"]
			}
			if _, err := b.bot.Send(replyToUser); err != nil {
				b.logger.Println("send error:", err)
				return err
			}
			continue
		}
		if update.Message.Text == "" || !b.verifyLink(update.Message.Text) {
			replyToUser.Text = b.config.Text["wrong_link"]
			if _, err := b.bot.Send(replyToUser); err != nil {
				b.logger.Println("send error:", err)
				return err
			}
			continue
		}
		if size, err := b.scraper.GetCountSymbols(update.Message.Text); err != nil {
			b.logger.Println("get count:", err)
			replyToUser.Text = fmt.Sprintf("%s: %v", b.config.Text["error"], err)
		} else {
			replyToUser.Text = strconv.Itoa(size)
		}
		if _, err := b.bot.Send(replyToUser); err != nil {
			b.logger.Println("send error:", err)
			return err
		}
	}
	return nil
}

func (b *BotAPI) verifyUser(user string) (state bool) {
	fmt.Println(b.config.AccessUsers, user)
	for _, u := range b.config.AccessUsers {
		if u == user {
			state = true
			break
		}
	}
	return
}

func (b *BotAPI) verifyLink(msg string) (state bool) {
	line := strings.Split(msg, "://")
	if len(line) == 2 {
		name := strings.Split(line[1], ".")
		switch name[0] {
		case b.config.Scraper.Medium, b.config.Scraper.WebSite:
			state = true
		}
	}
	return
}

func (b *BotAPI) configureLogger() error {
	level, err := logrus.ParseLevel(b.config.LogLevel)
	if err != nil {
		return err
	}
	b.logger.SetLevel(level)
	return nil
}
