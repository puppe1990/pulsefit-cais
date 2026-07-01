package store

import (
	"fmt"

	"github.com/puppe1990/pulsefit/internal/models"
)

func (s *SQLiteStore) ListExercises() ([]models.Exercise, error) {
	rows, err := s.db.Query(
		"SELECT id, name, muscle_group, equipment, instructions, image_url, video_url, created_at FROM exercises ORDER BY id",
	)
	if err != nil {
		return nil, fmt.Errorf("list exercises: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanExercises(rows)
}

func (s *SQLiteStore) FindExerciseByID(id int64) (models.Exercise, error) {
	row := s.db.QueryRow(
		"SELECT id, name, muscle_group, equipment, instructions, image_url, video_url, created_at FROM exercises WHERE id = ?",
		id,
	)
	var ex models.Exercise
	err := row.Scan(&ex.ID, &ex.Name, &ex.MuscleGroup, &ex.Equipment, &ex.Instructions, &ex.ImageURL, &ex.VideoURL, &ex.CreatedAt)
	if err != nil {
		return models.Exercise{}, fmt.Errorf("find exercise: %w", err)
	}
	return ex, nil
}

func (s *SQLiteStore) insertExercise(ex models.Exercise) (int64, error) {
	result, err := s.db.Exec(
		"INSERT INTO exercises (name, muscle_group, equipment, instructions, image_url, video_url) VALUES (?, ?, ?, ?, ?, ?)",
		ex.Name, ex.MuscleGroup, ex.Equipment, ex.Instructions, ex.ImageURL, ex.VideoURL,
	)
	if err != nil {
		return 0, fmt.Errorf("insert exercise: %w", err)
	}
	return result.LastInsertId()
}

func scanExercises(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]models.Exercise, error) {
	var items []models.Exercise
	for rows.Next() {
		var ex models.Exercise
		if err := rows.Scan(&ex.ID, &ex.Name, &ex.MuscleGroup, &ex.Equipment, &ex.Instructions, &ex.ImageURL, &ex.VideoURL, &ex.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan exercise: %w", err)
		}
		items = append(items, ex)
	}
	return items, rows.Err()
}
