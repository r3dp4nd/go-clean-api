BEGIN;

DROP INDEX IF EXISTS idx_products_sku_unique_active;

DROP INDEX IF EXISTS idx_products_deleted_at;

DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1
                       FROM pg_constraint
                       WHERE conname = 'products_sku_unique') THEN
            ALTER TABLE products
                ADD CONSTRAINT products_sku_unique
                    UNIQUE (sku);
        END IF;
    END
$$;

ALTER TABLE products
    DROP COLUMN IF EXISTS deleted_at;

COMMIT;