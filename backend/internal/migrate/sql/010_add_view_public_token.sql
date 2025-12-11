-- Add public sharing capability to views
ALTER TABLE views ADD COLUMN IF NOT EXISTS public_token VARCHAR(64) UNIQUE;
ALTER TABLE views ADD COLUMN IF NOT EXISTS is_public BOOLEAN DEFAULT false;

-- Create index for public token lookup
CREATE INDEX IF NOT EXISTS idx_views_public_token ON views(public_token) WHERE public_token IS NOT NULL;
