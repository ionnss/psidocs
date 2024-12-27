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

Você já está no caminho certo ao usar **bcrypt** como método de hashing para senhas e identificadores (CRP). Aqui estão alguns pontos importantes sobre segurança:

1. **Hash de Chave e CRP**:
    
    - Os dois hashes criados (um para a chave e outro para o CRP) tornam o acesso seguro e difícil de comprometer.
    - É importante **nunca armazenar a chave ou o CRP original em texto puro**.
2. **Salt Aleatório**:
    
    - O `salt` adicionado ao bcrypt precisa ser gerado aleatoriamente para cada hash gerado.
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

### 3. **Criação da Configuração do Psicólogo**

- O psicólogo insere os dados pessoais para contratos de pacientes e documentos psicológicos.
- O sistema armazena esses dados no banco de dados.

---

## **Arquitetura do Sistema**

### **1. Backend**

- **Linguagem**: Golang
- **Frameworks**: Use o `Mux` para criar as rotas do backend.
- **Autenticação**: `Gorrila Sessions` para gerenciar sessões e autenticação.
- **Criptografia**: `bcrypt` para hash das credenciais.
- **Banco de Dados**: PostgreSQL para armazenar informações dos psicólogos e documentos gerados.

### **2. Frontend**

- **Frameworks**:
    - **HTMX**: Para interatividade sem necessidade de JavaScript pesado.
    - **Bootstrap**: Para uma interface elegante e responsiva.
- **Templates**: Personalizados para exibir o painel do psicólogo e permitir o upload, criação e edição de documentos.

---

## Registro de psicólogos no Brasil

- https://cadastro.cfp.org.br/

Não existe uma api para consulta de numero de CPR, portanto, é necessário fazer a consulta manualmente. 

---
# Produção

## Acesso ao banco de dados pelo docker

```shell
docker exec -it psidocs-db-1 psql -U psicodocs -d superviso
```

---
# Desenvolvimento

## **Próximos Passos**



Vou criar um roadmap organizado dos próximos passos para o desenvolvimento do PsiDocs:

0. **Refatorar a página dashboard**
   - [X] Navbar sidebar
   - [X] UI (se comunicar com a UI index pois gostei do design)
     - [X] Background
     - [X] Cores
     - [X] Animações
     - [X] Responsividade
   - [ ] Tour guiado para novos usuários (?)
   - [ ] Melhorar a UI do usuário config
      - [ ] Alterar chave
         - [X] Estrutura básica
         - [ ] Verificação por email
      - [ ] Alterar CRP
         - [X] Estrutura básica
         - [ ] Verificação por email
      - [ ] Alterar email
         - [X] Estrutura básica
         - [ ] Verificação por email
  

1. **Sistema de Backup**
   - [ ] Script de backup automatizado no VPS
   - [ ] Rotação de backups (7 dias)
   - [ ] Monitoramento e alertas
   - [ ] Documentação do processo de restore

2. **Gestão de Pacientes**
   - [X] Tabela `pacientes` com:
     - [X] Dados básicos (nome, data_nascimento, cpf)
     - [X] Contato (email, telefone)
     - [X] Status (ativo/inativo)
     - [X] Vinculação com psicólogo
   - [X] CRUD completo de pacientes
   - [X] Interface intuitiva para gestão

3. **Configuração do Psicólogo**
   - [X] Tabela `users_config` com:
     - [X] Dados pessoais para contratos de pacientes
     - [X] Dados pessoais para documentos psicológicos
     - [X] Status (ativo/inativo)
     - [X] Vinculação com psicólogo
   - [X] CRUD completo de configuração
   - [ ] Interface intuitiva para gestão
   - [X] Recuperar chave do usuário página login

4. **Documentos Psicológicos**
   - [ ] Templates conforme Resolução CFP:
     - [ ] Declaração
     - [ ] Atestado
     - [ ] Relatório/Laudo
     - [ ] Parecer
   - [ ] Versionamento de documentos
   - [ ] Assinatura digital
   - [ ] Exportação em PDF

5. **Contratos e Termos**
   - [ ] Modelos de:
     - [ ] Contrato terapêutico
     - [ ] Termo de consentimento
     - [ ] Política de faltas
   - [ ] Personalização de modelos
   - [ ] Histórico de versões

6. **Melhorias de Segurança**
   - [ ] 2FA (email/app)
   - [ ] Logs de auditoria
   - [ ] Monitoramento de tentativas de invasão
   - [ ] Métricas de segurança

7. **Dashboard Aprimorado**
   - [X] Visão geral de pacientes
   - [ ] Documentos recentes
   - [ ] Alertas e notificações
   - [ ] Métricas e estatísticas

8. **Agenda e Sessões**
   - [ ] Calendário de atendimentos
   - [ ] Registro de sessões
   - [ ] Lembretes automáticos
   - [ ] Gestão de faltas

9. **Financeiro Básico**
   - [ ] Registro de pagamentos
   - [ ] Controle de inadimplência
   - [ ] Relatórios financeiros
   - [ ] Exportação para contabilidade

10. **Integrações**
   - [X] Envio de emails
   - [ ] WhatsApp para lembretes
   - [ ] Integração com calendário
   - [ ] Backup em nuvem (Google Drive/Dropbox)



---
# PRÓXIMO PASSO

Minha ideia é a seguinte.

