package workoutsession

import "time"

type WorkoutSession struct {
	ID                 string
	UserID             string
	ScheduledWorkoutID string
	PerformedAt        time.Time
	OverallNotes       string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type WorkoutSessionItem struct {
	ID               string
	WorkoutSessionID string
	ExerciseID       string
	Sets             int
	Reps             int
	Weight           float64
	Unit             string
	Notes            string
	CreatedAt        time.Time
}
