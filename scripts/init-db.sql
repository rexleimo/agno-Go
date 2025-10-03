-- AgentOS Database Initialization Script
-- This script creates tables for session management and agent metadata

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id VARCHAR(255) NOT NULL,
    user_id VARCHAR(255),
    name VARCHAR(255),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on agent_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_sessions_agent_id ON sessions(agent_id);

-- Create index on user_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);

-- Create index on created_at for sorting
CREATE INDEX IF NOT EXISTS idx_sessions_created_at ON sessions(created_at DESC);

-- Messages table (for storing conversation history)
CREATE TABLE IF NOT EXISTS messages (
    id BIGSERIAL PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,  -- system, user, assistant, tool
    content TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on session_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id);

-- Create index on created_at for sorting
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);

-- Agent runs table (for tracking agent executions)
CREATE TABLE IF NOT EXISTS agent_runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id VARCHAR(255) NOT NULL,
    session_id UUID REFERENCES sessions(id) ON DELETE SET NULL,
    input TEXT NOT NULL,
    output TEXT,
    status VARCHAR(50) NOT NULL,  -- running, completed, failed
    error TEXT,
    metadata JSONB DEFAULT '{}',
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE
);

-- Create index on agent_id for analytics
CREATE INDEX IF NOT EXISTS idx_agent_runs_agent_id ON agent_runs(agent_id);

-- Create index on session_id
CREATE INDEX IF NOT EXISTS idx_agent_runs_session_id ON agent_runs(session_id);

-- Create index on status
CREATE INDEX IF NOT EXISTS idx_agent_runs_status ON agent_runs(status);

-- Create index on started_at for time-based queries
CREATE INDEX IF NOT EXISTS idx_agent_runs_started_at ON agent_runs(started_at DESC);

-- Function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to update updated_at on sessions table
CREATE TRIGGER update_sessions_updated_at
    BEFORE UPDATE ON sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Sample data (optional, can be commented out in production)
-- INSERT INTO sessions (agent_id, user_id, name) VALUES
-- ('assistant', 'demo-user', 'Demo Session');

COMMENT ON TABLE sessions IS 'Stores session information for multi-turn conversations';
COMMENT ON TABLE messages IS 'Stores conversation history for each session';
COMMENT ON TABLE agent_runs IS 'Tracks agent execution history and performance';
