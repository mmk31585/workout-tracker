-- +goose Up
CREATE TABLE IF NOT EXISTS workout_plans
    (
        id      uuid PRIMARY KEY DEFAULT gen_random_uuid(),
        user_id uuid NOT NULL REFERENCES users(id) ON
        DELETE
            CASCADE                                                         ,
            title varchar(150) NOT NULL                                     ,
            description text                                                ,
            status varchar(30) NOT NULL DEFAULT 'active'                    ,
            created_at timestamptz NOT NULL DEFAULT now()                   ,
            updated_at timestamptz NOT NULL DEFAULT now()                   ,
            CONSTRAINT chk_workout_plans_title_not_empty CHECK (title != ''),
            CONSTRAINT chk_workout_plans_status CHECK (status IN ('active',
                                                                  'archived')) );
-- +goose Down
DROP TABLE IF EXISTS workout_plans;