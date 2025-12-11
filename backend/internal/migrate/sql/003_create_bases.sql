-- Migration: 003_create_bases
-- Description: Create bases and base_collaborators tables

-- Bases: Container for tables (like a spreadsheet file)
CREATE TABLE IF NOT EXISTS bases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for listing bases by creator
CREATE INDEX IF NOT EXISTS idx_bases_created_by ON bases(created_by);

-- Base Collaborators: Shared access to a base
CREATE TYPE collaborator_role AS ENUM ('owner', 'editor', 'viewer');

CREATE TABLE IF NOT EXISTS base_collaborators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    base_id UUID NOT NULL REFERENCES bases(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role collaborator_role NOT NULL DEFAULT 'viewer',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(base_id, user_id)
);

-- Index for listing collaborators on a base
CREATE INDEX IF NOT EXISTS idx_base_collaborators_base_id ON base_collaborators(base_id);
-- Index for listing bases a user has access to
CREATE INDEX IF NOT EXISTS idx_base_collaborators_user_id ON base_collaborators(user_id);
