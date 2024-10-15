package bot

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/drTragger/yakudza-cars-bot/internal/app/models"
	"github.com/yanzay/tbot/v2"
	"strconv"
	"time"
)

func (b *Bot) HandleMessage(m *tbot.Message) {
	if m.Chat.ID == b.config.GroupID {
		return
	}

	switch m.Text {
	// Handle main menu options
	case "Підібрати Авто":
		b.handleCarSelection(m) // Start the car selection process
	case "Відгуки":
		b.handleFeedback(m) // Handle feedback
	case "Бай Нау":
		b.sendMessage(m, "Перейдіть за посиланням:\nhttps://t.me/yakudzaoffer", nil)
	default:
		if m.Contact != nil {
			b.handlePhoneNumber(m)
			return
		}

		if m.Photo != nil && userStates[m.Chat.ID] == AwaitingPhoto {
			b.handleAdminPhoto(m)
			return
		}

		if m.Video != nil && userStates[m.Chat.ID] == AwaitingFeedbackVideo {
			b.handleFeedbackVideoInput(m)
			return
		}

		// Перевіряємо, чи є файл (документ)
		if m.Document != nil {
			if userStates[m.Chat.ID] == AwaitingFeedbackVideo {
				b.sendMessage(m, "Будь ласка, відправте відео, а не файл.", nil)
				return
			}
			b.sendMessage(m, "Будь ласка, відправте фото, а не файл.", nil)
			return
		}

		if userStates[m.Chat.ID] == AwaitingTitle {
			b.handleAdminTitleSelection(m)
			return
		}

		if userStates[m.Chat.ID] == AwaitingDescription {
			b.handleAdminDescriptionSelection(m)
			return
		}

		if userStates[m.Chat.ID] == AwaitingPrice {
			b.handleAdminPriceInput(m)
			return
		}

		if userStates[m.Chat.ID] == AwaitingFeedbackDescription {
			b.handleFeedbackDescriptionInput(m)
			return
		}

		// Check if the message matches a price selection
		for _, price := range prices {
			if m.Text == price.Title {
				b.handlePriceSelection(m, price) // Call the function to handle the price selection
				return
			}
		}

		// Check if the message matches a year selection
		years := generateYears(StartYear)
		for _, year := range years {
			if m.Text == year {
				b.handleYearSelection(m, year) // Call the function to handle the year selection
				return
			}
		}

		// If nothing matches, show a default message
		b.sendMessage(m, "Невідома команда. Будь ласка, оберіть варіант з меню.", nil)
	}
}

func (b *Bot) handleCarSelection(m *tbot.Message) {
	// Send a message with the reply keyboard for price selection
	b.sendMessage(m, "Будь ласка, оберіть ціновий діапазон:", b.generatePriceKeyboard())
}

func (b *Bot) generatePriceKeyboard() *tbot.ReplyKeyboardMarkup {
	priceKeyboard := &tbot.ReplyKeyboardMarkup{
		ResizeKeyboard:  true, // Make the keyboard fit the screen
		OneTimeKeyboard: true, // Keep the keyboard persistent
	}

	// Loop through the prices and arrange them into rows with 2 columns
	for i := 0; i < len(prices); i += 2 {
		var row []tbot.KeyboardButton

		// Add up to 2 prices in each row
		for j := i; j < i+2 && j < len(prices); j++ {
			row = append(row, tbot.KeyboardButton{
				Text: prices[j].Title, // The button text will be sent as the user's message
			})
		}

		// Append the row to the keyboard
		priceKeyboard.Keyboard = append(priceKeyboard.Keyboard, row)
	}

	return priceKeyboard
}

func (b *Bot) handlePriceSelection(m *tbot.Message, selectedPrice *models.PriceRange) {
	// Save the selected price
	if _, exists := carData[m.Chat.ID]; !exists {
		carData[m.Chat.ID] = &models.CarDetails{}
	}
	carData[m.Chat.ID].Price = selectedPrice

	// Proceed to the next step: ask for the car year
	b.askForCarYear(m)
}

