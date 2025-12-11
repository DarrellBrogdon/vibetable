-- Migration: 015_add_computed_field_types
-- Description: Add formula, rollup, lookup, and attachment field types

-- Add new enum values to field_type
ALTER TYPE field_type ADD VALUE IF NOT EXISTS 'formula';
ALTER TYPE field_type ADD VALUE IF NOT EXISTS 'rollup';
ALTER TYPE field_type ADD VALUE IF NOT EXISTS 'lookup';
ALTER TYPE field_type ADD VALUE IF NOT EXISTS 'attachment';
