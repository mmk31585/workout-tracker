package scheduledworkout

import "time"

type ScheduleRequest struct {
	UserID        string         `json:"user_id" validate:"required,uuid"`
	WorkoutPlanID string         `json:"workout_plan_id" validate:"required,uuid"`
	ScheduledAt   time.Time      `json:"scheduled_at" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	Status        ScheduleStatus `json:"status" validate:"required,oneof=scheduled in_progress completed cancelled"`
	Notes         *string        `json:"notes,omitempty" validate:"omitempty,max=2000"`
}

type WorkoutSessionItemRequest struct {
	ExerciseID string  `json:"exercise_id" validate:"required,uuid"`
	Sets       int     `json:"sets" validate:"required,min=1,max=100"`
	Reps       int     `json:"reps" validate:"required,min=1,max=1000"`
	Weight     float64 `json:"weight" validate:"required,min=0"`
	Unit       string  `json:"unit" validate:"required,oneof=kg lb"`
	Notes      *string `json:"notes,omitempty" validate:"omitempty,max=1000"`
}

type CompleteRequest struct {
	PerformedAt  time.Time                   `json:"performed_at" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	OverallNotes *string                     `json:"overall_notes,omitempty" validate:"omitempty,max=2000"`
	Items        []WorkoutSessionItemRequest `json:"items" validate:"required,min=1,dive"`
}

type ScheduleResponse struct {
	ID            string         `json:"id"`
	WorkoutPlanID string         `json:"workout_plan_id"`
	ScheduledAt   time.Time      `json:"scheduled_at"`
	Status        ScheduleStatus `json:"status"`
	Notes         *string        `json:"notes,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type WorkoutSessionItemResponse struct {
	ID         string    `json:"id"`
	ExerciseID string    `json:"exercise_id"`
	Sets       int       `json:"sets"`
	Reps       int       `json:"reps"`
	Weight     float64   `json:"weight"`
	Unit       string    `json:"unit"`
	Notes      *string   `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type CompleteResponse struct {
	ID                 string                       `json:"id"`
	ScheduledWorkoutID string                       `json:"scheduled_workout_id"`
	PerformedAt        time.Time                    `json:"performed_at"`
	OverallNotes       *string                      `json:"overall_notes,omitempty"`
	Items              []WorkoutSessionItemResponse `json:"items"`
	CreatedAt          time.Time                    `json:"created_at"`
	UpdatedAt          time.Time                    `json:"updated_at"`
}