func (b *Bot) askForCarYear(m *tbot.Message) {
	// Generate dynamic list of years from 2014 to the current year
	years := generateYears(StartYear)

	// Create a reply keyboard for selecting a year, with multiple columns
	yearKeyboard := &tbot.ReplyKeyboardMarkup{
		ResizeKeyboard:  true,
		OneTimeKeyboard: true,
	}

	// Loop through the years and arrange them into rows with 3 columns
	for i := 0; i < len(years); i += 3 {
		var row []tbot.KeyboardButton

		// Add up to 3 years in each row
		for j := i; j < i+3 && j < len(years); j++ {
			row = append(row, tbot.KeyboardButton{
				Text: years[j], // The button text will be sent as the user's message
			})
		}

		// Append the row to the keyboard
		yearKeyboard.Keyboard = append(yearKeyboard.Keyboard, row)
	}

	// Send a message asking the user to select a car year
	b.sendMessage(m, "Дякую! Тепер оберіть рік авто:", yearKeyboard)
}

func (b *Bot) handleYearSelection(m *tbot.Message, selectedYear string) {
	// Збереження вибраного року
	if _, exists := carData[m.Chat.ID]; !exists {
		carData[m.Chat.ID] = &models.CarDetails{}
	}
	carData[m.Chat.ID].Year = selectedYear

	// Підтвердження вибору автомобіля
	b.sendMessage(m, "Дякую! Ви обрали авто з ціною "+carData[m.Chat.ID].Price.Title+" і роком випуску "+carData[m.Chat.ID].Year, b.getMenuKeyboard())

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

	b.sendCarDetailsToGroup(user.Phone, carData[m.Chat.ID].Price.Title, carData[m.Chat.ID].Year)
	b.showCarOption(m)
}

func (b *Bot) handleFeedback(m *tbot.Message) {
	userStates[m.Chat.ID] = "show_feedback"

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
	// Remove any previous user state
	delete(userStates, m.Chat.ID)

	// Get the next feedback from the database
	feedback, err := b.storage.Feedback().GetNext(shownFeedbackIDs[m.Chat.ID])
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
	shownFeedbackIDs[m.Chat.ID] = append(shownFeedbackIDs[m.Chat.ID], feedback.ID)

	// Create inline keyboard without the "Хочу ще" button by default
	inlineKeyboard := &tbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]tbot.InlineKeyboardButton{},
	}

	// Check if there is more feedback
	_, err = b.storage.Feedback().GetNext(shownFeedbackIDs[m.Chat.ID])
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
}

func (b *Bot) requestPhoneNumber(m *tbot.Message) {
	// Створюємо клавіатуру з кнопкою для запиту номера телефону
	phoneKeyboard := &tbot.ReplyKeyboardMarkup{
		ResizeKeyboard:  true, // Підлаштувати клавіатуру під розмір екрану
		OneTimeKeyboard: true, // Сховати клавіатуру після використання
		Keyboard: [][]tbot.KeyboardButton{
			{
				tbot.KeyboardButton{
					Text:           "Поділитися номером телефону",
					RequestContact: true, // Запит номера телефону у користувача
				},
			},
		},
	}

	// Надсилаємо повідомлення з проханням поділитися номером телефону
	b.sendMessage(m, "Будь ласка, поділіться вашим номером телефону, щоб продовжити:", phoneKeyboard)
}

