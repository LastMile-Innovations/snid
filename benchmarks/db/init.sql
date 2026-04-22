-- SNID Database Benchmark Schema
-- This file initializes the benchmark database with all required tables

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS pgstattuple;

-- SNID (binary storage - 16 bytes)
CREATE TABLE IF NOT EXISTS snid_binary (
    id BYTEA PRIMARY KEY,
    data TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- SNID (UUID type storage - 16 bytes)
CREATE TABLE IF NOT EXISTS snid_uuid (
    id UUID PRIMARY KEY,
    data TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- UUIDv7 (baseline - 16 bytes)
CREATE TABLE IF NOT EXISTS uuidv7 (
    id UUID PRIMARY KEY,
    data TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- UUIDv4 (baseline - 16 bytes)
CREATE TABLE IF NOT EXISTS uuidv4 (
    id UUID PRIMARY KEY,
    data TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- ULID (text storage - 26 chars)
CREATE TABLE IF NOT EXISTS ulid (
    id TEXT PRIMARY KEY,
    data TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Sequential BIGINT (baseline best - 8 bytes)
CREATE TABLE IF NOT EXISTS sequential (
    id BIGSERIAL PRIMARY KEY,
    data TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes for query performance tests
CREATE INDEX IF NOT EXISTS idx_snid_binary_created_at ON snid_binary(created_at);
CREATE INDEX IF NOT EXISTS idx_snid_uuid_created_at ON snid_uuid(created_at);
CREATE INDEX IF NOT EXISTS idx_uuidv7_created_at ON uuidv7(created_at);
CREATE INDEX IF NOT EXISTS idx_uuidv4_created_at ON uuidv4(created_at);
CREATE INDEX IF NOT EXISTS idx_ulid_created_at ON ulid(created_at);
CREATE INDEX IF NOT EXISTS idx_sequential_created_at ON sequential(created_at);

-- Grant permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO benchmark;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO benchmark;
