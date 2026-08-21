-- +goose Up
CREATE TABLE IF NOT EXISTS exercises
    (
        id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
        name         varchar(150) NOT NULL                     ,
        description  text                                      ,
        category     varchar(80)                               ,
        muscle_group varchar(80)                               ,
        created_at   timestamptz NOT NULL DEFAULT now()        ,
        CONSTRAINT uq_exercises_name UNIQUE (name)
    )
;
-- +goose Down
DROP TABLE IF EXISTS exercises;