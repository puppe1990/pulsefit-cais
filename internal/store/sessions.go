package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/puppe1990/pulsefit/internal/models"
)

func (s *SQLiteStore) ListSessionsByUser(userID int64) ([]models.WorkoutSession, error) {
	rows, err := s.db.Query(`
		SELECT ws.id, ws.user_id, ws.routine_id, ws.routine_name, ws.started_at, ws.completed_at, ws.duration_seconds,
		       (SELECT COUNT(*) FROM exercise_logs el WHERE el.session_id = ws.id) AS exercise_count
		FROM workout_sessions ws
		WHERE ws.user_id = ?
		ORDER BY ws.started_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var items []models.WorkoutSession
	for rows.Next() {
		var sess models.WorkoutSession
		var routineID sql.NullInt64
		var completedAt sql.NullTime
		if err := rows.Scan(&sess.ID, &sess.UserID, &routineID, &sess.RoutineName, &sess.StartedAt, &completedAt, &sess.DurationSeconds, &sess.ExerciseCount); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		if routineID.Valid {
			sess.RoutineID = routineID.Int64
		}
		if completedAt.Valid {
			t := completedAt.Time
			sess.CompletedAt = &t
		}
		items = append(items, sess)
	}
	return items, rows.Err()
}

func (s *SQLiteStore) insertSession(userID, routineID int64, routineName string, startedAt time.Time, durationSeconds int) (int64, error) {
	result, err := s.db.Exec(
		`INSERT INTO workout_sessions (user_id, routine_id, routine_name, started_at, completed_at, duration_seconds)
		 VALUES (?, ?, ?, ?, datetime(?, '+' || ? || ' seconds'), ?)`,
		userID, routineID, routineName, startedAt, startedAt.Format("2006-01-02 15:04:05"), durationSeconds, durationSeconds,
	)
	if err != nil {
		return 0, fmt.Errorf("insert session: %w", err)
	}
	return result.LastInsertId()
}

func (s *SQLiteStore) insertExerciseLog(sessionID, exerciseID int64, name string, position int) error {
	_, err := s.db.Exec(
		"INSERT INTO exercise_logs (session_id, exercise_id, exercise_name, position) VALUES (?, ?, ?, ?)",
		sessionID, exerciseID, name, position,
	)
	if err != nil {
		return fmt.Errorf("insert exercise log: %w", err)
	}
	return nil
}