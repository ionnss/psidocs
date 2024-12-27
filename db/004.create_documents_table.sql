-- Tabela para armazenar os templates de documentos
CREATE TABLE IF NOT EXISTS document_templates (
    id SERIAL PRIMARY KEY,
    psicologo_id INTEGER REFERENCES users(id),
    nome VARCHAR(255) NOT NULL,
    descricao TEXT,
    tipo VARCHAR(50) NOT NULL, -- 'contract', 'laudo', 'parecer', 'relatorio', 'atestado', 'declaracao', 'anamnese'
    subtipo VARCHAR(50), -- 'presencial', 'online' (para contratos), outros subtipos específicos
    conteudo TEXT NOT NULL, -- template HTML com placeholders
    padrao BOOLEAN DEFAULT false, -- indica se é um template padrão do sistema
    versao VARCHAR(10), -- controle de versão do template
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

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