Agora que ja temos pacientes e todas as funções para cada paciente, podemos ir para a parte de documentos.

Os documentos seriam:
- Contratos (contrato psicologo/user)
- Documentos Psicológicos (laudos, relatórios, prontuários, anamnese, declarações...)

Pensei em ter um página somente para os documentos.

Ja temos no sidebar do dashboard um botão para documentos. 

Essa página de documentos conteria todos os tipos de documentos (separados por tipos: contratos e docs. psicologicos) 

eles seriam basicamente templates, que receberiam as informações dos pacientes cadastrados e do próprio user (contratos e documentos psicologicos geralmente contem dados das duas partes).

Pois daí, no perfil do paciente, teríamos seu contrato e seus documentos psicológicos. Que poderiam ser baixados como pdf, editados e etc. 

E os contratos poderia também sofrerem edições e serem enviados para os pacientes por email para assinatura. (mas acho que isso seria outra etapa devido a complexidade, se eu não estiver enganado).


1. Página de Documentos (Templates)
   - Divisão em duas categorias principais:
     - Contratos
     - Documentos Psicológicos (laudos, relatórios, prontuários, etc)
   - Cada categoria teria seus templates base
   - Os templates teriam campos dinâmicos para receber dados do:
     - Psicólogo (user)
     - Paciente

2. Integração com Perfil do Paciente
   - Seção de documentos no perfil
   - Possibilidade de gerar documentos a partir dos templates
   - Histórico de documentos gerados
   - Funcionalidades de:
     - Download em PDF
     - Edição
     - Visualização

3. Funcionalidades Futuras (próximas etapas)
   - Envio por email
   - Sistema de assinatura digital
   - Gestão de status dos documentos (enviado, assinado, etc)

---

Sim, exatamente! Vou fornecer um panorama detalhado de todo o sistema de documentos:

### 1. Estrutura de Dados

**Templates de Documentos** (`document_templates`):
```sql
- id
- psicologo_id (FK users)
- nome
- descricao
- tipo (contrato/psicologico)
- subtipo (laudo/relatorio/prontuario/etc)
- conteudo (HTML com placeholders)
- created_at
- updated_at
```

**Documentos Gerados** (`documents`):
```sql
- id
- template_id (FK document_templates)
- psicologo_id (FK users)
- paciente_id (FK patients)
- nome
- conteudo (HTML preenchido)
- status (rascunho/finalizado)
- created_at
- updated_at
```

### 2. Fluxos de Uso

**A. Gestão de Templates:**
1. Sistema carrega templates padrão da pasta `files`
2. Psicólogo pode:
   - Usar templates padrão
   - Criar novos templates
   - Personalizar templates existentes

**B. Geração de Documentos:**
1. No perfil do paciente:
   - Botão "Novo Documento"
   - Lista de documentos existentes
2. Ao gerar documento:
   - Seleciona template
   - Preenche dados específicos
   - Salva como rascunho ou finaliza
3. Documentos finalizados ficam no histórico do paciente

### 3. Interfaces

**A. Página de Documentos (Geral):**
- Lista de templates disponíveis
- Documentos recentes
- Filtros por tipo/status

**B. Perfil do Paciente:**
- Aba/seção de documentos
- Lista de documentos do paciente
- Botão para gerar novo documento

**C. Editor de Documentos:**
- Interface para edição
- Preview em tempo real
- Opções de salvar/finalizar

### 4. Funcionalidades Principais

1. **Templates:**
   - CRUD de templates
   - Importação de templates padrão
   - Personalização de conteúdo

2. **Documentos:**
   - Geração a partir de templates
   - Preenchimento automático de dados do paciente
   - Versionamento (rascunho/final)
   - Exportação (PDF)

3. **Integração:**
   - Vinculação com pacientes
   - Histórico por paciente
   - Busca e filtros

### 5. Considerações Técnicas

1. **Segurança:**
   - Acesso restrito por psicólogo
   - Proteção de dados sensíveis
   - Backup de documentos

2. **Performance:**
   - Cache de templates
   - Otimização de busca
   - Paginação de resultados

3. **UX:**
   - Interface intuitiva
   - Feedback claro
   - Confirmações importantes

### 6. Próximos Passos Sugeridos

1. Criar as tabelas no banco de dados
2. Implementar CRUD básico de templates
3. Desenvolver interface de listagem
4. Implementar geração de documentos
5. Integrar com perfil do paciente
6. Adicionar exportação PDF
7. Implementar busca e filtros

### Perguntas para Discussão:

1. Como você quer lidar com a personalização de templates? Total liberdade ou campos predefinidos?
2. Precisamos de um sistema de aprovação/revisão de documentos?
3. Que formato de exportação é prioritário? (PDF, DOC, etc)
4. Como você quer implementar o versionamento dos documentos?
5. Que tipos de dados do paciente devem ser preenchidos automaticamente?

Podemos começar por qualquer parte que você considere mais prioritária. Qual aspecto você gostaria de abordar primeiro?


"Então. 

O usuário do sistema (psicologo) sem a parte de documentos. A principio é lá que le pode criar e editar os modelos de contratos e documentos psicológicos. 

Mas minha dúvida é: como é que o usuário vai editar? Ele nao pode editar em html puro. Entende?  Ele não pode editar em markdown. 

Existe alguma forma de disponibilizar na web um editor? Existe algo para isso? Seria relativamente simples? 
"

https://summernote.org/getting-started/

