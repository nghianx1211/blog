CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    tags TEXT[] DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create GIN index for tags array for faster tag searches
CREATE INDEX idx_posts_tags ON posts USING GIN(tags);

-- Create index on created_at for ordering
CREATE INDEX idx_posts_created_at ON posts(created_at DESC);