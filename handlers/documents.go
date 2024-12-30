package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"psidocs/db"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var meses = map[string]string{
	"January":   "Janeiro",
	"February":  "Fevereiro",
	"March":     "Março",
	"April":     "Abril",
	"May":       "Maio",
	"June":      "Junho",
	"July":      "Julho",
	"August":    "Agosto",
	"September": "Setembro",
	"October":   "Outubro",
	"November":  "Novembro",
	"December":  "Dezembro",
}

// Função auxiliar para formatar datas
func formatDate(date string) string {
	if date == "" {
		return ""
	}
	// Converte a data do formato "2006-01-02" para time.Time
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	// Formata para o padrão brasileiro
	return t.Format("02/01/2006")
}

// DocumentTemplateHandler carrega o template do documento selecionado
//
// Recebe:
// - value: tipo do documento
// - document-type: tipo do documento
// - patient_id: ID do paciente
// - editor_contents: conteúdo do editor de documentos
// - data_inicio_ferias: data de início das férias
// - data_fim_ferias: data de fim das férias
// - faltas: número de faltas
// - data_inicio: data de início do tratamento
// - data_fim_tratamento: data de fim do tratamento
//
// Retorna:
// - template do documento preenchido com os dados recebidos
func DocumentTemplateHandler(w http.ResponseWriter, r *http.Request) {
	// Obter tipo do documento
	templatePath := r.URL.Query().Get("value")
	if templatePath == "" {
		templatePath = r.URL.Query().Get("document-type")
	}
	if templatePath == "" {
		http.Error(w, "Tipo de documento não especificado", http.StatusBadRequest)
		return
	}

	// Obter ID do paciente do contexto
	vars := mux.Vars(r)
	patientID := vars["id"]
	if patientID == "" {
		http.Error(w, "ID do paciente não fornecido", http.StatusBadRequest)
		return
	}

	// Obter dados do psicólogo da sessão
	email, crp, err := GetCurrentUserInfo(w, r)
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

	// Obter dados do psicólogo
	var psicologo struct {
		ID            int
		Nome          string
		CRP           string
		Email         string
		Cidade        string
		EstadoCivil   string
		Nacionalidade string
		CPF           string
		RG            string
		Endereco      string
		Numero        string
		CEP           string
		DDD           string
		Telefone      string
	}
	err = db.QueryRow(`
		SELECT 
			u.id, 
			CONCAT(ud.first_name, ' ', COALESCE(ud.middle_name || ' ', ''), ud.last_name) as nome,
			u.hash_crp as crp,
			u.email,
			ud.cidade,
			ud.estado_civil,
			ud.nacionalidade,
			ud.cpf,
			ud.rg,
			ud.endereco,
			ud.numero,
			ud.cep,
			ud.ddd,
			ud.telefone
		FROM users u
		LEFT JOIN users_data ud ON u.id = ud.user_id
		WHERE u.email = $1`,
		email,
	).Scan(
		&psicologo.ID, &psicologo.Nome, &psicologo.CRP, &psicologo.Email,
		&psicologo.Cidade, &psicologo.EstadoCivil, &psicologo.Nacionalidade,
		&psicologo.CPF, &psicologo.RG, &psicologo.Endereco, &psicologo.Numero,
		&psicologo.CEP, &psicologo.DDD, &psicologo.Telefone,
	)
	if err != nil {
		log.Printf("Erro ao obter dados do psicólogo: %v", err)
		http.Error(w, "Erro ao obter dados do psicólogo", http.StatusInternalServerError)
		return
	}

	// Obter dados do paciente
	var paciente Patient
	err = db.QueryRow(`
		SELECT 
			id, nome, email, cpf, data_nascimento, sexo,
			endereco, numero, bairro, cidade, estado, cep,
			estado_civil, nacionalidade, profissao,
			ddd, telefone, rg
		FROM patients 
		WHERE id = $1 AND psicologo_id = $2`,
		patientID, psicologo.ID,
	).Scan(
		&paciente.ID, &paciente.Nome, &paciente.Email, &paciente.CPF,
		&paciente.DataNascimento, &paciente.Sexo, &paciente.Endereco,
		&paciente.Numero, &paciente.Bairro, &paciente.Cidade,
		&paciente.Estado, &paciente.CEP, &paciente.EstadoCivil,
		&paciente.Nacionalidade, &paciente.Profissao,
		&paciente.DDD, &paciente.Telefone, &paciente.RG,
	)
	if err != nil {
		log.Printf("Erro ao obter dados do paciente: %v", err)
		http.Error(w, "Erro ao obter dados do paciente", http.StatusInternalServerError)
		return
	}

	// Ler o template do documento
	templateFile := fmt.Sprintf("templates/documents/%s.html", templatePath)
	content, err := os.ReadFile(templateFile)
	if err != nil {
		log.Printf("Erro ao ler template %s: %v", templateFile, err)
		http.Error(w, "Template não encontrado", http.StatusNotFound)
		return
	}

	// Preparar dados para o template
	data := map[string]interface{}{
		"PatientNome":           paciente.Nome,
		"PatientNacionalidade":  paciente.Nacionalidade,
		"PatientEstadoCivil":    paciente.EstadoCivil,
		"PatientRG":             paciente.RG,
		"PatientCPF":            paciente.CPF,
		"PatientCidade":         paciente.Cidade,
		"PatientRua":            paciente.Endereco,
		"PatientNumero":         paciente.Numero,
		"PatientCEP":            paciente.CEP,
		"PatientTelefone":       fmt.Sprintf("(%s) %s", paciente.DDD, paciente.Telefone),
		"PatientDataNascimento": formatDate(paciente.DataNascimento.Format("2006-01-02")),
		"PatientGenero":         paciente.Sexo,
		"PatientProfissao":      paciente.Profissao,
		"PatientEndereco":       fmt.Sprintf("%s, %s - %s, %s", paciente.Endereco, paciente.Numero, paciente.Bairro, paciente.Cidade),

		"PsicologoNome":          psicologo.Nome,
		"PsicologoNacionalidade": psicologo.Nacionalidade,
		"PsicologoEstadoCivil":   psicologo.EstadoCivil,
		"PsicologoCRP":           crp,
		"PsicologoRG":            psicologo.RG,
		"PsicologoCPF":           psicologo.CPF,
		"PsicologoCidade":        psicologo.Cidade,
		"PsicologoRua":           psicologo.Endereco,
		"PsicologoCEP":           psicologo.CEP,
		"PsicologoNumero":        psicologo.Numero,
		"PsicologoTelefone":      fmt.Sprintf("(%s) %s", psicologo.DDD, psicologo.Telefone),

		"DataAssinatura":   time.Now().Format("02/01/2006"),
		"DiaAssinatura":    psicologo.Cidade,
		"DiaMesAssinatura": time.Now().Format("02"),
		"MesAssinatura":    time.Now().Format("January"),
		"AnoAssinatura":    time.Now().Format("2006"),

		// Campos do formulário de contrato
		"Abordagem":           r.FormValue("abordagem"),
		"ValorSessao":         r.FormValue("valor_sessao"),
		"DiaSemana":           r.FormValue("dia_semana"),
		"HorarioSessao":       r.FormValue("horario_sessao"),
		"DataLimitePagamento": r.FormValue("data_limite_pagamento"),
		"MetodosPagamento":    r.FormValue("metodos_pagamento"),
		"DataInicioFerias":    r.FormValue("data_inicio_ferias"),
		"DataFimFerias":       r.FormValue("data_fim_ferias"),
		"DataFimTratamento":   r.FormValue("data_fim_tratamento"),
		"NumeroFaltas":        r.FormValue("numero_faltas"),
		"DataInicio":          r.FormValue("data_inicio"),

		// Campos específicos do atestado
		"DataInicialAvaliacao": formatDate(r.FormValue("data_inicial_avaliacao")),
		"DataFinalAvaliacao":   formatDate(r.FormValue("data_final_avaliacao")),
		"NaturezaAvaliacao":    r.FormValue("natureza_avaliacao"),
		"EstadoPsicologico":    r.FormValue("estado_psicologico"),
		"DataValidade":         formatDate(r.FormValue("data_validade")),
		"FinalidadeAtestado":   r.FormValue("finalidade_atestado"),
		"DataAtestado":         formatDate(r.FormValue("data_atestado")),

		// Campos específicos da declaração
		"DataInicialAtendimento": formatDate(r.FormValue("data_inicial_atendimento")),
		"DataFinalAtendimento":   formatDate(r.FormValue("data_final_atendimento")),
		"FrequenciaAtendimento":  r.FormValue("frequencia_atendimento"),
		"DuracaoSessoes":         r.FormValue("duracao_sessoes"),
		"FinalidadeDeclaracao":   r.FormValue("finalidade_declaracao"),
		"MotivoSolicitacao":      r.FormValue("motivo_solicitacao"),

		// Campos específicos do laudo
		"Solicitante":      r.FormValue("solicitante"),
		"Finalidade":       r.FormValue("finalidade"),
		"DescricaoDemanda": r.FormValue("descricao_demanda"),
		"Procedimento":     r.FormValue("procedimento"),
		"Analise":          r.FormValue("analise"),
		"Conclusao":        r.FormValue("conclusao"),
		"Referencias":      strings.Split(r.FormValue("referencias"), "\n"),

		// Campos específicos do relatório
		"DataInicioAcompanhamento":  formatDate(r.FormValue("data_inicio_acompanhamento")),
		"DataFimAcompanhamento":     formatDate(r.FormValue("data_fim_acompanhamento")),
		"DemandaInicial":            r.FormValue("demanda_inicial"),
		"EvolucaoProcesso":          r.FormValue("evolucao_processo"),
		"RecursosUtilizados":        r.FormValue("recursos_utilizados"),
		"ResultadosEncaminhamentos": r.FormValue("resultados_encaminhamentos"),
		"FinalidadeRelatorio":       r.FormValue("finalidade_relatorio"),
		"DataRelatorio":             time.Now().Format("02/01/2006"),

		// Campos específicos da anamnese
		"QueixaPrincipal":       strings.Split(r.FormValue("queixa_principal"), "\n"),
		"HistoriaProblema":      r.FormValue("historia_problema"),
		"Sintomatologia":        r.FormValue("sintomatologia"),
		"HistoricoMedico":       r.FormValue("historico_medico"),
		"TempoTrabalhoAtual":    r.FormValue("tempo_trabalho_atual"),
		"HistoricoOcupacional":  r.FormValue("historico_ocupacional"),
		"HistoriaPsicossocial":  r.FormValue("historia_psicossocial"),
		"TratamentosAnteriores": r.FormValue("tratamentos_anteriores"),
		"HabitosAlimentacao":    r.FormValue("habitos_alimentacao"),
		"HabitosExercicios":     r.FormValue("habitos_exercicios"),
		"HabitosSono":           r.FormValue("habitos_sono"),
		"HabitosSubstancias":    r.FormValue("habitos_substancias"),
		"OutrasInformacoes":     r.FormValue("outras_informacoes"),
		"DataAvaliacao":         time.Now().Format("02/01/2006"),
	}

	// Formata as datas antes de passar para o template
	if data["DataFimTratamento"] != nil && data["DataFimTratamento"].(string) != "" {
		data["DataFimTratamento"] = formatDate(data["DataFimTratamento"].(string))
	}
	if data["DataInicioFerias"] != nil && data["DataInicioFerias"].(string) != "" {
		data["DataInicioFerias"] = formatDate(data["DataInicioFerias"].(string))
	}
	if data["DataFimFerias"] != nil && data["DataFimFerias"].(string) != "" {
		data["DataFimFerias"] = formatDate(data["DataFimFerias"].(string))
	}
	if data["DataInicio"] != nil && data["DataInicio"].(string) != "" {
		data["DataInicio"] = formatDate(data["DataInicio"].(string))
	}

	// Formata a data de assinatura com mês em português
	now := time.Now()
	mes := meses[now.Format("January")]
	data["DataAssinatura"] = fmt.Sprintf("%s, %s de %s de %s",
		data["DiaAssinatura"],
		now.Format("02"),
		mes,
		now.Format("2006"))

	// Renderizar template
	tmpl, err := template.New("document").Parse(string(content))
	if err != nil {
		log.Printf("Erro ao parsear template: %v", err)
		http.Error(w, "Erro ao parsear template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Erro ao renderizar template: %v", err)
		http.Error(w, "Erro ao renderizar template", http.StatusInternalServerError)
		return
	}
}

