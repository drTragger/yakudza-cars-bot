package models

type CarOption struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	Year        string `json:"year"`
	PhotoID     string `json:"photoId"`
	CreatedAt   string `json:"createdAt"`
}
