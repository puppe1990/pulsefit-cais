package store

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/puppe1990/cais/pkg/cais/session"
)

type SessionStore struct {
	s *SQLiteStore
}

func NewSessionStore(s *SQLiteStore) session.Store {
	return &SessionStore{s: s}
}

func (ss *SessionStore) Create(userID int64) (string, error) {
	token, err := newSessionToken()
	if err != nil {
		return "", err
	}
	if _, err := ss.s.db.Exec("INSERT INTO sessions (token, user_id) VALUES (?, ?)", token, userID); err != nil {
		return "", fmt.Errorf("insert session: %w", err)
	}
	return token, nil
}

func (ss *SessionStore) Get(token string) (int64, bool) {
	var userID int64
	err := ss.s.db.QueryRow("SELECT user_id FROM sessions WHERE token = ?", token).Scan(&userID)
	if err != nil {
		return 0, false
	}
	return userID, true
}

func (ss *SessionStore) Delete(token string) {
	_, _ = ss.s.db.Exec("DELETE FROM sessions WHERE token = ?", token)
}

func newSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}