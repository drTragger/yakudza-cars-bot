package bot

import (
	"fmt"
	"github.com/drTragger/yakudza-cars-bot/internal/app/models"
	"github.com/yanzay/tbot/v2"
	"strconv"
	"time"
)

const (
	AwaitingTitle               = "awaiting_admin_car_title"
	AwaitingDescription         = "awaiting_admin_car_description"
	AwaitingPrice               = "awaiting_admin_price_selection"
	AwaitingYear                = "awaiting_admin_year_selection"
	AwaitingPhoto               = "awaiting_admin_photo"
	AwaitingFeedbackDescription = "awaiting_admin_feedback_description"
	AwaitingFeedbackVideo       = "awaiting_admin_feedback_video"
)

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

func (b *Bot) HandleAdmin(m *tbot.Message) {
	if m.Chat.ID == b.config.GroupID {
		return
	}

	// Перевіряємо, чи користувач є адміністратором
	if !b.ensureAdmin(m) {
		return
	}

	// Відправка повідомлення з інлайн-клавіатурою
	b.sendMessage(m, "Виберіть дію:", b.generateAdminOptionsKeyboard())
}

func (b *Bot) handleAddOptions(cq *tbot.CallbackQuery) {
	// Перевіряємо, чи є користувач адміністратором
	if !b.isAdmin(cq.Message) {
		b.sendMessage(cq.Message, "У вас немає прав для цієї команди.", nil)
		return
	}

	// Відправляємо повідомлення з вибором ціни
	b.editCallbackMessage(cq, "Введіть назву автомобіля:", nil)

	// Зберігаємо стан користувача
	userStates[cq.Message.Chat.ID] = AwaitingTitle
}

func (b *Bot) handleAdminTitleSelection(m *tbot.Message) {
	// Перевіряємо, чи є користувач адміністратором
	if !b.ensureAdmin(m) {
		return
	}

	if len(m.Text) > 255 {
		b.sendMessage(m, "Назва не має перевищувати 255 символів.", nil)
		return
	}

	carOption.Title = m.Text

	// Відправляємо повідомлення з вибором ціни
	b.sendMessage(m, "Введіть опис автомобіля:", nil)

	// Зберігаємо стан користувача
	userStates[m.Chat.ID] = AwaitingDescription
}

func (b *Bot) handleAdminDescriptionSelection(m *tbot.Message) {
	// Перевіряємо, чи є користувач адміністратором
	if !b.ensureAdmin(m) {
		return
	}

	if len(m.Text) > 500 {
		b.sendMessage(m, "Опис не має перевищувати 500 символів.", nil)
		return
	}

	carOption.Description = m.Text

	// Запитуємо користувача ввести ціну автомобіля вручну
	b.sendMessage(m, "Будь ласка, введіть ціну для автомобіля (у числовому форматі).", nil)

	// Зберігаємо стан користувача
	userStates[m.Chat.ID] = AwaitingPrice
}

func (b *Bot) handleAdminPriceInput(m *tbot.Message) {
	// Перевіряємо, чи є користувач адміністратором
	if !b.ensureAdmin(m) {
		return
	}

	// Перевіряємо, чи є введене значення коректною ціною
	price, err := strconv.Atoi(m.Text)
	if err != nil || price <= 0 {
		b.sendMessage(m, "Будь ласка, введіть коректну ціну (тільки цифри).", nil)
		return
	}

	// Зберігаємо ціну
	carOption.Price = price

	// Генеруємо список років з функції generateYears
	years := generateYears(StartYear)

	// Створюємо інлайн-клавіатуру для вибору року з чотирма колонками
	yearKeyboard := &tbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]tbot.InlineKeyboardButton{},
	}

	// Додаємо по чотири роки в кожен рядок клавіатури
	for i := 0; i < len(years); i += 4 {
		var row []tbot.InlineKeyboardButton

		// Додаємо перші 4 роки в рядок
		for j := i; j < i+4 && j < len(years); j++ {
			row = append(row, tbot.InlineKeyboardButton{
				Text:         years[j],
				CallbackData: fmt.Sprintf("year_%s", years[j]),
			})
		}

		// Додаємо рядок до клавіатури
		yearKeyboard.InlineKeyboard = append(yearKeyboard.InlineKeyboard, row)
	}

	// Запитуємо користувача вибрати рік
	b.sendMessage(m, "Тепер оберіть рік для автомобіля:", yearKeyboard)

	// Оновлюємо стан користувача
	userStates[m.Chat.ID] = AwaitingYear
}

