package exercise

import "time"

type Exercise struct {
	ID          string
	Name        string
	Description string
	Category    string
	MuscleGroup string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
