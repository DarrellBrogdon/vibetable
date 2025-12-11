-- Migration: 004_create_tables
-- Description: Create tables table (sheets within a base)

CREATE TABLE IF NOT EXISTS tables (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    base_id UUID NOT NULL REFERENCES bases(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for listing tables in a base
CREATE INDEX IF NOT EXISTS idx_tables_base_id ON tables(base_id);
-- Index for ordering tables
CREATE INDEX IF NOT EXISTS idx_tables_position ON tables(base_id, position);
