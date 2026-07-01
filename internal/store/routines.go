package store

import (
	"fmt"

	"github.com/puppe1990/pulsefit/internal/models"
)

func (s *SQLiteStore) ListRoutinesByUser(userID int64) ([]models.Routine, error) {
	rows, err := s.db.Query(
		"SELECT id, user_id, name, description, cover_image_url, emoji, color, created_at FROM routines WHERE user_id = ? ORDER BY id",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list routines: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []models.Routine
	for rows.Next() {
		var r models.Routine
		if err := rows.Scan(&r.ID, &r.UserID, &r.Name, &r.Description, &r.CoverImageURL, &r.Emoji, &r.Color, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan routine: %w", err)
		}
		items = append(items, r)
	}
	return items, rows.Err()
}

func (s *SQLiteStore) FindRoutineByID(id int64) (models.Routine, error) {
	var r models.Routine
	err := s.db.QueryRow(
		"SELECT id, user_id, name, description, cover_image_url, emoji, color, created_at FROM routines WHERE id = ?",
		id,
	).Scan(&r.ID, &r.UserID, &r.Name, &r.Description, &r.CoverImageURL, &r.Emoji, &r.Color, &r.CreatedAt)
	if err != nil {
		return models.Routine{}, fmt.Errorf("find routine: %w", err)
	}
	return r, nil
}

func (s *SQLiteStore) ListRoutineExercises(routineID int64) ([]models.Exercise, error) {
	rows, err := s.db.Query(`
		SELECT e.id, e.name, e.muscle_group, e.equipment, e.instructions, e.image_url, e.video_url, e.created_at
		FROM exercises e
		INNER JOIN routine_exercises re ON re.exercise_id = e.id
		WHERE re.routine_id = ?
		ORDER BY re.position`,
		routineID,
	)
	if err != nil {
		return nil, fmt.Errorf("list routine exercises: %w", err)
	}
	defer func() { _ = rows.Close() }()
	return scanExercises(rows)
}

func (s *SQLiteStore) CreateRoutine(userID int64, routine models.Routine, exerciseIDs []int64) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.Exec(
		"INSERT INTO routines (user_id, name, description, cover_image_url, emoji, color) VALUES (?, ?, ?, ?, ?, ?)",
		userID, routine.Name, routine.Description, routine.CoverImageURL, routine.Emoji, routine.Color,
	)
	if err != nil {
		return 0, fmt.Errorf("insert routine: %w", err)
	}
	routineID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	for i, exID := range exerciseIDs {
		if _, err := tx.Exec(
			"INSERT INTO routine_exercises (routine_id, exercise_id, position) VALUES (?, ?, ?)",
			routineID, exID, i,
		); err != nil {
			return 0, fmt.Errorf("link exercise: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return routineID, nil
}