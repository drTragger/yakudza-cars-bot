package utils

import (
	"fmt"
	"github.com/drTragger/yakudza-cars-bot/internal/app/models"
	"github.com/yanzay/tbot/v2"
)

var Prices = []*models.PriceRange{
	{Title: "До 10 000$", Min: 0, Max: 9999},
	{Title: "10 000$ - 15 000$", Min: 10000, Max: 14999},
	{Title: "15 000$ - 20 000$", Min: 15000, Max: 19999},
	{Title: "20 000$ - 30 000$", Min: 20000, Max: 29999},
	{Title: "30 000$ - 35 000$", Min: 30000, Max: 34999},
	{Title: "35 000$ +", Min: 35000, Max: 1000000},
}

func GetContactKeyboard() *tbot.ReplyKeyboardMarkup {
	return &tbot.ReplyKeyboardMarkup{
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
}

func GetMenuKeyboard() *tbot.ReplyKeyboardMarkup {
	return &tbot.ReplyKeyboardMarkup{
		ResizeKeyboard:  true,
		OneTimeKeyboard: false, // Keep the keyboard after use
		Keyboard: [][]tbot.KeyboardButton{
			{
				{Text: "Підібрати Авто"}, // Button to start the car selection process
				{Text: "Відгуки"},        // Button to allow feedback
			},
			{
				{Text: "Бай Нау"}, // Third button at the bottom
			},
		},
	}
}

func GetPriceKeyboard() *tbot.ReplyKeyboardMarkup {
	priceKeyboard := &tbot.ReplyKeyboardMarkup{
		ResizeKeyboard:  true, // Make the keyboard fit the screen
		OneTimeKeyboard: true, // Keep the keyboard persistent
	}

	// Loop through the prices and arrange them into rows with 2 columns
	for i := 0; i < len(Prices); i += 2 {
		var row []tbot.KeyboardButton

		// Add up to 2 prices in each row
		for j := i; j < i+2 && j < len(Prices); j++ {
			row = append(row, tbot.KeyboardButton{
				Text: Prices[j].Title, // The button text will be sent as the user's message
			})
		}

		// Append the row to the keyboard
		priceKeyboard.Keyboard = append(priceKeyboard.Keyboard, row)
	}

	return priceKeyboard
}

func GetAdminMenuKeyboard() *tbot.InlineKeyboardMarkup {
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

func GetDeleteOptionKeyboard(optionID int) *tbot.InlineKeyboardMarkup {
	return &tbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]tbot.InlineKeyboardButton{
			{
				{Text: "Видалити ❌", CallbackData: fmt.Sprintf("delete_option_%d", optionID)},
			},
		},
	}
}

func GetDeleteFeedbackKeyboard(feedbackID int) *tbot.InlineKeyboardMarkup {
	return &tbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]tbot.InlineKeyboardButton{
			{
				{Text: "Видалити ❌", CallbackData: fmt.Sprintf("delete_feedback_%d", feedbackID)},
			},
		},
	}
}
