package workoutplan

import "time"

type WorkoutPlan struct {
	ID          string    `db:"id"`
	UserID      string    `db:"user_id"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	Status      string    `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type WorkoutPlanItem struct {
	ID            string    `db:"id"`
	WorkoutPlanID string    `db:"workout_plan_id"`
	ExerciseID    string    `db:"exercise_id"`
	OrderIndex    int       `db:"order_index"`
	Sets          int       `db:"sets"`
	Reps          int       `db:"reps"`
	Weight        float64   `db:"weight"`
	Unit          string    `db:"unit"`
	CreatedAt     time.Time `db:"created_at"`
}
