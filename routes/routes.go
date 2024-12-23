package routes

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"
)

func ConfigureRoutes(r *mux.Router, db *sql.DB) {
	// Servir arquivos estáticos
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Rota para a página inicial
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

}
