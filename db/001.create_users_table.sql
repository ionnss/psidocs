-- Criar tabela de usuários
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    hash_crp VARCHAR(255) NOT NULL UNIQUE,
    hash_chave VARCHAR(255) NOT NULL,
    salt_chave VARCHAR(255) NOT NULL,
    salt_crp VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Criar tabela de configurações de usuários
CREATE TABLE IF NOT EXISTS users_data (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    first_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),
    last_name VARCHAR(100) NOT NULL,
    cpf VARCHAR(14) NOT NULL,
    rg VARCHAR(20) NOT NULL,
    date_of_birth DATE NOT NULL,
    ddd VARCHAR(3) NOT NULL,
    telefone VARCHAR(10) NOT NULL,
    whatsapp BOOLEAN DEFAULT false,
    endereco VARCHAR(255) NOT NULL,
    numero VARCHAR(10) NOT NULL,
    bairro VARCHAR(100) NOT NULL,
    cidade VARCHAR(100) NOT NULL,
    estado VARCHAR(2) NOT NULL,
    cep VARCHAR(9) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Criar tabela de planos
CREATE TABLE IF NOT EXISTS users_plan (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    plan_name VARCHAR (100) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    duration VARCHAR(100) NOT NULL, -- mensal, trimestral, anual
    status VARCHAR(20) NOT NULL DEFAULT 'ativo' CHECK (status IN ('ativo', 'inativo')),
    last_payment_date DATE,
    next_payment_date DATE,
    auto_renewal BOOLEAN DEFAULT true,
    cancel_date DATE, 
    cancellation_policy TEXT, -- texto explicativo da política de cancelamento
    payment_method VARCHAR(100) NOT NULL, -- pix, boleto, cartão, etc
    payment_status VARCHAR(20) NOT NULL DEFAULT 'pendente' CHECK (payment_status IN ('pendente', 'pago', 'cancelado')),
    transaction_id VARCHAR(100), -- id da transação no gateway de pagamento
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Criar índices
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_data_user_id ON users_data(user_id); 
CREATE INDEX IF NOT EXISTS idx_users_plan_user_id ON users_plan(user_id); 
