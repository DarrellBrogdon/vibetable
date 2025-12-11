-- Migration: 011_add_view_types
-- Description: Add calendar and gallery view types to the view_type enum

ALTER TYPE view_type ADD VALUE IF NOT EXISTS 'calendar';
ALTER TYPE view_type ADD VALUE IF NOT EXISTS 'gallery';
