package exercise

import "time"

type Exercise struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	Category    *string   `db:"category"`
	MuscleGroup *string   `db:"muscle_group"`
	CreatedAt   time.Time `db:"created_at"`
}
