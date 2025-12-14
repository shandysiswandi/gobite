package pkgmail

import (
	"context"
	"errors"
	"fmt"
	"net/smtp"
	"strings"
)

var (
	ErrSMTPHostPortRequired = errors.New("smtp host and port are required")
	ErrSMTPNoRecipients     = errors.New("no recipients provided")
	ErrSMTPNoSender         = errors.New("no sender provided")
)

type SMTP struct {
	addr        string
	host        string
	defaultFrom string
	auth        smtp.Auth
}

type SMTPConfig struct {
	Host     string // e.g., "localhost"
	Port     int    // e.g., 1025 for MailHog
	Username string
	Password string
	From     string // default From if Message.From is empty
}

func NewSMTP(cfg SMTPConfig) (*SMTP, error) {
	if cfg.Host == "" || cfg.Port == 0 {
		return nil, ErrSMTPHostPortRequired
	}

	var auth smtp.Auth
	if cfg.Username != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}

	return &SMTP{
		addr:        fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		host:        cfg.Host,
		defaultFrom: cfg.From,
		auth:        auth,
	}, nil
}

func (s *SMTP) Send(ctx context.Context, msg Message) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	recipients := append([]string{}, msg.To...)
	recipients = append(recipients, msg.Cc...)
	recipients = append(recipients, msg.Bcc...)

	if len(recipients) == 0 {
		return ErrSMTPNoRecipients
	}

	from := msg.From
	if from == "" {
		from = s.defaultFrom
	}
	if from == "" {
		return ErrSMTPNoSender
	}

	body, contentType := buildBody(msg)

	var headers []string
	headers = append(headers, fmt.Sprintf("From: %s", from))
	headers = append(headers, fmt.Sprintf("To: %s", strings.Join(msg.To, ", ")))
	if len(msg.Cc) > 0 {
		headers = append(headers, fmt.Sprintf("Cc: %s", strings.Join(msg.Cc, ", ")))
	}
	headers = append(headers, fmt.Sprintf("Subject: %s", msg.Subject))
	headers = append(headers, "MIME-Version: 1.0")
	headers = append(headers, fmt.Sprintf("Content-Type: %s", contentType))

	raw := strings.Join(headers, "\r\n") + "\r\n\r\n" + body

	if err := ctx.Err(); err != nil {
		return err
	}

	return smtp.SendMail(s.addr, s.auth, from, recipients, []byte(raw))
}

func (s *SMTP) Close() error {
	return nil
}

func buildBody(msg Message) (body string, contentType string) {
	if msg.HTMLBody != "" && msg.TextBody != "" {
		boundary := "gobite-boundary"
		var sb strings.Builder
		sb.WriteString("This is a multipart message in MIME format.\r\n")
		sb.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		sb.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		sb.WriteString("\r\n")
		sb.WriteString(msg.TextBody)
		sb.WriteString("\r\n")
		sb.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		sb.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		sb.WriteString("\r\n")
		sb.WriteString(msg.HTMLBody)
		sb.WriteString("\r\n")
		sb.WriteString(fmt.Sprintf("--%s--", boundary))
		return sb.String(), fmt.Sprintf("multipart/alternative; boundary=%s", boundary)
	}

	if msg.HTMLBody != "" {
		return msg.HTMLBody, "text/html; charset=UTF-8"
	}

	return msg.TextBody, "text/plain; charset=UTF-8"
}
