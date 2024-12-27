// Package handlers é um pacote que contém os handlers da aplicação
package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"psidocs/db"
	"sync"
	"text/template"
	"time"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

// Store é a sessão do usuário
var Store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))

// LoginAttempt armazena informações sobre tentativas de login
type LoginAttempt struct {
	Count     int       // Número de tentativas
	FirstTry  time.Time // Primeira tentativa
	BlockedAt time.Time // Quando foi bloqueado
}

const (
	MaxLoginAttempts = 5                // Máximo de tentativas
	BlockDuration    = 15 * time.Minute // Duração do bloqueio
)

var (
	loginAttempts = make(map[string]*LoginAttempt) // Mapa de tentativas por IP
	loginMutex    sync.RWMutex                     // Mutex para acesso concorrente
)

// checkRateLimit verifica se o IP está bloqueado
func checkRateLimit(ip string) bool {
	loginMutex.Lock()
	defer loginMutex.Unlock()

	now := time.Now()
	attempt, exists := loginAttempts[ip]

	// Se não existe, cria nova entrada
	if !exists {
		loginAttempts[ip] = &LoginAttempt{
			Count:    0,
			FirstTry: now,
		}
		return true
	}

	// Se está bloqueado, verifica se já pode tentar novamente
	if !attempt.BlockedAt.IsZero() {
		if now.Sub(attempt.BlockedAt) < BlockDuration {
			return false // Ainda bloqueado
		}
		// Reseta as tentativas após o período de bloqueio
		attempt.Count = 0
		attempt.FirstTry = now
		attempt.BlockedAt = time.Time{}
		return true
	}

	// Se passou 24h desde a primeira tentativa, reseta o contador
	if now.Sub(attempt.FirstTry) > 24*time.Hour {
		attempt.Count = 0
		attempt.FirstTry = now
		return true
	}

	// Se excedeu o limite de tentativas, bloqueia
	if attempt.Count >= MaxLoginAttempts {
		attempt.BlockedAt = now
		return false
	}

	return true
}

// incrementLoginAttempt incrementa o contador de tentativas
func incrementLoginAttempt(ip string) {
	loginMutex.Lock()
	defer loginMutex.Unlock()

	if attempt, exists := loginAttempts[ip]; exists {
		attempt.Count++
	}
}

// checkCRPHash verifica se o CRP está correto
//
// Recebe:
// - crp: o CRP a ser verificado
// - hashBase64: o hash do CRP armazenado no banco em base64
// - salt: o salt do CRP armazenado no banco de dados
//
// Retorna:
// - true: se o CRP está correto
// - false: se o CRP está incorreto
func checkCRPHash(crp, hashBase64, salt string) bool {
	// Decodifica o hash de base64 para bytes
	hashBytes, err := base64.StdEncoding.DecodeString(hashBase64)
	if err != nil {
		return false
	}

	// Combina o CRP com o salt da mesma forma que foi feito na criação
	saltedCRP := crp + salt

	// Compara usando bcrypt
	err = bcrypt.CompareHashAndPassword(hashBytes, []byte(saltedCRP))
	return err == nil
}

