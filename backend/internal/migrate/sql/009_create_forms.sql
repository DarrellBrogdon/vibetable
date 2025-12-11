-- Migration: 009_create_forms
-- Description: Create forms and form_fields tables for Form View feature

CREATE TABLE IF NOT EXISTS forms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_id UUID NOT NULL REFERENCES tables(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    public_token VARCHAR(64) UNIQUE,
    is_active BOOLEAN DEFAULT true,
    success_message TEXT DEFAULT 'Thank you for your submission!',
    redirect_url TEXT,
    submit_button_text VARCHAR(100) DEFAULT 'Submit',
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS form_fields (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    form_id UUID NOT NULL REFERENCES forms(id) ON DELETE CASCADE,
    field_id UUID NOT NULL REFERENCES fields(id) ON DELETE CASCADE,
    label VARCHAR(255),
    help_text TEXT,
    is_required BOOLEAN DEFAULT false,
    is_visible BOOLEAN DEFAULT true,
    position INT NOT NULL,
    UNIQUE(form_id, field_id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_forms_table ON forms(table_id);
CREATE INDEX IF NOT EXISTS idx_forms_token ON forms(public_token);
CREATE INDEX IF NOT EXISTS idx_form_fields_form ON form_fields(form_id, position);
