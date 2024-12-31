package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"psidocs/db"
	"time"

	"github.com/gorilla/mux"
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
	RG             string
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
	EstadoCivil    string
	Nacionalidade  string
	Profissao      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// CreatePatientHandler processa o formulário de criação de paciente
//
// Recebe:
// - w: o writer do response
// - r: o request
//
// Retorna:
// - void
//
// Descrição:
// - Processa o formulário de criação de paciente
// - Renderiza o template de criação de paciente
// - Insere o paciente no banco de dados
// - Retorna uma mensagem de sucesso ou erro
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
			cpf, rg, data_nascimento, sexo,
			endereco, numero, bairro,
			cidade, estado, cep,
			observacoes, status, estado_civil, nacionalidade, profissao
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`,
		psicologoID,
		r.FormValue("nome"),
		r.FormValue("email"),
		r.FormValue("ddd"),
		r.FormValue("telefone"),
		r.FormValue("whatsapp") == "on",
		cpf,
		r.FormValue("rg"),
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
		r.FormValue("estado_civil"),
		r.FormValue("nacionalidade"),
		r.FormValue("profissao"),
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
			htmx.ajax('GET', '/patients', {target: '#content-area'});
		}, 2000);
	</script>`))
}

// ListPatientsHandler lista os pacientes do psicólogo
//
// Recebe:
// - w: o writer do response
// - r: o request
//
// Retorna:
// - void
//
// Descrição:
// - Lista os pacientes do psicólogo
// - Renderiza o template de lista de pacientes
// - Retorna uma mensagem de sucesso ou erro
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
	if r.Header.Get("HX-Target") == "patients-tbody" {
		// Renderizar apenas as linhas da tabela
		w.Header().Set("Content-Type", "text/html")
		for _, p := range patients {
			fmt.Fprintf(w, `
				<tr>
					<td>%s</td>
					<td>%s</td>
					<td>
						%s%s
						%s
					</td>
					<td>
						<span class="badge %s">
							%s
						</span>
					</td>
					<td>
						<div class="btn-group btn-group-sm">
							<button class="btn btn-info" 
									hx-get="/patients/%d"
									hx-target="#content-area">
								<i class="bi bi-eye-fill"></i>
							</button>
							<button class="btn btn-warning"
									hx-get="/patients/%d/edit"
									hx-target="#content-area">
								<i class="bi bi-pencil-fill"></i>
							</button>
							%s
						</div>
					</td>
				</tr>`,
				p.Nome,
				p.CPF,
				func() string {
					if p.DDD != "" {
						return fmt.Sprintf("(%s) ", p.DDD)
					}
					return ""
				}(),
				p.Telefone,
				func() string {
					if p.WhatsApp {
						return `<i class="bi bi-whatsapp text-success"></i>`
					}
					return ""
				}(),
				func() string {
					if p.Status == "ativo" {
						return "bg-success"
					}
					return "bg-warning"
				}(),
				p.Status,
				p.ID,
				p.ID,
				func() string {
					if p.Status == "ativo" {
						return fmt.Sprintf(`
							<button class="btn btn-danger"
									hx-post="/patients/%d/archive"
									hx-confirm="Tem certeza que deseja arquivar este paciente?"
									hx-target="#patients-message-area">
								<i class="bi bi-archive-fill"></i>
							</button>`, p.ID)
					}
					return fmt.Sprintf(`
							<button class="btn btn-success"
									hx-post="/patients/%d/unarchive"
									hx-confirm="Tem certeza que deseja reativar este paciente?"
									hx-target="#patients-message-area">
								<i class="bi bi-arrow-counterclockwise"></i>
							</button>`, p.ID)
				}(),
			)
		}
		if len(patients) == 0 {
			fmt.Fprintf(w, `<tr><td colspan="5" class="text-center">Nenhum paciente encontrado</td></tr>`)
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
		"templates/view/partials/dashboard_content.html",
		"templates/view/partials/patients_lists.html",
	))
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Erro ao renderizar template: %v", err)
		http.Error(w, "Erro ao renderizar template", http.StatusInternalServerError)
	}
}

