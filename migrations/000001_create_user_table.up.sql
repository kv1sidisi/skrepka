CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY ,
    google_id VARCHAR(255) UNIQUE NOT NULL ,
    email VARCHAR(255) UNIQUE NOT NULL ,
    name VARCHAR(255),
    avatar_url TEXT,
    created_at TIMESTAMPTZ NOT NULL ,
    updated_at TIMESTAMPTZ NOT NULL
)