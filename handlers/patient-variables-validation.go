package handlers

import (
	"fmt"
	"regexp"
	"strings"
)

// Constantes para validação de pacientes
const (
	MaxNomeLength     = 255
	MaxEnderecoLength = 255
	MaxBairroLength   = 255
	MaxCidadeLength   = 255
	DDDLength         = 3
	MaxTelLength      = 20
	CPFLength         = 14 // Incluindo pontos e traço
	CEPLength         = 9  // Incluindo traço
)

// ValidatePatientNome valida o nome do paciente
func ValidatePatientNome(nome string) error {
	nome = strings.TrimSpace(nome)

	if nome == "" {
		return fmt.Errorf("nome não pode estar vazio")
	}

	if len(nome) > MaxNomeLength {
		return fmt.Errorf("nome não pode ter mais que %d caracteres", MaxNomeLength)
	}

	// Verifica se contém apenas letras, espaços e acentos
	match, _ := regexp.MatchString(`^[a-zA-ZÀ-ÿ\s]+$`, nome)
	if !match {
		return fmt.Errorf("nome deve conter apenas letras")
	}

	return nil
}

// ValidatePatientCPF valida o CPF do paciente
func ValidatePatientCPF(cpf string) error {
	if cpf == "" {
		return nil // CPF é opcional
	}

	cpf = strings.TrimSpace(cpf)

	if len(cpf) != CPFLength {
		return fmt.Errorf("CPF deve ter %d caracteres incluindo pontos e traço", CPFLength)
	}

	// Verifica formato XXX.XXX.XXX-XX
	match, _ := regexp.MatchString(`^\d{3}\.\d{3}\.\d{3}-\d{2}$`, cpf)
	if !match {
		return fmt.Errorf("formato de CPF inválido, deve ser XXX.XXX.XXX-XX")
	}

	return nil
}

// ValidatePatientDDD valida o DDD do paciente
func ValidatePatientDDD(ddd string) error {
	ddd = strings.TrimSpace(ddd)

	if len(ddd) != DDDLength {
		return fmt.Errorf("DDD deve ter %d dígitos", DDDLength)
	}

	// Verifica se contém apenas números
	match, _ := regexp.MatchString(`^\d{3}$`, ddd)
	if !match {
		return fmt.Errorf("DDD deve conter apenas números")
	}

	return nil
}

// ValidatePatientTelefone valida o telefone do paciente
func ValidatePatientTelefone(telefone string) error {
	telefone = strings.TrimSpace(telefone)

	if len(telefone) > MaxTelLength {
		return fmt.Errorf("telefone não pode ter mais que %d caracteres", MaxTelLength)
	}

	// Verifica se contém apenas números
	match, _ := regexp.MatchString(`^\d+$`, telefone)
	if !match {
		return fmt.Errorf("telefone deve conter apenas números")
	}

	// Verifica se tem pelo menos 8 dígitos
	if len(telefone) < 8 {
		return fmt.Errorf("telefone deve ter pelo menos 8 dígitos")
	}

	return nil
}

// ValidatePatientEmail valida o email do paciente
func ValidatePatientEmail(email string) error {
	email = strings.TrimSpace(email)

	if email == "" {
		return fmt.Errorf("email não pode estar vazio")
	}

	if len(email) > MaxEmailLength {
		return fmt.Errorf("email não pode ter mais que %d caracteres", MaxEmailLength)
	}

	// Verifica formato
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("formato de email inválido")
	}

	return nil
}

// ValidatePatientSexo valida o sexo do paciente
func ValidatePatientSexo(sexo string) error {
	sexo = strings.TrimSpace(sexo)

	if sexo == "" {
		return fmt.Errorf("sexo não pode estar vazio")
	}

	if sexo != "M" && sexo != "F" && sexo != "O" {
		return fmt.Errorf("sexo deve ser M, F ou O")
	}

	return nil
}

// ValidatePatientCEP valida o CEP do paciente
func ValidatePatientCEP(cep string) error {
	cep = strings.TrimSpace(cep)

	if len(cep) != CEPLength {
		return fmt.Errorf("CEP deve ter %d caracteres incluindo traço", CEPLength)
	}

	// Verifica formato XXXXX-XXX
	match, _ := regexp.MatchString(`^\d{5}-\d{3}$`, cep)
	if !match {
		return fmt.Errorf("formato de CEP inválido, deve ser XXXXX-XXX")
	}

	return nil
}

// ValidatePatientEstado valida o estado do paciente
func ValidatePatientEstado(estado string) error {
	estado = strings.TrimSpace(estado)

	if estado == "" {
		return fmt.Errorf("estado não pode estar vazio")
	}

	if !validUFs[estado] {
		return fmt.Errorf("estado inválido")
	}

	return nil
}

// ValidatePatientEndereco valida o endereço do paciente
func ValidatePatientEndereco(endereco string) error {
	endereco = strings.TrimSpace(endereco)

	if endereco == "" {
		return fmt.Errorf("endereço não pode estar vazio")
	}

	if len(endereco) > MaxEnderecoLength {
		return fmt.Errorf("endereço não pode ter mais que %d caracteres", MaxEnderecoLength)
	}

	return nil
}

// ValidatePatientBairro valida o bairro do paciente
func ValidatePatientBairro(bairro string) error {
	bairro = strings.TrimSpace(bairro)

	if bairro == "" {
		return fmt.Errorf("bairro não pode estar vazio")
	}

	if len(bairro) > MaxBairroLength {
		return fmt.Errorf("bairro não pode ter mais que %d caracteres", MaxBairroLength)
	}

	return nil
}

// ValidatePatientCidade valida a cidade do paciente
func ValidatePatientCidade(cidade string) error {
	cidade = strings.TrimSpace(cidade)

	if cidade == "" {
		return fmt.Errorf("cidade não pode estar vazia")
	}

	if len(cidade) > MaxCidadeLength {
		return fmt.Errorf("cidade não pode ter mais que %d caracteres", MaxCidadeLength)
	}

	return nil
}

// ValidatePatientNumero valida o número do endereço do paciente
func ValidatePatientNumero(numero string) error {
	numero = strings.TrimSpace(numero)

	if numero == "" {
		return fmt.Errorf("número não pode estar vazio")
	}

	if len(numero) > 10 {
		return fmt.Errorf("número não pode ter mais que 10 caracteres")
	}

	return nil
}
