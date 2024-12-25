// Package handlers é um pacote que contém os handlers da aplicação
//
// Fornece:
// - Autenticação de usuários (psicólogos)
// - Criação de usuários (psicólogos)
// - Criação de senhas (psicólogos)
// - Login de usuários (psicólogos)
// - Logout de usuários (psicólogos)
// - Criação de senhas (psicólogos)
package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
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

	// Preparar statement para select
	stmt, err := db.Prepare("SELECT hash_chave, salt_chave FROM users WHERE hash_crp = $1")
	if err != nil {
		http.Error(w, "Erro ao preparar consulta", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	// Executar a consulta
	var hashChave, saltChave string
	err = stmt.QueryRow(crp).Scan(&hashChave, &saltChave)

	if err == sql.ErrNoRows {
		// Verificar se o email já existe
		var emailExists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&emailExists)
		if err != nil {
			http.Error(w, "Erro ao verificar email", http.StatusInternalServerError)
			return
		}
		if emailExists {
			incrementLoginAttempt(ip) // Incrementa tentativa em caso de erro
			http.Error(w, "Email já cadastrado", http.StatusConflict)
			return
		}

		// Preparar statement para insert
		insertStmt, err := db.Prepare("INSERT INTO users (hash_crp, hash_chave, salt_chave, salt_crp, email) VALUES ($1, $2, $3, $4, $5)")
		if err != nil {
			http.Error(w, "Erro ao preparar inserção", http.StatusInternalServerError)
			return
		}
		defer insertStmt.Close()

		// Registra novo usuário
		saltChave = generateSalt()
		saltCrp := generateSalt()
		hashChave = hashPassword(chave, saltChave)

		// Insere o novo usuário no banco de dados
		_, err = insertStmt.Exec(crp, hashChave, saltChave, saltCrp, email)
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

	} else if err != nil {
		http.Error(w, "Database error query", http.StatusInternalServerError)
		return
	} else {
		// Login de usuário
		if checkPasswordHash(chave, hashChave, saltChave) {
			// Buscar email do usuário
			var userEmail string
			err = db.QueryRow("SELECT email FROM users WHERE hash_crp = $1", crp).Scan(&userEmail)
			if err != nil {
				http.Error(w, "Erro ao buscar dados do usuário", http.StatusInternalServerError)
				return
			}

			// Set session values
			session.Values["authenticated"] = true
			session.Values["crp"] = crp
			session.Values["email"] = userEmail
			session.Save(r, w)

			// Adiciona headers no-cache
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")

			// Redireciona para o dashboard
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		} else {
			incrementLoginAttempt(ip) // Incrementa tentativa em caso de senha incorreta
			http.Error(w, "Senha incorreta", http.StatusUnauthorized)
		}
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
// - hash: o hash da senha
func hashPassword(password, salt string) string {
	saltedPassword := password + salt
	hash, err := bcrypt.GenerateFromPassword([]byte(saltedPassword), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
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

	// Combina senha+salt
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
	session, _ := Store.Get(r, "psidocs-session")

	// Preparar dados do template
	data := map[string]interface{}{
		"Authenticated": session.Values["authenticated"],
		"Email":         session.Values["email"],
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
