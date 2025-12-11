-- Migration: Create attachments table
-- This migration adds support for file attachments on records

CREATE TABLE IF NOT EXISTS attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    record_id UUID NOT NULL REFERENCES records(id) ON DELETE CASCADE,
    field_id UUID NOT NULL REFERENCES fields(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    size_bytes BIGINT NOT NULL,
    storage_key VARCHAR(500) NOT NULL,      -- Path in storage (local or S3)
    thumbnail_key VARCHAR(500),              -- Path to thumbnail for images
    width INT,                               -- Image width (if applicable)
    height INT,                              -- Image height (if applicable)
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Index for finding attachments by record and field
CREATE INDEX IF NOT EXISTS idx_attachments_record_field ON attachments(record_id, field_id);

-- Index for finding attachments by storage key (for cleanup)
CREATE INDEX IF NOT EXISTS idx_attachments_storage_key ON attachments(storage_key);