func (b *Bot) handleAdminYearSelection(cq *tbot.CallbackQuery) {
	// Перевіряємо, чи є користувач адміністратором
	if !b.ensureAdmin(cq.Message) {
		return
	}

	// Отримуємо вибраний рік
	selectedYear := cq.Data[5:] // year_2018, year_2019, etc.
	b.logger.Infof("Рік обраний: %s", selectedYear)

	// Зберігаємо вибраний рік
	carOption.Year = selectedYear

	// Запитуємо користувача надіслати фото автомобіля
	b.editCallbackMessage(cq, "Будь ласка, надішліть фото автомобіля.", nil)

	// Оновлюємо стан користувача
	userStates[cq.Message.Chat.ID] = AwaitingPhoto
}

func (b *Bot) handleAdminPhoto(m *tbot.Message) {
	// Перевіряємо, чи є користувач адміністратором
	if !b.ensureAdmin(m) {
		return
	}

	location, err := time.LoadLocation("Europe/Kyiv")
	if err != nil {
		b.logger.Info("Failed to load location: ", err.Error())
	}

	carOption.PhotoID = m.Photo[0].FileID // Зберігаємо FileID фото
	carOption.CreatedAt = time.Now().In(location).Format(TimeLayout)
	err = b.storage.CarOption().Create(carOption)
	if err != nil {
		b.logger.Error("Failed to create new car option: ", err.Error())
		b.sendMessage(m, "Щось пішло не так. Спробуйте ще раз.", nil)
		return
	}
	b.logger.Infof("Фото отримано: %s", m.Photo[0].FileID)

	// Відправляємо підтвердження
	b.sendMessage(m, "Цей автомобіль успішно збережено.", b.generateAdminOptionsKeyboard())

	// Очищуємо стан користувача
	delete(userStates, m.Chat.ID)
}

func (b *Bot) handleViewOptions(cq *tbot.CallbackQuery) {
	// Перевіряємо, чи користувач є адміністратором
	if !b.ensureAdmin(cq.Message) {
		return
	}

	// Отримуємо всі варіанти автомобілів з бази даних
	options, err := b.storage.CarOption().GetAll()
	if err != nil {
		b.logger.Error("Failed to get car options: ", err.Error())
		b.sendCallbackMessage(cq, "Щось пішло не так. Спробуйте ще раз.", nil)
		return
	}

	// Перевіряємо, чи є варіанти авто
	if len(options) == 0 {
		b.sendCallbackMessage(cq, "Немає доступних варіантів авто.", nil)
		return
	}

	// Створюємо канал для помилок
	errChan := make(chan error, len(options))

	// Відправляємо кожен варіант авто асинхронно
	for _, option := range options {
		go func(option *models.CarOption) {
			// Створюємо повідомлення з ціною і роком
			message := fmt.Sprintf(
				"%d. %s\n\n📝Опис: %s\n\n💵Ціна: %d\n📅Рік: %s",
				option.ID, option.Title, option.Description, option.Price, option.Year,
			)

			// Створюємо інлайн-клавіатуру з кнопкою "Видалити"
			deleteKeyboard := &tbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]tbot.InlineKeyboardButton{
					{
						{Text: "Видалити ❌", CallbackData: fmt.Sprintf("delete_option_%d", option.ID)},
					},
				},
			}

			// Відправляємо фото автомобіля, якщо воно є
			if option.PhotoID != "" {
				_, err := b.client.SendPhoto(
					cq.Message.Chat.ID,
					option.PhotoID,
					tbot.OptCaption(message),
					tbot.OptInlineKeyboardMarkup(deleteKeyboard),
				)
				errChan <- err
			} else {
				// Якщо фото немає, відправляємо просто текстове повідомлення
				_, err := b.client.SendMessage(cq.Message.Chat.ID, message, tbot.OptInlineKeyboardMarkup(deleteKeyboard))
				errChan <- err
			}
		}(option)
	}

	// Обробляємо помилки після завершення відправлення всіх повідомлень
	for i := 0; i < len(options); i++ {
		if err := <-errChan; err != nil {
			b.logger.Error("Failed to send car option: ", err.Error())
		}
	}

	// Закриваємо канал після завершення операцій
	close(errChan)
}

