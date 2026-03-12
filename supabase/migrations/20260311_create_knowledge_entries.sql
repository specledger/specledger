-- Migration: Create knowledge_entries table for Session-to-Knowledge Memory Pipeline
-- Feature: 607-session-memory-pipeline

CREATE TABLE IF NOT EXISTS knowledge_entries (
    id TEXT PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    title TEXT NOT NULL CHECK (char_length(title) BETWEEN 1 AND 200),
    description TEXT NOT NULL CHECK (char_length(description) BETWEEN 1 AND 5000),
    tags TEXT[] NOT NULL CHECK (array_length(tags, 1) BETWEEN 1 AND 10),
    source_session_id TEXT,
    source_branch TEXT,
    score_recurrence REAL NOT NULL DEFAULT 0 CHECK (score_recurrence BETWEEN 0 AND 10),
    score_impact REAL NOT NULL DEFAULT 0 CHECK (score_impact BETWEEN 0 AND 10),
    score_specificity REAL NOT NULL DEFAULT 0 CHECK (score_specificity BETWEEN 0 AND 10),
    composite_score REAL NOT NULL DEFAULT 0 CHECK (composite_score BETWEEN 0 AND 10),
    status TEXT NOT NULL DEFAULT 'candidate' CHECK (status IN ('candidate', 'promoted', 'archived')),
    recurrence_count INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for project-scoped queries
CREATE INDEX IF NOT EXISTS idx_knowledge_entries_project_id ON knowledge_entries(project_id);

-- Index for status filtering
CREATE INDEX IF NOT EXISTS idx_knowledge_entries_status ON knowledge_entries(status);

-- Index for composite score ordering
CREATE INDEX IF NOT EXISTS idx_knowledge_entries_composite ON knowledge_entries(composite_score DESC);

-- Enable Row Level Security
ALTER TABLE knowledge_entries ENABLE ROW LEVEL SECURITY;

-- RLS Policy: Users can read entries for projects they belong to
CREATE POLICY knowledge_entries_select ON knowledge_entries
    FOR SELECT
    USING (
        project_id IN (
            SELECT p.id FROM projects p
            WHERE p.id = knowledge_entries.project_id
        )
    );

-- RLS Policy: Users can insert entries for their projects
CREATE POLICY knowledge_entries_insert ON knowledge_entries
    FOR INSERT
    WITH CHECK (
        project_id IN (
            SELECT p.id FROM projects p
            WHERE p.id = knowledge_entries.project_id
        )
    );

-- RLS Policy: Users can update their own entries
CREATE POLICY knowledge_entries_update ON knowledge_entries
    FOR UPDATE
    USING (
        project_id IN (
            SELECT p.id FROM projects p
            WHERE p.id = knowledge_entries.project_id
        )
    );

-- RLS Policy: Users can delete their own entries
CREATE POLICY knowledge_entries_delete ON knowledge_entries
    FOR DELETE
    USING (
        project_id IN (
            SELECT p.id FROM projects p
            WHERE p.id = knowledge_entries.project_id
        )
    );

-- Enable upsert on conflict (for sync merge-duplicates)
-- PostgREST uses this with Prefer: resolution=merge-duplicates