// AuthHandler é um handler que fará a autenticação de usuários
//
// Recebe:
// - w: o writer do response
// - r: o request
func AuthHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica rate limit
	ip := r.RemoteAddr
	if !checkRateLimit(ip) {
		http.Error(w, "Muitas tentativas de login. Tente novamente em 15 minutos.", http.StatusTooManyRequests)
		return
	}

	// Get session
	session, err := Store.Get(r, "psidocs-session")
	if err != nil {
		http.Error(w, "Erro ao obter sessão", http.StatusInternalServerError)
		return
	}

	// Variáveis de entrada (sanitizadas)
	crp := SanitizeInput(r.FormValue("crp"))
	chave := SanitizeInput(r.FormValue("chave"))
	email := SanitizeInput(r.FormValue("email"))

	// Validação dos inputs
	if err := ValidateCRP(crp); err != nil {
		incrementLoginAttempt(ip) // Incrementa tentativa em caso de erro
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := ValidateEmail(email); err != nil {
		incrementLoginAttempt(ip)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := ValidateChave(chave); err != nil {
		incrementLoginAttempt(ip)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Conectar ao banco de dados
	db, err := db.Connect()
	if err != nil {
		http.Error(w, "Erro ao conectar ao banco de dados", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Primeiro, verificar se o usuário existe
	var userExists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&userExists)
	if err != nil {
		http.Error(w, "Erro ao verificar usuário", http.StatusInternalServerError)
		return
	}

	if userExists {
		// Usuário existe - FLUXO DE LOGIN
		log.Printf("Iniciando fluxo de login para email: %s", email)

		// Buscar todos os dados necessários de uma vez
		var hashCrpArmazenado, saltCrp, hashChave, saltChave string
		err = db.QueryRow(`
			SELECT hash_crp, salt_crp, hash_chave, salt_chave 
			FROM users 
			WHERE email = $1
		`, email).Scan(&hashCrpArmazenado, &saltCrp, &hashChave, &saltChave)

		if err != nil {
			log.Printf("Erro ao buscar dados do usuário: %v", err)
			http.Error(w, "Erro ao verificar credenciais", http.StatusInternalServerError)
			return
		}

		log.Printf("Debug - Verificação do CRP:")
		log.Printf("- CRP fornecido: %s", crp)
		log.Printf("- Salt CRP: %s", saltCrp)

		// Verificar CRP usando bcrypt
		if !checkCRPHash(crp, hashCrpArmazenado, saltCrp) {
			log.Printf("CRP incorreto para usuário com email %s", email)
			incrementLoginAttempt(ip)
			http.Error(w, "CRP ou chave incorretos", http.StatusUnauthorized)
			return
		}

		log.Printf("CRP verificado com sucesso, verificando chave")

		// Verificar a chave
		if !checkPasswordHash(chave, hashChave, saltChave) {
			log.Printf("Chave incorreta para usuário com email %s", email)
			incrementLoginAttempt(ip)
			http.Error(w, "CRP ou chave incorretos", http.StatusUnauthorized)
			return
		}

		log.Printf("Login bem sucedido para usuário com email %s", email)

		// Login bem sucedido
		session.Values["authenticated"] = true
		session.Values["crp"] = crp
		session.Values["email"] = email
		session.Save(r, w)

		// Adiciona headers no-cache
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		// Redireciona para o dashboard
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)

	} else {
		// Usuário não existe - FLUXO DE REGISTRO
		// Gerar salts
		saltChave := generateSalt()
		saltCrp := generateSalt()

		// Gerar hashes
		hashChave := hashPassword(chave, saltChave)
		hashCrp := hashPassword(crp, saltCrp)

		// Inserir novo usuário
		_, err = db.Exec("INSERT INTO users (hash_crp, hash_chave, salt_chave, salt_crp, email) VALUES ($1, $2, $3, $4, $5)",
			hashCrp, hashChave, saltChave, saltCrp, email)
		if err != nil {
			http.Error(w, "Erro ao registrar novo usuário", http.StatusInternalServerError)
			return
		}

		// Set session values
		session.Values["authenticated"] = true
		session.Values["crp"] = crp
		session.Values["email"] = email
		session.Save(r, w)

		// Adiciona headers no-cache
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		// Redireciona para o dashboard
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

// LogoutHandler handles user logout
//
// Recebe:
// - w: o writer do response
// - r: o request
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "psidocs-session")
	if err != nil {
		http.Error(w, "Erro ao obter sessão", http.StatusInternalServerError)
		return
	}

	// Revoga a autenticação do usuário
	session.Values["authenticated"] = false
	session.Values["crp"] = ""
	session.Values["email"] = ""
	session.Options.MaxAge = -1
	session.Save(r, w)

	// Redirecionar para a página inicial
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// generateSalt gera um salt aleatório para a senha
//
// Retorna:
// - salt: o salt da senha
func generateSalt() string {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(salt)
}

// hashPassword gera um hash para a senha
//
// Recebe:
// - password: a senha a ser hasheada
// - salt: o salt da senha
//
// Retorna:
// - hash: o hash da senha em base64
func hashPassword(password, salt string) string {
	// Combina a senha com o salt
	saltedPassword := password + salt

	// Gera o hash usando bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(saltedPassword), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	// Retorna o hash em base64
	return base64.StdEncoding.EncodeToString(hash)
}

// checkPasswordHash verifica se a senha está correta
//
// Recebe:
// - password: a senha a ser verificada
// - hashBase64: o hash da senha armazenado no banco em base64
// - salt: o salt da senha armazenado no banco de dados
//
// Retorna:
// - true: se a senha está correta
// - false: se a senha está incorreta
func checkPasswordHash(password, hashBase64, salt string) bool {
	// Decodifica o hash de base64 para bytes
	hashBytes, err := base64.StdEncoding.DecodeString(hashBase64)
	if err != nil {
		return false
	}

	// Combina a senha com o salt da mesma forma que foi feito na criação
	saltedPassword := password + salt

	// Compara usando bcrypt
	err = bcrypt.CompareHashAndPassword(hashBytes, []byte(saltedPassword))
	return err == nil
}

// AuthMiddleware é um middleware que verifica se o usuário está autenticado
//
// Recebe:
// - next: o próximo handler a ser executado
//
// Retorna:
// - http.Handler: o handler autenticado
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, _ := Store.Get(r, "psidocs-session")
		if auth, ok := session.Values["authenticated"].(bool); ok && auth {
			// Usuário autenticado
			// Injeta dados no contexto
			next.ServeHTTP(w, r)
		} else {
			// Não autenticado
			// Redireciona ou continua dependendo da rota
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	})
}

// DashboardHandler renderiza o dashboard do usuário
//
// Recebe:
// - w: o writer do response
// - r: o request
func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "psidocs-session")
	if err != nil {
		http.Error(w, "Erro ao obter sessão", http.StatusInternalServerError)
		return
	}

	// Preparar dados do template
	data := map[string]interface{}{
		"Authenticated": session.Values["authenticated"],
		"Email":         session.Values["email"],
		"CRP":           session.Values["crp"],
	}

	// Se for uma requisição HTMX, renderiza só o conteúdo
	if r.Header.Get("HX-Request") == "true" {
		tmpl := template.Must(template.ParseFiles("templates/view/partials/dashboard_content.html"))
		tmpl.Execute(w, data)
		return
	}

	// Se não for HTMX, renderiza o layout completo
	tmpl := template.Must(template.ParseFiles(
		"templates/view/dashboard_layout.html",
		"templates/view/partials/dashboard_content.html",
	))
	tmpl.Execute(w, data)
}

