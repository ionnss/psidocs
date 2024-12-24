package routes

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
)

func ConfigureRoutes(r *mux.Router, db *sql.DB) {
	// Servir arquivos estáticos
	fs := http.FileServer(http.Dir("templates/statics"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	// Rota para a página inicial
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "templates/view/index.html")
	})

}
