-- Tabela para armazenar os documentos gerados para cada paciente
CREATE TABLE IF NOT EXISTS documents (
    id SERIAL PRIMARY KEY,
    psicologo_id INTEGER REFERENCES users(id),
    paciente_id INTEGER REFERENCES patients(id),
    tipo VARCHAR(50) NOT NULL, -- 'contrato_servicos', 'atestado', etc
    nome VARCHAR(255) NOT NULL,
    conteudo TEXT NOT NULL,
    requer_assinatura BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

