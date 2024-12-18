package bot

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/yanzay/tbot/v2"
	"strconv"
)

func (b *Bot) handleContactUs(cq *tbot.CallbackQuery) {
	err := b.client.DeleteMessage(cq.Message.Chat.ID, cq.Message.MessageID)
	if err != nil {
		b.logger.Error("Failed to delete a message: ", err.Error())
	}

	chatId, err := strconv.Atoi(cq.Message.Chat.ID)
	if err != nil {
		b.logger.Error("Failed to convert chatId to int: ", err.Error())
		b.sendMessage(cq.Message, "Щось пішло не так. Спробуйте ще раз.", nil)
		return
	}

	user, err := b.storage.User().FindByChatId(chatId)
	if errors.Is(err, sql.ErrNoRows) {
		b.requestPhoneNumber(cq.Message)
		return
	} else if err != nil {
		b.logger.Error("Failed to find user: ", err.Error())
		b.sendMessage(cq.Message, "Щось пішло не так. Спробуйте ще раз.", nil)
		return
	}

	_, err = b.client.SendMessage(
		b.config.Admin.GroupID,
		fmt.Sprintf("‼️Клієнт не знайшов для себе варіант‼️\n+%s", user.Phone),
	)
	if err != nil {
		b.logger.Error("Failed to send message to the group (could not find an option): ", err.Error())
		return
	}

	b.sendMessage(cq.Message, "Дякуємо, ми скоро з вами звʼяжемось.", nil)
}