// SaveDocumentHandler salva o documento preenchido
func SaveDocumentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Método não permitido", http.StatusMethodNotAllowed)
		return
	}

	// Obter dados do formulário
	err := r.ParseForm()
	if err != nil {
		log.Printf("Erro ao processar formulário: %v", err)
		http.Error(w, "Erro ao processar formulário", http.StatusBadRequest)
		return
	}

	// Log de todos os campos recebidos
	log.Printf("Campos recebidos no formulário:")
	for key, values := range r.Form {
		log.Printf("Campo %s: %v", key, values)
	}

	// Obter tipo do documento e ID do paciente
	documentType := r.FormValue("value")
	patientID := r.FormValue("patient_id")

	if documentType == "" || patientID == "" {
		log.Printf("Campos obrigatórios não preenchidos - type: %v, patient: %v",
			documentType != "", patientID != "")
		http.Error(w, "Campos obrigatórios não preenchidos", http.StatusBadRequest)
		return
	}

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
		SELECT id, nome 
		FROM patients 
		WHERE id = $1 AND psicologo_id = $2`,
		patientID, psicologoID,
	).Scan(&patient.ID, &patient.Nome)
	if err != nil {
		log.Printf("Erro ao obter dados do paciente: %v", err)
		http.Error(w, "Erro ao obter dados do paciente", http.StatusInternalServerError)
		return
	}

	// Gerar nome do documento
	documentName := generateDocumentName(documentType, patient.Nome)
	log.Printf("Nome do documento gerado: %s", documentName)

	// Ler o template do documento
	templateFile := fmt.Sprintf("templates/documents/%s.html", documentType)
	content, err := os.ReadFile(templateFile)
	if err != nil {
		log.Printf("Erro ao ler template %s: %v", templateFile, err)
		http.Error(w, "Template não encontrado", http.StatusNotFound)
		return
	}

	// Preparar dados para o template
	data := map[string]interface{}{
		"PatientNome":          patient.Nome,
		"PatientNacionalidade": r.FormValue("nacionalidade"),
		"PatientEstadoCivil":   r.FormValue("estado_civil"),
		"PatientRG":            r.FormValue("rg"),
		"PatientCPF":           r.FormValue("cpf"),
		"PatientCidade":        r.FormValue("cidade"),
		"PatientRua":           r.FormValue("endereco"),
		"PatientNumero":        r.FormValue("numero"),
		"PatientCEP":           r.FormValue("cep"),
		"PatientTelefone":      fmt.Sprintf("(%s) %s", r.FormValue("ddd"), r.FormValue("telefone")),

		// Campos do formulário
		"Abordagem":           r.FormValue("abordagem"),
		"ValorSessao":         r.FormValue("valor_sessao"),
		"DiaSemana":           r.FormValue("dia_semana"),
		"HorarioSessao":       r.FormValue("horario_sessao"),
		"DataLimitePagamento": r.FormValue("data_limite_pagamento"),
		"MetodosPagamento":    r.FormValue("metodos_pagamento"),
		"DataInicioFerias":    formatDate(r.FormValue("data_inicio_ferias")),
		"DataFimFerias":       formatDate(r.FormValue("data_fim_ferias")),
		"DataFimTratamento":   formatDate(r.FormValue("data_fim_tratamento")),
		"NumeroFaltas":        r.FormValue("numero_faltas"),
		"DataInicio":          formatDate(r.FormValue("data_inicio")),
		"Frequencia":          r.FormValue("frequencia"),

		// Campos específicos do atestado
		"DataInicialAvaliacao": formatDate(r.FormValue("data_inicial_avaliacao")),
		"DataFinalAvaliacao":   formatDate(r.FormValue("data_final_avaliacao")),
		"NaturezaAvaliacao":    r.FormValue("natureza_avaliacao"),
		"EstadoPsicologico":    r.FormValue("estado_psicologico"),
		"DataValidade":         formatDate(r.FormValue("data_validade")),
		"FinalidadeAtestado":   r.FormValue("finalidade_atestado"),
		"DataAtestado":         formatDate(r.FormValue("data_atestado")),

		// Campos específicos da declaração
		"DataInicialAtendimento": formatDate(r.FormValue("data_inicial_atendimento")),
		"DataFinalAtendimento":   formatDate(r.FormValue("data_final_atendimento")),
		"FrequenciaAtendimento":  r.FormValue("frequencia_atendimento"),
		"DuracaoSessoes":         r.FormValue("duracao_sessoes"),
		"FinalidadeDeclaracao":   r.FormValue("finalidade_declaracao"),
		"MotivoSolicitacao":      r.FormValue("motivo_solicitacao"),

		// Campos específicos do laudo
		"Solicitante":      r.FormValue("solicitante"),
		"Finalidade":       r.FormValue("finalidade"),
		"DescricaoDemanda": r.FormValue("descricao_demanda"),
		"Procedimento":     r.FormValue("procedimento"),
		"Analise":          r.FormValue("analise"),
		"Conclusao":        r.FormValue("conclusao"),
		"Referencias":      strings.Split(r.FormValue("referencias"), "\n"),

		// Campos específicos do relatório
		"DataInicioAcompanhamento":  formatDate(r.FormValue("data_inicio_acompanhamento")),
		"DataFimAcompanhamento":     formatDate(r.FormValue("data_fim_acompanhamento")),
		"DemandaInicial":            r.FormValue("demanda_inicial"),
		"EvolucaoProcesso":          r.FormValue("evolucao_processo"),
		"RecursosUtilizados":        r.FormValue("recursos_utilizados"),
		"ResultadosEncaminhamentos": r.FormValue("resultados_encaminhamentos"),
		"FinalidadeRelatorio":       r.FormValue("finalidade_relatorio"),
		"DataRelatorio":             time.Now().Format("02/01/2006"),

		// Campos específicos da anamnese
		"QueixaPrincipal":       strings.Split(r.FormValue("queixa_principal"), "\n"),
		"HistoriaProblema":      r.FormValue("historia_problema"),
		"Sintomatologia":        r.FormValue("sintomatologia"),
		"HistoricoMedico":       r.FormValue("historico_medico"),
		"TempoTrabalhoAtual":    r.FormValue("tempo_trabalho_atual"),
		"HistoricoOcupacional":  r.FormValue("historico_ocupacional"),
		"HistoriaPsicossocial":  r.FormValue("historia_psicossocial"),
		"TratamentosAnteriores": r.FormValue("tratamentos_anteriores"),
		"HabitosAlimentacao":    r.FormValue("habitos_alimentacao"),
		"HabitosExercicios":     r.FormValue("habitos_exercicios"),
		"HabitosSono":           r.FormValue("habitos_sono"),
		"HabitosSubstancias":    r.FormValue("habitos_substancias"),
		"OutrasInformacoes":     r.FormValue("outras_informacoes"),
		"DataAvaliacao":         time.Now().Format("02/01/2006"),
	}

	// Renderizar template
	var buffer bytes.Buffer
	tmpl, err := template.New("document").Parse(string(content))
	if err != nil {
		log.Printf("Erro ao parsear template: %v", err)
		http.Error(w, "Erro ao parsear template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(&buffer, data)
	if err != nil {
		log.Printf("Erro ao renderizar template: %v", err)
		http.Error(w, "Erro ao renderizar template", http.StatusInternalServerError)
		return
	}

	// Determinar se requer assinatura baseado no tipo
	requerAssinatura := strings.HasPrefix(documentType, "contracts/")

	// Inserir documento
	_, err = db.Exec(`
		INSERT INTO documents (
			psicologo_id, paciente_id, tipo, nome, 
			conteudo, requer_assinatura
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		psicologoID,
		patientID,
		documentType,
		documentName,
		buffer.String(),
		requerAssinatura,
	)

	if err != nil {
		log.Printf("Erro ao salvar documento: %v", err)
		w.Write([]byte(`<div class="alert alert-danger alert-dismissible fade show" role="alert">
			<i class="bi bi-exclamation-triangle-fill me-2"></i>
			Erro ao salvar documento. Por favor, tente novamente.
			<button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Fechar"></button>
		</div>`))
		return
	}

	// Retornar mensagem de sucesso
	w.Write([]byte(`<div class="alert alert-success alert-dismissible fade show" role="alert">
		<i class="bi bi-check-circle-fill me-2"></i>
		Documento salvo com sucesso!
		<button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Fechar"></button>
	</div>
	<script>
		setTimeout(function() {
			htmx.ajax('GET', '/patients/' + document.querySelector('[name="patient_id"]').value, {target: '#content-area'});
		}, 2000);
	</script>`))
}

