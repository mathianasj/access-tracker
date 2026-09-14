-- Migration: Create access_requests table
-- Version: 001
-- Description: Initial schema for access requests

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE access_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    requester VARCHAR(255) NOT NULL,
    system_resource VARCHAR(255) NOT NULL,
    access_level VARCHAR(100) NOT NULL CHECK (access_level IN ('read', 'write', 'admin')),
    justification TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'denied')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_access_requests_requester ON access_requests(requester);
CREATE INDEX idx_access_requests_status ON access_requests(status);
CREATE INDEX idx_access_requests_created_at ON access_requests(created_at);