func (b *Bot) handlePhoneNumber(m *tbot.Message) {
	// Ensure `m.Contact` is not nil
	if m.Contact == nil {
		b.logger.Error("No contact information provided")
		b.sendMessage(m, "Не було надано контактної інформації.", nil)
		return
	}

	// Get the phone number from the message
	phoneNumber := m.Contact.PhoneNumber

	// Use channels to manage asynchronous error handling
	errChan := make(chan error, 2)

	// Asynchronously fetch location and create a user
	go func() {
		location, err := time.LoadLocation("Europe/Kyiv")
		if err != nil {
			b.logger.Info("Failed to load location: ", err.Error())
			errChan <- err
			return
		}

		// Create the user in the database
		err = b.storage.User().Create(&models.User{
			ChatId:    m.Chat.ID,
			Phone:     phoneNumber,
			CreatedAt: time.Now().In(location).Format(TimeLayout),
		})

		errChan <- err
	}()

	// Handle state for showing feedback
	if userStates[m.Chat.ID] == "show_feedback" {
		b.showFeedback(m)
		return
	}

	// Asynchronously update `carData` with the phone number
	go func() {
		carDataMutex.Lock() // Lock before writing to the map
		defer carDataMutex.Unlock()

		// Ensure `carData` for this user exists
		if _, exists := carData[m.Chat.ID]; !exists {
			carData[m.Chat.ID] = &models.CarDetails{}
		}
		carData[m.Chat.ID].Phone = phoneNumber

		// Send confirmation to the user
		b.sendMessage(m, "Дякую! Ми отримали ваш номер телефону: "+phoneNumber+".", b.getMenuKeyboard())

		errChan <- nil
	}()

	// Send details to the group asynchronously
	go func() {
		carDataMutex.Lock()
		defer carDataMutex.Unlock()

		b.sendCarDetailsToGroup(phoneNumber, carData[m.Chat.ID].Price.Title, carData[m.Chat.ID].Year)
	}()

	// Process potential errors from goroutines
	for i := 0; i < 2; i++ {
		if err := <-errChan; err != nil {
			b.logger.Error("Error occurred: ", err.Error())
		}
	}

	// Close the error channel after all operations
	close(errChan)

	// Display car options after operations complete
	b.showCarOption(m)
}

func (b *Bot) sendCarDetailsToGroup(phoneNumber string, price string, year string) {
	// Формуємо повідомлення для групи
	groupMessage := fmt.Sprintf(
		"Користувач +%s\nОбрав параметри:\nЦіна: %s\nРік: %s",
		phoneNumber, price, year,
	)

	_, err := b.client.SendMessage(b.config.GroupID, groupMessage)
	if err != nil {
		b.logger.Error("Failed to send message to the group: ", err.Error())
	}
}

func (b *Bot) showCarOption(m *tbot.Message) {
	selectedCar, exists := carData[m.Chat.ID]
	if !exists {
		b.sendMessage(m, "Ви не обрали жодного параметру. Будь ласка, пройдіть опитування:", b.generatePriceKeyboard())
		return
	}

	carOption, err := b.storage.CarOption().GetByDetails(selectedCar, shownOptionIDs[m.Chat.ID])
	if errors.Is(err, sql.ErrNoRows) {
		b.logger.Info("Немає варіантів авто.")
		b.sendMessage(m, "Більше немає варіантів авто.", nil)
		return
	} else if err != nil {
		b.logger.Error("Failed to get car option: ", err.Error())
		b.sendMessage(m, "Щось пішло не так. Спробуйте ще раз.", b.generatePriceKeyboard())
		return
	}
	shownOptionIDs[m.Chat.ID] = append(shownOptionIDs[m.Chat.ID], carOption.ID)

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
	otherCarOption, err := b.storage.CarOption().GetByDetails(selectedCar, shownOptionIDs[m.Chat.ID])
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

	// Відправлення повідомлення з інформацією про авто та інлайн-клавіатурою
	if carOption.PhotoID != "" {
		// Якщо є фото авто, відправляємо його
		_, err := b.client.SendPhoto(
			m.Chat.ID,
			carOption.PhotoID,
			tbot.OptCaption(message), // Форматоване повідомлення
			tbot.OptInlineKeyboardMarkup(inlineKeyboard), // Додаємо інлайн-клавіатуру
		)
		if err != nil {
			b.logger.Error("Failed to send photo to the group: ", err.Error())
		}
	} else {
		// Якщо фото немає, просто відправляємо текстове повідомлення
		_, err := b.client.SendMessage(
			m.Chat.ID,
			message,
			tbot.OptInlineKeyboardMarkup(inlineKeyboard), // Додаємо інлайн-клавіатуру
		)
		if err != nil {
			b.logger.Error("Failed to send message to the group: ", err.Error())
		}
	}
}
