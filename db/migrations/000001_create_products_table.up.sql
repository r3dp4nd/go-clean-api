BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS products
(
    id          UUID PRIMARY KEY        DEFAULT gen_random_uuid(),

    name        TEXT           NOT NULL,
    description TEXT           NOT NULL,
    price       NUMERIC(12, 2) NOT NULL,

    created_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

    CONSTRAINT products_name_not_blank CHECK (char_length(btrim(name)) > 0),
    CONSTRAINT products_description_not_null CHECK (description IS NOT NULL),
    CONSTRAINT products_price_non_negative CHECK (price >= 0)
);

CREATE INDEX IF NOT EXISTS idx_products_name ON products (name);
CREATE INDEX IF NOT EXISTS idx_products_price ON products (price);
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products (created_at);
CREATE INDEX IF NOT EXISTS idx_products_updated_at ON products (updated_at);

COMMIT;