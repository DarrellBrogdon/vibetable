-- Migration: 016_create_automations
-- Description: Create automations tables for triggers and actions

-- Trigger types: record_created, record_updated, record_deleted, field_value_changed, scheduled
-- Action types: send_email, update_record, create_record, send_webhook

CREATE TABLE IF NOT EXISTS automations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    base_id UUID NOT NULL REFERENCES bases(id) ON DELETE CASCADE,
    table_id UUID NOT NULL REFERENCES tables(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    enabled BOOLEAN NOT NULL DEFAULT true,

    -- Trigger configuration
    trigger_type VARCHAR(50) NOT NULL,
    trigger_config JSONB NOT NULL DEFAULT '{}',

    -- Action configuration
    action_type VARCHAR(50) NOT NULL,
    action_config JSONB NOT NULL DEFAULT '{}',

    -- Metadata
    created_by UUID NOT NULL REFERENCES users(id),
    last_triggered_at TIMESTAMP WITH TIME ZONE,
    run_count INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Automation run history
CREATE TABLE IF NOT EXISTS automation_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    automation_id UUID NOT NULL REFERENCES automations(id) ON DELETE CASCADE,

    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, running, success, failed

    -- Trigger context
    trigger_record_id UUID REFERENCES records(id) ON DELETE SET NULL,
    trigger_data JSONB,

    -- Execution result
    result JSONB,
    error TEXT,

    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_automations_base_id ON automations(base_id);
CREATE INDEX IF NOT EXISTS idx_automations_table_id ON automations(table_id);
CREATE INDEX IF NOT EXISTS idx_automations_enabled ON automations(enabled) WHERE enabled = true;
CREATE INDEX IF NOT EXISTS idx_automation_runs_automation_id ON automation_runs(automation_id);
CREATE INDEX IF NOT EXISTS idx_automation_runs_status ON automation_runs(status);
CREATE INDEX IF NOT EXISTS idx_automation_runs_started_at ON automation_runs(started_at);
