-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS orders
(
    order_id             SERIAL PRIMARY KEY,
    customer_id          INT       NOT NULL,
    expiration_time      TIMESTAMP NOT NULL,
    received_time        TIMESTAMP,
    received_by_customer BOOLEAN DEFAULT FALSE,
    refunded             BOOLEAN DEFAULT FALSE,
    package              TEXT,
    weight               FLOAT     NOT NULL,
    cost                 FLOAT     NOT NULL,
    package_cost         INT       NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
