package models

import "time"

type Routine struct {
	ID            int64
	UserID        int64
	Name          string
	Description   string
	CoverImageURL string
	Emoji         string
	Color         string
	CreatedAt     time.Time
}