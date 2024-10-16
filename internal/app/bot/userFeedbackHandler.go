package bot

import (
	"database/sql"
	"errors"
	"github.com/yanzay/tbot/v2"
	"strconv"
)

func (b *Bot) handleFeedback(m *tbot.Message) {
	b.setUserState(m.Chat.ID, "show_feedback")

	chatId, err := strconv.Atoi(m.Chat.ID)
	if err != nil {
		b.logger.Error("Failed to convert chatId to int: ", err.Error())
		b.sendMessage(m, "Щось пішло не так. Спробуйте ще раз.", nil)
		return
	}

	_, err = b.storage.User().FindByChatId(chatId)
	if errors.Is(err, sql.ErrNoRows) {
		// Запит номера телефону
		b.requestPhoneNumber(m)
		return
	} else if err != nil {
		b.logger.Error("Failed to find user by chat id: ", err.Error())
		b.sendMessage(m, "Щось пішло не так. Спробуйте ще раз.", nil)
		return
	}

	b.showFeedback(m)
}

func (b *Bot) showFeedback(m *tbot.Message) {
	shownFeedbacks := b.getShownFeedbackIDs(m.Chat.ID)

	// Get the next feedback from the database
	feedback, err := b.storage.Feedback().GetNext(shownFeedbacks)
	if errors.Is(err, sql.ErrNoRows) {
		// No more feedback to show
		b.sendMessage(m, "Більше немає відгуків.", nil)
		return
	} else if err != nil {
		// Handle any other errors
		b.logger.Error("Failed to get feedback: ", err.Error())
		b.sendMessage(m, "Щось пішло не так. Спробуйте ще раз.", nil)
		return
	}

	// Track the shown feedback ID to avoid repetition
	shownFeedbacks = append(shownFeedbacks, feedback.ID)
	b.setShownFeedbackIDs(m.Chat.ID, shownFeedbacks)

	// Create inline keyboard without the "Хочу ще" button by default
	inlineKeyboard := &tbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]tbot.InlineKeyboardButton{},
	}

	// Check if there is more feedback
	_, err = b.storage.Feedback().GetNext(shownFeedbacks)
	if err == nil {
		// If there is more feedback, add the "Хочу ще" button
		inlineKeyboard.InlineKeyboard = append(inlineKeyboard.InlineKeyboard, []tbot.InlineKeyboardButton{
			{Text: "Хочу ще", CallbackData: "more_feedback"},
		})
	} else if !errors.Is(err, sql.ErrNoRows) {
		// If an unexpected error occurs, log it
		b.logger.Error("Failed to check for more feedback: ", err.Error())
	}

	// Send the feedback video with the inline keyboard
	_, err = b.client.SendVideo(
		m.Chat.ID,
		feedback.VideoFileID,
		tbot.OptCaption(feedback.Description),
		tbot.OptInlineKeyboardMarkup(inlineKeyboard),
	)
	if err != nil {
		b.logger.Error("Failed to send feedback video: ", err.Error())
		b.sendMessage(m, "Щось пішло не так. Спробуйте ще раз.", nil)
		return
	}

	// Remove any previous user state
	b.deleteUserState(m.Chat.ID)
}
