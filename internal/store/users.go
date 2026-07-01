package store

import (
	"database/sql"
	"fmt"

	"github.com/puppe1990/pulsefit/internal/models"
)

func (s *SQLiteStore) CreateUser(u models.User) (int64, error) {
	result, err := s.db.Exec(
		"INSERT INTO users (email, password_hash, display_name, photo_url) VALUES (?, ?, ?, ?)",
		u.Email, u.PasswordHash, u.DisplayName, u.PhotoURL,
	)
	if err != nil {
		return 0, fmt.Errorf("insert user: %w", err)
	}
	return result.LastInsertId()
}

func (s *SQLiteStore) FindUserByEmail(email string) (models.User, error) {
	return s.scanUser(s.db.QueryRow(
		"SELECT id, email, password_hash, display_name, photo_url, created_at FROM users WHERE email = ?",
		email,
	))
}

func (s *SQLiteStore) FindUserByID(id int64) (models.User, error) {
	return s.scanUser(s.db.QueryRow(
		"SELECT id, email, password_hash, display_name, photo_url, created_at FROM users WHERE id = ?",
		id,
	))
}

func (s *SQLiteStore) scanUser(row *sql.Row) (models.User, error) {
	var u models.User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.DisplayName, &u.PhotoURL, &u.CreatedAt)
	if err != nil {
		return models.User{}, fmt.Errorf("scan user: %w", err)
	}
	return u, nil
}