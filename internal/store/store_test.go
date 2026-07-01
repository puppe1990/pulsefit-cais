package store

import (
	"testing"

	"github.com/puppe1990/cais/pkg/cais/session"
	"github.com/puppe1990/pulsefit/internal/models"
)

func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	s, err := NewSQLiteStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestStore_Migrations(t *testing.T) {
	_ = newTestStore(t)
}

func TestStore_CreateUser_FindUserByEmail(t *testing.T) {
	s := newTestStore(t)
	hash, err := session.HashPassword("secret")
	if err != nil {
		t.Fatal(err)
	}

	id, err := s.CreateUser(models.User{
		Email:        "athlete@example.com",
		PasswordHash: hash,
		DisplayName:  "Athlete",
		PhotoURL:     "https://example.com/photo.svg",
	})
	if err != nil {
		t.Fatal(err)
	}
	if id == 0 {
		t.Fatal("id = 0")
	}

	u, err := s.FindUserByEmail("athlete@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if u.ID != id || u.DisplayName != "Athlete" {
		t.Fatalf("user = %+v, want id %d", u, id)
	}
	if !session.VerifyPassword(u.PasswordHash, "secret") {
		t.Error("password hash should verify")
	}
}

func TestStore_ListExercises_empty(t *testing.T) {
	s := newTestStore(t)
	items, err := s.ListExercises()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("len = %d, want 0", len(items))
	}
}

func TestStore_SeedDemo(t *testing.T) {
	s := newTestStore(t)
	if err := s.SeedDemo(); err != nil {
		t.Fatal(err)
	}

	exercises, err := s.ListExercises()
	if err != nil {
		t.Fatal(err)
	}
	if len(exercises) != 6 {
		t.Fatalf("exercises = %d, want 6", len(exercises))
	}

	u, err := s.FindUserByEmail("demo@pulsefit.local")
	if err != nil {
		t.Fatal(err)
	}
	if u.DisplayName != "Demo Athlete" {
		t.Errorf("display name = %q", u.DisplayName)
	}
	if !session.VerifyPassword(u.PasswordHash, "demo") {
		t.Error("demo password should be 'demo'")
	}

	routines, err := s.ListRoutinesByUser(u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(routines) < 3 {
		t.Fatalf("routines = %d, want at least 3", len(routines))
	}

	// idempotent
	if err := s.SeedDemo(); err != nil {
		t.Fatal(err)
	}
	exercises2, _ := s.ListExercises()
	if len(exercises2) != 6 {
		t.Errorf("seed should be idempotent, got %d exercises", len(exercises2))
	}
}

func TestStore_CreateRoutine_ListRoutineExercises(t *testing.T) {
	s := newTestStore(t)
	if err := s.SeedDemo(); err != nil {
		t.Fatal(err)
	}
	u, err := s.FindUserByEmail("demo@pulsefit.local")
	if err != nil {
		t.Fatal(err)
	}
	exercises, err := s.ListExercises()
	if err != nil {
		t.Fatal(err)
	}

	routineID, err := s.CreateRoutine(u.ID, models.Routine{
		Name:  "Test Routine",
		Emoji: "💪",
		Color: "bg-blue-600",
	}, []int64{exercises[0].ID, exercises[1].ID})
	if err != nil {
		t.Fatal(err)
	}

	routine, err := s.FindRoutineByID(routineID)
	if err != nil {
		t.Fatal(err)
	}
	if routine.Name != "Test Routine" {
		t.Errorf("name = %q", routine.Name)
	}

	linked, err := s.ListRoutineExercises(routineID)
	if err != nil {
		t.Fatal(err)
	}
	if len(linked) != 2 {
		t.Fatalf("linked exercises = %d, want 2", len(linked))
	}
	if linked[0].ID != exercises[0].ID {
		t.Errorf("first exercise id = %d, want %d", linked[0].ID, exercises[0].ID)
	}
}

func TestStore_ListSessionsByUser(t *testing.T) {
	s := newTestStore(t)
	if err := s.SeedDemo(); err != nil {
		t.Fatal(err)
	}
	u, err := s.FindUserByEmail("demo@pulsefit.local")
	if err != nil {
		t.Fatal(err)
	}

	sessions, err := s.ListSessionsByUser(u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) < 2 {
		t.Fatalf("sessions = %d, want at least 2", len(sessions))
	}
	if sessions[0].RoutineName == "" {
		t.Error("routine name should be set")
	}
}