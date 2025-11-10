-- +goose Up
-- +goose StatementBegin
CREATE TABLE comments
(
    id          BIGSERIAL NOT NULL PRIMARY KEY,
    user_id     BIGSERIAL NOT NULL,
    sku         BIGSERIAL NOT NULL,
    comment     TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE comments;
-- +goose StatementEnd
