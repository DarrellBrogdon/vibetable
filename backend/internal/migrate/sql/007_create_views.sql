-- Migration: 007_create_views
-- Description: Create views table (saved view configurations)

CREATE TYPE view_type AS ENUM ('grid', 'kanban');

CREATE TABLE IF NOT EXISTS views (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_id UUID NOT NULL REFERENCES tables(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    view_type view_type NOT NULL DEFAULT 'grid',
    config JSONB DEFAULT '{}',
    position INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for listing views in a table
CREATE INDEX IF NOT EXISTS idx_views_table_id ON views(table_id);
-- Index for ordering views
CREATE INDEX IF NOT EXISTS idx_views_position ON views(table_id, position);
