package handlers

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateToken gera um token seguro para reset de senha
// Recebe:
// - w: o ResponseWriter para enviar a resposta
// - r: o Request para obter os dados do formulário
// Retorna:
// - token: o token gerado
// - error: um erro se ocorrer
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
