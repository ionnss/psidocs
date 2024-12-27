package handlers

type ContractData struct {
	// Dados do Contratante
	ClienteNome          string
	ClienteNacionalidade string
	ClienteEstadoCivil   string
	ClienteRG            string
	ClienteCPF           string
	ClienteCidade        string
	ClienteRua           string
	ClienteNumero        string
	ClienteCEP           string
	ClienteTelefone      string

	// Dados do Contratado
	PsicologoNome          string
	PsicologoNacionalidade string
	PsicologoEstadoCivil   string
	PsicologoCRP           string
	PsicologoRG            string
	PsicologoCPF           string
	PsicologoCidade        string
	PsicologoRua           string
	PsicologoCEP           string
	PsicologoNumero        string
	PsicologoTelefone      string

	// Dados do Contrato
	Abordagem           string
	DataAssinatura      string
	DataFimTratamento   string
	DiaSemana           string
	HorarioSessao       string
	ValorSessao         string
	DataLimitePagamento string
	MetodosPagamento    string
	DataInicioFerias    string
	DataFimFerias       string
	NumeroFaltas        string

	// Dados da Assinatura
	DiaAssinatura    string
	DiaMesAssinatura string
	MesAssinatura    string
	AnoAssinatura    string
}
