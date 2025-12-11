-- Migration: 002_create_auth_tables
-- Description: Create sessions and magic_links tables for passwordless auth

-- Sessions: Active login sessions
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for token lookups (auth middleware)
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
-- Index for cleanup of expired sessions
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

-- Magic Links: Pending email verification tokens
CREATE TABLE IF NOT EXISTS magic_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for token lookups (verification)
CREATE INDEX IF NOT EXISTS idx_magic_links_token ON magic_links(token);
-- Index for cleanup of expired links
CREATE INDEX IF NOT EXISTS idx_magic_links_expires_at ON magic_links(expires_at);
