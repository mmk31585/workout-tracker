package workoutplan

import "time"

type CreateWorkoutPlanRequest struct {
	Title       string                    `json:"title" validate:"required,min=1,max=150"`
	Description string                    `json:"description" validate:"max=2000"`
	Items       []CreateWorkoutPlanItemRequest `json:"items" validate:"dive,required"`
}

type CreateWorkoutPlanItemRequest struct {
	ExerciseID string  `json:"exercise_id" validate:"required,uuid"`
	OrderIndex int     `json:"order_index" validate:"min=0"`
	Sets       int     `json:"sets" validate:"required,min=1,max=100"`
	Reps       int     `json:"reps" validate:"required,min=1,max=1000"`
	Weight     float64 `json:"weight" validate:"required,min=0"`
	Unit       string  `json:"unit" validate:"required,oneof=kg lb"`
}

type UpdateWorkoutPlanRequest struct {
	Title       *string                          `json:"title,omitempty" validate:"omitempty,min=1,max=150"`
	Description *string                          `json:"description,omitempty" validate:"omitempty,max=2000"`
	Status      *string                          `json:"status,omitempty" validate:"omitempty,oneof=active archived"`
	Items       []UpdateWorkoutPlanItemRequest  `json:"items,omitempty" validate:"omitempty,dive"`
}

type UpdateWorkoutPlanItemRequest struct {
	ID         string   `json:"id" validate:"required,uuid"`
	ExerciseID string   `json:"exercise_id" validate:"required,uuid"`
	OrderIndex int      `json:"order_index" validate:"min=0"`
	Sets       int      `json:"sets" validate:"min=1,max=100"`
	Reps       int      `json:"reps" validate:"min=1,max=1000"`
	Weight     float64  `json:"weight" validate:"min=0"`
	Unit       string   `json:"unit" validate:"oneof=kg lb"`
}

type WorkoutPlanResponse struct {
	ID          string                   `json:"id"`
	Title       string                   `json:"title"`
	Description string                   `json:"description"`
	Status      string                   `json:"status"`
	Items       []WorkoutPlanItemResponse `json:"items"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
}

type WorkoutPlanItemResponse struct {
	ID         string  `json:"id"`
	ExerciseID string  `json:"exercise_id"`
	OrderIndex int     `json:"order_index"`
	Sets       int     `json:"sets"`
	Reps       int     `json:"reps"`
	Weight     float64 `json:"weight"`
	Unit       string  `json:"unit"`
	CreatedAt  time.Time `json:"created_at"`
}