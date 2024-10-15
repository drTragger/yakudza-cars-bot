package storage

import (
	"fmt"
	"github.com/drTragger/yakudza-cars-bot/internal/app/models"
	"strings"
)

type FeedbackRepository struct {
	storage *Storage
}

var (
	tableFeedback = "feedback"
)

func (fr *FeedbackRepository) Create(f *models.Feedback) error {
	query := fmt.Sprintf("INSERT INTO %s (description, video_file_id, created_at) VALUES (?, ?, ?)", tableFeedback)
	stmt, err := fr.storage.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(f.Description, f.VideoFileID, f.CreatedAt)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	f.ID = int(id)
	return nil
}

func (fr *FeedbackRepository) GetAll() ([]*models.Feedback, error) {
	// Створюємо SELECT запит для отримання всіх відгуків
	query := fmt.Sprintf("SELECT * FROM %s", tableFeedback)
	rows, err := fr.storage.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Створюємо зріз для збереження відгуків
	var feedback []*models.Feedback

	// Проходимо по всіх рядках і додаємо їх до зрізу feedback
	for rows.Next() {
		var fb models.Feedback
		err := rows.Scan(&fb.ID, &fb.Description, &fb.VideoFileID, &fb.CreatedAt)
		if err != nil {
			return nil, err
		}
		feedback = append(feedback, &fb)
	}

	// Перевіряємо на помилки після закінчення роботи з рядками
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return feedback, nil
}

func (fr *FeedbackRepository) Delete(feedbackID int) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableFeedback)
	stmt, err := fr.storage.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(feedbackID)
	if err != nil {
		return err
	}

	return nil
}

func (fr *FeedbackRepository) GetNext(shownIDs []int) (*models.Feedback, error) {
	// Base query to select feedback
	query := fmt.Sprintf("SELECT * FROM %s", tableFeedback)

	// If shownIDs is not empty, add a condition to exclude those IDs
	if len(shownIDs) > 0 {
		// Create placeholders for each shownID (e.g., "?, ?, ?, ...")
		placeholders := make([]string, len(shownIDs))
		for i := range placeholders {
			placeholders[i] = "?"
		}
		// Add the NOT IN clause to the query
		query += fmt.Sprintf(" WHERE id NOT IN (%s)", strings.Join(placeholders, ","))
	}

	// Limit the results to return only one feedback
	query += " ORDER BY RAND() LIMIT 1"

	// Prepare the query
	stmt, err := fr.storage.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// Create argument slice for executing the query
	var args []interface{}
	for _, id := range shownIDs {
		args = append(args, id)
	}

	// Execute the query and fetch the result
	row := stmt.QueryRow(args...)

	// Create a variable to hold the feedback result
	var fb models.Feedback

	// Scan the result into the feedback struct
	err = row.Scan(&fb.ID, &fb.Description, &fb.VideoFileID, &fb.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Return the feedback result
	return &fb, nil
}
