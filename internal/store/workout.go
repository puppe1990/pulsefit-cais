package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/puppe1990/pulsefit/internal/models"
)

func (s *SQLiteStore) StartWorkout(userID, routineID int64) (int64, error) {
	routine, err := s.FindRoutineByID(routineID)
	if err != nil {
		return 0, fmt.Errorf("find routine: %w", err)
	}
	if routine.UserID != userID {
		return 0, fmt.Errorf("routine not owned by user")
	}

	exercises, err := s.ListRoutineExercises(routineID)
	if err != nil {
		return 0, err
	}
	if len(exercises) == 0 {
		return 0, fmt.Errorf("routine has no exercises")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.Exec(
		`INSERT INTO workout_sessions (user_id, routine_id, routine_name, started_at, duration_seconds)
		 VALUES (?, ?, ?, CURRENT_TIMESTAMP, 0)`,
		userID, routineID, routine.Name,
	)
	if err != nil {
		return 0, fmt.Errorf("insert session: %w", err)
	}
	sessionID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	for i, ex := range exercises {
		if _, err := tx.Exec(
			"INSERT INTO exercise_logs (session_id, exercise_id, exercise_name, position) VALUES (?, ?, ?, ?)",
			sessionID, ex.ID, ex.Name, i,
		); err != nil {
			return 0, fmt.Errorf("insert exercise log: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return sessionID, nil
}

func (s *SQLiteStore) FindWorkoutSession(id int64) (models.WorkoutSession, error) {
	var sess models.WorkoutSession
	var routineID sql.NullInt64
	var completedAt sql.NullTime
	err := s.db.QueryRow(`
		SELECT id, user_id, routine_id, routine_name, started_at, completed_at, duration_seconds
		FROM workout_sessions WHERE id = ?`, id,
	).Scan(&sess.ID, &sess.UserID, &routineID, &sess.RoutineName, &sess.StartedAt, &completedAt, &sess.DurationSeconds)
	if err != nil {
		return models.WorkoutSession{}, fmt.Errorf("find session: %w", err)
	}
	if routineID.Valid {
		sess.RoutineID = routineID.Int64
	}
	if completedAt.Valid {
		t := completedAt.Time
		sess.CompletedAt = &t
	}
	return sess, nil
}

func (s *SQLiteStore) ListExerciseLogs(sessionID int64) ([]models.ExerciseLog, error) {
	rows, err := s.db.Query(
		"SELECT id, session_id, exercise_id, exercise_name, position FROM exercise_logs WHERE session_id = ? ORDER BY position",
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("list exercise logs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []models.ExerciseLog
	for rows.Next() {
		var log models.ExerciseLog
		var exerciseID sql.NullInt64
		if err := rows.Scan(&log.ID, &log.SessionID, &exerciseID, &log.ExerciseName, &log.Position); err != nil {
			return nil, fmt.Errorf("scan exercise log: %w", err)
		}
		if exerciseID.Valid {
			log.ExerciseID = exerciseID.Int64
		}
		items = append(items, log)
	}
	return items, rows.Err()
}

func (s *SQLiteStore) AddSet(exerciseLogID int64, weight float64, reps int, completed bool) (int64, error) {
	var position int
	err := s.db.QueryRow("SELECT COALESCE(MAX(position), -1) + 1 FROM set_logs WHERE exercise_log_id = ?", exerciseLogID).Scan(&position)
	if err != nil {
		return 0, fmt.Errorf("next position: %w", err)
	}

	result, err := s.db.Exec(
		"INSERT INTO set_logs (exercise_log_id, weight, reps, completed, position) VALUES (?, ?, ?, ?, ?)",
		exerciseLogID, weight, reps, boolInt(completed), position,
	)
	if err != nil {
		return 0, fmt.Errorf("insert set: %w", err)
	}
	return result.LastInsertId()
}

func (s *SQLiteStore) ListSets(exerciseLogID int64) ([]models.SetLog, error) {
	rows, err := s.db.Query(
		"SELECT id, exercise_log_id, weight, reps, completed, position FROM set_logs WHERE exercise_log_id = ? ORDER BY position",
		exerciseLogID,
	)
	if err != nil {
		return nil, fmt.Errorf("list sets: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []models.SetLog
	for rows.Next() {
		var set models.SetLog
		var completedInt int
		if err := rows.Scan(&set.ID, &set.ExerciseLogID, &set.Weight, &set.Reps, &completedInt, &set.Position); err != nil {
			return nil, fmt.Errorf("scan set: %w", err)
		}
		set.Completed = completedInt == 1
		items = append(items, set)
	}
	return items, rows.Err()
}

func (s *SQLiteStore) FinishWorkout(sessionID int64) error {
	sess, err := s.FindWorkoutSession(sessionID)
	if err != nil {
		return err
	}
	if sess.CompletedAt != nil {
		return fmt.Errorf("session already finished")
	}
	duration := int(time.Since(sess.StartedAt).Seconds())
	if duration < 1 {
		duration = 1
	}
	_, err = s.db.Exec(
		"UPDATE workout_sessions SET completed_at = CURRENT_TIMESTAMP, duration_seconds = ? WHERE id = ?",
		duration, sessionID,
	)
	if err != nil {
		return fmt.Errorf("finish session: %w", err)
	}
	return nil
}

func (s *SQLiteStore) WorkoutSummary(sessionID int64) (models.WorkoutSummary, error) {
	sess, err := s.FindWorkoutSession(sessionID)
	if err != nil {
		return models.WorkoutSummary{}, err
	}
	var total, completed int
	err = s.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(CASE WHEN completed = 1 THEN 1 ELSE 0 END), 0)
		FROM set_logs sl
		INNER JOIN exercise_logs el ON el.id = sl.exercise_log_id
		WHERE el.session_id = ?`, sessionID,
	).Scan(&total, &completed)
	if err != nil {
		return models.WorkoutSummary{}, fmt.Errorf("count sets: %w", err)
	}
	return models.WorkoutSummary{Session: sess, TotalSets: total, CompletedSets: completed}, nil
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}