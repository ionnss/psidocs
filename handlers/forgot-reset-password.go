package handlers

import (
	"fmt"
	"log"
	"net/http"
	"psidocs/db"
	"text/template"
	"time"

	"github.com/gorilla/mux"
)

// ForgotPasswordHandler processa solicitações de recuperação de senha
func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Renderizar o formulário de recuperação
		tmpl, err := template.ParseFiles("templates/view/forgot_password.html")
		if err != nil {
			log.Printf("Erro ao carregar template: %v", err)
			http.Error(w, "Erro ao carregar página", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == "POST" {
		// Obter dados do formulário
		crp := SanitizeInput(r.FormValue("crp"))
		email := SanitizeInput(r.FormValue("email"))

		// Validar inputs
		if err := ValidateCRP(crp); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := ValidateEmail(email); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Conectar ao banco
		db, err := db.Connect()
		if err != nil {
			log.Printf("Erro ao conectar ao banco: %v", err)
			http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		// Verificar se usuário existe
		var userID int
		var hashCrpArmazenado, saltCrp string

		// Primeiro, buscar o usuário pelo email e obter o salt e hash do CRP
		err = db.QueryRow(`
			SELECT id, hash_crp, salt_crp 
			FROM users 
			WHERE email = $1
		`, email).Scan(&userID, &hashCrpArmazenado, &saltCrp)

		if err != nil {
			// Não informamos o erro específico por segurança
			log.Printf("Usuário não encontrado para email %s: %v", email, err)
			http.Error(w, "Se os dados estiverem corretos, você receberá um email com instruções", http.StatusOK)
			return
		}

		// Verificar se o CRP está correto
		if !checkCRPHash(crp, hashCrpArmazenado, saltCrp) {
			log.Printf("CRP incorreto para usuário %d", userID)
			http.Error(w, "Se os dados estiverem corretos, você receberá um email com instruções", http.StatusOK)
			return
		}

		// Gerar token
		token, err := GenerateToken()
		if err != nil {
			log.Printf("Erro ao gerar token: %v", err)
			http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
			return
		}
		log.Printf("Token gerado com sucesso: %s", token)

		// Salvar token no banco
		expiresAt := time.Now().Add(15 * time.Minute)
		_, err = db.Exec(`
			INSERT INTO password_resets (user_id, token, expires_at)
			VALUES ($1, $2, $3)
		`, userID, token, expiresAt)
		if err != nil {
			log.Printf("Erro ao salvar token: %v", err)
			http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
			return
		}
		log.Printf("Token salvo no banco para usuário %d", userID)

		// Preparar dados do email
		data := map[string]string{
			"Title":      "Recuperação de Chave de Acesso",
			"Message":    "Você solicitou a recuperação da sua chave de acesso. Clique no botão abaixo para criar uma nova chave. Este link é válido por 15 minutos.",
			"ButtonText": "Criar Nova Chave",
			"ButtonLink": fmt.Sprintf("https://localhost/reset-password/%s", token),
		}

		// Enviar email
		log.Printf("Tentando enviar email para %s", email)
		err = SendEmail(email, "Recuperação de Chave de Acesso - PSIDOCS", data)
		if err != nil {
			log.Printf("Erro ao enviar email: %v", err)
			http.Error(w, "Erro ao enviar email de recuperação", http.StatusInternalServerError)
			return
		}
		log.Printf("Email enviado com sucesso para %s", email)

		// Resposta de sucesso
		w.Write([]byte("Se os dados estiverem corretos, você receberá um email com instruções"))
	}
}

// ResetPasswordHandler processa solicitações de redefinição de senha
func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	// Obter token da URL
	vars := mux.Vars(r)
	token := vars["token"]

	// Conectar ao banco
	db, err := db.Connect()
	if err != nil {
		log.Printf("Erro ao conectar ao banco: %v", err)
		http.Error(w, "Erro interno do servidor", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	if r.Method == "GET" {
		// Verificar se o token é válido
		var userID int
		var expiresAt time.Time
		var used bool
		err = db.QueryRow(`
			SELECT user_id, expires_at, used 
			FROM password_resets 
			WHERE token = $1
		`, token).Scan(&userID, &expiresAt, &used)

		if err != nil {
			log.Printf("Token não encontrado: %v", err)
			http.Error(w, "Link de recuperação inválido", http.StatusBadRequest)
			return
		}

		// Verificar se o token já foi usado
		if used {
			http.Error(w, "Este link já foi utilizado", http.StatusBadRequest)
			return
		}

		// Verificar se o token expirou
		if time.Now().After(expiresAt) {
			http.Error(w, "Este link expirou", http.StatusBadRequest)
			return
		}

		// Renderizar formulário para nova senha
		tmpl, err := template.ParseFiles("templates/view/reset_password.html")
		if err != nil {
			log.Printf("Erro ao carregar template: %v", err)
			http.Error(w, "Erro ao carregar página", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == "POST" {
		// Verificar token novamente
		var userID int
		var expiresAt time.Time
		var used bool
		err = db.QueryRow(`
			SELECT user_id, expires_at, used 
			FROM password_resets 
			WHERE token = $1
		`, token).Scan(&userID, &expiresAt, &used)

		if err != nil {
			log.Printf("Token não encontrado: %v", err)
			http.Error(w, "Link de recuperação inválido", http.StatusBadRequest)
			return
		}

		if used {
			http.Error(w, "Este link já foi utilizado", http.StatusBadRequest)
			return
		}

		if time.Now().After(expiresAt) {
			http.Error(w, "Este link expirou", http.StatusBadRequest)
			return
		}

		// Obter e validar nova senha
		novaChave := SanitizeInput(r.FormValue("chave"))
		confirmarChave := SanitizeInput(r.FormValue("confirmar_chave"))

		if novaChave != confirmarChave {
			http.Error(w, "As chaves não coincidem", http.StatusBadRequest)
			return
		}

		if err := ValidateChave(novaChave); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Gerar novo salt e hash para a senha
		saltChave := generateSalt()
		hashChave := hashPassword(novaChave, saltChave)

		// Atualizar senha do usuário
		_, err = db.Exec(`
			UPDATE users 
			SET hash_chave = $1, salt_chave = $2 
			WHERE id = $3
		`, hashChave, saltChave, userID)
		if err != nil {
			log.Printf("Erro ao atualizar senha: %v", err)
			http.Error(w, "Erro ao atualizar senha", http.StatusInternalServerError)
			return
		}

		// Marcar token como usado
		_, err = db.Exec(`
			UPDATE password_resets 
			SET used = true 
			WHERE token = $1
		`, token)
		if err != nil {
			log.Printf("Erro ao marcar token como usado: %v", err)
		}

		// Redirecionar para login com mensagem de sucesso
		http.Redirect(w, r, "/?msg=senha-alterada", http.StatusSeeOther)
	}
}
