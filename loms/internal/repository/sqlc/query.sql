-- name: AddOrderToOrders :one
INSERT INTO orders (user_id) VALUES ($1) returning id;

-- name: AddOrderToOrdersItems :exec
INSERT INTO orders_items (order_id, sku, count) VALUES
($1, $2, $3),
($4, $5, $6),
($7, $8, $9),
($10, $11, $12),
($13, $14, $15);

-- name: SetStatusOrder :exec
UPDATE orders SET status = $1 WHERE id = $2;

-- name: GetInfoOrders :many
SELECT user_id, status FROM orders WHERE id = $1;

-- name: GetInfoOrdersItems :many
SELECT sku, count FROM orders_items WHERE order_id = $1;

-- name: GetStocksBySku :many
SELECT total_count, reserved FROM stocks WHERE sku = $1;

-- name: GetStocksBySkuForUpdate :many
SELECT total_count, reserved FROM stocks WHERE sku = $1 FOR UPDATE;

-- name: ReserveStockBySku :exec
UPDATE stocks SET reserved = $1 WHERE sku = $2;

-- name: ReserveRemove :exec
UPDATE stocks SET total_count = $1, reserved = $2 WHERE sku = $3;

-- name: GetReservedStocksBySku :many
SELECT reserved FROM stocks WHERE sku = $1;

-- name: ReserveCancel :exec
UPDATE stocks SET reserved = $1 WHERE sku = $2;

-- name: DeleteOrders :exec
DELETE FROM orders WHERE id = $1;

-- name: DeleteOrdersItems :exec
DELETE FROM orders_items WHERE order_id = $1;

-- name: AddOutbox :exec
INSERT INTO outbox (topic, key, payload) VALUES ($1, $2, $3);

-- name: GetNewMsgOutbox :many
SELECT id, topic, key, payload FROM outbox WHERE status = 'new' ORDER BY created_at ASC, id ASC LIMIT 100 FOR UPDATE SKIP LOCKED;

-- name: UpdateStatusMsgOutbox :exec
UPDATE outbox SET status=$1, sent_at=now() WHERE id = $2;