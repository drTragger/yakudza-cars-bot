package bot

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/drTragger/yakudza-cars-bot/internal/app"
	"github.com/drTragger/yakudza-cars-bot/internal/app/models"
	"github.com/drTragger/yakudza-cars-bot/internal/app/utils"
	"github.com/yanzay/tbot/v2"
	"strconv"
)

func (b *Bot) handleCarSelection(m *tbot.Message) {
	// Send a message with the reply keyboard for price selection
	b.sendMessage(m, "Будь ласка, оберіть ціновий діапазон:", utils.GetPriceKeyboard())
}

func (b *Bot) handlePriceSelection(m *tbot.Message, selectedPrice *models.PriceRange) {
	b.setCarData(m.Chat.ID, &models.CarDetails{Price: selectedPrice})

	b.askForCarYear(m)
}

func (b *Bot) askForCarYear(m *tbot.Message) {
	years := generateYears(app.CarOptionsStartYear)

	yearKeyboard := &tbot.ReplyKeyboardMarkup{
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	}

	for i := 0; i < len(years); i += 3 {
		var row []tbot.KeyboardButton

		for j := i; j < i+3 && j < len(years); j++ {
			row = append(row, tbot.KeyboardButton{
				Text: years[j],
			})
		}

		yearKeyboard.Keyboard = append(yearKeyboard.Keyboard, row)
	}

	b.sendMessage(m, "Дякую! Тепер оберіть рік авто:", yearKeyboard)
}

func (b *Bot) handleYearSelection(m *tbot.Message, selectedYear string) {
	carData := b.getCarData(m.Chat.ID)
	carData.Year = selectedYear

	// Підтвердження вибору автомобіля
	b.sendMessage(m, fmt.Sprintf(
		"Дякую! Ви обрали авто з ціною %s і роком випуску %s",
		carData.Price.Title, carData.Year),
		utils.GetMenuKeyboard(),
	)

	chatId, err := strconv.Atoi(m.Chat.ID)
	if err != nil {
		b.logger.Error("Failed to convert chatId to int: ", err.Error())
		b.sendMessage(m, "Щось пішло не так. Спробуйте ще раз.", nil)
		return
	}

	user, err := b.storage.User().FindByChatId(chatId)
	if errors.Is(err, sql.ErrNoRows) {
		// Запит номера телефону
		b.requestPhoneNumber(m)
		return
	} else if err != nil {
		b.logger.Error("Failed to find user by chat id: ", err.Error())
		b.sendMessage(m, "Щось пішло не так. Спробуйте ще раз.", nil)
		return
	}

	b.sendCarDetailsToGroup(user.Phone, carData.Price.Title, carData.Year)
	b.showCarOption(m)
}

func (b *Bot) sendCarDetailsToGroup(phoneNumber string, price string, year string) {
	// Формуємо повідомлення для групи
	groupMessage := fmt.Sprintf(
		"Користувач +%s\nОбрав параметри:\nЦіна: %s\nРік: %s",
		phoneNumber, price, year,
	)

	_, err := b.client.SendMessage(b.config.Admin.GroupID, groupMessage)
	if err != nil {
		b.logger.Error("Failed to send message to the group: ", err.Error())
	}
}

