-- +goose Up
CREATE INDEX IF
NOT EXISTS idx_workout_plans_user_id ON workout_plans
    (
        user_id
    )
;
CREATE INDEX IF
NOT EXISTS idx_workout_plan_items_plan_id ON workout_plan_items
    (
        workout_plan_id
    )
;
CREATE INDEX IF
NOT EXISTS idx_workout_plan_items_exercise_id ON workout_plan_items
    (
        exercise_id
    )
;
CREATE INDEX IF
NOT EXISTS idx_scheduled_workouts_user_id ON scheduled_workouts
    (
        user_id
    )
;
CREATE INDEX IF
NOT EXISTS idx_scheduled_workouts_plan_id ON scheduled_workouts
    (
        workout_plan_id
    )
;
CREATE INDEX IF
NOT EXISTS idx_scheduled_workouts_scheduled_at ON scheduled_workouts
    (
        scheduled_at
    )
;
CREATE INDEX IF
NOT EXISTS idx_scheduled_workouts_status ON scheduled_workouts
    (
        status
    )
;
CREATE INDEX IF
NOT EXISTS idx_workout_sessions_user_id ON workout_sessions
    (
        user_id
    )
;
CREATE INDEX IF
NOT EXISTS idx_workout_session_items_session_id ON workout_session_items
    (
        workout_session_id
    )
;
CREATE INDEX IF
NOT EXISTS idx_workout_session_items_exercise_id ON workout_session_items
    (
        exercise_id
    )
;
-- +goose Down
DROP INDEX IF EXISTS idx_workout_session_items_exercise_id;
DROP INDEX IF EXISTS idx_workout_session_items_session_id;
DROP INDEX IF EXISTS idx_workout_sessions_user_id;
DROP INDEX IF EXISTS idx_scheduled_workouts_status;
DROP INDEX IF EXISTS idx_scheduled_workouts_scheduled_at;
DROP INDEX IF EXISTS idx_scheduled_workouts_plan_id;
DROP INDEX IF EXISTS idx_scheduled_workouts_user_id;
DROP INDEX IF EXISTS idx_workout_plan_items_exercise_id;
DROP INDEX IF EXISTS idx_workout_plan_items_plan_id;
DROP INDEX IF EXISTS idx_workout_plans_user_id;