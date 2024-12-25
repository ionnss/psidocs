// Package routes é um pacote que contém as rotas da aplicação
//
// Fornece:
// - Configuração de rotas
// - Serviço de arquivos estáticos
// - Autenticação de usuários
// - Criação de usuários
// - Criação de senhas
// - Login de usuários
// - Logout de usuários
// - Criação de senhas
package routes

import (
	"database/sql"
	"net/http"
	"psidocs/handlers"

	"github.com/gorilla/mux"
)

// ConfigureRoutes configura as rotas da aplicação
//
// Recebe:
// - r: o router da aplicação
// - db: o banco de dados da aplicação
func ConfigureRoutes(r *mux.Router, db *sql.DB) {

	// Servir arquivos estáticos
	//
	// Serve os arquivos estáticos da aplicação
	fs := http.FileServer(http.Dir("templates/statics"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// Rotas públicas
	//
	// Rota para a página inicial
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/view/index.html")
	})

	// Rota auth
	//
	// Rota para autenticar um usuário
	r.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		handlers.AuthHandler(w, r)
	}).Methods("POST")

	// Rota logout
	//
	// Rota para deslogar um usuário
	r.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		handlers.LogoutHandler(w, r)
	}).Methods("POST")
}
