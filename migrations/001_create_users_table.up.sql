-- Migration: Create users table
-- Created: 2025-02-05
-- Description: Initial users table with proper indexing and constraints

BEGIN;

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    telegram_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create unique index on telegram_id for fast lookups
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_telegram_id 
ON users (telegram_id);

-- Create index on created_at for time-based queries
CREATE INDEX IF NOT EXISTS idx_users_created_at 
ON users (created_at DESC);

-- Add constraints
ALTER TABLE users 
ADD CONSTRAINT chk_users_telegram_id_not_empty 
CHECK (LENGTH(TRIM(telegram_id)) > 0);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER trigger_users_updated_at 
    BEFORE UPDATE ON users 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

COMMIT;