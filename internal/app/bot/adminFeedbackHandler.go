package bot

import (
	"fmt"
	"github.com/drTragger/yakudza-cars-bot/internal/app"
	"github.com/drTragger/yakudza-cars-bot/internal/app/models"
	"github.com/drTragger/yakudza-cars-bot/internal/app/utils"
	"github.com/yanzay/tbot/v2"
	"strconv"
	"time"
)

// Step 1
func (b *Bot) handleAdminAddFeedback(cq *tbot.CallbackQuery) {
	if !b.ensureAdmin(cq.Message) {
		b.sendMessage(cq.Message, "У вас немає прав для цієї команди.", nil)
		return
	}

	b.editMessage(cq.Message, "Будь ласка, введіть опис відгуку.", nil)

	b.setUserState(cq.Message.Chat.ID, app.AwaitingFeedbackDescription)
}

// Step 2
func (b *Bot) handleFeedbackDescriptionInput(m *tbot.Message) {
	if !b.ensureAdmin(m) {
		b.sendMessage(m, "У вас немає прав для цієї команди.", nil)
		return
	}

	if len(m.Text) > 600 {
		b.sendMessage(m, "Опис не має перевищувати 600 символів.", nil)
	}

	b.setFeedback(m.Chat.ID, &models.Feedback{Description: m.Text})

	b.sendMessage(m, "Тепер надішліть відео для відгуку.", nil)

	b.setUserState(m.Chat.ID, app.AwaitingFeedbackVideo)
}

// Step 3
func (b *Bot) handleFeedbackVideoInput(m *tbot.Message) {
	if b.getUserState(m.Chat.ID) != app.AwaitingFeedbackVideo {
		return
	}

	if m.Video == nil || m.Video.FileID == "" {
		b.sendMessage(m, "Будь ласка, надішліть відео для відгуку.", nil)
		return
	}

	feedback := b.getFeedback(m.Chat.ID)
	feedback.VideoFileID = m.Video.FileID

	location, err := time.LoadLocation("Europe/Kyiv")
	if err != nil {
		b.logger.Info("Failed to load location: ", err.Error())
	}

	feedback.CreatedAt = time.Now().In(location).Format(app.TimeLayout)

	err = b.storage.Feedback().Create(feedback)

	if err != nil {
		b.logger.Error("Failed to save feedback: ", err.Error())
		b.sendMessage(m, "Щось пішло не так. Спробуйте ще раз.", nil)
		return
	}

	b.sendMessage(m, "Дякуємо! Ваш відгук був успішно збережений.", nil)

	b.deleteUserState(m.Chat.ID)
	b.deleteFeedback(m.Chat.ID)
}

func (b *Bot) handleAdminViewFeedback(cq *tbot.CallbackQuery) {
	if !b.ensureAdmin(cq.Message) {
		return
	}

	feedbackList, err := b.storage.Feedback().GetAll()
	if err != nil {
		b.logger.Error("Failed to get all feedback: ", err.Error())
		b.sendMessage(cq.Message, "Щось пішло не так. Спробуйте ще раз.", nil)
		return
	}

	if len(feedbackList) == 0 {
		b.sendMessage(cq.Message, "Немає доступних відгуків.", nil)
		return
	}

	errChan := make(chan error, len(feedbackList))

	for _, feedback := range feedbackList {
		go func(fb *models.Feedback) {
			createdAt, err := time.Parse(app.TimeLayout, fb.CreatedAt) // Replace 'TimeLayout' with your actual time layout
			if err != nil {
				b.logger.Error("Failed to parse date: ", err.Error())
				return
			}

			formattedDate := createdAt.Format("02.01.2006 о 15:04")

			_, err = b.client.SendVideo(
				cq.Message.Chat.ID,
				fb.VideoFileID,
				tbot.OptCaption(fmt.Sprintf("%s\n\nЗавантажено %s", feedback.Description, formattedDate)),
				tbot.OptInlineKeyboardMarkup(utils.GetDeleteFeedbackKeyboard(fb.ID)),
			)
			errChan <- err
		}(feedback)
	}

	for i := 0; i < len(feedbackList); i++ {
		if err := <-errChan; err != nil {
			b.logger.Error("Failed to send feedback: ", err.Error())
		}
	}

	close(errChan)
}

func (b *Bot) handleAdminDeleteFeedback(cq *tbot.CallbackQuery) {
	if !b.ensureAdmin(cq.Message) {
		return
	}

	feedbackIDStr := cq.Data[16:]
	feedbackID, err := strconv.Atoi(feedbackIDStr)
	if err != nil {
		b.logger.Error("Failed to convert feedback ID to int: ", err.Error())
		return
	}

	errChan := make(chan error, 2)

	go func() {
		err := b.storage.Feedback().Delete(feedbackID)
		errChan <- err
	}()

	go func() {
		err := b.client.DeleteMessage(cq.Message.Chat.ID, cq.Message.MessageID)
		errChan <- err
	}()

	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			b.logger.Error("Failed to complete operation: ", err.Error())
			return
		}
	}

	close(errChan)
}
