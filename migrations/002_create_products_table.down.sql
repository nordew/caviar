-- Migration rollback: Drop products and related tables
-- Created: 2025-02-05

BEGIN;

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_products_updated_at ON products;
DROP TRIGGER IF EXISTS trigger_product_variants_updated_at ON product_variants;

-- Drop indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_products_slug;
DROP INDEX CONCURRENTLY IF EXISTS idx_product_variants_product_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_product_variants_product_mass;

-- Drop tables (CASCADE will handle foreign key constraints)
DROP TABLE IF EXISTS product_variants CASCADE;
DROP TABLE IF EXISTS products CASCADE;

COMMIT;