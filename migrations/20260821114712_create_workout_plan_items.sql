-- +goose Up
CREATE TABLE IF NOT EXISTS workout_plan_items
    (
        id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
        workout_plan_id uuid NOT NULL REFERENCES workout_plans(id) ON
        DELETE
            CASCADE                                                ,
            exercise_id uuid NOT NULL REFERENCES exercises(id)     ,
            order_index int NOT NULL DEFAULT 0                     ,
            sets int NOT NULL                                      ,
            reps int NOT NULL                                      ,
            weight numeric(10,2) NOT NULL                          ,
            unit varchar(20) NOT NULL DEFAULT 'kg'                 ,
            created_at timestamptz NOT NULL DEFAULT now()          ,
            CONSTRAINT chk_wpi_sets CHECK (sets               > 0) ,
            CONSTRAINT chk_wpi_reps CHECK (reps               > 0) ,
            CONSTRAINT chk_wpi_weight CHECK (weight           >= 0),
            CONSTRAINT chk_wpi_order_index CHECK (order_index >= 0) );
-- +goose Down
DROP TABLE IF EXISTS workout_plan_items;