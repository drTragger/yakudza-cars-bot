package bot

import (
	"fmt"
	"github.com/drTragger/yakudza-cars-bot/internal/app"
	"github.com/drTragger/yakudza-cars-bot/storage"
	"github.com/sirupsen/logrus"
	"github.com/yanzay/tbot/v2"
	"log"
	"time"
)

func (b *Bot) configureLoggerField() error {
	logLevel, err := logrus.ParseLevel(b.config.LoggerLevel)
	if err != nil {
		return err
	}
	b.logger.SetLevel(logLevel)
	return nil
}

func (b *Bot) configureRouterField() {
	b.bot.HandleMessage("/start", b.StartHandler)
	b.bot.HandleMessage("/admin", b.HandleAdmin)
	b.bot.HandleMessage("", b.HandleMessage)
	b.bot.HandleCallback(b.HandleCallback)
}

func (b *Bot) configureStorageField() error {
	newStorage := storage.New(b.config.Storage)
	if err := newStorage.Open(); err != nil {
		return err
	}
	b.storage = newStorage
	return nil
}

func (b *Bot) sendMessage(m *tbot.Message, msg string, opts interface{}) *tbot.Message {
	b.logHandler(m, msg)
	var message *tbot.Message

	switch v := opts.(type) {
	case *tbot.InlineKeyboardMarkup:
		message = handleMessageError(b.client.SendMessage(m.Chat.ID, msg, tbot.OptInlineKeyboardMarkup(v)))
	case *tbot.ReplyKeyboardMarkup:
		message = handleMessageError(b.client.SendMessage(m.Chat.ID, msg, tbot.OptReplyKeyboardMarkup(v)))
	default:
		message = handleMessageError(b.client.SendMessage(m.Chat.ID, msg))
	}

	return message
}

func (b *Bot) editMessage(m *tbot.Message, msg string, opts interface{}) *tbot.Message {
	b.logHandler(m, msg)
	var message *tbot.Message

	switch v := opts.(type) {
	case *tbot.InlineKeyboardMarkup:
		message = handleMessageError(b.client.EditMessageText(m.Chat.ID, m.MessageID, msg, tbot.OptInlineKeyboardMarkup(v)))
	default:
		message = handleMessageError(b.client.EditMessageText(m.Chat.ID, m.MessageID, msg))
	}

	return message
}

func (b *Bot) logHandler(m *tbot.Message, answer string) {
	location, err := time.LoadLocation("Europe/Kyiv")
	if err != nil {
		b.logger.Info("Failed loading location ", err.Error())
	}

	b.logger.Infof(
		"%s\nChat ID: %s\nMessage: %s\nAnswer: %s",
		time.Now().In(location).Format(app.TimeLayout), m.Chat.ID, m.Text, answer,
	)
}

func handleMessageError(message *tbot.Message, err error) *tbot.Message {
	if err != nil {
		log.Printf("Message: %s\nError: %s", message.Text, err.Error())
	}
	return message
}

func generateYears(startYear int) []string {
	currentYear := time.Now().Year()
	var years []string

	for year := startYear; year <= currentYear; year++ {
		years = append(years, fmt.Sprintf("%d", year))
	}

	return years
}
