package app

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/pkg/errors"

	"github.com/b2r2/character-counter-bot/internal/scrape"
	api "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
)

type BotAPI struct {
	config  *Config
	logger  *logrus.Logger
	bot     *api.BotAPI
	scrape  *scrape.Scrape
	updates api.UpdatesChannel
}

func New(config *Config) (*BotAPI, error) {
	b := &BotAPI{
		config: config,
		logger: logrus.New(),
		scrape: scrape.New(config.Scraper),
	}
	bot, err := api.NewBotAPI(config.Token)
	if err != nil {
		return nil, fmt.Errorf("bot initialization error %w\n", err)
	}
	b.bot = bot
	b.bot.Debug = config.BotLogLevel
	return b, nil
}

func (b *BotAPI) Run() error {
	goScrape := map[bool]func() error{
		true:  b.configureWebhook,
		false: b.configureUpdates,
	}
	if err := b.configureLogger(); err != nil {
		logrus.Errorln("configure logger:", err)
		return err
	}

	if err := goScrape[b.config.Webhook.IsWebhook](); err != nil {
		return err
	}
	b.logger.Infof("Authorized on account %s, debuging mode: %t", b.bot.Self.UserName, b.config.BotLogLevel)
	// TODO: error for webhook
	return errors.Wrap(b.handleUpdates(), "handle updates:")
}

func (b *BotAPI) configureWebhook() error {
	if err := b.removeWebhook(); err != nil {
		return err
	}
	if _, err := b.bot.SetWebhook(api.NewWebhookWithCert(
		b.config.Webhook.Cert+b.config.Token, "cert.pem",
	)); err != nil {
		b.logger.Errorln("configure webhook: set webhook:", err)
		return err
	}
	info, err := b.bot.GetWebhookInfo()
	if err != nil {
		b.logger.Errorln("configure webhook: get webhook:", err)
		return err
	}
	if info.LastErrorDate != 0 {
		b.logger.Errorln("configure webhook: last error date:", err)
		return fmt.Errorf("callback failed: %s\n", info.LastErrorMessage)
	}
	b.updates = b.bot.ListenForWebhook("/" + b.config.Token)

	errCh := make(chan error)
	go func() {
		defer close(errCh)
		if err := http.ListenAndServeTLS(b.config.Webhook.Addr, "cert.pem", "key.pem", nil); err != nil {
			errCh <- err
		}
	}()

	return <-errCh
}

func (b *BotAPI) removeWebhook() error {
	if _, err := b.bot.RemoveWebhook(); err != nil {
		b.logger.Errorln("configure webhook: remove webhook:", err)
		return err
	}
	return nil
}

func (b *BotAPI) configureUpdates() error {
	if err := b.removeWebhook(); err != nil {
		return err
	}

	u := api.NewUpdate(0)
	u.Timeout = 60
	updates, err := b.bot.GetUpdatesChan(u)
	if err != nil {
		b.logger.Errorln("get updates:", err)
		return err
	}
	b.updates = updates
	return nil
}

func (b *BotAPI) configureLogger() error {
	level, err := logrus.ParseLevel(b.config.LogLevel)
	if err != nil {
		return err
	}
	b.logger.SetLevel(level)
	return nil
}

func (b *BotAPI) handleUpdates() error {
	for update := range b.updates {
		if update.Message == nil {
			continue
		}
		user := update.Message.From.UserName
		userID := int64(update.Message.From.ID)
		replyToUser := api.NewMessage(userID, "")
		if !verifyUser(user, b.config.AccessUsers...) {
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
		if update.Message.Text == "" || !verifyLink(
			update.Message.Text,
			b.config.Scraper.Medium,
			b.config.Scraper.WebSite) {
			replyToUser.Text = b.config.Text["wrong_link"]
			if _, err := b.bot.Send(replyToUser); err != nil {
				b.logger.Println("send error:", err)
				return err
			}
			continue
		}
		if size, err := b.scrape.GetCountSymbols(update.Message.Text); err != nil {
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
