-- Migration: Add request_id for tracing
-- Version: 002
-- Description: Add request_id UUID for end-to-end request tracing

ALTER TABLE access_requests ADD COLUMN request_id UUID DEFAULT uuid_generate_v4();

CREATE INDEX idx_access_requests_request_id ON access_requests(request_id);