// GetPatientHandler mostra os detalhes de um paciente específico
func GetPatientHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("GetPatientHandler iniciado - Método: %s", r.Method)

	// Obter ID do paciente da URL
	vars := mux.Vars(r)
	patientID := vars["id"]

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

	// Buscar dados do paciente
	var patient Patient
	err = db.QueryRow(`
		SELECT 
			id, psicologo_id, nome, email, 
			ddd, telefone, whatsapp,
			cpf, rg, data_nascimento, sexo,
			endereco, numero, bairro,
			cidade, estado, cep,
			observacoes, status,
			estado_civil, nacionalidade, profissao,
			created_at, updated_at
		FROM patients 
		WHERE id = $1 AND psicologo_id = $2`,
		patientID, psicologoID,
	).Scan(
		&patient.ID, &patient.PsicologoID, &patient.Nome, &patient.Email,
		&patient.DDD, &patient.Telefone, &patient.WhatsApp,
		&patient.CPF, &patient.RG, &patient.DataNascimento, &patient.Sexo,
		&patient.Endereco, &patient.Numero, &patient.Bairro,
		&patient.Cidade, &patient.Estado, &patient.CEP,
		&patient.Observacoes, &patient.Status,
		&patient.EstadoCivil, &patient.Nacionalidade, &patient.Profissao,
		&patient.CreatedAt, &patient.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		log.Printf("Paciente não encontrado ou não pertence ao psicólogo: %v", err)
		http.Error(w, "Paciente não encontrado", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Erro ao buscar paciente: %v", err)
		http.Error(w, "Erro ao buscar paciente", http.StatusInternalServerError)
		return
	}

	// Preparar dados para o template
	data := map[string]interface{}{
		"Patient": patient,
	}

	// Se for uma requisição HTMX
	if r.Header.Get("HX-Request") == "true" {
		tmpl := template.Must(template.ParseFiles("templates/view/partials/patients_profile.html"))
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
		"templates/view/partials/patients_profile.html",
	))
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Erro ao renderizar template: %v", err)
		http.Error(w, "Erro ao renderizar template", http.StatusInternalServerError)
	}
}

// UpdatePatientHandler processa a edição de um paciente
func UpdatePatientHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("UpdatePatientHandler iniciado - Método: %s", r.Method)

	// Obter ID do paciente da URL
	vars := mux.Vars(r)
	patientID := vars["id"]

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

	// Se for GET, mostra o formulário de edição
	if r.Method == "GET" {
		// Buscar dados do paciente
		var patient Patient
		err = db.QueryRow(`
			SELECT 
				id, psicologo_id, nome, email, 
				ddd, telefone, whatsapp,
				cpf, rg, data_nascimento, sexo,
				endereco, numero, bairro,
				cidade, estado, cep,
				observacoes, status,
				estado_civil, nacionalidade, profissao
			FROM patients 
			WHERE id = $1 AND psicologo_id = $2`,
			patientID, psicologoID,
		).Scan(
			&patient.ID, &patient.PsicologoID, &patient.Nome, &patient.Email,
			&patient.DDD, &patient.Telefone, &patient.WhatsApp,
			&patient.CPF, &patient.RG, &patient.DataNascimento, &patient.Sexo,
			&patient.Endereco, &patient.Numero, &patient.Bairro,
			&patient.Cidade, &patient.Estado, &patient.CEP,
			&patient.Observacoes, &patient.Status,
			&patient.EstadoCivil, &patient.Nacionalidade, &patient.Profissao,
		)

		if err == sql.ErrNoRows {
			log.Printf("Paciente não encontrado ou não pertence ao psicólogo: %v", err)
			http.Error(w, "Paciente não encontrado", http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("Erro ao buscar paciente: %v", err)
			http.Error(w, "Erro ao buscar paciente", http.StatusInternalServerError)
			return
		}

		// Preparar dados para o template
		data := map[string]interface{}{
			"Patient": patient,
		}

		// Se for uma requisição HTMX
		if r.Header.Get("HX-Request") == "true" {
			tmpl := template.Must(template.ParseFiles("templates/view/partials/patients_edit_form.html"))
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
			"templates/view/partials/patients_edit_form.html",
		))
		err = tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Erro ao renderizar template: %v", err)
			http.Error(w, "Erro ao renderizar template", http.StatusInternalServerError)
		}
		return
	}

	// Se for POST, processa a atualização
	if r.Method == "POST" {
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

		// Verificar se CPF já existe para outro paciente deste psicólogo
		cpf := r.FormValue("cpf")
		if cpf != "" {
			var exists bool
			err = db.QueryRow(`
				SELECT EXISTS(
					SELECT 1 FROM patients 
					WHERE cpf = $1 AND psicologo_id = $2 AND id != $3
				)`, cpf, psicologoID, patientID).Scan(&exists)
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

		// Atualizar paciente
		_, err = db.Exec(`
			UPDATE patients SET
				nome = $1, email = $2,
				ddd = $3, telefone = $4, whatsapp = $5,
				cpf = $6, rg = $7, data_nascimento = $8, sexo = $9,
				endereco = $10, numero = $11, bairro = $12,
				cidade = $13, estado = $14, cep = $15,
				observacoes = $16,
				estado_civil = $17, nacionalidade = $18, profissao = $19,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $20 AND psicologo_id = $21`,
			r.FormValue("nome"),
			r.FormValue("email"),
			r.FormValue("ddd"),
			r.FormValue("telefone"),
			r.FormValue("whatsapp") == "on",
			cpf,
			r.FormValue("rg"),
			dataNascimento,
			r.FormValue("sexo"),
			r.FormValue("endereco"),
			r.FormValue("numero"),
			r.FormValue("bairro"),
			r.FormValue("cidade"),
			r.FormValue("estado"),
			r.FormValue("cep"),
			r.FormValue("observacoes"),
			r.FormValue("estado_civil"),
			r.FormValue("nacionalidade"),
			r.FormValue("profissao"),
			patientID,
			psicologoID,
		)

		if err != nil {
			log.Printf("Erro ao atualizar paciente: %v", err)
			w.Write([]byte(`<div class="alert alert-danger alert-dismissible fade show" role="alert">
				<i class="bi bi-exclamation-triangle-fill me-2"></i>
				Erro ao atualizar paciente. Por favor, tente novamente.
				<button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Fechar"></button>
			</div>`))
			return
		}

		// Retornar mensagem de sucesso e redirecionar via HTMX
		w.Write([]byte(`<div class="alert alert-success alert-dismissible fade show" role="alert">
			<i class="bi bi-check-circle-fill me-2"></i>
			Paciente atualizado com sucesso!
			<button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Fechar"></button>
		</div>
		<script>
			setTimeout(function() {
				htmx.ajax('GET', '/patients/` + patientID + `', {target: '#content-area'});
			}, 2000);
		</script>`))
		return
	}

	http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
}

func ArchivePatientHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("ArchivePatientHandler iniciado - Método: %s", r.Method)

	if r.Method != "POST" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Obter ID do paciente da URL
	vars := mux.Vars(r)
	patientID := vars["id"]

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

	// Atualizar status do paciente para inativo
	result, err := db.Exec(`
		UPDATE patients 
		SET status = 'inativo', updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND psicologo_id = $2`,
		patientID, psicologoID,
	)

	if err != nil {
		log.Printf("Erro ao arquivar paciente: %v", err)
		w.Write([]byte(`<div class="alert alert-danger alert-dismissible fade show" role="alert">
			<i class="bi bi-exclamation-triangle-fill me-2"></i>
			Erro ao arquivar paciente. Por favor, tente novamente.
			<button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Fechar"></button>
		</div>`))
		return
	}

	// Verificar se algum registro foi atualizado
	rows, err := result.RowsAffected()
	if err != nil {
		log.Printf("Erro ao verificar linhas afetadas: %v", err)
		http.Error(w, "Erro ao verificar atualização", http.StatusInternalServerError)
		return
	}

	if rows == 0 {
		log.Printf("Nenhum paciente encontrado para arquivar")
		w.Write([]byte(`<div class="alert alert-warning alert-dismissible fade show" role="alert">
			<i class="bi bi-exclamation-triangle-fill me-2"></i>
			Paciente não encontrado
			<button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Fechar"></button>
		</div>`))
		return
	}

	// Retornar mensagem de sucesso e recarregar o perfil
	w.Write([]byte(`<div class="alert alert-success alert-dismissible fade show" role="alert">
		<i class="bi bi-check-circle-fill me-2"></i>
		Paciente arquivado com sucesso!
		<button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Fechar"></button>
	</div>
	<script>
		setTimeout(function() {
			htmx.ajax('GET', '/patients/' + ` + patientID + `, {target: '#content-area'});
		}, 2000);
	</script>`))
}

func UnarchivePatientHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("UnarchivePatientHandler iniciado - Método: %s", r.Method)

	if r.Method != "POST" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Obter ID do paciente da URL
	vars := mux.Vars(r)
	patientID := vars["id"]

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

	// Atualizar status do paciente para ativo
	result, err := db.Exec(`
		UPDATE patients 
		SET status = 'ativo', updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND psicologo_id = $2`,
		patientID, psicologoID,
	)

	if err != nil {
		log.Printf("Erro ao desarquivar paciente: %v", err)
		w.Write([]byte(`<div class="alert alert-danger alert-dismissible fade show" role="alert">
			<i class="bi bi-exclamation-triangle-fill me-2"></i>
			Erro ao desarquivar paciente. Por favor, tente novamente.
			<button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Fechar"></button>
		</div>`))
		return
	}

	// Verificar se algum registro foi atualizado
	rows, err := result.RowsAffected()
	if err != nil {
		log.Printf("Erro ao verificar linhas afetadas: %v", err)
		http.Error(w, "Erro ao verificar atualização", http.StatusInternalServerError)
		return
	}

	if rows == 0 {
		log.Printf("Nenhum paciente encontrado para desarquivar")
		w.Write([]byte(`<div class="alert alert-warning alert-dismissible fade show" role="alert">
			<i class="bi bi-exclamation-triangle-fill me-2"></i>
			Paciente não encontrado
			<button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Fechar"></button>
		</div>`))
		return
	}

	// Retornar mensagem de sucesso e recarregar o perfil
	w.Write([]byte(`<div class="alert alert-success alert-dismissible fade show" role="alert">
		<i class="bi bi-check-circle-fill me-2"></i>
		Paciente desarquivado com sucesso!
		<button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Fechar"></button>
	</div>
	<script>
		setTimeout(function() {
			htmx.ajax('GET', '/patients/' + ` + patientID + `, {target: '#content-area'});
		}, 2000);
	</script>`))
}

// formatDocumentType retorna o tipo formatado do documento
func formatDocumentType(tipo string) string {
	switch tipo {
	case "anamnese":
		return "Anamnese"
	case "atestado":
		return "Atestado"
	case "declaracao":
		return "Declaração"
	case "laudo":
		return "Laudo"
	case "relatorio":
		return "Relatório"
	case "presencial":
		return "Contrato Presencial"
	case "online":
		return "Contrato Online"
	default:
		return tipo
	}
}

