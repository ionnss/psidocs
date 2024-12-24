# PsiDocs

Plataforma de gerenciamento de documentos e contratos para psicólogos
- **Documentos Psicológicos**
	- *Laudos*
	- *Relatórios*
- **Contratos dos pacientes**
	- *Auxílio jurídico*
	- *Criação*
	- *Envio para assinatura*

A plataforma não exige cadastro. Pois ao acessar pela primeira vez, já é um cadastro. Caso já esteja cadastrado, reconhece e devolve o painel do psicólogo.

Você acessa por uma senha e registro de classe (CRP), que serão criptografados.
```shell
Chave: A!l#40501993
CRP: 936729
```

Esse conjunto será coletado como `Chave` e `CRP`, gerando um hash para cada um utilizando bcrypt
```shell
---------------------
HEX-output-Hash-Chave:
---------------------
15c7139ad02ee51dc4ba64d48dbf0312b251957acc2ed8ca2f9788a66aba12cd

------------------------------
Encoded-Output-Form-Hash-Chave: 
------------------------------
$argon2id$v=19$m=64,t=3,p=3$eXlYN2Z0SnZUeTI1S0dzSQ$FccTmtAu5R3EumTUjb8DErJRlXrMLtjKL5eIpmq6Es0
```

```shell
-------------------
HEX-output-Hash-CRP:
-------------------
dbf799a6d5f873cbfe0a13d2f3c76eeb1c68820ef01da53e9697fad882064264

----------------------------
Encoded-Output-Form-Hash-CRP:
----------------------------
$argon2id$v=19$m=64,t=3,p=3$eXlYN2Z0SnZUeTI1S0dzSQ$2/eZptX4c8v+ChPS88du6xxogg7wHaU+lpf62IIGQmQ
```

---
Seu projeto **PsiDocs** tem uma proposta sólida e atende a uma necessidade real para psicólogos ao simplificar a gestão de documentos e contratos, especialmente com a inclusão de elementos como **laudos**, **relatórios**, e **contratos**. Aqui está uma análise e algumas sugestões para o desenvolvimento e organização:

---

## **Características Relevantes**

1. **Documentos Psicológicos**:
    
    - Funcionalidade de geração de **laudos** e **relatórios** pode ser implementada usando templates personalizáveis, permitindo que psicólogos modifiquem facilmente as informações antes de gerar o documento final.
2. **Contratos dos Pacientes**:
    
    - Auxílio jurídico: Pode ser realizado com modelos pré-desenhados de contratos que podem ser personalizados.
    - Envio para assinatura: A integração com serviços de assinatura eletrônica como **DocuSign** ou similares seria ideal. Para uma solução simplificada, o envio pode ser realizado via e-mail com um botão de "Aceitar Contrato".
3. **Cadastro Automático**:
    
    - Essa abordagem (sem necessidade de cadastro manual) é inovadora e prática. Para implementá-la, basta registrar o par **Chave** e **CRP** como identificador único (hashes gerados pelo Argon2id), vinculando-os ao painel do psicólogo na primeira interação.

---

## **Segurança e Criptografia**

Você já está no caminho certo ao usar **Argon2id** como método de hashing para senhas e identificadores (CRP). Aqui estão alguns pontos importantes sobre segurança:

1. **Hash de Chave e CRP**:
    
    - Os dois hashes criados (um para a chave e outro para o CRP) tornam o acesso seguro e difícil de comprometer.
    - É importante **nunca armazenar a chave ou o CRP original em texto puro**.
2. **Salt Aleatório**:
    
    - O `salt` adicionado ao Argon2id precisa ser gerado aleatoriamente para cada hash gerado.
    - Certifique-se de armazenar o `salt` com segurança, pois ele é necessário para verificar o hash.
3. **Acesso e Reconhecimento Automático**:
    
    - Quando o usuário acessa o sistema, você pode verificar os hashes da `Chave` e do `CRP` em um banco de dados para autenticação.
    - Um token JWT pode ser usado para manter a sessão do usuário, garantindo acesso contínuo sem precisar refazer o login.

---

## **Fluxo Inicial do PsiDocs**

### 1. **Primeiro Acesso (Cadastro Automático)**:

- O psicólogo insere a Chave e o CRP.
- Você gera os hashes usando Argon2id e armazena os valores (HEX ou Encoded) no banco de dados junto com o Salt.
- Cria o painel inicial do psicólogo no sistema.

### 2. **Acesso Posterior**:

- O psicólogo insere novamente a Chave e o CRP.
- O sistema gera os hashes e os compara com os armazenados no banco de dados.
- Caso coincida, o painel personalizado do psicólogo é carregado.

---

## **Arquitetura do Sistema**

### **1. Backend**

- **Linguagem**: Golang
- **Frameworks**: Use o `Gin` ou `Fiber` para criar as rotas do backend.
- **Autenticação**: Argon2id para hash das credenciais + JWT para gerenciar sessões.
- **Banco de Dados**: MySQL, PostgreSQL ou SQLite para armazenar informações dos psicólogos e documentos gerados.

### **2. Frontend**

- **Frameworks**:
    - **HTMX**: Para interatividade sem necessidade de JavaScript pesado.
    - **Bootstrap**: Para uma interface elegante e responsiva.
- **Templates**: Personalizados para exibir o painel do psicólogo e permitir o upload, criação e edição de documentos.

---

## **Exemplo de Banco de Dados**

```sql
-- Tabela de usuários
CREATE TABLE psicologos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    hash_chave TEXT NOT NULL,
    hash_crp TEXT NOT NULL,
    salt_chave TEXT NOT NULL,
    salt_crp TEXT NOT NULL
);

-- Tabela de documentos
CREATE TABLE documentos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    psicologo_id INTEGER NOT NULL,
    tipo TEXT NOT NULL, -- 'laudo', 'relatorio', ou 'contrato'
    conteudo TEXT NOT NULL,
    data_criacao TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (psicologo_id) REFERENCES psicologos (id)
);
```

---

## **Próximos Passos**

1. **Autenticação Básica**:
    
    - Configure as rotas para criar e validar os hashes no backend.
    - Utilize o banco de dados para armazenar e verificar os dados.
2. **Criação de Documentos**:
    
    - Implemente uma funcionalidade para psicólogos criarem documentos a partir de templates pré-definidos.
    - Adicione suporte para download em formatos como PDF.
3. **Segurança**:
    
    - Configure HTTPS no servidor para proteger os dados transmitidos.
    - Armazene os hashes e salts de forma segura no banco de dados.
4. **Interface do Usuário**:
    
    - Desenvolva o painel do psicólogo com funcionalidades intuitivas para acessar e gerenciar documentos.

Se precisar de ajuda em alguma etapa específica, posso ajudar com exemplos de código ou estratégias! 😊


---
## Registro de psicólogos no Brasil

- https://cadastro.cfp.org.br/

Não existe uma api para consulta de numero de CPR, portanto, é necessário fazer a consulta manualmente. 

