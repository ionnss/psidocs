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
	"html/template"
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

	// Health check
	r.HandleFunc("/health", HealthCheckHandler)

	// Rota para a página inicial
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Obter sessão
		session, _ := handlers.Store.Get(r, "psidocs-session")

		// Se autenticado, redireciona para /dashboard
		if auth, ok := session.Values["authenticated"].(bool); ok && auth {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}

		// Se não autenticado, mostra a página inicial
		tmpl := template.Must(template.ParseFiles("templates/view/index.html"))
		tmpl.Execute(w, nil)
	})

	// Rota auth
	//
	// Rota para autenticar um usuário
	r.HandleFunc("/auth", handlers.AuthHandler).Methods("POST")

	// Rota logout
	//
	// Rota para deslogar um usuário
	r.HandleFunc("/logout", handlers.LogoutHandler).Methods("POST")

	// Rota dashboard (protegida)
	//
	// Rota para a página dashboard
	r.Handle("/dashboard", handlers.AuthMiddleware(http.HandlerFunc(handlers.DashboardHandler))).Methods("GET")
	r.HandleFunc("/dashboard", handlers.AuthHandler).Methods("POST")
	r.Handle("/dashboard/configuracoes", handlers.AuthMiddleware(http.HandlerFunc(handlers.UpdateUserConfigHandler))).Methods("GET", "POST")
	r.Handle("/dashboard/credenciais", handlers.AuthMiddleware(http.HandlerFunc(handlers.UpdateUserCredentialsHandler))).Methods("GET", "POST")
}

// HealthCheckHandler retorna 200 OK para health checks
//
// Recebe:
// - w: o writer do response
// - r: o request
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
