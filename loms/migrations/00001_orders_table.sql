-- +goose Up
-- +goose StatementBegin
CREATE TABLE orders (
    id      BIGSERIAL PRIMARY KEY,
    user_id int8      not null, 
    status  text      not null default 'new'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE orders;
-- +goose StatementEnd
