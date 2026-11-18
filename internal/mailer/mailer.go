package mailer

import (
	"fmt"
	"log"
	"net/smtp"

	"laboratory-management/config"
)

// SendActivationEmail sends an account activation link to the invited user.
// When SMTP_HOST is not configured, it logs the activation URL for local development.
func SendActivationEmail(toEmail, userName, activationURL string) error {
	cfg := config.Load()
	subject := "Convite de acesso — Laboratório de Informática CETEP"
	body := fmt.Sprintf(
		"Olá, %s!\n\n"+
			"Você foi cadastrado(a) no Sistema de Agendamento do Laboratório de Informática (CETEP Bacia do Paramirim).\n\n"+
			"Para finalizar seu cadastro e definir sua senha de acesso, abra o link abaixo:\n"+
			"%s\n\n"+
			"Este link é válido por 48 horas.\n",
		userName,
		activationURL,
	)

	if cfg.SMTPHost == "" {
		log.Printf("[MAILER DEV] Para: %s | Assunto: %s | Link de ativação: %s", toEmail, subject, activationURL)
		return nil
	}

	addr := fmt.Sprintf("%s:%s", cfg.SMTPHost, cfg.SMTPPort)
	msg := []byte(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		cfg.SMTPFrom, toEmail, subject, body))

	var auth smtp.Auth
	if cfg.SMTPUser != "" {
		auth = smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost)
	}

	if err := smtp.SendMail(addr, auth, cfg.SMTPFrom, []string{toEmail}, msg); err != nil {
		return fmt.Errorf("error sending activation email: %w", err)
	}
	return nil
}
