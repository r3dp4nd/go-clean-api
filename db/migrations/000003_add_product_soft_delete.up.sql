BEGIN;

ALTER TABLE products
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

ALTER TABLE products
    DROP CONSTRAINT IF EXISTS products_sku_unique;

CREATE UNIQUE INDEX IF NOT EXISTS idx_products_sku_unique_active
    ON products (sku)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_products_deleted_at
    ON products (deleted_at);

COMMIT;