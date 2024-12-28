CREATE TABLE IF NOT EXISTS patients (
    id SERIAL PRIMARY KEY,
    psicologo_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    nome VARCHAR(255) NOT NULL,
    estado_civil VARCHAR(100) NOT NULL,
    nacionalidade VARCHAR(100) NOT NULL,
    profissao VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    ddd VARCHAR(3) NOT NULL,
    telefone VARCHAR(20) NOT NULL,
    whatsapp BOOLEAN NOT NULL DEFAULT FALSE,
    cpf VARCHAR(14) UNIQUE,
    data_nascimento DATE,
    sexo VARCHAR(1) NOT NULL CHECK (sexo IN ('M', 'F', 'O')),
    endereco VARCHAR(255),
    numero VARCHAR(10),
    bairro VARCHAR(255),
    cidade VARCHAR(255),
    estado VARCHAR(2) NOT NULL CHECK (estado IN ('AC', 'AL', 'AP', 'AM', 'BA', 'CE', 'DF', 'ES', 'GO', 'MA', 'MT', 'MS', 'MG', 'PA', 'PB', 'PR', 'PE', 'PI', 'RJ', 'RN', 'RS', 'RO', 'RR', 'SC', 'SP', 'SE', 'TO')),
    cep VARCHAR(9),
    status VARCHAR(20) DEFAULT 'ativo' CHECK (status IN ('ativo', 'inativo', 'arquivado')),
    observacoes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Índices para melhorar performance
CREATE INDEX IF NOT EXISTS idx_patients_psicologo ON patients(psicologo_id);
CREATE INDEX IF NOT EXISTS idx_patients_status ON patients(status);
CREATE INDEX IF NOT EXISTS idx_patients_cpf ON patients(cpf);

-- Trigger para atualizar updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

DO $$ 
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_patients_updated_at') THEN
        CREATE TRIGGER update_patients_updated_at
            BEFORE UPDATE ON patients
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
    END IF;
END $$; 