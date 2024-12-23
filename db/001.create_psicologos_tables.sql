CREATE TABLE IF NOT EXISTS psicologos (
    id SERIAL PRIMARY KEY,
    hash_chave TEXT NOT NULL,
    hash_crp TEXT NOT NULL,
    salt_chave TEXT NOT NULL,
    salt_crp TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_psicologos_hash_crp ON psicologos(hash_crp);