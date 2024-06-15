-- +goose Up
-- +goose StatementBegin
CREATE TABLE orders
(
    order_id             SERIAL PRIMARY KEY,
    customer_id          INT       NOT NULL,
    expiration_time      TIMESTAMP NOT NULL,
    received_time        TIMESTAMP,
    received_by_customer BOOLEAN DEFAULT FALSE,
    refunded             BOOLEAN DEFAULT FALSE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
