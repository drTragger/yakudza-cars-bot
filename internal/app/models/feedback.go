package models

type Feedback struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	VideoFileID string `json:"videoFileID"`
	CreatedAt   string `json:"createdAt"`
}
