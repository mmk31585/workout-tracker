package workoutsession

import "time"

type WorkoutSession struct {
	ID                 string     `db:"id"`
	UserID             string     `db:"user_id"`
	ScheduledWorkoutID string     `db:"scheduled_workout_id"`
	PerformedAt        *time.Time `db:"performed_at"`
	OverallNotes       *string    `db:"overall_notes"`
	CreatedAt          time.Time  `db:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at"`
}

type WorkoutSessionItem struct {
	ID               string    `db:"id"`
	WorkoutSessionID string    `db:"workout_session_id"`
	ExerciseID       string    `db:"exercise_id"`
	Sets             int       `db:"sets"`
	Reps             int       `db:"reps"`
	Weight           float64   `db:"weight"`
	Unit             string    `db:"unit"`
	Notes            *string   `db:"notes"`
	CreatedAt        time.Time `db:"created_at"`
}
