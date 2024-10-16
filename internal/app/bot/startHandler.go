package bot

import (
	"github.com/drTragger/yakudza-cars-bot/internal/app/utils"
	"github.com/yanzay/tbot/v2"
)

func (b *Bot) StartHandler(m *tbot.Message) {
	if m.Chat.ID == b.config.Admin.GroupID {
		return
	}

	b.sendMessage(m, "Будь ласка, оберіть варіант з меню:", utils.GetMenuKeyboard())
}
