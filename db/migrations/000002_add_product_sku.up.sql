BEGIN;

ALTER TABLE products
    ADD COLUMN IF NOT EXISTS sku TEXT;

UPDATE products
SET sku = 'LEGACY-' || replace(id::text, '-', '')
WHERE sku IS NULL
   OR char_length(btrim(sku)) = 0;

ALTER TABLE products
    ALTER COLUMN sku SET NOT NULL;

DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1
                       FROM pg_constraint
                       WHERE conname = 'products_sku_not_blank') THEN
            ALTER TABLE products
                ADD CONSTRAINT products_sku_not_blank
                    CHECK (char_length(btrim(sku)) > 0);
        END IF;

        IF NOT EXISTS (SELECT 1
                       FROM pg_constraint
                       WHERE conname = 'products_sku_unique') THEN
            ALTER TABLE products
                ADD CONSTRAINT products_sku_unique
                    UNIQUE (sku);
        END IF;
    END
$$;

COMMIT;