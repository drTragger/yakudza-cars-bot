package bot

import (
	"github.com/drTragger/yakudza-cars-bot/internal/app/utils"
	"github.com/yanzay/tbot/v2"
)

func (b *Bot) HandleAdmin(m *tbot.Message) {
	if m.Chat.ID == b.config.Admin.GroupID {
		return
	}

	if !b.ensureAdmin(m) {
		return
	}

	b.sendMessage(m, "Виберіть дію:", utils.GetAdminMenuKeyboard())
}

func (b *Bot) isAdmin(m *tbot.Message) bool {
	for _, id := range b.config.Admin.AdminIDs {
		if id == m.Chat.ID {
			return true
		}
	}
	return false
}

func (b *Bot) ensureAdmin(m *tbot.Message) bool {
	if !b.isAdmin(m) {
		b.sendMessage(m, "У вас немає прав для цієї команди.", nil)
		return false
	}
	return true
}
