package models

import "time"

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	DisplayName  string
	PhotoURL     string
	CreatedAt    time.Time
}
