package handlers

import (
	"html/template"
	"log"
	"net/http"
	"psidocs/db"
	"time"
)

// Document representa a estrutura base de um documento
type Document struct {
	ID          int
	PsicologoID int
	Nome        string
	Descricao   string
	Tipo        string // "contrato" ou "psicologico"
	Subtipo     string // para documentos psicológicos: "laudo", "relatorio", "prontuario", etc
	Conteudo    string // template do documento em HTML
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ListDocumentsHandler lista os templates de documentos do psicólogo
func ListDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("ListDocumentsHandler iniciado - Método: %s", r.Method)

	// Obter ID do psicólogo da sessão
	email, _, err := GetCurrentUserInfo(w, r)
	if err != nil {
		log.Printf("Erro ao obter informações do usuário: %v", err)
		http.Error(w, "Erro ao obter informações do usuário", http.StatusUnauthorized)
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

	// Obter ID do psicólogo
	var psicologoID int
	err = db.QueryRow("SELECT id FROM users WHERE email = $1", email).Scan(&psicologoID)
	if err != nil {
		log.Printf("Erro ao obter ID do psicólogo: %v", err)
		http.Error(w, "Erro ao obter ID do psicólogo", http.StatusInternalServerError)
		return
	}

	// TODO: Buscar documentos do banco de dados
	// Por enquanto, usando dados de exemplo
	contracts := []Document{
		{
			ID:        1,
			Nome:      "Contrato Padrão",
			Descricao: "Contrato base para atendimento psicológico",
			Tipo:      "contrato",
			UpdatedAt: time.Now(),
		},
	}

	psychologicalDocs := []Document{
		{
			ID:        2,
			Nome:      "Laudo Psicológico",
			Descricao: "Template para laudos",
			Tipo:      "psicologico",
			Subtipo:   "laudo",
			UpdatedAt: time.Now(),
		},
	}

	// Preparar dados para o template
	data := map[string]interface{}{
		"Contracts":         contracts,
		"PsychologicalDocs": psychologicalDocs,
	}

	// Se for uma requisição HTMX
	if r.Header.Get("HX-Request") == "true" {
		tmpl := template.Must(template.ParseFiles("templates/view/partials/documents_lists.html"))
		err = tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Erro ao renderizar template: %v", err)
			http.Error(w, "Erro ao renderizar template", http.StatusInternalServerError)
		}
		return
	}

	// Se não for HTMX, renderiza o layout completo
	tmpl := template.Must(template.ParseFiles(
		"templates/view/dashboard_layout.html",
		"templates/view/partials/dashboard_content.html",
		"templates/view/partials/documents_lists.html",
	))
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Erro ao renderizar template: %v", err)
		http.Error(w, "Erro ao renderizar template", http.StatusInternalServerError)
	}
}