func (b *Bot) handleDeleteOption(cq *tbot.CallbackQuery) {
	// Перевіряємо, чи користувач є адміністратором
	if !b.ensureAdmin(cq.Message) {
		return
	}

	id := cq.Data[14:]
	carOptionId, err := strconv.Atoi(id)
	if err != nil {
		b.logger.Error("Failed to convert carOptionId to int: ", err.Error())
		b.sendCallbackMessage(cq, "Щось пішло не так. Спробуйте ще раз.", nil)
		return
	}

	// Канали для результатів операцій
	errChan := make(chan error, 2)

	// Видаляємо опцію з бази даних асинхронно
	go func() {
		_, err := b.storage.CarOption().Delete(carOptionId)
		errChan <- err
	}()

	// Видаляємо повідомлення з чату асинхронно
	go func() {
		errChan <- b.client.DeleteMessage(cq.Message.Chat.ID, cq.Message.MessageID)
	}()

	// Обробляємо помилки після виконання обох goroutines
	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			b.logger.Error("Failed to complete operation: ", err.Error())
			b.sendCallbackMessage(cq, "Щось пішло не так. Спробуйте ще раз.", nil)
			return
		}
	}

	// Закриваємо канал після завершення операцій
	close(errChan)
}

func (b *Bot) handleAdminAddFeedback(cq *tbot.CallbackQuery) {
	// Перевіряємо, чи користувач є адміністратором
	if !b.ensureAdmin(cq.Message) {
		b.sendCallbackMessage(cq, "У вас немає прав для цієї команди.", nil)
		return
	}

	// Надсилаємо повідомлення адміністраторам з запитом на введення відгуку з відео
	b.editCallbackMessage(cq, "Будь ласка, введіть опис відгуку.", nil)

	// Оновлюємо стан користувача на очікування введення відгуку
	userStates[cq.Message.Chat.ID] = AwaitingFeedbackDescription
}

func (b *Bot) handleFeedbackDescriptionInput(m *tbot.Message) {
	// Перевіряємо, чи користувач є адміністратором
	if !b.ensureAdmin(m) {
		b.sendMessage(m, "У вас немає прав для цієї команди.", nil)
		return
	}

	if len(m.Text) > 100 {
		b.sendMessage(m, "Опис не має перевищувати 100 символів.", nil)
	}

	feedbackToAdd.Description = m.Text

	b.sendMessage(m, "Тепер надішліть відео для відгуку.", nil)

	userStates[m.Chat.ID] = AwaitingFeedbackVideo
}

func (b *Bot) handleFeedbackVideoInput(m *tbot.Message) {
	// Перевіряємо, чи користувач у стані очікування введення відгуку
	if userStates[m.Chat.ID] != AwaitingFeedbackVideo {
		return
	}

	// Перевіряємо, чи надіслано відео
	if m.Video == nil || m.Video.FileID == "" {
		b.sendMessage(m, "Будь ласка, надішліть відео для відгуку.", nil)
		return
	}

	feedbackToAdd.VideoFileID = m.Video.FileID

	location, err := time.LoadLocation("Europe/Kyiv")
	if err != nil {
		b.logger.Info("Failed to load location: ", err.Error())
	}

	feedbackToAdd.CreatedAt = time.Now().In(location).Format(TimeLayout)

	// Зберігаємо відео у базі даних або надсилаємо до групи разом із текстом відгуку
	err = b.storage.Feedback().Create(feedbackToAdd)

	if err != nil {
		b.logger.Error("Failed to save feedback: ", err.Error())
		b.sendMessage(m, "Щось пішло не так. Спробуйте ще раз.", nil)
		return
	}

	// Відправляємо підтвердження адміністратору
	b.sendMessage(m, "Дякуємо! Ваш відгук був успішно збережений.", nil)

	// Очищуємо стан користувача
	delete(userStates, m.Chat.ID)
}

