// Package handlers é um pacote que contém os handlers da aplicação
package handlers

import (
	"bytes"
	"fmt"
	"net/smtp"
	"os"
	"text/template"
)

// SendEmail sends an email to the given address with the given subject and data
//
// The data is used to fill in the template.
//
// Receives:
// - to: the email address to send the email to (user email)
// - subject: the subject of the email
// - data: the data to fill in the template with dynamic content
//
// Returns:
// - error: an error if the email fails to send or if parameters are invalid
func SendEmail(to, subject string, data map[string]string) error {
	// Validate parameters
	if to == "" || subject == "" || data == nil {
		return fmt.Errorf("invalid parameters: email, subject and data are required")
	}

	// Get env variables
	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("EMAIL_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	// Validate env variables
	if from == "" || password == "" || smtpHost == "" || smtpPort == "" {
		return fmt.Errorf("missing email configuration in environment variables")
	}

	// Create authentication
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Parse HTML template
	tmpl, err := template.ParseFiles("templates/email/email-sender.html")
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	// Execute the template with data
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	// Create message with proper headers
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n%s",
		from, to, subject, body.String()))

	// Send email
	err = smtp.SendMail(fmt.Sprintf("%s:%s", smtpHost, smtpPort), auth, from, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}
	return nil
}

// Exemplo de uso
//
//data := map[string]string{
//    "Title": "Alteração de Email",
//    "Message": "Seu email foi alterado com sucesso...",
//    "ButtonText": "Acessar Dashboard", // opcional
//    "ButtonLink": "https://psidocs.com/dashboard", // opcional
//}
//SendEmail(userEmail, "Confirmação de Alteração", data)
