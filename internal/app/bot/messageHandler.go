package bot

import (
	"github.com/drTragger/yakudza-cars-bot/internal/app"
	"github.com/drTragger/yakudza-cars-bot/internal/app/utils"
	"github.com/yanzay/tbot/v2"
)

var menuHandlers = map[string]func(*Bot, *tbot.Message){
	"Підібрати Авто": (*Bot).handleCarSelection,
	"Відгуки":        (*Bot).handleFeedback,
	"Бай Нау": func(b *Bot, m *tbot.Message) {
		b.sendMessage(m, "Перейдіть за посиланням:\nhttps://t.me/yakudzaoffer", nil)
	},
}

func (b *Bot) HandleMessage(m *tbot.Message) {
	if m.Chat.ID == b.config.Admin.GroupID {
		return
	}

	userState := b.getUserState(m.Chat.ID)

	if handler, exists := menuHandlers[m.Text]; exists {
		handler(b, m)
		return
	}

	b.handleNonMenuMessage(m, userState)
}

func (b *Bot) handleNonMenuMessage(m *tbot.Message, userState string) {
	// Handle various types of content based on the user state
	switch {
	case m.Contact != nil && userState == app.AwaitingPhone:
		b.handlePhoneNumber(m)
	case m.Photo != nil && userState == app.AwaitingCarPhoto:
		b.handleAdminPhotoInput(m)
	case m.Video != nil && userState == app.AwaitingFeedbackVideo:
		b.handleFeedbackVideoInput(m)
	case m.Document != nil:
		b.handleDocument(m, userState)
	case userState == app.AwaitingCarTitle:
		b.handleAdminTitleInput(m)
	case userState == app.AwaitingCarDescription:
		b.handleAdminDescriptionInput(m)
	case userState == app.AwaitingCarPrice:
		b.handleAdminPriceInput(m)
	case userState == app.AwaitingFeedbackDescription:
		b.handleFeedbackDescriptionInput(m)
	default:
		// If nothing matches, try price or year selection
		b.handleSelection(m)
	}
}

func (b *Bot) handleDocument(m *tbot.Message, userState string) {
	if userState == app.AwaitingFeedbackVideo {
		b.sendMessage(m, "Будь ласка, відправте відео, а не файл.", nil)
	} else {
		b.sendMessage(m, "Будь ласка, відправте фото, а не файл.", nil)
	}
}

func (b *Bot) handleSelection(m *tbot.Message) {
	// Check if the message matches a price selection
	for _, price := range utils.Prices {
		if m.Text == price.Title {
			b.handlePriceSelection(m, price)
			return
		}
	}

	// Check if the message matches a year selection
	years := generateYears(app.CarOptionsStartYear)
	for _, year := range years {
		if m.Text == year {
			b.handleYearSelection(m, year)
			return
		}
	}

	// If nothing matches, show a default message
	b.sendMessage(m, "Невідома команда. Будь ласка, оберіть варіант з меню.", nil)
}