// UserConfig representa os dados de configuração do usuário
type UserConfig struct {
	FirstName   string
	MiddleName  string
	LastName    string
	DateOfBirth time.Time
	CPF         string
	RG          string
	DDD         string
	Telefone    string
	WhatsApp    bool
	Endereco    string
	Numero      string
	Bairro      string
	Cidade      string
	Estado      string
	CEP         string
}

// CreateUserConfig cria a configuração do usuário (dados pessoais para contratos de pacientes e documentos psicológicos)
//
// Recebe:
// - w: o writer do response
// - r: o request
func UpdateUserConfigHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("UpdateUserConfigHandler iniciado - Método: %s", r.Method)

	// Verificar sessão
	session, err := Store.Get(r, "psidocs-session")
	if err != nil {
		log.Printf("Erro ao obter sessão: %v", err)
		http.Error(w, "Erro ao obter sessão", http.StatusInternalServerError)
		return
	}

	// Verificar autenticação
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		log.Printf("Usuário não autenticado - ok: %v, auth: %v", ok, auth)
		http.Error(w, "Usuário não autenticado", http.StatusUnauthorized)
		return
	}

	// Conectar ao banco
	db, err := db.Connect()
	if err != nil {
		log.Printf("Erro ao conectar ao banco: %v", err)
		http.Error(w, "Erro ao conectar ao banco de dados", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Obter ID do usuário
	var userID int
	email := session.Values["email"]
	log.Printf("Buscando ID do usuário para email: %v", email)
	err = db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&userID)
	if err != nil {
		log.Printf("Erro ao obter ID do usuário: %v", err)
		http.Error(w, "Erro ao obter ID do usuário", http.StatusInternalServerError)
		return
	}
	log.Printf("ID do usuário encontrado: %d", userID)

	if r.Method == "GET" {
		log.Printf("Processando requisição GET")
		// Buscar dados existentes
		var config UserConfig
		err = db.QueryRow(`
			SELECT first_name, middle_name, last_name, 
				   cpf, rg, date_of_birth,
				   ddd, telefone, whatsapp,
				   endereco, numero, bairro,
				   cidade, estado, cep
			FROM users_data
			WHERE user_id = $1
		`, userID).Scan(
			&config.FirstName, &config.MiddleName, &config.LastName,
			&config.CPF, &config.RG, &config.DateOfBirth,
			&config.DDD, &config.Telefone, &config.WhatsApp,
			&config.Endereco, &config.Numero, &config.Bairro,
			&config.Cidade, &config.Estado, &config.CEP,
		)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Erro ao buscar dados: %v", err)
			http.Error(w, "Erro ao buscar dados", http.StatusInternalServerError)
			return
		}
		log.Printf("Dados encontrados: %+v", config)

		// Se for uma requisição HTMX, renderiza só o conteúdo
		if r.Header.Get("HX-Request") == "true" {
			log.Printf("Renderizando template para requisição HTMX")
			tmpl := template.Must(template.ParseFiles("templates/view/partials/user_data.html"))
			if err := tmpl.Execute(w, config); err != nil {
				log.Printf("Erro ao renderizar template: %v", err)
				http.Error(w, "Erro ao renderizar template", http.StatusInternalServerError)
				return
			}
			return
		}

		// Se não for HTMX, renderiza o layout completo
		log.Printf("Renderizando layout completo")
		tmpl := template.Must(template.ParseFiles(
			"templates/view/dashboard_layout.html",
			"templates/view/partials/user_data.html",
		))
		if err := tmpl.Execute(w, config); err != nil {
			log.Printf("Erro ao renderizar template: %v", err)
			http.Error(w, "Erro ao renderizar template", http.StatusInternalServerError)
			return
		}
		log.Printf("Template renderizado com sucesso")
		return

	} else if r.Method == "POST" {
		// Processar formulário
		whatsappValue := r.FormValue("whatsapp") == "on"
		dateOfBirth, err := time.Parse("2006-01-02", r.FormValue("date_of_birth"))
		if err != nil {
			http.Error(w, "Data de nascimento inválida", http.StatusBadRequest)
			return
		}

		// Verificar se já existe configuração
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users_data WHERE user_id = $1)", userID).Scan(&exists)
		if err != nil {
			http.Error(w, "Erro ao verificar configuração existente", http.StatusInternalServerError)
			return
		}

		if exists {
			// Update
			_, err = db.Exec(`
				UPDATE users_data 
				SET first_name = $1, middle_name = $2, last_name = $3, 
					cpf = $4, rg = $5, date_of_birth = $6,
					ddd = $7, telefone = $8, whatsapp = $9,
					endereco = $10, numero = $11, bairro = $12,
					cidade = $13, estado = $14, cep = $15
				WHERE user_id = $16`,
				r.FormValue("first_name"),
				r.FormValue("middle_name"),
				r.FormValue("last_name"),
				r.FormValue("cpf"),
				r.FormValue("rg"),
				dateOfBirth,
				r.FormValue("ddd"),
				r.FormValue("telefone"),
				whatsappValue,
				r.FormValue("endereco"),
				r.FormValue("numero"),
				r.FormValue("bairro"),
				r.FormValue("cidade"),
				r.FormValue("estado"),
				r.FormValue("cep"),
				userID,
			)
		} else {
			// Insert
			_, err = db.Exec(`
				INSERT INTO users_data (
					user_id, first_name, middle_name, last_name,
					cpf, rg, date_of_birth,
					ddd, telefone, whatsapp,
					endereco, numero, bairro,
					cidade, estado, cep, plan
				) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`,
				userID,
				r.FormValue("first_name"),
				r.FormValue("middle_name"),
				r.FormValue("last_name"),
				r.FormValue("cpf"),
				r.FormValue("rg"),
				dateOfBirth,
				r.FormValue("ddd"),
				r.FormValue("telefone"),
				whatsappValue,
				r.FormValue("endereco"),
				r.FormValue("numero"),
				r.FormValue("bairro"),
				r.FormValue("cidade"),
				r.FormValue("estado"),
				r.FormValue("cep"),
				"free", // Plano inicial gratuito
			)
		}

		if err != nil {
			http.Error(w, "Erro ao salvar configurações", http.StatusInternalServerError)
			return
		}

		// Resposta baseada no tipo de requisição
		if r.Header.Get("HX-Request") == "true" {
			w.Write([]byte(`<div class="alert alert-success alert-dismissible fade show" role="alert">
				<i class="bi bi-check-circle-fill me-2"></i>
				Configurações salvas com sucesso!
				<button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Fechar"></button>
			</div>`))
		} else {
			http.Redirect(w, r, "/dashboard/dados_pessoais", http.StatusSeeOther)
		}
	}
}

