package bot

import (
	"github.com/drTragger/yakudza-cars-bot/internal/app"
	"github.com/drTragger/yakudza-cars-bot/internal/app/models"
	"github.com/drTragger/yakudza-cars-bot/internal/app/utils"
	"github.com/yanzay/tbot/v2"
	"strings"
	"time"
)

func (b *Bot) requestPhoneNumber(m *tbot.Message) {
	b.sendMessage(m, "Будь ласка, поділіться вашим номером телефону, щоб продовжити:", utils.GetContactKeyboard())

	b.setUserState(m.Chat.ID, app.AwaitingPhone)
}

func (b *Bot) handlePhoneNumber(m *tbot.Message) {
	// Ensure `m.Contact` is not nil
	if m.Contact == nil {
		b.logger.Error("No contact information provided")
		b.sendMessage(m, "Не було надано контактної інформації.", nil)
		return
	}

	carData := b.getCarData(m.Chat.ID)

	// Get the phone number from the message
	phoneNumber := strings.TrimLeft(m.Contact.PhoneNumber, "+")

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
			CreatedAt: time.Now().In(location).Format(app.TimeLayout),
		})

		errChan <- err
	}()

	// Handle state for showing feedback
	if b.getUserState(m.Chat.ID) == "show_feedback" {
		b.showFeedback(m)
		return
	}

	// Asynchronously update `carData` with the phone number
	go func() {
		carData.Phone = phoneNumber
		b.setCarData(m.Chat.ID, carData)

		// Send confirmation to the user
		b.sendMessage(m, "Дякую! Ми отримали ваш номер телефону: "+phoneNumber+".", utils.GetMenuKeyboard())

		errChan <- nil
	}()

	// Send details to the group asynchronously
	go b.sendCarDetailsToGroup(phoneNumber, carData.Price.Title, carData.Year)

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

	b.deleteUserState(m.Chat.ID)
}
