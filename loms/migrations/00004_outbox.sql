-- +goose Up
-- +goose StatementBegin
CREATE TABLE if not exists outbox
(
    id         BIGSERIAL   PRIMARY KEY,
    topic      text        NOT NULL,
    key        text,
    payload    jsonb       NOT NULL,
    status     text        NOT NULL DEFAULT 'new', -- new | sent | error
    created_at timestamptz NOT NULL DEFAULT now(),
    sent_at    timestamptz
);
CREATE INDEX ON outbox (status, created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE outbox;
-- +goose StatementEnd
