package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/drTragger/yakudza-cars-bot/internal/app/models"
	"strings"
)

type CarOptionRepository struct {
	storage *Storage
}

var (
	tableCarOptions = "car_options"
)

func (cr *CarOptionRepository) Create(co *models.CarOption) error {
	query := fmt.Sprintf("INSERT INTO %s (title, description, price, year, photo_id, created_at) VALUES (?, ?, ?, ?, ?, ?)", tableCarOptions)
	stmt, err := cr.storage.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(co.Title, co.Description, co.Price, co.Year, co.PhotoID, co.CreatedAt)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	co.ID = int(id)
	return nil
}

func (cr *CarOptionRepository) GetAll() ([]*models.CarOption, error) {
	// Створюємо SELECT запит для отримання всіх варіантів автомобілів
	query := fmt.Sprintf("SELECT * FROM %s ORDER BY created_at", tableCarOptions)
	rows, err := cr.storage.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Створюємо зріз для збереження результатів
	var carOptions []*models.CarOption

	// Проходимо по всіх рядках і додаємо їх до зрізу carOptions
	for rows.Next() {
		var co models.CarOption
		err := rows.Scan(&co.ID, &co.Title, &co.Description, &co.Price, &co.Year, &co.PhotoID, &co.CreatedAt)
		if err != nil {
			return nil, err
		}
		carOptions = append(carOptions, &co)
	}

	// Перевіряємо на помилки після закінчення роботи з рядками
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return carOptions, nil
}

func (cr *CarOptionRepository) GetByID(id int) (*models.CarOption, error) {
	// Створюємо запит для отримання варіанту автомобіля за ID
	query := fmt.Sprintf("SELECT * FROM %s WHERE id=?", tableCarOptions)

	// Підготовка запиту
	stmt, err := cr.storage.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// Створюємо змінну для збереження результату
	var co models.CarOption

	// Виконуємо запит і скануємо результат у структуру carOption
	err = stmt.QueryRow(id).Scan(&co.ID, &co.Title, &co.Description, &co.Price, &co.Year, &co.PhotoID, &co.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("car option with ID %d not found", id)
		}
		return nil, err
	}

	return &co, nil
}

func (cr *CarOptionRepository) Delete(id int) (*models.CarOption, error) {
	// Спочатку отримаємо запис, який будемо видаляти, щоб повернути його після видалення
	carOption, err := cr.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("could not find car option with ID %d: %v", id, err)
	}

	// Створюємо запит для видалення
	query := fmt.Sprintf("DELETE FROM %s WHERE id=?", tableCarOptions)
	stmt, err := cr.storage.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// Виконуємо запит
	res, err := stmt.Exec(id)
	if err != nil {
		return nil, err
	}

	// Перевіряємо, скільки рядків було видалено
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("car option with ID %d not found", id)
	}

	// Повертаємо видалений запис
	return carOption, nil
}

func (cr *CarOptionRepository) GetByDetails(carDetails *models.CarDetails, shownIDs []int) (*models.CarOption, error) {
	// Якщо немає показаних ID, створюємо простий запит без виключення shownIDs
	query := fmt.Sprintf(
		"SELECT * FROM %s WHERE (price BETWEEN ? AND ?) AND year >= ?",
		tableCarOptions,
	)

	// Якщо показані ID існують, додаємо умову `NOT IN`
	if len(shownIDs) > 0 {
		// Створюємо місце для знаків запитання для кожного ID
		placeholders := make([]string, len(shownIDs))
		for i := range placeholders {
			placeholders[i] = "?"
		}
		// Додаємо умову `NOT IN` у запит
		query += fmt.Sprintf(" AND id NOT IN (%s)", strings.Join(placeholders, ","))
	}

	// Додаємо обмеження результатів
	query += " ORDER BY RAND() LIMIT 1"

	// Підготовка запиту
	stmt, err := cr.storage.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// Створюємо змінну для збереження результату
	var co models.CarOption

	// Створюємо масив аргументів для виконання запиту
	args := []interface{}{carDetails.Price.Min, carDetails.Price.Max, carDetails.Year}
	for _, id := range shownIDs {
		args = append(args, id)
	}

	// Виконуємо запит і скануємо результат у структуру carOption
	err = stmt.QueryRow(args...).Scan(&co.ID, &co.Title, &co.Description, &co.Price, &co.Year, &co.PhotoID, &co.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &co, nil
}
