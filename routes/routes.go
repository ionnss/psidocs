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

	// Rota para a página de recuperação de senha
	r.HandleFunc("/forgot-password", handlers.ForgotPasswordHandler).Methods("GET", "POST")
	r.HandleFunc("/reset-password/{token}", handlers.ResetPasswordHandler).Methods("GET", "POST")

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
	r.Handle("/dashboard/dados_pessoais", handlers.AuthMiddleware(http.HandlerFunc(handlers.UpdateUserConfigHandler))).Methods("GET", "POST")
	r.Handle("/dashboard/credenciais", handlers.AuthMiddleware(http.HandlerFunc(handlers.UpdateUserCredentialsHandler))).Methods("GET", "POST")
	// Rotas de pacientes
	r.Handle("/patients", handlers.AuthMiddleware(http.HandlerFunc(handlers.ListPatientsHandler))).Methods("GET")
	r.Handle("/patients/create", handlers.AuthMiddleware(http.HandlerFunc(handlers.CreatePatientHandler))).Methods("GET", "POST")
	r.Handle("/patients/{id:[0-9]+}", handlers.AuthMiddleware(http.HandlerFunc(handlers.GetPatientProfileHandler))).Methods("GET")
	r.Handle("/patients/{id:[0-9]+}/edit", handlers.AuthMiddleware(http.HandlerFunc(handlers.UpdatePatientHandler))).Methods("GET", "POST")
	r.Handle("/patients/{id:[0-9]+}/archive", handlers.AuthMiddleware(http.HandlerFunc(handlers.ArchivePatientHandler))).Methods("POST")
	r.Handle("/patients/{id:[0-9]+}/unarchive", handlers.AuthMiddleware(http.HandlerFunc(handlers.UnarchivePatientHandler))).Methods("POST")
	// Rotas de documentos
	r.Handle("/patients/{id:[0-9]+}/documents/editor", handlers.AuthMiddleware(http.HandlerFunc(handlers.DocumentEditorHandler))).Methods("GET")
	r.Handle("/patients/{id:[0-9]+}/documents/personalized/editor", handlers.AuthMiddleware(http.HandlerFunc(handlers.PersonalizedDocumentEditorHandler))).Methods("GET")
	r.Handle("/patients/{id:[0-9]+}/documents/template", handlers.AuthMiddleware(http.HandlerFunc(handlers.DocumentTemplateHandler))).Methods("GET")
	r.Handle("/documents/save", handlers.AuthMiddleware(http.HandlerFunc(handlers.SaveDocumentHandler))).Methods("POST")
	r.Handle("/documents/template-content", handlers.AuthMiddleware(http.HandlerFunc(handlers.TemplateContentHandler))).Methods("GET")
	r.Handle("/documents/{id:[0-9]+}/preview", handlers.AuthMiddleware(http.HandlerFunc(handlers.DocumentPreviewHandler))).Methods("GET")
	r.Handle("/documents/{id:[0-9]+}", handlers.AuthMiddleware(http.HandlerFunc(handlers.DeleteDocumentHandler))).Methods("DELETE")
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
