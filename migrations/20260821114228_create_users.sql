-- +goose Up
CREATE TABLE IF NOT EXISTS users
    (
        id            uuid PRIMARY KEY DEFAULT gen_random_uuid()                     ,
        email         varchar(255) NOT NULL UNIQUE                                   ,
        password_hash varchar(255) NOT NULL                                          ,
        display_name  varchar(100) NOT NULL                                          ,
        created_at    timestamptz NOT NULL DEFAULT now()                             ,
        updated_at    timestamptz NOT NULL DEFAULT now()                             ,
        CONSTRAINT chk_users_email_not_empty CHECK (btrim(email)               <> ''),
        CONSTRAINT chk_users_display_name_not_empty CHECK (btrim(display_name) <> '')
    )
;
-- +goose Down
DROP TABLE IF EXISTS users;