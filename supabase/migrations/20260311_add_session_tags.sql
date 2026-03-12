-- Add tags column to sessions table for session categorization
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS tags text[] DEFAULT ARRAY[]::text[];

-- GIN index for efficient array contains queries (cs. operator)
CREATE INDEX IF NOT EXISTS idx_sessions_tags ON sessions USING GIN (tags);
