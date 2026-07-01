package models

type WorkoutSummary struct {
	Session       WorkoutSession
	TotalSets     int
	CompletedSets int
}