// UpdateUserCredentialsHandler lida com a exibição e atualização das credenciais do usuário
//
// Recebe:
// - w: o writer do response
// - r: o request
func UpdateUserCredentialsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("UpdateUserCredentialsHandler iniciado - Método: %s", r.Method)

	// Verificar sessão
	session, err := Store.Get(r, "psidocs-session")
	if err != nil {
		log.Printf("Erro ao obter sessão: %v", err)
		http.Error(w, "Erro ao obter sessão", http.StatusInternalServerError)
		return
	}

	// Verificar autenticação
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		log.Printf("Usuário não autenticado - ok: %v, auth: %v", ok, auth)
		http.Error(w, "Usuário não autenticado", http.StatusUnauthorized)
		return
	}

	if r.Method == "GET" {
		log.Printf("Processando requisição GET")
		// Preparar dados do template
		data := map[string]interface{}{
			"Email": session.Values["email"],
			"CRP":   session.Values["crp"],
		}

		// Se for uma requisição HTMX, renderiza só o conteúdo
		if r.Header.Get("HX-Request") == "true" {
			log.Printf("Renderizando template para requisição HTMX")
			tmpl := template.Must(template.ParseFiles("templates/view/partials/user_credentials.html"))
			if err := tmpl.Execute(w, data); err != nil {
				log.Printf("Erro ao renderizar template: %v", err)
				http.Error(w, "Erro ao renderizar template", http.StatusInternalServerError)
				return
			}
			return
		}

		// Se não for HTMX, renderiza o layout completo
		log.Printf("Renderizando layout completo")
		tmpl := template.Must(template.ParseFiles(
			"templates/view/dashboard_layout.html",
			"templates/view/partials/user_credentials.html",
		))
		if err := tmpl.Execute(w, data); err != nil {
			log.Printf("Erro ao renderizar template: %v", err)
			http.Error(w, "Erro ao renderizar template", http.StatusInternalServerError)
			return
		}
		return
	}

	if r.Method == "POST" {
		// Conectar ao banco
		db, err := db.Connect()
		if err != nil {
			http.Error(w, "Erro ao conectar ao banco de dados", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		// Obter ID do usuário atual
		var userID int
		email := session.Values["email"].(string)
		err = db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&userID)
		if err != nil {
			http.Error(w, "Erro ao obter dados do usuário", http.StatusInternalServerError)
			return
		}

		// Verificar qual campo está sendo atualizado
		newEmail := r.FormValue("email")
		newCRP := r.FormValue("crp")
		newChave := r.FormValue("chave")
		chaveAtual := r.FormValue("chave_atual")

		// Atualizar email
		if newEmail != "" && newEmail != email {
			if err := ValidateEmail(newEmail); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// Verificar se novo email já existe
			var exists bool
			err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND id != $2)", newEmail, userID).Scan(&exists)
			if err != nil {
				http.Error(w, "Erro ao verificar email", http.StatusInternalServerError)
				return
			}
			if exists {
				http.Error(w, "Email já está em uso", http.StatusConflict)
				return
			}

			// Atualizar email
			_, err = db.Exec("UPDATE users SET email = $1 WHERE id = $2", newEmail, userID)
			if err != nil {
				http.Error(w, "Erro ao atualizar email", http.StatusInternalServerError)
				return
			}

			// Atualizar sessão
			session.Values["email"] = newEmail
			session.Save(r, w)
		}

		// Atualizar CRP
		if newCRP != "" && newCRP != session.Values["crp"] {
			if err := ValidateCRP(newCRP); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// Gerar novo salt e hash para o CRP
			saltCrp := generateSalt()
			hashCrp := hashPassword(newCRP, saltCrp)

			// Verificar se novo CRP já existe
			var exists bool
			err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE hash_crp = $1 AND id != $2)", hashCrp, userID).Scan(&exists)
			if err != nil {
				http.Error(w, "Erro ao verificar CRP", http.StatusInternalServerError)
				return
			}
			if exists {
				http.Error(w, "CRP já está em uso", http.StatusConflict)
				return
			}

			// Atualizar CRP
			_, err = db.Exec("UPDATE users SET hash_crp = $1, salt_crp = $2 WHERE id = $3", hashCrp, saltCrp, userID)
			if err != nil {
				http.Error(w, "Erro ao atualizar CRP", http.StatusInternalServerError)
				return
			}

			// Atualizar sessão
			session.Values["crp"] = newCRP
			session.Save(r, w)
		}

		// Atualizar chave
		if newChave != "" {
			log.Printf("Tentando atualizar chave para usuário %d", userID)

			if err := ValidateChave(newChave); err != nil {
				log.Printf("Erro na validação da nova chave: %v", err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// Verificar chave atual
			var hashChaveAtual, saltChaveAtual string
			err = db.QueryRow("SELECT hash_chave, salt_chave FROM users WHERE id = $1", userID).Scan(&hashChaveAtual, &saltChaveAtual)
			if err != nil {
				log.Printf("Erro ao buscar chave atual: %v", err)
				http.Error(w, "Erro ao verificar chave atual", http.StatusInternalServerError)
				return
			}

			log.Printf("Debug - Valores recebidos do formulário:")
			log.Printf("- chave_atual: %v (vazio: %v)", chaveAtual, chaveAtual == "")
			log.Printf("- nova_chave: %v (vazio: %v)", newChave, newChave == "")

			log.Printf("Debug - Valores do banco:")
			log.Printf("- Salt atual: %s", saltChaveAtual)
			log.Printf("- Hash atual: %s", hashChaveAtual)

			// Gerar hash da chave atual fornecida para comparação
			hashTeste := hashPassword(chaveAtual, saltChaveAtual)
			log.Printf("Debug - Hash gerado com a chave atual fornecida: %s", hashTeste)
			log.Printf("Debug - Os hashes são iguais? %v", hashTeste == hashChaveAtual)

			// Verificar se a chave atual está correta
			resultado := checkPasswordHash(chaveAtual, hashChaveAtual, saltChaveAtual)
			log.Printf("Debug - Resultado do checkPasswordHash: %v", resultado)

			if !resultado {
				log.Printf("Chave atual incorreta para usuário %d", userID)
				http.Error(w, "Chave atual incorreta", http.StatusUnauthorized)
				return
			}

			// Se chegou aqui, a chave atual está correta
			log.Printf("Chave atual verificada com sucesso, procedendo com atualização")

			// Gerar novo salt e hash para a nova chave
			saltChave := generateSalt()
			hashChave := hashPassword(newChave, saltChave)

			// Atualizar chave
			_, err = db.Exec("UPDATE users SET hash_chave = $1, salt_chave = $2 WHERE id = $3",
				hashChave, saltChave, userID)
			if err != nil {
				log.Printf("Erro ao atualizar chave: %v", err)
				http.Error(w, "Erro ao atualizar chave", http.StatusInternalServerError)
				return
			}

			log.Printf("Chave atualizada com sucesso para usuário %d", userID)
		}

		// Resposta baseada no tipo de requisição
		if r.Header.Get("HX-Request") == "true" {
			w.Write([]byte(`<div class="alert alert-success alert-dismissible fade show" role="alert">
				<i class="bi bi-check-circle-fill me-2"></i>
				Credenciais atualizadas com sucesso!
				<button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Fechar"></button>
			</div>`))
		} else {
			http.Redirect(w, r, "/dashboard/credenciais", http.StatusSeeOther)
		}
	}
}

