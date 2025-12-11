-- Migration: 005_create_fields
-- Description: Create fields table (column definitions)

CREATE TYPE field_type AS ENUM (
    'text',
    'number',
    'checkbox',
    'date',
    'single_select',
    'multi_select',
    'linked_record'
);

CREATE TABLE IF NOT EXISTS fields (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_id UUID NOT NULL REFERENCES tables(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    field_type field_type NOT NULL DEFAULT 'text',
    options JSONB DEFAULT '{}',
    position INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for listing fields in a table
CREATE INDEX IF NOT EXISTS idx_fields_table_id ON fields(table_id);
-- Index for ordering fields
CREATE INDEX IF NOT EXISTS idx_fields_position ON fields(table_id, position);