// generateDocumentName gera o nome do documento baseado no tipo e nome do paciente
func generateDocumentName(documentType, patientName string) string {
	// Extrair categoria e tipo do documento
	parts := strings.Split(documentType, "/")
	var categoria, tipo string
	if len(parts) == 2 {
		switch parts[0] {
		case "contracts":
			categoria = "Contrato"
			if parts[1] == "presencial" {
				tipo = "Presencial"
			} else if parts[1] == "online" {
				tipo = "Online"
			}
		case "psychological-documents":
			categoria = "Documento"
			switch parts[1] {
			case "anamnese":
				tipo = "Anamnese"
			case "atestado":
				tipo = "Atestado"
			case "declaracao":
				tipo = "Declaração"
			case "laudo":
				tipo = "Laudo"
			case "relatorio":
				tipo = "Relatório"
			}
		}
	}

	// Formatar data atual
	now := time.Now()
	data := now.Format("02/01/2006")

	return fmt.Sprintf("%s - %s - %s - %s", categoria, tipo, patientName, data)
}

// DocumentEditorHandler renderiza a página do editor de documentos
func DocumentEditorHandler(w http.ResponseWriter, r *http.Request) {
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

	// Obter dados do paciente do contexto atual
	vars := mux.Vars(r)
	patientID := vars["id"]

	var patient Patient
	err = db.QueryRow(`
		SELECT id, nome 
		FROM patients 
		WHERE id = $1 AND psicologo_id = $2`,
		patientID, psicologoID,
	).Scan(&patient.ID, &patient.Nome)

	if err != nil {
		log.Printf("Erro ao obter dados do paciente: %v", err)
		http.Error(w, "Erro ao obter dados do paciente", http.StatusInternalServerError)
		return
	}

	// Preparar dados para o template
	data := map[string]interface{}{
		"Patient": patient,
	}

	// Se for uma requisição HTMX
	if r.Header.Get("HX-Request") == "true" {
		tmpl := template.Must(template.ParseFiles("templates/view/partials/documents_editor.html"))
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
		"templates/view/partials/documents_editor.html",
	))
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Erro ao renderizar template: %v", err)
		http.Error(w, "Erro ao renderizar template", http.StatusInternalServerError)
	}
}

// PersonalizedDocumentEditorHandler renderiza a página do editor personalizado de documentos
func PersonalizedDocumentEditorHandler(w http.ResponseWriter, r *http.Request) {
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

	// Obter dados do paciente do contexto atual
	vars := mux.Vars(r)
	patientID := vars["id"]

	var patient Patient
	err = db.QueryRow(`
		SELECT id, nome 
		FROM patients 
		WHERE id = $1 AND psicologo_id = $2`,
		patientID, psicologoID,
	).Scan(&patient.ID, &patient.Nome)

	if err != nil {
		log.Printf("Erro ao obter dados do paciente: %v", err)
		http.Error(w, "Erro ao obter dados do paciente", http.StatusInternalServerError)
		return
	}

	// Preparar dados para o template
	data := map[string]interface{}{
		"Patient": patient,
	}

	// Se for uma requisição HTMX
	if r.Header.Get("HX-Request") == "true" {
		tmpl := template.Must(template.ParseFiles("templates/view/partials/documents_personalized_editor.html"))
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
		"templates/view/partials/documents_personalized_editor.html",
	))
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Erro ao renderizar template: %v", err)
		http.Error(w, "Erro ao renderizar template", http.StatusInternalServerError)
	}
}
