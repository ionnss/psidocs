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
	"fmt"
	"net/http"
	"psidocs/db"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

// Store é a sessão do usuário
var Store = sessions.NewCookieStore([]byte("your-secret-key-here"))

// AuthHandler é um handler que fará a autenticação de usuários
//
// Recebe:
// - w: o writer da resposta HTTP
// - r: o request HTTP
func AuthHandler(w http.ResponseWriter, r *http.Request) {
	// Get session
	session, err := Store.Get(r, "psidocs-session")
	if err != nil {
		http.Error(w, "Erro ao obter sessão", http.StatusInternalServerError)
		return
	}

	// Variáveis de entrada
	crp := r.FormValue("crp")
	chave := r.FormValue("chave")

	// Conectar ao banco de dados
	db, err := db.Connect()
	if err != nil {
		http.Error(w, "Erro ao conectar ao banco de dados", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Preparar statement para select
	//
	// Prepara uma consulta para selecionar o hash da chave e o salt do usuário
	stmt, err := db.Prepare("SELECT hash_chave, salt_chave FROM users WHERE hash_crp = $1")
	if err != nil {
		http.Error(w, "Erro ao preparar consulta", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	// Executar a consulta
	//
	// Executa a consulta para selecionar o hash da chave e o salt do usuário
	var hashChave, saltChave string
	err = stmt.QueryRow(crp).Scan(&hashChave, &saltChave)

	if err == sql.ErrNoRows {
		// Preparar statement para insert
		//
		// Prepara uma consulta para inserir um novo usuário
		insertStmt, err := db.Prepare("INSERT INTO users (hash_crp, hash_chave, salt_chave, salt_crp) VALUES ($1, $2, $3, $4)")
		if err != nil {
			http.Error(w, "Erro ao preparar inserção", http.StatusInternalServerError)
			return
		}
		defer insertStmt.Close()

		// Registra novo usuário
		//
		// Gera um salt aleatório para a senha
		saltChave = generateSalt()
		saltCrp := generateSalt() // Gera salt para o CRP
		hashChave = hashPassword(chave, saltChave)

		// Insere o novo usuário no banco de dados
		_, err = insertStmt.Exec(crp, hashChave, saltChave, saltCrp)
		if err != nil {
			http.Error(w, "Erro ao registrar novo usuário", http.StatusInternalServerError)
			return
		}

		// Set session values
		session.Values["authenticated"] = true
		session.Values["crp"] = crp
		session.Save(r, w)

		fmt.Fprintf(w, "Usuário registrado com sucesso")
	} else if err != nil {
		http.Error(w, "Database error query", http.StatusInternalServerError)
		return
	} else {
		// Login de usuário
		//
		// Verifica se a senha está correta
		if checkPasswordHash(chave, hashChave, saltChave) {
			// Set session values
			session.Values["authenticated"] = true
			session.Values["crp"] = crp
			session.Save(r, w)

			fmt.Fprintf(w, "Login realizado com sucesso")
		} else {
			http.Error(w, "Senha incorreta", http.StatusUnauthorized)
		}
	}
}

// LogoutHandler handles user logout
//
// Recebe:
// - w: o writer da resposta HTTP
// - r: o request HTTP
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := Store.Get(r, "psidocs-session")
	if err != nil {
		http.Error(w, "Erro ao obter sessão", http.StatusInternalServerError)
		return
	}

	// Revoga a autenticação do usuário
	session.Values["authenticated"] = false
	session.Values["crp"] = ""
	session.Options.MaxAge = -1 // This will delete the cookie

	// Salva a sessão
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, "Erro ao fazer logout", http.StatusInternalServerError)
		return
	}

	// Redireciona para a página inicial
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
// - hash: o hash da senha armazenado no banco de dados
// - salt: o salt da senha armazenado no banco de dados
//
// Retorna:
// - true: se a senha está correta
// - false: se a senha está incorreta
func checkPasswordHash(password, hash, salt string) bool {
	saltedPassword := password + salt
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(saltedPassword))
	return err == nil
}
