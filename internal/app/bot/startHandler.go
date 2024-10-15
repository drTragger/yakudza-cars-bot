package bot

import (
	"github.com/yanzay/tbot/v2"
)

func (b *Bot) StartHandler(m *tbot.Message) {
	if m.Chat.ID == b.config.GroupID {
		return
	}

	// Send a message with the main menu keyboard
	b.sendMessage(m, "Будь ласка, оберіть варіант з меню:", b.getMenuKeyboard())
}

func (b *Bot) getMenuKeyboard() *tbot.ReplyKeyboardMarkup {
	return &tbot.ReplyKeyboardMarkup{
		ResizeKeyboard:  true,
		OneTimeKeyboard: false, // Keep the keyboard after use
		Keyboard: [][]tbot.KeyboardButton{
			{
				{Text: "Підібрати Авто"}, // Button to start the car selection process
				{Text: "Відгуки"},        // Button to allow feedback
			},
			{
				{Text: "Бай Нау"}, // Third button at the bottom
			},
		},
	}
}
