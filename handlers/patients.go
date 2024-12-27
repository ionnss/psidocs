package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"psidocs/db"
	"time"
)

// Patient representa a estrutura de dados de um paciente
type Patient struct {
	ID             int
	PsicologoID    int
	Nome           string
	Email          string
	DDD            string
	Telefone       string
	WhatsApp       bool
	CPF            string
	DataNascimento time.Time
	Sexo           string
	Endereco       string
	Numero         string
	Bairro         string
	Cidade         string
	Estado         string
	CEP            string
	Status         string
	Observacoes    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// CreatePatientHandler processa o formulário de criação de paciente
func CreatePatientHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("CreatePatientHandler iniciado - Método: %s", r.Method)

	if r.Method == "GET" {
		// Se for uma requisição HTMX, renderiza só o conteúdo
		if r.Header.Get("HX-Request") == "true" {
			log.Printf("Renderizando template para requisição HTMX")
			tmpl := template.Must(template.ParseFiles("templates/view/partials/patients_registration_form.html"))
			if err := tmpl.Execute(w, nil); err != nil {
				log.Printf("Erro ao renderizar template: %v", err)
				http.Error(w, "Erro ao renderizar template", http.StatusInternalServerError)
				return
			}
			return
		}

		// Se não for HTMX, renderiza o layout completo
		log.Printf("Renderizando layout completo")
		tmpl := template.Must(template.ParseFiles(
			"templates/view/dashboard_layout.html",
			"templates/view/partials/patients_registration_form.html",
		))
		if err := tmpl.Execute(w, nil); err != nil {
			log.Printf("Erro ao renderizar template: %v", err)
			http.Error(w, "Erro ao renderizar template", http.StatusInternalServerError)
			return
		}
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

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

	// Processar dados do formulário
	dataNascimento, err := time.Parse("2006-01-02", r.FormValue("data_nascimento"))
	if err != nil {
		log.Printf("Erro ao processar data de nascimento: %v", err)
		w.Write([]byte(`<div class="alert alert-danger alert-dismissible fade show" role="alert">
			<i class="bi bi-exclamation-triangle-fill me-2"></i>
			Data de nascimento inválida
			<button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Fechar"></button>
		</div>`))
		return
	}

	// Verificar se CPF já existe para este psicólogo
	cpf := r.FormValue("cpf")
	if cpf != "" {
		var exists bool
		err = db.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM patients 
				WHERE cpf = $1 AND psicologo_id = $2
			)`, cpf, psicologoID).Scan(&exists)
		if err != nil {
			log.Printf("Erro ao verificar CPF: %v", err)
			http.Error(w, "Erro ao verificar CPF", http.StatusInternalServerError)
			return
		}
		if exists {
			w.Write([]byte(`<div class="alert alert-danger alert-dismissible fade show" role="alert">
				<i class="bi bi-exclamation-triangle-fill me-2"></i>
				CPF já cadastrado para outro paciente
				<button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Fechar"></button>
			</div>`))
			return
		}
	}

	// Criar novo paciente
	_, err = db.Exec(`
		INSERT INTO patients (
			psicologo_id, nome, email, 
			ddd, telefone, whatsapp,
			cpf, data_nascimento, sexo,
			endereco, numero, bairro,
			cidade, estado, cep,
			observacoes, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`,
		psicologoID,
		r.FormValue("nome"),
		r.FormValue("email"),
		r.FormValue("ddd"),
		r.FormValue("telefone"),
		r.FormValue("whatsapp") == "on",
		cpf,
		dataNascimento,
		r.FormValue("sexo"),
		r.FormValue("endereco"),
		r.FormValue("numero"),
		r.FormValue("bairro"),
		r.FormValue("cidade"),
		r.FormValue("estado"),
		r.FormValue("cep"),
		r.FormValue("observacoes"),
		"ativo", // Status inicial
	)

	if err != nil {
		log.Printf("Erro ao inserir paciente: %v", err)
		w.Write([]byte(`<div class="alert alert-danger alert-dismissible fade show" role="alert">
			<i class="bi bi-exclamation-triangle-fill me-2"></i>
			Erro ao cadastrar paciente. Por favor, tente novamente.
			<button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Fechar"></button>
		</div>`))
		return
	}

	// Retornar mensagem de sucesso
	w.Write([]byte(`<div class="alert alert-success alert-dismissible fade show" role="alert">
		<i class="bi bi-check-circle-fill me-2"></i>
		Paciente cadastrado com sucesso!
		<button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Fechar"></button>
	</div>
	<script>
		setTimeout(function() {
			window.location.href = "/patients";
		}, 2000);
	</script>`))
}

func UpdatePatientHandler(w http.ResponseWriter, r *http.Request) {

}

func GetPatientHandler(w http.ResponseWriter, r *http.Request) {

}

func DeletePatientHandler(w http.ResponseWriter, r *http.Request) {

}

// ListPatientsHandler lista os pacientes do psicólogo
func ListPatientsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("ListPatientsHandler iniciado - Método: %s", r.Method)

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

	// Preparar query base
	query := `
		SELECT id, nome, cpf, ddd, telefone, whatsapp, status
		FROM patients 
		WHERE psicologo_id = $1`
	args := []interface{}{psicologoID}
	argCount := 1

	// Adicionar filtros se fornecidos
	search := r.URL.Query().Get("search")
	if search != "" {
		query += fmt.Sprintf(" AND (nome ILIKE $%d OR cpf ILIKE $%d)", argCount+1, argCount+1)
		args = append(args, "%"+search+"%")
		argCount++
	}

	status := r.URL.Query().Get("status")
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argCount+1)
		args = append(args, status)
	}

	// Ordenar por nome
	query += " ORDER BY nome"

	// Executar query
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Erro ao buscar pacientes: %v", err)
		http.Error(w, "Erro ao buscar pacientes", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Processar resultados
	var patients []Patient
	for rows.Next() {
		var p Patient
		err := rows.Scan(&p.ID, &p.Nome, &p.CPF, &p.DDD, &p.Telefone, &p.WhatsApp, &p.Status)
		if err != nil {
			log.Printf("Erro ao processar paciente: %v", err)
			continue
		}
		patients = append(patients, p)
	}

	// Preparar dados do template
	data := map[string]interface{}{
		"Patients": patients,
	}

	// Se for uma requisição HTMX para a tabela
	if r.Header.Get("HX-Target") == "patients-table" {
		tmpl := template.Must(template.ParseFiles("templates/view/partials/patients_lists.html"))
		err = tmpl.ExecuteTemplate(w, "patients-table", data)
		if err != nil {
			log.Printf("Erro ao renderizar tabela: %v", err)
			http.Error(w, "Erro ao renderizar tabela", http.StatusInternalServerError)
		}
		return
	}

	// Se for uma requisição HTMX normal
	if r.Header.Get("HX-Request") == "true" {
		tmpl := template.Must(template.ParseFiles("templates/view/partials/patients_lists.html"))
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
		"templates/view/partials/patients_lists.html",
	))
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Erro ao renderizar template: %v", err)
		http.Error(w, "Erro ao renderizar template", http.StatusInternalServerError)
	}
}

func ArchivePatientHandler(w http.ResponseWriter, r *http.Request) {

}

func UnarchivePatientHandler(w http.ResponseWriter, r *http.Request) {

}
