-- LangLite Ingestion Database Schema

-- Create schema if it doesn't exist
CREATE SCHEMA IF NOT EXISTS langlite;

-- Set search path to use the schema
SET search_path TO langlite, public;

-- Traces table
CREATE TABLE IF NOT EXISTS traces (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    metadata JSONB,
    tags TEXT[],
    user_id VARCHAR(255),
    session_id VARCHAR(255),
    start_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    end_time TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Generations table
CREATE TABLE IF NOT EXISTS generations (
    id VARCHAR(255) PRIMARY KEY,
    trace_id VARCHAR(255) NOT NULL REFERENCES traces(id) ON DELETE CASCADE,
    name VARCHAR(255),
    input TEXT NOT NULL,
    output TEXT,
    model VARCHAR(255) NOT NULL,
    prompt_tokens INTEGER,
    completion_tokens INTEGER,
    total_tokens INTEGER,
    metadata JSONB,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    end_time TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Spans table
CREATE TABLE IF NOT EXISTS spans (
    id VARCHAR(255) PRIMARY KEY,
    trace_id VARCHAR(255) NOT NULL REFERENCES traces(id) ON DELETE CASCADE,
    parent_id VARCHAR(255) REFERENCES spans(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50),
    metadata JSONB,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    end_time TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Events table
CREATE TABLE IF NOT EXISTS events (
    id VARCHAR(255) PRIMARY KEY,
    trace_id VARCHAR(255) NOT NULL REFERENCES traces(id) ON DELETE CASCADE,
    span_id VARCHAR(255) REFERENCES spans(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    level VARCHAR(20) NOT NULL DEFAULT 'info',
    message TEXT NOT NULL,
    metadata JSONB,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Scores table
CREATE TABLE IF NOT EXISTS scores (
    id VARCHAR(255) PRIMARY KEY,
    trace_id VARCHAR(255) REFERENCES traces(id) ON DELETE CASCADE,
    generation_id VARCHAR(255) REFERENCES generations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    value DECIMAL(5,4) NOT NULL CHECK (value >= 0 AND value <= 1),
    source VARCHAR(50) NOT NULL DEFAULT 'human',
    comment TEXT,
    metadata JSONB,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT scores_reference_check CHECK (
        (trace_id IS NOT NULL AND generation_id IS NULL) OR
        (trace_id IS NULL AND generation_id IS NOT NULL) OR
        (trace_id IS NOT NULL AND generation_id IS NOT NULL)
    )
);

-- Indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_traces_user_id ON traces(user_id);
CREATE INDEX IF NOT EXISTS idx_traces_session_id ON traces(session_id);
CREATE INDEX IF NOT EXISTS idx_traces_start_time ON traces(start_time);

CREATE INDEX IF NOT EXISTS idx_generations_trace_id ON generations(trace_id);
CREATE INDEX IF NOT EXISTS idx_generations_model ON generations(model);
CREATE INDEX IF NOT EXISTS idx_generations_start_time ON generations(start_time);

CREATE INDEX IF NOT EXISTS idx_spans_trace_id ON spans(trace_id);
CREATE INDEX IF NOT EXISTS idx_spans_parent_id ON spans(parent_id);
CREATE INDEX IF NOT EXISTS idx_spans_type ON spans(type);
CREATE INDEX IF NOT EXISTS idx_spans_start_time ON spans(start_time);

CREATE INDEX IF NOT EXISTS idx_events_trace_id ON events(trace_id);
CREATE INDEX IF NOT EXISTS idx_events_span_id ON events(span_id);
CREATE INDEX IF NOT EXISTS idx_events_level ON events(level);
CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);

CREATE INDEX IF NOT EXISTS idx_scores_trace_id ON scores(trace_id);
CREATE INDEX IF NOT EXISTS idx_scores_generation_id ON scores(generation_id);
CREATE INDEX IF NOT EXISTS idx_scores_name ON scores(name);
CREATE INDEX IF NOT EXISTS idx_scores_timestamp ON scores(timestamp);
