package provider

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

// SMTPProvider implements email sending via SMTP
type SMTPProvider struct {
	host     string
	port     string
	username string
	password string
	from     string
	auth     smtp.Auth
}

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

func NewSMTPProvider(cfg SMTPConfig) (*SMTPProvider, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("email: smtp host is required")
	}
	if cfg.Port == "" {
		cfg.Port = "587" // Default to TLS port
	}
	if cfg.From == "" {
		return nil, fmt.Errorf("email: from address is required")
	}

	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)

	return &SMTPProvider{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.Username,
		password: cfg.Password,
		from:     cfg.From,
		auth:     auth,
	}, nil
}

func (p *SMTPProvider) Send(ctx context.Context, to, subject, body string, metadata map[string]string) error {
	addr := fmt.Sprintf("%s:%s", p.host, p.port)

	// Build email message
	msg := []byte(fmt.Sprintf("From: %s\r\n", p.from) +
		fmt.Sprintf("To: %s\r\n", to) +
		fmt.Sprintf("Subject: %s\r\n", subject) +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		body + "\r\n")

	// Add metadata as headers if provided
	if len(metadata) > 0 {
		headerBytes := make([]byte, 0)
		for k, v := range metadata {
			// Convert key to header format (X-Key-Name)
			headerKey := strings.ToUpper(k[:1]) + k[1:]
			headerBytes = append(headerBytes, []byte(fmt.Sprintf("X-%s: %s\r\n", headerKey, v))...)
		}
		msg = append(headerBytes, msg...)
	}

	err := smtp.SendMail(addr, p.auth, p.from, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("email: smtp send: %w", err)
	}

	return nil
}

func (p *SMTPProvider) Name() string {
	return "smtp"
}

