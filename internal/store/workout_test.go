package store

import (
	"testing"
)

func TestStore_StartWorkout_createsSessionAndLogs(t *testing.T) {
	s := newTestStore(t)
	if err := s.SeedDemo(); err != nil {
		t.Fatal(err)
	}
	u, err := s.FindUserByEmail("demo@pulsefit.local")
	if err != nil {
		t.Fatal(err)
	}
	routines, err := s.ListRoutinesByUser(u.ID)
	if err != nil || len(routines) == 0 {
		t.Fatal("expected routines")
	}

	sessionID, err := s.StartWorkout(u.ID, routines[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if sessionID == 0 {
		t.Fatal("session id = 0")
	}

	sess, err := s.FindWorkoutSession(sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if sess.CompletedAt != nil {
		t.Error("session should be in progress")
	}
	if sess.RoutineName != routines[0].Name {
		t.Errorf("routine name = %q", sess.RoutineName)
	}

	logs, err := s.ListExerciseLogs(sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) == 0 {
		t.Fatal("expected exercise logs")
	}
}

func TestStore_AddSet_andFinishWorkout(t *testing.T) {
	s := newTestStore(t)
	if err := s.SeedDemo(); err != nil {
		t.Fatal(err)
	}
	u, _ := s.FindUserByEmail("demo@pulsefit.local")
	routines, _ := s.ListRoutinesByUser(u.ID)
	sessionID, err := s.StartWorkout(u.ID, routines[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	logs, err := s.ListExerciseLogs(sessionID)
	if err != nil {
		t.Fatal(err)
	}

	setID, err := s.AddSet(logs[0].ID, 60, 10, true)
	if err != nil {
		t.Fatal(err)
	}
	if setID == 0 {
		t.Fatal("set id = 0")
	}

	sets, err := s.ListSets(logs[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(sets) != 1 || sets[0].Reps != 10 {
		t.Fatalf("sets = %+v", sets)
	}

	if err := s.FinishWorkout(sessionID); err != nil {
		t.Fatal(err)
	}
	sess, err := s.FindWorkoutSession(sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if sess.CompletedAt == nil {
		t.Error("expected completed_at")
	}
	if sess.DurationSeconds <= 0 {
		t.Error("expected positive duration")
	}

	summary, err := s.WorkoutSummary(sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if summary.TotalSets != 1 {
		t.Errorf("total sets = %d, want 1", summary.TotalSets)
	}
}