func (b *Bot) showCarOption(m *tbot.Message) {
	selectedCar := b.getCarData(m.Chat.ID)
	if selectedCar == nil || *selectedCar == (models.CarDetails{}) {
		b.sendMessage(m, "Ви не обрали жодного параметру. Будь ласка, пройдіть опитування:", utils.GetPriceKeyboard())
		return
	}

	shownOptions := b.getShownOptionIDs(m.Chat.ID)

	carOption, err := b.storage.CarOption().GetByDetails(selectedCar, shownOptions)
	if errors.Is(err, sql.ErrNoRows) {
		b.logger.Info("Немає варіантів авто.")
		b.sendMessage(m, "Ще більше варіантів Вам запропонує наш менеджер, він скоро з Вами звʼяжеться 🔜", nil)
		b.deleteCarData(m.Chat.ID)
		return
	} else if err != nil {
		b.logger.Error("Failed to get car option: ", err.Error())
		b.sendMessage(m, "Щось пішло не так. Спробуйте ще раз.", utils.GetPriceKeyboard())
		return
	}

	shownOptions = append(shownOptions, carOption.ID)
	b.setShownOptionIDs(m.Chat.ID, shownOptions)

	// Створення повідомлення з даними про авто
	message := fmt.Sprintf(
		"🚗 %s\n\n%s\n\n💵 Ціна: %d$\n📅 Рік: %s",
		carOption.Title, carOption.Description, carOption.Price, carOption.Year,
	)

	// Створюємо інлайн-клавіатуру з кнопкою "Хочу це авто"
	inlineKeyboard := &tbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]tbot.InlineKeyboardButton{
			{
				{Text: "Хочу це авто", CallbackData: fmt.Sprintf("select_car_%d", carOption.ID)},
			},
		},
	}

	// Перевірка наявності інших варіантів авто
	otherCarOption, err := b.storage.CarOption().GetByDetails(selectedCar, shownOptions)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Якщо немає інших варіантів авто
			b.logger.Info("Немає інших варіантів авто.")
		} else {
			// Якщо сталася інша помилка
			b.logger.Error("Failed to check for other cars: ", err.Error())
			b.sendMessage(m, "Щось пішло не так. Спробуйте ще раз.", nil)
			return
		}
	}

	// Додаємо кнопку "Хочу ще", якщо є інші варіанти авто
	if otherCarOption != nil {
		inlineKeyboard.InlineKeyboard[0] = append(inlineKeyboard.InlineKeyboard[0], tbot.InlineKeyboardButton{
			Text:         "Хочу ще",
			CallbackData: "more_cars",
		})
	}

	// Якщо є фото авто, відправляємо його
	_, err = b.client.SendPhoto(
		m.Chat.ID,
		carOption.PhotoID,
		tbot.OptCaption(message), // Форматоване повідомлення
		tbot.OptInlineKeyboardMarkup(inlineKeyboard), // Додаємо інлайн-клавіатуру
	)
	if err != nil {
		b.logger.Error("Failed to send photo to the group: ", err.Error())
	}

	if otherCarOption == nil {
		contactKeyboard := &tbot.InlineKeyboardMarkup{
			InlineKeyboard: [][]tbot.InlineKeyboardButton{
				{
					{Text: "Звʼязатися", CallbackData: "contact_us"},
				},
			},
		}

		b.sendMessage(m, "Оберіть один з наданих варіантів, або натисніть звʼязатися.", contactKeyboard)
	}
}

func (b *Bot) handleSelectCar(cq *tbot.CallbackQuery) {
	id := cq.Data[11:]
	carOptionId, err := strconv.Atoi(id)
	if err != nil {
		b.logger.Error("Failed to convert carOptionId to int: ", err.Error())
		b.sendMessage(cq.Message, "Щось пішло не так. Спробуйте ще раз.", nil)
		return
	}

	// Канали для результатів
	carOptionChan := make(chan *models.CarOption)
	userChan := make(chan *models.User)
	errChan := make(chan error, 2)

	// Асинхронне отримання даних про автомобіль
	go func() {
		carOption, err := b.storage.CarOption().GetByID(carOptionId)
		if err != nil {
			errChan <- err
			return
		}
		carOptionChan <- carOption
	}()

	// Асинхронне отримання даних користувача
	go func() {
		chatId, err := strconv.Atoi(cq.Message.Chat.ID)
		if err != nil {
			errChan <- err
			return
		}

		user, err := b.storage.User().FindByChatId(chatId)
		if err != nil {
			errChan <- err
			return
		}
		userChan <- user
	}()

	// Перевіряємо результати роботи горутин
	var carOption *models.CarOption
	var user *models.User

	for i := 0; i < 2; i++ {
		select {
		case carOption = <-carOptionChan:
		case user = <-userChan:
		case err = <-errChan:
			b.logger.Error("Error occurred: ", err.Error())
			b.sendMessage(cq.Message, "Щось пішло не так. Спробуйте ще раз.", nil)
			return
		}
	}

	// Видаляємо повідомлення з варіантом автомобіля, яке вибрав користувач
	go func() {
		err = b.client.DeleteMessage(cq.Message.Chat.ID, cq.Message.MessageID)
		if err != nil {
			b.logger.Error("Failed to delete message: ", err.Error())
		}
	}()

	// Форматування повідомлення для групи
	groupMessage := fmt.Sprintf(
		"+%s:\n%d. %s",
		user.Phone, carOption.ID, carOption.Title,
	)

	// Відправлення повідомлення з фото в групу
	go func() {
		_, err = b.client.SendPhoto(b.config.Admin.GroupID, carOption.PhotoID, tbot.OptCaption(groupMessage))
		if err != nil {
			b.logger.Error("Failed to send message to the group: ", err.Error())
			b.sendMessage(cq.Message, "Щось пішло не так. Спробуйте ще раз.", nil)
			return
		}
	}()

	// Надсилаємо подяку користувачеві після вибору автомобіля
	thankYouMessage := "Дякуємо за ваш вибір! Ми зателефонуємо вам найближчим часом для уточнення деталей."
	b.sendMessage(cq.Message, thankYouMessage, nil)
}
