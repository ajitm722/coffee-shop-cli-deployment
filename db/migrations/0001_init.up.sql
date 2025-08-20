-- db/migrations/0001_init.up.sql

-- Always drop existing tables so we start fresh
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS menu;

-- Recreate menu table
CREATE TABLE menu (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    price_cents INT NOT NULL
);

-- Recreate orders table
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    customer_name TEXT NOT NULL,
    items TEXT[] NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
