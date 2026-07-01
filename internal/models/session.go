package models

import "time"

type WorkoutSession struct {
	ID              int64
	UserID          int64
	RoutineID       int64
	RoutineName     string
	StartedAt       time.Time
	CompletedAt     *time.Time
	DurationSeconds int
	ExerciseCount   int
}
