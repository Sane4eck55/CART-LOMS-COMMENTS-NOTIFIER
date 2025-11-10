-- +goose Up
-- +goose StatementBegin
INSERT INTO stocks (sku, total_count, reserved) VALUES
(111111, 300, 34),
(22222, 255, 11),
(33333, 450, 0); 
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM stocks WHERE sku in(111111, 22222, 33333);
-- +goose StatementEnd