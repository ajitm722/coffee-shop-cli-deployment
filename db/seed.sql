-- db/seed.sql
INSERT INTO menu (name, price_cents) VALUES
    ('Espresso', 250),
    ('Latte', 350),
    ('Cappuccino', 300);

INSERT INTO orders (customer_name, items) VALUES
    ('John Doe', ARRAY['Espresso', 'Latte']);
