package bot

import (
	"fmt"
	"github.com/drTragger/yakudza-cars-bot/storage"
	"github.com/sirupsen/logrus"
	"github.com/yanzay/tbot/v2"
	"log"
	"strconv"
	"time"
)

const TimeLayout = "2006-01-02 15:04:05" // Random date and time

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
	// Логування повідомлення
	b.LogHandler(m, msg)
	var message *tbot.Message

	// Перевірка типу opts через type assertion
	switch v := opts.(type) {
	case *tbot.InlineKeyboardMarkup:
		// Якщо opts є типом InlineKeyboardMarkup
		message = handleMessageError(b.client.SendMessage(m.Chat.ID, msg, tbot.OptInlineKeyboardMarkup(v)))
	case *tbot.ReplyKeyboardMarkup:
		// Якщо opts є типом ReplyKeyboardMarkup
		message = handleMessageError(b.client.SendMessage(m.Chat.ID, msg, tbot.OptReplyKeyboardMarkup(v)))
	default:
		// Якщо opts == nil або невідомий тип
		message = handleMessageError(b.client.SendMessage(m.Chat.ID, msg))
	}

	// Збереження ID останнього повідомлення
	userStates[m.Chat.ID+"_last_msg"] = fmt.Sprintf("%d", message.MessageID)
	return message
}

func (b *Bot) sendCallbackMessage(cq *tbot.CallbackQuery, msg string, opts *tbot.ReplyKeyboardMarkup) *tbot.Message {
	//handleChatActionError(b.client.SendChatAction(cq.Message.Chat.ID, tbot.ActionTyping))
	//time.Sleep(500 * time.Millisecond)
	b.LogCallbackHandler(cq, msg)
	if opts != nil {
		return handleMessageError(b.client.SendMessage(cq.Message.Chat.ID, msg, tbot.OptReplyKeyboardMarkup(opts)))
	} else {
		return handleMessageError(b.client.SendMessage(cq.Message.Chat.ID, msg))
	}
}

func (b *Bot) editCallbackMessage(cq *tbot.CallbackQuery, msg string, opts interface{}) *tbot.Message {
	b.LogCallbackHandler(cq, msg)
	var message *tbot.Message

	// Перевірка типу opts через type assertion
	switch v := opts.(type) {
	case *tbot.InlineKeyboardMarkup:
		// Якщо opts є типом InlineKeyboardMarkup
		message = handleMessageError(b.client.EditMessageText(cq.Message.Chat.ID, cq.Message.MessageID, msg, tbot.OptInlineKeyboardMarkup(v)))
	default:
		// Якщо opts == nil або невідомий тип
		message = handleMessageError(b.client.EditMessageText(cq.Message.Chat.ID, cq.Message.MessageID, msg))
	}

	return message
}

func generateYears(startYear int) []string {
	currentYear := time.Now().Year()
	var years []string

	// Loop from startYear to the current year
	for year := startYear; year <= currentYear; year++ {
		years = append(years, fmt.Sprintf("%d", year)) // Convert year to string
	}

	return years
}

func (b *Bot) deleteLastMessage(chatId string) {
	if lastMsgID, exists := userStates[chatId+"_last_msg"]; exists {
		msgID, _ := strconv.Atoi(lastMsgID)
		if err := b.client.DeleteMessage(chatId, msgID); err != nil {
			log.Printf("Failed to delete last message %d\nError: %s", msgID, err.Error())
		}
	}
}
