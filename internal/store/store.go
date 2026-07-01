package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/puppe1990/cais/pkg/cais/devlog"
	"github.com/puppe1990/cais/pkg/cais/sqllog"
	_ "modernc.org/sqlite"

	"github.com/puppe1990/pulsefit/internal/models"
)

type Store interface {
	CreateUser(models.User) (int64, error)
	FindUserByEmail(email string) (models.User, error)
	FindUserByID(id int64) (models.User, error)
	ListExercises() ([]models.Exercise, error)
	FindExerciseByID(id int64) (models.Exercise, error)
	ListRoutinesByUser(userID int64) ([]models.Routine, error)
	FindRoutineByID(id int64) (models.Routine, error)
	ListRoutineExercises(routineID int64) ([]models.Exercise, error)
	CreateRoutine(userID int64, routine models.Routine, exerciseIDs []int64) (int64, error)
	ListSessionsByUser(userID int64) ([]models.WorkoutSession, error)
	StartWorkout(userID, routineID int64) (int64, error)
	FindWorkoutSession(id int64) (models.WorkoutSession, error)
	ListExerciseLogs(sessionID int64) ([]models.ExerciseLog, error)
	AddSet(exerciseLogID int64, weight float64, reps int, completed bool) (int64, error)
	ListSets(exerciseLogID int64) ([]models.SetLog, error)
	FinishWorkout(sessionID int64) error
	WorkoutSummary(sessionID int64) (models.WorkoutSummary, error)
	SeedDemo() error
	Close() error
}

type SQLiteStore struct {
	db *sqllog.DB
}

func NewSQLiteStore(dsn string, env string) (*SQLiteStore, error) {
	if dsn != ":memory:" {
		dir := filepath.Dir(dsn)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := applyMigrations(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	cfg := sqllog.Config{Enabled: sqllog.EnabledForEnv(env)}
	if cfg.Enabled {
		cfg.Writer = devlog.MirrorDefault(os.Stdout)
	}
	return &SQLiteStore{db: sqllog.Wrap(db, cfg)}, nil
}

func (s *SQLiteStore) DB() *sql.DB {
	return s.db.Raw()
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
