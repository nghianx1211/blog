CREATE TABLE activity_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    action VARCHAR(50) NOT NULL,
    post_id UUID NOT NULL,
    logged_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);

-- Create index on post_id for faster lookups
CREATE INDEX idx_activity_logs_post_id ON activity_logs(post_id);

-- Create index on logged_at for chronological ordering
CREATE INDEX idx_activity_logs_logged_at ON activity_logs(logged_at DESC);