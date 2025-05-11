package services

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/smtp"
	"path/filepath"
	"time"
)

type EmailService interface {
	SendPasswordResetEmail(toEmail, resetLink string) error
}

type emailService struct {
	smtpHost           string
	smtpPort           int
	smtpUsername       string
	smtpPassword       string
	fromEmail          string
	templatePath       string
	logoURL            string
	timeOutDuration    int
	insecureSkipVerify bool
	useTLS             bool
}

func NewEmailService(smtpHost string, smtpPort, timeOutDuration int, smtpUsername, smtpPassword, fromEmail, templatePath, logoURL string, insecureSkipVerify, useTLS bool) EmailService {
	return &emailService{
		smtpHost:           smtpHost,
		smtpPort:           smtpPort,
		smtpUsername:       smtpUsername,
		smtpPassword:       smtpPassword,
		fromEmail:          fromEmail,
		templatePath:       templatePath,
		logoURL:            logoURL,
		timeOutDuration:    timeOutDuration,
		useTLS:             useTLS,             // Use TLS if port is 465
		insecureSkipVerify: insecureSkipVerify, // Set to true if you want to skip TLS verification
	}
}

func (s *emailService) SendPasswordResetEmail(toEmail, resetLink string) error {
	log.Printf("Sending password reset email to %s", toEmail)

	subject := "Reset Password - E-Meeting"

	htmlBody, err := s.renderEmailTemplate(resetLink)
	if err != nil {
		log.Printf("Template rendering error: %v", err)
		return err
	}

	message := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\n"+
			"MIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		s.fromEmail, toEmail, subject, htmlBody)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.timeOutDuration)*time.Second)
	defer cancel()

	client, err := s.connectSMTP(ctx)
	if err != nil {
		return err
	}
	defer client.Quit()

	auth := smtp.PlainAuth("", s.smtpUsername, s.smtpPassword, s.smtpHost)
	if err := client.Auth(auth); err != nil {
		log.Printf("Auth error: %v", err)
		return err
	}

	if err := client.Mail(s.fromEmail); err != nil {
		return err
	}
	if err := client.Rcpt(toEmail); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		log.Printf("Close error: %v", err)
		return err
	}
	log.Printf("Password reset email successfully sent to %s", toEmail)
	return nil
}

func (s *emailService) connectSMTP(ctx context.Context) (*smtp.Client, error) {
	address := fmt.Sprintf("%s:%d", s.smtpHost, s.smtpPort)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: s.insecureSkipVerify,
		ServerName:         s.smtpHost,
	}

	var conn net.Conn
	var err error

	if s.useTLS {
		log.Println("Connecting via TLS (port 465)")
		conn, err = tls.Dial("tcp", address, tlsConfig)
		if err != nil {
			log.Printf("TLS Dial error: %v", err)
			return nil, err
		}
		log.Println("Connected via TLS (port 465)")
	} else {
		dialer := &net.Dialer{}
		log.Println("Connecting via plain (STARTTLS route)")
		conn, err = dialer.DialContext(ctx, "tcp", address)
		if err != nil {
			log.Printf("DialContext error: %v", err)
			return nil, err
		}
		log.Println("Connected via plain (STARTTLS route)")
	}

	client, err := smtp.NewClient(conn, s.smtpHost)
	if err != nil {
		log.Printf("SMTP client error: %v", err)
		return nil, err
	}

	if !s.useTLS {
		if err := client.StartTLS(tlsConfig); err != nil {
			log.Printf("StartTLS error: %v", err)
			return nil, err
		}
		log.Println("Upgraded to TLS via STARTTLS")
	}

	return client, nil
}

func (s *emailService) renderEmailTemplate(resetLink string) (string, error) {
	tmpl, err := template.ParseFiles(filepath.Clean(s.templatePath))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, map[string]string{
		"ResetLink": resetLink,
		"LogoURL":   s.logoURL,
		"Year":      fmt.Sprintf("%d", time.Now().Year()),
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
