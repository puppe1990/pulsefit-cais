package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/puppe1990/pulsefit/internal/models"
)

func routineCard(r models.Routine) RoutineCard {
	return RoutineCard{
		ID:    strconv.FormatInt(r.ID, 10),
		Name:  r.Name,
		Emoji: r.Emoji,
		Color: r.Color,
	}
}

func exerciseCard(ex models.Exercise) ExerciseCard {
	return ExerciseCard{
		ID:           strconv.FormatInt(ex.ID, 10),
		Name:         ex.Name,
		MuscleGroup:  ex.MuscleGroup,
		Equipment:    ex.Equipment,
		Instructions: ex.Instructions,
		ImageURL:     ex.ImageURL,
	}
}

func historySession(sess models.WorkoutSession) HistorySession {
	mins := sess.DurationSeconds / 60
	return HistorySession{
		ID:          strconv.FormatInt(sess.ID, 10),
		RoutineName: sess.RoutineName,
		Date:        sess.StartedAt.Format("Jan 2, 2006"),
		Duration:    fmt.Sprintf("%dm", mins),
		ExerciseCnt: sess.ExerciseCount,
	}
}

func firstName(displayName string) string {
	parts := strings.Fields(displayName)
	if len(parts) == 0 {
		return "Athlete"
	}
	return parts[0]
}

var muscleGroupCategories = map[string]SearchCategory{
	"Chest":     {Name: "Chest", Color: "bg-red-500", Image: "https://images.unsplash.com/photo-1571019614242-c5c5dee9f50b?w=200&q=80"},
	"Back":      {Name: "Back", Color: "bg-green-700", Image: "https://images.unsplash.com/photo-1603287611837-f212781b0a8c?w=200&q=80"},
	"Legs":      {Name: "Legs", Color: "bg-blue-600", Image: "https://images.unsplash.com/photo-1434682772747-f16d3ea162c3?w=200&q=80"},
	"Shoulders": {Name: "Shoulders", Color: "bg-purple-600", Image: "https://images.unsplash.com/photo-1541534741688-6078c65b5a33?w=200&q=80"},
	"Arms":      {Name: "Arms", Color: "bg-orange-500", Image: "https://images.unsplash.com/photo-1581009146145-b5ef050c2e1e?w=200&q=80"},
	"Core":      {Name: "Core", Color: "bg-yellow-600", Image: "https://images.unsplash.com/photo-1571019613454-1cb2f99b2d8b?w=200&q=80"},
}

func categoriesFromExercises(exercises []models.Exercise) []SearchCategory {
	seen := make(map[string]bool)
	var cats []SearchCategory
	order := []string{"Chest", "Back", "Legs", "Shoulders", "Arms", "Core"}
	for _, ex := range exercises {
		seen[ex.MuscleGroup] = true
	}
	for _, name := range order {
		if seen[name] {
			if cat, ok := muscleGroupCategories[name]; ok {
				cats = append(cats, cat)
			}
		}
	}
	return cats
}

func formatDuration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	d := time.Duration(seconds) * time.Second
	return fmt.Sprintf("%dm", int(d.Minutes()))
}

func formatElapsed(seconds int) string {
	if seconds < 0 {
		seconds = 0
	}
	return fmt.Sprintf("%d:%02d", seconds/60, seconds%60)
}