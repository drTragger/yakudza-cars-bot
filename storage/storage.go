package storage

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

type Storage struct {
	config              *Config
	db                  *sql.DB
	userRepository      *UserRepository
	carOptionRepository *CarOptionRepository
	feedbackRepository  *FeedbackRepository
}

func New(config *Config) *Storage {
	return &Storage{
		config: config,
	}
}

func (storage *Storage) Open() error {
	dataSourceName := storage.config.User + ":" + storage.config.Password + "@/" + storage.config.DataBase
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}
	storage.db = db
	log.Println("Database connection has been set up successfully")
	return nil
}

func (storage *Storage) Close() {
	if err := storage.db.Close(); err != nil {
		log.Println("Error during closing DB connection: ", err)
	}
}

func (storage *Storage) User() *UserRepository {
	if storage.userRepository != nil {
		return storage.userRepository
	}
	storage.userRepository = &UserRepository{
		storage: storage,
	}
	return storage.userRepository
}

func (storage *Storage) CarOption() *CarOptionRepository {
	if storage.carOptionRepository != nil {
		return storage.carOptionRepository
	}
	storage.carOptionRepository = &CarOptionRepository{
		storage: storage,
	}
	return storage.carOptionRepository
}

func (storage *Storage) Feedback() *FeedbackRepository {
	if storage.feedbackRepository != nil {
		return storage.feedbackRepository
	}
	storage.feedbackRepository = &FeedbackRepository{
		storage: storage,
	}
	return storage.feedbackRepository
}
