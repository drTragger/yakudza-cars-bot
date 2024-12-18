package bot

import (
	"github.com/drTragger/yakudza-cars-bot/internal/app/models"
	"github.com/patrickmn/go-cache"
	"github.com/yanzay/tbot/v2"
)

// Car data
func (b *Bot) setCarData(chatID string, carData *models.CarDetails) {
	b.cache.Set(chatID+"_carData", carData, cache.DefaultExpiration)
}

func (b *Bot) getCarData(chatID string) *models.CarDetails {
	if data, found := b.cache.Get(chatID + "_carData"); found {
		return data.(*models.CarDetails)
	}
	return nil
}

func (b *Bot) deleteCarData(chatID string) {
	b.cache.Delete(chatID + "_carData")
}

// Shown car option IDs
func (b *Bot) setShownOptionIDs(chatID string, ids []int) {
	b.cache.Set(chatID+"_shownOptionIDs", ids, cache.DefaultExpiration)
}

func (b *Bot) getShownOptionIDs(chatID string) []int {
	if ids, found := b.cache.Get(chatID + "_shownOptionIDs"); found {
		return ids.([]int)
	}
	return []int{}
}

func (b *Bot) setSelectedCar(chatID string, cq *tbot.CallbackQuery) {
	b.cache.Set(chatID+"_selectedCar", cq, cache.DefaultExpiration)
}

func (b *Bot) getSelectedCar(chatID string) *tbot.CallbackQuery {
	if cq, found := b.cache.Get(chatID + "_selectedCar"); found {
		return cq.(*tbot.CallbackQuery)
	}
	return nil
}

func (b *Bot) deleteSelectedCar(chatID string) {
	b.cache.Delete(chatID + "_selectedCar")
}

// User states
func (b *Bot) setUserState(chatID string, state string) {
	b.cache.Set(chatID+"_state", state, cache.DefaultExpiration)
}

func (b *Bot) getUserState(chatID string) string {
	if state, found := b.cache.Get(chatID + "_state"); found {
		return state.(string)
	}
	return ""
}

func (b *Bot) deleteUserState(chatID string) {
	b.cache.Delete(chatID + "_state")
}

// Car options
func (b *Bot) setCarOption(chatID string, option *models.CarOption) {
	b.cache.Set(chatID+"_carOption", option, cache.DefaultExpiration)
}

func (b *Bot) getCarOption(chatID string) *models.CarOption {
	if data, found := b.cache.Get(chatID + "_carOption"); found {
		return data.(*models.CarOption)
	}
	return nil
}

func (b *Bot) deleteCarOption(chatID string) {
	b.cache.Delete(chatID + "_carOption")
}

// Feedback
func (b *Bot) setFeedback(chatID string, feedback *models.Feedback) {
	b.cache.Set(chatID+"_feedback", feedback, cache.DefaultExpiration)
}

func (b *Bot) getFeedback(chatID string) *models.Feedback {
	if data, found := b.cache.Get(chatID + "_feedback"); found {
		return data.(*models.Feedback)
	}
	return nil
}

func (b *Bot) deleteFeedback(chatID string) {
	b.cache.Delete(chatID + "_feedback")
}

// Shown feedback
func (b *Bot) setShownFeedbackIDs(chatID string, ids []int) {
	b.cache.Set(chatID+"_shownFeedbackIDs", ids, cache.DefaultExpiration)
}

func (b *Bot) getShownFeedbackIDs(chatID string) []int {
	if ids, found := b.cache.Get(chatID + "_shownFeedbackIDs"); found {
		return ids.([]int)
	}
	return []int{}
}
