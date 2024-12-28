package main

import (
	"log"

	"net/http"
	"psidocs/db"
	"psidocs/handlers"
	"psidocs/routes"

	"github.com/gorilla/mux"
)

func main() {
	// Conecta ao banco de dados
	conn, err := db.Connect()
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	defer conn.Close()

	// Executa as migrações
	db.ExecuteMigrations(conn)

	// Configura o roteador
	r := mux.NewRouter()
	routes.ConfigureRoutes(r, conn)

	// Rotas de documentos
	r.HandleFunc("/documents/template", handlers.DocumentTemplateHandler)
	r.HandleFunc("/documents/save", handlers.SaveDocumentHandler)

	// Servir arquivos estáticos
	fs := http.FileServer(http.Dir("static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// Inicia o servidor
	log.Println("Servidor iniciado na porta 8080 em http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}
