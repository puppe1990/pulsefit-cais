package models

import "time"

type Exercise struct {
	ID           int64
	Name         string
	MuscleGroup  string
	Equipment    string
	Instructions string
	ImageURL     string
	VideoURL     string
	CreatedAt    time.Time
}
