-- Drop triggers
DROP TRIGGER IF EXISTS update_orders_updated_at ON orders;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_orders_status;
DROP INDEX IF EXISTS idx_orders_customer_phone;
DROP INDEX IF EXISTS idx_orders_country;
DROP INDEX IF EXISTS idx_orders_created_at;
DROP INDEX IF EXISTS idx_orders_order_number;

DROP INDEX IF EXISTS idx_order_items_order_id;
DROP INDEX IF EXISTS idx_order_items_product_id;
DROP INDEX IF EXISTS idx_order_items_variant_id;

-- Drop tables
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;