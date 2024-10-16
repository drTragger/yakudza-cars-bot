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
func (b *Bot) handleAddOption(cq *tbot.CallbackQuery) {
	if !b.isAdmin(cq.Message) {
		b.sendMessage(cq.Message, "У вас немає прав для цієї команди.", nil)
		return
	}

	b.editMessage(cq.Message, "Введіть назву автомобіля:", nil)

	b.setUserState(cq.Message.Chat.ID, app.AwaitingCarTitle)
}

// Step 2
func (b *Bot) handleAdminTitleInput(m *tbot.Message) {
	if !b.ensureAdmin(m) {
		return
	}

	if len(m.Text) > 255 {
		b.sendMessage(m, "Назва не має перевищувати 255 символів.", nil)
		return
	}

	b.setCarOption(m.Chat.ID, &models.CarOption{Title: m.Text})

	b.sendMessage(m, "Введіть опис автомобіля:", nil)

	b.setUserState(m.Chat.ID, app.AwaitingCarDescription)
}

// Step 3
func (b *Bot) handleAdminDescriptionInput(m *tbot.Message) {
	if !b.ensureAdmin(m) {
		return
	}

	if len(m.Text) > 500 {
		b.sendMessage(m, "Опис не має перевищувати 500 символів.", nil)
		return
	}

	carOption := b.getCarOption(m.Chat.ID)
	carOption.Description = m.Text
	b.setCarOption(m.Chat.ID, carOption)

	b.sendMessage(m, "Будь ласка, введіть ціну для автомобіля (у числовому форматі).", nil)

	b.setUserState(m.Chat.ID, app.AwaitingCarPrice)
}

// Step 4
func (b *Bot) handleAdminPriceInput(m *tbot.Message) {
	if !b.ensureAdmin(m) {
		return
	}

	price, err := strconv.Atoi(m.Text)
	if err != nil || price <= 0 {
		b.sendMessage(m, "Будь ласка, введіть коректну ціну (тільки цифри).", nil)
		return
	}

	carOption := b.getCarOption(m.Chat.ID)
	carOption.Price = price
	b.setCarOption(m.Chat.ID, carOption)

	years := generateYears(app.CarOptionsStartYear)

	yearKeyboard := &tbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]tbot.InlineKeyboardButton{},
	}

	for i := 0; i < len(years); i += 4 {
		var row []tbot.InlineKeyboardButton

		for j := i; j < i+4 && j < len(years); j++ {
			row = append(row, tbot.InlineKeyboardButton{
				Text:         years[j],
				CallbackData: fmt.Sprintf("year_%s", years[j]),
			})
		}

		yearKeyboard.InlineKeyboard = append(yearKeyboard.InlineKeyboard, row)
	}

	b.sendMessage(m, "Тепер оберіть рік для автомобіля:", yearKeyboard)

	b.setUserState(m.Chat.ID, app.AwaitingCarYear)
}

// Step 5
func (b *Bot) handleAdminYearSelection(cq *tbot.CallbackQuery) {
	if !b.ensureAdmin(cq.Message) {
		return
	}

	selectedYear := cq.Data[5:]
	b.logger.Infof("Рік обраний: %s", selectedYear)

	carOption := b.getCarOption(cq.Message.Chat.ID)
	carOption.Year = selectedYear
	b.setCarOption(cq.Message.Chat.ID, carOption)

	b.editMessage(cq.Message, "Будь ласка, надішліть фото автомобіля.", nil)

	b.setUserState(cq.Message.Chat.ID, app.AwaitingCarPhoto)
}

// Step 6
func (b *Bot) handleAdminPhotoInput(m *tbot.Message) {
	if !b.ensureAdmin(m) {
		return
	}

	carOption := b.getCarOption(m.Chat.ID)

	location, err := time.LoadLocation("Europe/Kyiv")
	if err != nil {
		b.logger.Info("Failed to load location: ", err.Error())
	}

	carOption.PhotoID = m.Photo[0].FileID
	carOption.CreatedAt = time.Now().In(location).Format(app.TimeLayout)
	err = b.storage.CarOption().Create(carOption)
	if err != nil {
		b.logger.Error("Failed to create new car option: ", err.Error())
		b.sendMessage(m, "Щось пішло не так. Спробуйте ще раз.", nil)
		return
	}
	b.logger.Infof("Фото отримано: %s", m.Photo[0].FileID)

	b.sendMessage(m, "Цей автомобіль успішно збережено.", utils.GetAdminMenuKeyboard())

	b.deleteUserState(m.Chat.ID)
	b.deleteCarOption(m.Chat.ID)
}

func (b *Bot) handleViewOptions(cq *tbot.CallbackQuery) {
	if !b.ensureAdmin(cq.Message) {
		return
	}

	options, err := b.storage.CarOption().GetAll()
	if err != nil {
		b.logger.Error("Failed to get car options: ", err.Error())
		b.sendMessage(cq.Message, "Щось пішло не так. Спробуйте ще раз.", nil)
		return
	}

	if len(options) == 0 {
		b.sendMessage(cq.Message, "Немає доступних варіантів авто.", nil)
		return
	}

	errChan := make(chan error, len(options))

	for _, option := range options {
		go func(option *models.CarOption) {
			message := fmt.Sprintf(
				"%d. %s\n\n📝Опис:\n%s\n\n💵Ціна: %d$\n📅Рік: %s",
				option.ID, option.Title, option.Description, option.Price, option.Year,
			)

			_, err := b.client.SendPhoto(
				cq.Message.Chat.ID,
				option.PhotoID,
				tbot.OptCaption(message),
				tbot.OptInlineKeyboardMarkup(utils.GetDeleteOptionKeyboard(option.ID)),
			)
			errChan <- err
		}(option)
	}

	for i := 0; i < len(options); i++ {
		if err := <-errChan; err != nil {
			b.logger.Error("Failed to send car option: ", err.Error())
		}
	}

	close(errChan)
}

func (b *Bot) handleDeleteOption(cq *tbot.CallbackQuery) {
	if !b.ensureAdmin(cq.Message) {
		return
	}

	id := cq.Data[14:]
	carOptionId, err := strconv.Atoi(id)
	if err != nil {
		b.logger.Error("Failed to convert carOptionId to int: ", err.Error())
		b.sendMessage(cq.Message, "Щось пішло не так. Спробуйте ще раз.", nil)
		return
	}

	errChan := make(chan error, 2)

	go func() {
		_, err := b.storage.CarOption().Delete(carOptionId)
		errChan <- err
	}()

	go func() {
		errChan <- b.client.DeleteMessage(cq.Message.Chat.ID, cq.Message.MessageID)
	}()

	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			b.logger.Error("Failed to complete operation: ", err.Error())
			b.sendMessage(cq.Message, "Щось пішло не так. Спробуйте ще раз.", nil)
			return
		}
	}

	close(errChan)
}
