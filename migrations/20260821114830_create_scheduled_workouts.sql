-- +goose Up
CREATE TABLE IF NOT EXISTS scheduled_workouts
    (
        id      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
        user_id uuid NOT NULL REFERENCES users(id) ON
        DELETE
            CASCADE                                                   ,
            workout_plan_id uuid NOT NULL REFERENCES workout_plans(id),
            scheduled_at timestamptz NOT NULL                         ,
            status text NOT NULL DEFAULT 'scheduled'                  ,
            notes text                                                ,
            created_at timestamptz NOT NULL DEFAULT now()             ,
            updated_at timestamptz NOT NULL DEFAULT now()             ,
            CONSTRAINT chk_scheduled_workouts_status CHECK (status IN ('scheduled',
                                                                       'in_progress',
                                                                       'completed',
                                                                       'cancelled')) );
-- +goose Down
DROP TABLE IF EXISTS scheduled_workouts;