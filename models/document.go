package models

import "time"

// Document representa a estrutura de dados de um documento
type Document struct {
	ID               int
	PsicologoID      int
	PacienteID       int
	Tipo             string
	Nome             string
	Descricao        string
	Conteudo         string
	RequerAssinatura bool
	PatientName      string // Nome do paciente (usado para exibição)
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
