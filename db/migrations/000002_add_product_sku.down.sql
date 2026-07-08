BEGIN;

ALTER TABLE products
    DROP CONSTRAINT IF EXISTS products_sku_unique;

ALTER TABLE products
    DROP CONSTRAINT IF EXISTS products_sku_not_blank;

ALTER TABLE products
    DROP COLUMN IF EXISTS sku;

COMMIT;