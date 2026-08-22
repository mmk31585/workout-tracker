package scheduledworkout

import "time"

type ScheduledWorkout struct {
	ID            string
	UserID        string
	WorkoutPlanID string
	ScheduledAt   time.Time
	Status        ScheduleStatus
	Notes         *string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
