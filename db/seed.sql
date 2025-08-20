-- db/seed.sql
-- Ensure menu items are only inserted once.

INSERT INTO menu (name, price_cents) VALUES
    ('Espresso', 250),
    ('Double Espresso', 350),
    ('Latte', 350),
    ('Cappuccino', 300),
    ('Mocha', 400),
    ('Hot Chocolate', 400),
    ('White Chocolate', 450),
    ('Matcha', 450),
    ('Chai Latte', 400)
