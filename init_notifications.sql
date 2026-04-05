CREATE TABLE IF NOT EXISTS notifications (
    id          VARCHAR(36) PRIMARY KEY,
    order_id    VARCHAR(36) NOT NULL,
    user_id     VARCHAR(36) NOT NULL,
    message     TEXT        NOT NULL,
    created_at  TIMESTAMP   NOT NULL
);