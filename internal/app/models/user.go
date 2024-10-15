package models

type User struct {
	ID        int    `json:"id"`
	Phone     string `json:"phone"`
	ChatId    string `json:"chatId"`
	CreatedAt string `json:"createdAt"`
}
