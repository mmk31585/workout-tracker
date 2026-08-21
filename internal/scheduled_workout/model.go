package scheduledworkout

import "time"

type ScheduledWorkout struct {
	ID            string
	UserID        string
	WorkoutPlanID string
	ScheduledAt   time.Time
	Status        string
	Notes         string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
