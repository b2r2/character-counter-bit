package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/b2r2/character-counter-bot/internal/scrape"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
)

// BotAPI ...
type BotAPI struct {
	config  *Config
	logger  *logrus.Logger
	bot     *tgbotapi.BotAPI
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
		return err
	}

	if err := b.configureBot(); err != nil {
		return err
	}

	b.configureScraper()

	logrus.Infof("Authorized on account %s, debuging mode: %t", b.bot.Self.UserName, b.config.BotLogLevel)

	if err := b.handleUpdates(); err != nil {
		return err
	}
	return nil
}

func (b *BotAPI) configureBot() error {
	bot, err := tgbotapi.NewBotAPI(b.config.Token)
	if err != nil {
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
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := b.bot.GetUpdatesChan(u)
	if err != nil {
		return err
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}
		user := update.Message.From.UserName
		userID := int64(update.Message.From.ID)
		replyToUser := tgbotapi.NewMessage(userID, "")
		if !b.verifyUser(user) {
			replyToUser.Text = "401 Unauthorized"
			b.bot.Send(replyToUser)
			continue
		}
		if update.Message.IsCommand() {
			if update.Message.Command() == "start" {
				replyToUser.Text = "Привет! Я помогу тебе подсчитать количество символов в статье! Скинь мне ссылку на статью и я скажу сколько там символов 😉"
			} else {
				replyToUser.Text = "К сожалению, я не знаю такую команду 😭\nОднако, ты можешь скинуть мне ссылку на статью и я скажу сколько там символов 😉"
			}
			b.bot.Send(replyToUser)
			continue
		}
		if update.Message.Text == "" || !b.verifyLink(update.Message.Text) {
			replyToUser.Text = "Мне бы ссылочку на статью, а не вот этот вот всё"
			b.bot.Send(replyToUser)
			continue
		}
		if size, err := b.scraper.GetCountSymbols(update.Message.Text); err != nil {
			replyToUser.Text = fmt.Sprintf("Что-то пошло не так: %v", err)
		} else {
			replyToUser.Text = strconv.Itoa(size)
		}
		b.bot.Send(replyToUser)
	}
	return nil
}

func (b *BotAPI) verifyUser(user string) (state bool) {
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
		case scrape.MEDIUM, b.config.Scraper.WebSite:
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
