package handlers

type Profile struct {
	DisplayName string
	PhotoURL    string
	Email       string
}

type RoutineCard struct {
	ID    string
	Name  string
	Emoji string
	Color string
}

type ExerciseCard struct {
	ID           string
	Name         string
	MuscleGroup  string
	Equipment    string
	Instructions string
	ImageURL     string
}

type HistorySession struct {
	ID          string
	RoutineName string
	Date        string
	Duration    string
	ExerciseCnt int
}

type SearchCategory struct {
	Name  string
	Color string
	Image string
}