// GetCurrentUserInfo obtém o email e CRP do usuário atualmente autenticado
//
// Receives:
// - w: http.ResponseWriter para retornar erros HTTP se necessário
// - r: *http.Request para acessar a sessão
//
// Returns:
// - email: email do usuário autenticado
// - crp: CRP do usuário autenticado
// - error: erro se houver falha ao obter os dados ou se o usuário não estiver autenticado
func GetCurrentUserInfo(w http.ResponseWriter, r *http.Request) (string, string, error) {
	// Verificar sessão
	session, err := Store.Get(r, "psidocs-session")
	if err != nil {
		log.Printf("Erro ao obter sessão: %v", err)
		http.Error(w, "Erro ao obter sessão", http.StatusInternalServerError)
		return "", "", fmt.Errorf("erro ao obter sessão: %v", err)
	}

	// Verificar autenticação
	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		log.Printf("Usuário não autenticado - ok: %v, auth: %v", ok, auth)
		http.Error(w, "Usuário não autenticado", http.StatusUnauthorized)
		return "", "", fmt.Errorf("usuário não autenticado")
	}

	email, emailOk := session.Values["email"].(string)
	crp, crpOk := session.Values["crp"].(string)

	if !emailOk || !crpOk {
		return "", "", fmt.Errorf("email ou CRP não encontrados na sessão")
	}

	return email, crp, nil
}
