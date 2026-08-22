package workoutplan

import "time"

type WorkoutPlan struct {
	ID          string
	UserID      string
	Title       string
	Description string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
type WorkoutPlanItem struct {
	ID            string
	WorkoutPlanID string
	ExerciseID    string
	OrderIndex    int
	Sets          int
	Reps          int
	Weight        float64
	Unit          string
	CreatedAt     time.Time
}