// GetPatientProfileHandler carrega o perfil do paciente
func GetPatientProfileHandler(w http.ResponseWriter, r *http.Request) {
	// Obter ID do paciente da URL
	vars := mux.Vars(r)
	patientID := vars["id"]

	// Obter dados do psicólogo da sessão
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

	// Obter dados do paciente
	var patient Patient
	err = db.QueryRow(`
		SELECT 
			id, nome, email, cpf, data_nascimento, sexo,
			endereco, numero, bairro, cidade, estado, cep,
			estado_civil, nacionalidade, profissao,
			ddd, telefone, whatsapp, rg, status, observacoes,
			created_at, updated_at
		FROM patients 
		WHERE id = $1 AND psicologo_id = $2`,
		patientID, psicologoID,
	).Scan(
		&patient.ID, &patient.Nome, &patient.Email, &patient.CPF,
		&patient.DataNascimento, &patient.Sexo, &patient.Endereco,
		&patient.Numero, &patient.Bairro, &patient.Cidade,
		&patient.Estado, &patient.CEP, &patient.EstadoCivil,
		&patient.Nacionalidade, &patient.Profissao,
		&patient.DDD, &patient.Telefone, &patient.WhatsApp,
		&patient.RG, &patient.Status, &patient.Observacoes,
		&patient.CreatedAt, &patient.UpdatedAt,
	)

	if err != nil {
		log.Printf("Erro ao obter dados do paciente: %v", err)
		http.Error(w, "Erro ao obter dados do paciente", http.StatusInternalServerError)
		return
	}

	// Converter timestamps para timezone do Brasil
	brazilLoc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		log.Printf("Erro ao carregar timezone do Brasil: %v", err)
		// Não retornamos erro aqui para não quebrar a exibição
		// Apenas logamos o erro e mantemos UTC
	} else {
		patient.CreatedAt = patient.CreatedAt.In(brazilLoc)
		patient.UpdatedAt = patient.UpdatedAt.In(brazilLoc)
	}

	// Log para debug
	log.Printf("Datas do paciente (BR) - Criação: %v, Atualização: %v",
		patient.CreatedAt.Format("02/01/2006 15:04"),
		patient.UpdatedAt.Format("02/01/2006 15:04"))

	// Obter contratos do paciente
	rows, err := db.Query(`
		SELECT 
			d.id, 
			d.tipo, 
			d.nome, 
			d.conteudo, 
			d.requer_assinatura, 
			d.created_at, 
			d.updated_at,
			p.nome as patient_name
		FROM documents d
		JOIN patients p ON d.paciente_id = p.id
		WHERE d.paciente_id = $1 AND d.psicologo_id = $2 
		AND (
			d.tipo LIKE 'contracts/%' OR 
			d.tipo IN ('presencial', 'online')
		)
		ORDER BY d.updated_at DESC`,
		patientID, psicologoID,
	)
	if err != nil {
		log.Printf("Erro ao obter contratos: %v", err)
		http.Error(w, "Erro ao obter contratos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var contracts []Document
	for rows.Next() {
		var doc Document
		err := rows.Scan(
			&doc.ID,
			&doc.Tipo,
			&doc.Nome,
			&doc.Conteudo,
			&doc.RequerAssinatura,
			&doc.CreatedAt,
			&doc.UpdatedAt,
			&doc.PatientName,
		)
		if err != nil {
			log.Printf("Erro ao ler contrato: %v", err)
			continue
		}
		// Formatar o tipo do documento
		doc.Tipo = formatDocumentType(doc.Tipo)
		// Converter timestamps do documento para timezone do Brasil
		if brazilLoc != nil { // Só converte se o timezone foi carregado com sucesso
			doc.CreatedAt = doc.CreatedAt.In(brazilLoc)
			doc.UpdatedAt = doc.UpdatedAt.In(brazilLoc)
		}
		contracts = append(contracts, doc)
	}

	// Obter documentos psicológicos do paciente
	rows, err = db.Query(`
		SELECT 
			d.id, 
			d.tipo, 
			d.nome, 
			d.conteudo, 
			d.requer_assinatura, 
			d.created_at, 
			d.updated_at,
			p.nome as patient_name
		FROM documents d
		JOIN patients p ON d.paciente_id = p.id
		WHERE d.paciente_id = $1 AND d.psicologo_id = $2 
		AND (
			d.tipo LIKE 'psychological-documents/%' OR 
			d.tipo IN ('anamnese', 'atestado', 'declaracao', 'laudo', 'relatorio')
		)
		ORDER BY d.updated_at DESC`,
		patientID, psicologoID,
	)
	if err != nil {
		log.Printf("Erro ao obter documentos psicológicos: %v", err)
		http.Error(w, "Erro ao obter documentos psicológicos", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var psychologicalDocs []Document
	for rows.Next() {
		var doc Document
		err := rows.Scan(
			&doc.ID,
			&doc.Tipo,
			&doc.Nome,
			&doc.Conteudo,
			&doc.RequerAssinatura,
			&doc.CreatedAt,
			&doc.UpdatedAt,
			&doc.PatientName,
		)
		if err != nil {
			log.Printf("Erro ao ler documento psicológico: %v", err)
			continue
		}
		// Formatar o tipo do documento
		doc.Tipo = formatDocumentType(doc.Tipo)
		// Converter timestamps do documento para timezone do Brasil
		if brazilLoc != nil { // Só converte se o timezone foi carregado com sucesso
			doc.CreatedAt = doc.CreatedAt.In(brazilLoc)
			doc.UpdatedAt = doc.UpdatedAt.In(brazilLoc)
		}
		psychologicalDocs = append(psychologicalDocs, doc)
	}

	// Preparar dados para o template
	data := map[string]interface{}{
		"Patient":           patient,
		"Contracts":         contracts,
		"PsychologicalDocs": psychologicalDocs,
	}

	// Se for uma requisição HTMX
	if r.Header.Get("HX-Request") == "true" {
		tmpl := template.Must(template.ParseFiles("templates/view/partials/patients_profile.html"))
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
		"templates/view/partials/patients_profile.html",
	))
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Erro ao renderizar template: %v", err)
		http.Error(w, "Erro ao renderizar template", http.StatusInternalServerError)
	}
}
