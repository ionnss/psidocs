-- Tabela para armazenar os documentos gerados para cada paciente
CREATE TABLE IF NOT EXISTS documents (
    id SERIAL PRIMARY KEY,
    template_id INTEGER REFERENCES document_templates(id),
    psicologo_id INTEGER REFERENCES users(id),
    paciente_id INTEGER REFERENCES patients(id),
    nome VARCHAR(255) NOT NULL,
    conteudo TEXT NOT NULL, -- HTML com dados preenchidos
    status VARCHAR(50) NOT NULL DEFAULT 'rascunho', -- 'rascunho', 'finalizado'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


