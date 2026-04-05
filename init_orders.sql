CREATE TABLE IF NOT EXISTS orders (
    id         VARCHAR(36) PRIMARY KEY,
    user_id    VARCHAR(36) NOT NULL,
    item       VARCHAR(255) NOT NULL,
    status     VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS outbox (
    id         VARCHAR(36) PRIMARY KEY,
    order_id   VARCHAR(36) NOT NULL REFERENCES orders(id), -- FK на orders
    event_type VARCHAR(100) NOT NULL,
    payload    TEXT NOT NULL,        -- JSON события целиком
    sent       BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL
);