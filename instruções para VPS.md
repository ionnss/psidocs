# Guia Completo: Deploy do PsiDocs

## Índice
1. [Preparação](#1-preparação)
2. [Compra e Configuração da VPS](#2-compra-e-configuração-da-vps)
3. [Conexão com a VPS](#3-conexão-com-a-vps)
4. [Preparação do Servidor](#4-preparação-do-servidor)
5. [Deploy da Aplicação](#5-deploy-da-aplicação)
6. [Configuração do Domínio](#6-configuração-do-domínio)
7. [Configuração do HTTPS](#7-configuração-do-https)
8. [Manutenção](#8-manutenção)

## 1. Preparação

### 1.1 Requisitos
- Domínio comprado (psidocs.com)
- VPS da Hostinger (será comprada)
- Projeto com Docker configurado
- Computador com Mac OS

### 1.2 Programas Necessários
1. **Terminal**
   - Já vem instalado no Mac (Aplicativos > Utilitários > Terminal)
   - Usado para conectar na VPS via SSH
   
2. **FileZilla** (opcional)
   - Download: https://filezilla-project.org/
   - Usado para transferir arquivos para a VPS
   - Alternativa: usar comandos `scp` ou `rsync` no Terminal

## 2. Compra e Configuração da VPS

### 2.1 Compra
1. Acesse hostinger.com.br
2. Vá em "VPS Hosting"
3. Escolha o plano:
   - Recomendado: 2GB RAM, 1 vCPU
   - Sistema: Ubuntu (última versão LTS)

### 2.2 Após a Compra
Você receberá um email com:
- IP do servidor
- Nome de usuário (root)
- Senha inicial
- Instruções de acesso

⚠️ **IMPORTANTE**: Guarde estas informações em local seguro!

## 3. Conexão com a VPS

### 3.1 Usando Terminal
1. Abra o Terminal (Cmd + Espaço, digite "Terminal")
2. Digite o comando:
   ```bash
   ssh root@[IP-DO-SERVIDOR]
   ```
3. Na primeira conexão, aparecerá uma mensagem de fingerprint, digite "yes"
4. Digite a senha recebida por email

### 3.2 Primeiro Acesso
1. O sistema pedirá para trocar a senha
2. Digite a senha atual
3. Digite a nova senha (2x)
4. ⚠️ Use uma senha forte e guarde em local seguro!

### 3.3 Dica de Segurança (Opcional)
Para não precisar digitar a senha sempre:
```bash
# No seu Mac
ssh-keygen -t rsa -b 4096
ssh-copy-id root@[IP-DO-SERVIDOR]
```

## 4. Preparação do Servidor

### 4.1 Atualização do Sistema
```bash
# Atualizar lista de pacotes
sudo apt update

# Instalar atualizações
sudo apt upgrade -y
```

### 4.2 Instalação do Docker
```bash
# Instalar Docker
sudo apt install docker.io -y

# Instalar Docker Compose
sudo apt install docker-compose -y

# Iniciar Docker
sudo systemctl start docker

# Configurar Docker para iniciar com o sistema
sudo systemctl enable docker

# Verificar instalação
docker --version
docker-compose --version
```

## 5. Deploy da Aplicação

### 5.1 Transferência de Arquivos
**Opção 1 - Usando Terminal (Recomendada):**
```bash
# Primeiro, crie a pasta no servidor
ssh root@[IP-DO-SERVIDOR] "mkdir -p /root/psidocs"

# No seu Mac, vá para a pasta do projeto
cd /Users/ions/psidocs

# Transfira todos os arquivos
scp -r ./* root@[IP-DO-SERVIDOR]:/root/psidocs/

# Transfira também o arquivo .env (que está oculto)
scp .env root@[IP-DO-SERVIDOR]:/root/psidocs/
```

**Opção 2 - Usando FileZilla:**
1. Abra o FileZilla
2. Configure a conexão:
   - Host: sftp://[IP-DO-SERVIDOR]
   - Username: root
   - Password: [sua-senha]
   - Port: 22
3. No painel esquerdo (Local), navegue até:
   `/Users/ions/psidocs`
4. No painel direito (Remoto), navegue até:
   `/root/psidocs`
5. Selecione todos os arquivos (Ctrl+A ou Cmd+A) incluindo o .env
6. Arraste os arquivos do painel esquerdo para o direito

### 5.2 Verificação dos Arquivos
Após a transferência, conecte no servidor e verifique se tudo foi transferido corretamente:
```bash
# Conectar no servidor
ssh root@[IP-DO-SERVIDOR]

# Listar arquivos transferidos
cd /root/psidocs
ls -la

# Você deve ver:
# - docker-compose.yml
# - Dockerfile
# - .env
# - e outros arquivos do projeto
```

### 5.2 Configuração do Ambiente
```bash
# Entrar na pasta do projeto
cd /root/psidocs

# Criar arquivo de ambiente
cp .env.example .env

# Editar configurações
nano .env
```

### 5.3 Iniciar Aplicação
```bash
# Construir e iniciar containers
docker-compose up -d

# Verificar se está rodando
docker ps

# Ver logs (se necessário)
docker-compose logs -f
```

## 6. Configuração do Domínio

### 6.1 No Painel da Hostinger
1. Acesse o painel de controle
2. Vá em "Domínios" → psidocs.com
3. Procure "Gerenciar DNS" ou "Zona DNS"
4. Configure registros:
   ```
   Tipo: A
   Nome: @
   Valor: [IP-DO-SERVIDOR]
   TTL: 300

   Tipo: A
   Nome: www
   Valor: [IP-DO-SERVIDOR]
   TTL: 300
   ```

### 6.2 Verificação
- A propagação DNS pode levar até 24 horas
- Verifique em: https://www.whatsmydns.net/

## 7. Configuração do HTTPS

### 7.1 Instalação do Nginx Proxy Manager
```bash
# Criar diretório
mkdir -p /root/nginx-proxy
cd /root/nginx-proxy

# Criar docker-compose.yml
nano docker-compose.yml
```

Conteúdo do docker-compose.yml:
```yaml
version: '3'
services:
  npm:
    image: 'jc21/nginx-proxy-manager:latest'
    restart: always
    ports:
      - '80:80'
      - '443:443'
      - '81:81'
    volumes:
      - ./data:/data
      - ./letsencrypt:/etc/letsencrypt
```

```bash
# Iniciar Nginx Proxy Manager
docker-compose up -d
```

### 7.2 Configuração do Proxy
1. Acesse: http://[IP-DO-SERVIDOR]:81
2. Login inicial:
   - Email: admin@example.com
   - Senha: changeme
3. Troque a senha quando solicitado
4. Adicione novo proxy host:
   - Domain: psidocs.com
   - Forward IP: IP-DO-SERVIDOR
   - Forward Port: [PORTA-DA-APLICAÇÃO]
   - SSL: Let's Encrypt
   - Force SSL: Sim

## 8. Manutenção

### 8.1 Comandos Úteis
```bash
# Ver containers rodando
docker ps

# Ver logs
docker-compose logs -f

# Reiniciar containers
docker-compose restart

# Parar tudo
docker-compose down

# Iniciar tudo
docker-compose up -d

# Atualizar aplicação
git pull
docker-compose down
docker-compose up -d --build
```

### 8.2 Monitoramento
- Verifique logs regularmente
- Monitore uso de recursos:
  ```bash
  # Uso de CPU e memória
  htop
  
  # Uso de disco
  df -h
  ```

### 8.3 Backup
```bash
# Backup dos volumes
docker run --rm -v psidocs_db_data:/dbdata -v $(pwd):/backup ubuntu tar czf /backup/db_backup.tar.gz /dbdata
```

### 8.4 Segurança
- Mantenha o sistema atualizado
- Use senhas fortes
- Monitore logs de acesso
- Faça backups regulares

## Suporte
- Hostinger tem suporte 24/7 em português
- Para problemas com:
  - VPS/Domínio: Contate Hostinger
  - Aplicação: Verifique logs e documentação

## Troubleshooting

### Problemas Comuns

1. **Aplicação não inicia:**
   ```bash
   # Verificar logs
   docker-compose logs -f
   ```

2. **Domínio não funciona:**
   - Verifique configurações DNS
   - Aguarde propagação (até 24h)
   - Teste com ping psidocs.com

3. **Certificado SSL falha:**
   - Verifique se portas 80/443 estão livres
   - Confirme configurações DNS
   - Verifique logs do Nginx Proxy Manager