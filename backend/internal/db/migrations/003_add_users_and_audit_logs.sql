-- Migration: Create users table
-- Version: 003
-- Description: Add users and audit_logs tables for auth and audit trail

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    access_request_id UUID REFERENCES access_requests(id),
    action VARCHAR(50) NOT NULL,
    old_value VARCHAR(50),
    new_value VARCHAR(50),
    changed_by VARCHAR(255) NOT NULL,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_access_request_id ON audit_logs(access_request_id);
CREATE INDEX idx_audit_logs_changed_at ON audit_logs(changed_at);
