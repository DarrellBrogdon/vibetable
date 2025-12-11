-- Migration: 008_add_record_color
-- Description: Add color field to records for visual highlighting

ALTER TABLE records ADD COLUMN IF NOT EXISTS color VARCHAR(20) DEFAULT NULL;
