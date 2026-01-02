package mail

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/smtp"
	"strings"
)

var (
	// ErrSMTPHostPortRequired is returned when Host/Port are missing.
	ErrSMTPHostPortRequired = errors.New("smtp host and port are required")
	// ErrSMTPNoRecipients is returned when To/Cc/Bcc are all empty.
	ErrSMTPNoRecipients = errors.New("no recipients provided")
	// ErrSMTPNoSender is returned when both Message.From and the configured default From are empty.
	ErrSMTPNoSender = errors.New("no sender provided")
)

// SMTP is a Mail implementation backed by net/smtp.
type SMTP struct {
	addr        string
	host        string
	defaultFrom string
	auth        smtp.Auth
}

// SMTPConfig configures the SMTP implementation.
type SMTPConfig struct {
	// Host is the SMTP server hostname.
	Host string
	// Port is the SMTP server port.
	Port int
	// Username is the SMTP authentication username.
	Username string
	// Password is the SMTP authentication password.
	Password string
	// From is the default sender when Message.From is empty.
	From string
}

// NewSMTP constructs an SMTP mail sender.
func NewSMTP(cfg SMTPConfig) (*SMTP, error) {
	if cfg.Host == "" || cfg.Port == 0 {
		return nil, ErrSMTPHostPortRequired
	}

	var auth smtp.Auth
	if cfg.Username != "" && cfg.Password != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}

	return &SMTP{
		addr:        fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		host:        cfg.Host,
		defaultFrom: cfg.From,
		auth:        auth,
	}, nil
}

// Send delivers a message over SMTP.
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

// Close implements io.Closer for interface compatibility.
func (s *SMTP) Close() error {
	return nil
}

func buildBody(msg Message) (body string, contentType string) {
	if msg.HTMLBody != "" && msg.TextBody != "" {
		boundary := multipartBoundary()
		var sb strings.Builder
		sb.WriteString("This is a multipart message in MIME format.\r\n")
		fmt.Fprintf(&sb, "--%s\r\n", boundary)
		sb.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		sb.WriteString("\r\n")
		sb.WriteString(msg.TextBody)
		sb.WriteString("\r\n")
		fmt.Fprintf(&sb, "--%s\r\n", boundary)
		sb.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		sb.WriteString("\r\n")
		sb.WriteString(msg.HTMLBody)
		sb.WriteString("\r\n")
		fmt.Fprintf(&sb, "--%s--", boundary)
		return sb.String(), fmt.Sprintf("multipart/alternative; boundary=%s", boundary)
	}

	if msg.HTMLBody != "" {
		return msg.HTMLBody, "text/html; charset=UTF-8"
	}

	return msg.TextBody, "text/plain; charset=UTF-8"
}

func multipartBoundary() string {
	var b [12]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "gobite-boundary-fallback"
	}
	return "gobite-boundary-" + hex.EncodeToString(b[:])
}
