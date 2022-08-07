BEGIN;
CREATE TABLE IF NOT EXISTS users(
    id SERIAL PRIMARY KEY,
    username VARCHAR(128) UNIQUE,
    password VARCHAR(128)
);

CREATE TABLE IF NOT EXISTS balance(
    id SERIAL PRIMARY KEY,
    user_id INTEGER UNIQUE,
    balance FLOAT,
    spent_all_time FLOAT
);

CREATE TABLE IF NOT EXISTS orders(
    id  SERIAL PRIMARY KEY,
    user_id BIGINT,
    number BIGINT UNIQUE,
    upload_time TIMESTAMP WITH TIME ZONE,
    accrual BIGINT,
    status VARCHAR(64)
);

CREATE TABLE IF NOT EXISTS withdrawals(
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT,
    sum FLOAT,
    processed_at TIMESTAMP WITH TIME ZONE
);

COMMIT;