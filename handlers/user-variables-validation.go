package handlers

import (
	"fmt"
	"regexp"
	"strings"
)

// Constantes para validação
const (
	MaxEmailLength    = 255
	MinPasswordLength = 10
	MaxPasswordLength = 72 // Limite do bcrypt
	CRPLength         = 7
)

// UFs válidas do Brasil
var validUFs = map[string]bool{
	"AC": true, "AL": true, "AP": true, "AM": true, "BA": true,
	"CE": true, "DF": true, "ES": true, "GO": true, "MA": true,
	"MT": true, "MS": true, "MG": true, "PA": true, "PB": true,
	"PR": true, "PE": true, "PI": true, "RJ": true, "RN": true,
	"RS": true, "RO": true, "RR": true, "SC": true, "SP": true,
	"SE": true, "TO": true,
}

// ValidateCRP valida o formato do CRP
func ValidateCRP(crp string) error {
	// Remove espaços
	crp = strings.TrimSpace(crp)

	// Verifica comprimento
	if len(crp) != CRPLength {
		return fmt.Errorf("CRP deve ter exatamente %d caracteres", CRPLength)
	}

	// Verifica formato (5 números + 2 letras)
	match, _ := regexp.MatchString(`^\d{5}[A-Z]{2}$`, crp)
	if !match {
		return fmt.Errorf("formato de CRP inválido, deve ser 5 números seguidos de UF em maiúsculo")
	}

	// Verifica se UF é válida
	uf := crp[5:]
	if !validUFs[uf] {
		return fmt.Errorf("UF inválida no CRP")
	}

	return nil
}

// ValidateEmail valida o formato do email
func ValidateEmail(email string) error {
	// Remove espaços
	email = strings.TrimSpace(email)

	// Verifica comprimento
	if len(email) == 0 {
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

// ValidateChave valida a chave de acesso
func ValidateChave(chave string) error {
	// Remove espaços
	chave = strings.TrimSpace(chave)

	// Verifica comprimento
	if len(chave) < MinPasswordLength {
		return fmt.Errorf("chave deve ter no mínimo %d caracteres", MinPasswordLength)
	}
	if len(chave) > MaxPasswordLength {
		return fmt.Errorf("chave não pode ter mais que %d caracteres", MaxPasswordLength)
	}

	// Verifica complexidade
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(chave)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(chave)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(chave)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*]`).MatchString(chave)

	if !hasUpper {
		return fmt.Errorf("chave deve conter pelo menos uma letra maiúscula")
	}
	if !hasLower {
		return fmt.Errorf("chave deve conter pelo menos uma letra minúscula")
	}
	if !hasNumber {
		return fmt.Errorf("chave deve conter pelo menos um número")
	}
	if !hasSpecial {
		return fmt.Errorf("chave deve conter pelo menos um caractere especial (!@#$%%^&*)")
	}

	// Verifica caracteres permitidos
	validChars := regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*]+$`)
	if !validChars.MatchString(chave) {
		return fmt.Errorf("chave contém caracteres não permitidos")
	}

	return nil
}

// SanitizeInput remove caracteres perigosos de uma string
func SanitizeInput(input string) string {
	// Remove espaços extras
	input = strings.TrimSpace(input)

	// Remove caracteres de controle
	input = regexp.MustCompile(`[\x00-\x1F\x7F]`).ReplaceAllString(input, "")

	return input
}
