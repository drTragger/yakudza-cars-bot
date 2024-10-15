package bot

import (
	"fmt"
	"github.com/drTragger/yakudza-cars-bot/internal/app/models"
	"github.com/yanzay/tbot/v2"
	"strconv"
)

func (b *Bot) HandleCallback(cq *tbot.CallbackQuery) {
	switch {
	case cq.Data == "add_option":
		// Відповідь для "Додати опцію"
		b.handleAddOptions(cq)
	case cq.Data == "view_options":
		b.handleViewOptions(cq)
	case cq.Data == "add_feedback":
		b.handleAdminAddFeedback(cq)
	case cq.Data == "view_feedback":
		b.handleAdminViewFeedback(cq)
	case cq.Data == "more_cars":
		b.showCarOption(cq.Message)
	case cq.Data == "more_feedback":
		b.showFeedback(cq.Message)
	case cq.Data[:5] == "year_":
		b.handleAdminYearSelection(cq)
	case cq.Data[:11] == "select_car_":
		b.handleSelectCar(cq)
	case cq.Data[:14] == "delete_option_":
		b.handleDeleteOption(cq)
	case cq.Data[:16] == "delete_feedback_":
		b.handleAdminDeleteFeedback(cq)

	default:
		// Обробка невідомих варіантів
		b.sendCallbackMessage(cq, "Невідома дія.", nil)
	}
}

func (b *Bot) handleSelectCar(cq *tbot.CallbackQuery) {
	id := cq.Data[11:]
	carOptionId, err := strconv.Atoi(id)
	if err != nil {
		b.logger.Error("Failed to convert carOptionId to int: ", err.Error())
		b.sendCallbackMessage(cq, "Щось пішло не так. Спробуйте ще раз.", nil)
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
			b.sendCallbackMessage(cq, "Щось пішло не так. Спробуйте ще раз.", nil)
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
		_, err = b.client.SendPhoto(b.config.GroupID, carOption.PhotoID, tbot.OptCaption(groupMessage))
		if err != nil {
			b.logger.Error("Failed to send message to the group: ", err.Error())
			b.sendCallbackMessage(cq, "Щось пішло не так. Спробуйте ще раз.", nil)
			return
		}
	}()

	// Надсилаємо подяку користувачеві після вибору автомобіля
	thankYouMessage := "Дякуємо за ваш вибір! Ми зателефонуємо вам найближчим часом для уточнення деталей."
	b.sendMessage(cq.Message, thankYouMessage, nil)
}
