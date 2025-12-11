-- Migration: 006_create_records
-- Description: Create records table (rows of data)

CREATE TABLE IF NOT EXISTS records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_id UUID NOT NULL REFERENCES tables(id) ON DELETE CASCADE,
    values JSONB DEFAULT '{}',
    position INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for listing records in a table
CREATE INDEX IF NOT EXISTS idx_records_table_id ON records(table_id);
-- Index for ordering records
CREATE INDEX IF NOT EXISTS idx_records_position ON records(table_id, position);
-- GIN index for querying JSONB values (filtering)
CREATE INDEX IF NOT EXISTS idx_records_values ON records USING GIN(values);
