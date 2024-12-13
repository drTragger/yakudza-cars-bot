package bot

import (
	"database/sql"
	"errors"
	"github.com/drTragger/yakudza-cars-bot/internal/app/utils"
	"github.com/yanzay/tbot/v2"
)

func (b *Bot) handleFeedback(m *tbot.Message) {
	b.showFeedback(m)
}

func (b *Bot) showFeedback(m *tbot.Message) {
	shownFeedbacks := b.getShownFeedbackIDs(m.Chat.ID)

	feedback, err := b.storage.Feedback().GetNext(shownFeedbacks)
	if errors.Is(err, sql.ErrNoRows) {
		b.sendInstagramMessage(m)
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
	_, nextFeedbackErr := b.storage.Feedback().GetNext(shownFeedbacks)
	if nextFeedbackErr == nil {
		// If there is more feedback, add the "Хочу ще" button
		inlineKeyboard.InlineKeyboard = append(inlineKeyboard.InlineKeyboard, []tbot.InlineKeyboardButton{
			{Text: "Хочу ще", CallbackData: "more_feedback"},
		})
	} else if !errors.Is(nextFeedbackErr, sql.ErrNoRows) {
		// If an unexpected error occurs, log it
		b.logger.Error("Failed to check for more feedback: ", nextFeedbackErr.Error())
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

	if errors.Is(nextFeedbackErr, sql.ErrNoRows) {
		b.sendInstagramMessage(m)
		return
	}

	// Remove any previous user state
	b.deleteUserState(m.Chat.ID)
}

func (b *Bot) sendInstagramMessage(m *tbot.Message) {
	b.sendMessage(
		m,
		"Якщо хочеш детальніше познйомитися з нами та побачити більше авто, як ми привезли — переходь в наш Instagram.",
		utils.GetInstagramKeyboard(),
	)

	b.sendMessage(m, "Оберіть наступну дію ⚙️", utils.GetMenuKeyboard())
}
