package services

import (
	"fmt"
	"log"
	"net/smtp"
)

type EmailService interface {
	SendPasswordResetEmail(toEmail, resetLink string) error
}

type emailService struct {
	smtpHost     string
	smtpPort     int
	smtpUsername string
	smtpPassword string
	fromEmail    string
}

func NewEmailService(smtpHost string, smtpPort int, smtpUsername, smtpPassword, fromEmail string) EmailService {
	return &emailService{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUsername: smtpUsername,
		smtpPassword: smtpPassword,
		fromEmail:    fromEmail,
	}
}

func (s *emailService) SendPasswordResetEmail(toEmail, resetLink string) error {
	subject := "Reset Password - E-Meeting"
	body := fmt.Sprintf(`
		Hello,
		
		You have requested to reset your password. Please click the link below to reset your password:
		
		%s
		
		This link will expire in 24 hours.
		
		If you did not request this password reset, please ignore this email.
		
		Best regards,
		E-Meeting Team
	`, resetLink)

	message := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", s.fromEmail, toEmail, subject, body)

	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)
	addr := fmt.Sprintf("%s:%d", s.smtpHost, s.smtpPort)

	err := smtp.SendMail(addr, auth, s.fromEmail, []string{toEmail}, []byte(message))
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}

	return nil
}
