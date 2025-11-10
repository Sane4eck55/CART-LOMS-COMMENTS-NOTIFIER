-- +goose Up
-- +goose StatementBegin
CREATE TABLE orders_items(
    id       bigserial primary key,
    order_id bigserial not null,
    sku      int8      not null,
    count    bigint    CHECK (count >= 0 AND count <= 4294967295)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE orders_items;
-- +goose StatementEnd