func (b *Bot) handleAdminViewFeedback(cq *tbot.CallbackQuery) {
	// Перевіряємо, чи користувач є адміністратором
	if !b.ensureAdmin(cq.Message) {
		return
	}

	// Отримуємо всі відгуки з бази даних
	feedbackList, err := b.storage.Feedback().GetAll()
	if err != nil {
		b.logger.Error("Failed to get all feedback: ", err.Error())
		b.sendCallbackMessage(cq, "Щось пішло не так. Спробуйте ще раз.", nil)
		return
	}

	// Перевіряємо, чи є відгуки
	if len(feedbackList) == 0 {
		b.sendCallbackMessage(cq, "Немає доступних відгуків.", nil)
		return
	}

	// Створюємо канал для помилок
	errChan := make(chan error, len(feedbackList))

	// Відправляємо кожен відгук асинхронно
	for _, feedback := range feedbackList {
		go func(fb *models.Feedback) {
			// Створюємо інлайн-клавіатуру з кнопкою "Видалити"
			deleteKeyboard := &tbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]tbot.InlineKeyboardButton{
					{
						{Text: "Видалити ❌", CallbackData: fmt.Sprintf("delete_feedback_%d", fb.ID)},
					},
				},
			}

			createdAt, err := time.Parse(TimeLayout, fb.CreatedAt) // Replace 'TimeLayout' with your actual time layout
			if err != nil {
				b.logger.Error("Failed to parse date: ", err.Error())
				return
			}

			// Format the time in a more user-friendly way
			formattedDate := createdAt.Format("02.01.2006 о 15:04")

			// Якщо є відео, відправляємо його
			_, err = b.client.SendVideo(
				cq.Message.Chat.ID,
				fb.VideoFileID,
				tbot.OptCaption(fmt.Sprintf("%s\n\nЗавантажено %s", feedback.Description, formattedDate)),
				tbot.OptInlineKeyboardMarkup(deleteKeyboard),
			)
			errChan <- err
		}(feedback)
	}

	// Перевіряємо помилки після завершення відправки повідомлень
	for i := 0; i < len(feedbackList); i++ {
		if err := <-errChan; err != nil {
			b.logger.Error("Failed to send feedback: ", err.Error())
		}
	}

	// Закриваємо канал
	close(errChan)
}

func (b *Bot) handleAdminDeleteFeedback(cq *tbot.CallbackQuery) {
	// Check if the user is an admin
	if !b.ensureAdmin(cq.Message) {
		return
	}

	// Extract the feedback ID from the callback data
	feedbackIDStr := cq.Data[16:] // Assuming "delete_feedback_<id>"
	feedbackID, err := strconv.Atoi(feedbackIDStr)
	if err != nil {
		b.logger.Error("Failed to convert feedback ID to int: ", err.Error())
		return
	}

	// Create error channel for handling operations
	errChan := make(chan error, 2)

	// Run feedback deletion and message deletion in parallel
	go func() {
		err := b.storage.Feedback().Delete(feedbackID)
		errChan <- err
	}()

	go func() {
		err := b.client.DeleteMessage(cq.Message.Chat.ID, cq.Message.MessageID)
		errChan <- err
	}()

	// Collect results from both operations
	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			b.logger.Error("Failed to complete operation: ", err.Error())
			return
		}
	}

	// Close the error channel
	close(errChan)
}

func (b *Bot) generateAdminOptionsKeyboard() *tbot.InlineKeyboardMarkup {
	// Створення інлайн-клавіатури з опціями
	return &tbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]tbot.InlineKeyboardButton{
			{
				{Text: "Додати опцію", CallbackData: "add_option"},
				{Text: "Переглянути опції", CallbackData: "view_options"},
			},
			{
				{Text: "Додати відгук", CallbackData: "add_feedback"},
				{Text: "Переглянути відгуки", CallbackData: "view_feedback"},
			},
		},
	}
}
