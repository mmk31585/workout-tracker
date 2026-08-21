-- +goose Up
CREATE TABLE IF NOT EXISTS workout_sessions
    (
        id      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
        user_id uuid NOT NULL REFERENCES users(id) ON
        DELETE
            CASCADE                                                                    ,
            scheduled_workout_id uuid NOT NULL UNIQUE REFERENCES scheduled_workouts(id),
            performed_at timestamptz NOT NULL                                          ,
            overall_notes text                                                         ,
            created_at timestamptz NOT NULL DEFAULT now()                              ,
            updated_at timestamptz NOT NULL DEFAULT now() );
-- +goose Down
DROP TABLE IF EXISTS workout_sessions;