-- Setup test data for authentication
-- This should be run after the main schema.sql

SET search_path TO langlite, public;

-- Insert a test project
INSERT INTO projects (id, name, description, created_at, updated_at) 
VALUES (
    'test-project-1', 
    'Test Project', 
    'A test project for development', 
    NOW(), 
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- Insert a test API key
-- Key: test-key-123 (plain text for testing)
-- Hash: SHA256 of "test-key-123"
INSERT INTO api_keys (
    id, 
    project_id, 
    key_hash, 
    name, 
    rate_limit_per_minute, 
    rate_limit_per_hour, 
    is_active, 
    created_at, 
    updated_at
) VALUES (
    'api-key-1',
    'test-project-1',
    '625faa3fbbc3d2bd9d6ee7678d04cc5339cb33dc68d9b58451853d60046e226a', -- SHA256 of "test-key-123"
    'Test API Key',
    1000,
    10000,
    true,
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;
