package models

type ExerciseLog struct {
	ID           int64
	SessionID    int64
	ExerciseID   int64
	ExerciseName string
	Position     int
}

type SetLog struct {
	ID            int64
	ExerciseLogID int64
	Weight        float64
	Reps          int
	Completed     bool
	Position      int
}