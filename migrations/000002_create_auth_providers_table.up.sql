CREATE TABLE auth_providers (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    provider_name VARCHAR(50) NOT NULL,
    provider_id VARCHAR(255) NOT NULL,
    UNIQUE (provider_name,provider_id)
)