-- Migration rollback: Drop users table
-- Created: 2025-02-05

BEGIN;

-- Drop trigger and function
DROP TRIGGER IF EXISTS trigger_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX CONCURRENTLY IF EXISTS idx_users_telegram_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_created_at;

-- Drop table
DROP TABLE IF EXISTS users CASCADE;

COMMIT;