package storage

import (
	"fmt"
	"github.com/drTragger/yakudza-cars-bot/internal/app/models"
)

type UserRepository struct {
	storage *Storage
}

var (
	tableUsers = "users"
)

func (ur *UserRepository) Create(u *models.User) error {
	query := fmt.Sprintf("INSERT INTO %s (chat_id, phone, created_at) VALUES (?, ?, ?)", tableUsers)
	stmt, err := ur.storage.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(u.ChatId, u.Phone, u.CreatedAt)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	u.ID = int(id)
	return nil
}

func (ur *UserRepository) UserExists(chatId string) (bool, error) {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE chat_id=?)", tableUsers)
	row := ur.storage.db.QueryRow(query, chatId)
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func (ur *UserRepository) FindById(id int) (*models.User, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id=?", tableUsers)
	user := models.User{}
	row := ur.storage.db.QueryRow(query, id)
	if err := row.Scan(&user.ID, &user.ChatId, &user.Phone); err != nil {
		return nil, err
	}
	return &user, nil
}

func (ur *UserRepository) FindByChatId(chatId int) (*models.User, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE chat_id=?", tableUsers)
	user := models.User{}
	row := ur.storage.db.QueryRow(query, chatId)
	if err := row.Scan(&user.ID, &user.ChatId, &user.Phone, &user.CreatedAt); err != nil {
		return nil, err
	}
	return &user, nil
}
