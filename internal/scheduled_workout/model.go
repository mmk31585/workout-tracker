package scheduledworkout

import "time"

type ScheduledWorkout struct {
	ID            string         `db:"id"`
	UserID        string         `db:"user_id"`
	WorkoutPlanID string         `db:"workout_plan_id"`
	ScheduledAt   time.Time      `db:"scheduled_at"`
	Status        ScheduleStatus `db:"status"`
	Notes         *string        `db:"notes"`
	CreatedAt     time.Time      `db:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at"`
}
