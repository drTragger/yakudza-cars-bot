package bot

import (
	"github.com/yanzay/tbot/v2"
)

var callbackHandlers = map[string]func(*Bot, *tbot.CallbackQuery){
	"add_option":    (*Bot).handleAddOption,
	"view_options":  (*Bot).handleViewOptions,
	"add_feedback":  (*Bot).handleAdminAddFeedback,
	"view_feedback": (*Bot).handleAdminViewFeedback,
	"more_cars":     func(b *Bot, cq *tbot.CallbackQuery) { b.showCarOption(cq.Message) },
	"more_feedback": func(b *Bot, cq *tbot.CallbackQuery) { b.showFeedback(cq.Message) },
	"contact_us":    (*Bot).handleContactUs,
}

func (b *Bot) HandleCallback(cq *tbot.CallbackQuery) {
	if handler, exists := callbackHandlers[cq.Data]; exists {
		handler(b, cq)
		return
	}

	switch {
	case len(cq.Data) >= 5 && cq.Data[:5] == "year_":
		b.handleAdminYearSelection(cq)
	case len(cq.Data) >= 11 && cq.Data[:11] == "select_car_":
		b.handleSelectCar(cq)
	case len(cq.Data) >= 14 && cq.Data[:14] == "delete_option_":
		b.handleDeleteOption(cq)
	case len(cq.Data) >= 16 && cq.Data[:16] == "delete_feedback_":
		b.handleAdminDeleteFeedback(cq)
	default:
		b.sendMessage(cq.Message, "Невідома дія.", nil)
	}
}
