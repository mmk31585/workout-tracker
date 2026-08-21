-- +goose Up
CREATE TABLE IF NOT EXISTS workout_session_items
    (
        id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
        workout_session_id uuid NOT NULL REFERENCES workout_sessions(id) ON
        DELETE
            CASCADE                                           ,
            exercise_id uuid NOT NULL REFERENCES exercises(id),
            sets int NOT NULL                                 ,
            reps int NOT NULL                                 ,
            weight numeric(10,2) NOT NULL                     ,
            unit varchar(20) NOT NULL DEFAULT 'kg'            ,
            notes text                                        ,
            created_at timestamptz NOT NULL DEFAULT now()     ,
            CONSTRAINT chk_wsi_sets CHECK (sets     > 0)      ,
            CONSTRAINT chk_wsi_reps CHECK (reps     > 0)      ,
            CONSTRAINT chk_wsi_weight CHECK (weight >= 0) );
-- +goose Down
DROP TABLE IF EXISTS workout_session_items;