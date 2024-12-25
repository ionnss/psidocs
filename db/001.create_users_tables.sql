CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    hash_chave TEXT NOT NULL,
    hash_crp TEXT NOT NULL,
    salt_chave TEXT NOT NULL,
    salt_crp TEXT NOT NULL,
    email TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_hash_crp ON users(hash_crp);