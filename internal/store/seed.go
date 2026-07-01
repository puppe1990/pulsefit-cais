package store

import (
	"fmt"
	"time"

	"github.com/puppe1990/cais/pkg/cais/session"

	"github.com/puppe1990/pulsefit/internal/models"
)

const demoEmail = "demo@pulsefit.local"

var defaultExercises = []models.Exercise{
	{Name: "Barbell Bench Press", MuscleGroup: "Chest", Equipment: "Barbell", Instructions: "Lie on a flat bench. Grip the barbell with hands slightly wider than shoulder-width. Lower the bar to your mid-chest, then press it back up.", ImageURL: "https://images.unsplash.com/photo-1534438327276-14e5300c3a48?w=400&q=80"},
	{Name: "Bent Over Rows", MuscleGroup: "Back", Equipment: "Barbell", Instructions: "Bend at your hips and knees, keeping your back flat. Pull the bar toward your lower chest.", ImageURL: "https://images.unsplash.com/photo-1541534741688-6078c65b5a33?w=400&q=80"},
	{Name: "Overhead Press", MuscleGroup: "Shoulders", Equipment: "Barbell", Instructions: "Stand tall with feet shoulder-width apart. Press the bar straight up until arms are fully extended.", ImageURL: "https://images.unsplash.com/photo-1590487988256-9ed24133863e?w=400&q=80"},
	{Name: "Barbell Squat", MuscleGroup: "Legs", Equipment: "Barbell", Instructions: "Place bar on upper back. Squat down by sitting your hips back.", ImageURL: "https://images.unsplash.com/photo-1534368420009-621bf3424584?w=400&q=80"},
	{Name: "Deadlift", MuscleGroup: "Back", Equipment: "Barbell", Instructions: "Stand with feet mid-foot under the bar. Lift the bar by standing up.", ImageURL: "https://images.unsplash.com/photo-1517836357463-d25dfeac3438?w=400&q=80"},
	{Name: "Dumbbell Bicep Curls", MuscleGroup: "Arms", Equipment: "Dumbbell", Instructions: "Curl the weights toward your shoulders, keeping elbows tucked in.", ImageURL: "https://images.unsplash.com/photo-1581009146145-b5ef050c2e1e?w=400&q=80"},
}

var demoRoutines = []struct {
	model        models.Routine
	exerciseIdxs []int
}{
	{models.Routine{Name: "Upper Body Power", Emoji: "💪", Color: "bg-blue-600"}, []int{0, 1, 2}},
	{models.Routine{Name: "Leg Day (Heavy)", Emoji: "🦵", Color: "bg-red-600"}, []int{3, 4}},
	{models.Routine{Name: "Pull Day", Emoji: "🏗️", Color: "bg-green-600"}, []int{1, 4, 5}},
}

func (s *SQLiteStore) SeedDemo() error {
	if err := s.seedExercises(); err != nil {
		return err
	}
	userID, err := s.seedDemoUser()
	if err != nil {
		return err
	}
	if err := s.seedDemoRoutines(userID); err != nil {
		return err
	}
	return s.seedDemoSessions(userID)
}

func (s *SQLiteStore) seedExercises() error {
	count, err := s.countExercises()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	for _, ex := range defaultExercises {
		if _, err := s.insertExercise(ex); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteStore) seedDemoUser() (int64, error) {
	u, err := s.FindUserByEmail(demoEmail)
	if err == nil {
		return u.ID, nil
	}
	hash, err := session.HashPassword("demo")
	if err != nil {
		return 0, err
	}
	return s.CreateUser(models.User{
		Email:        demoEmail,
		PasswordHash: hash,
		DisplayName:  "Demo Athlete",
		PhotoURL:     "https://api.dicebear.com/7.x/avataaars/svg?seed=demo",
	})
}

func (s *SQLiteStore) seedDemoRoutines(userID int64) error {
	count, err := s.countRoutinesByUser(userID)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	exercises, err := s.ListExercises()
	if err != nil {
		return err
	}
	if len(exercises) < 6 {
		return fmt.Errorf("expected 6 exercises, got %d", len(exercises))
	}

	for _, dr := range demoRoutines {
		ids := make([]int64, len(dr.exerciseIdxs))
		for i, idx := range dr.exerciseIdxs {
			ids[i] = exercises[idx].ID
		}
		if _, err := s.CreateRoutine(userID, dr.model, ids); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteStore) seedDemoSessions(userID int64) error {
	count, err := s.countSessionsByUser(userID)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	routines, err := s.ListRoutinesByUser(userID)
	if err != nil {
		return err
	}
	if len(routines) < 2 {
		return fmt.Errorf("expected demo routines")
	}

	sessions := []struct {
		routineIdx int
		daysAgo    int
		duration   int
	}{
		{0, 3, 42 * 60},
		{1, 6, 55 * 60},
	}

	for _, seed := range sessions {
		r := routines[seed.routineIdx]
		started := time.Now().AddDate(0, 0, -seed.daysAgo)
		sessionID, err := s.insertSession(userID, r.ID, r.Name, started, seed.duration)
		if err != nil {
			return err
		}
		exercises, err := s.ListRoutineExercises(r.ID)
		if err != nil {
			return err
		}
		for i, ex := range exercises {
			if err := s.insertExerciseLog(sessionID, ex.ID, ex.Name, i); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *SQLiteStore) countExercises() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM exercises").Scan(&count)
	return count, err
}

func (s *SQLiteStore) countRoutinesByUser(userID int64) (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM routines WHERE user_id = ?", userID).Scan(&count)
	return count, err
}

func (s *SQLiteStore) countSessionsByUser(userID int64) (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM workout_sessions WHERE user_id = ?", userID).Scan(&count)
	return count, err
}
