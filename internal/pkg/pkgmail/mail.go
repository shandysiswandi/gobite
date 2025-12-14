// Package pkgmail defines contracts for sending email messages.
package pkgmail

import (
	"context"
	"io"
)

// Message represents an email payload.
type Message struct {
	From     string   // optional explicit sender; fallback depends on implementation
	To       []string // required recipients
	Cc       []string
	Bcc      []string
	Subject  string
	TextBody string // plain-text body; preferred when HTMLBody is empty
	HTMLBody string // optional HTML alternative
}

// Mail abstracts an email provider.
type Mail interface {
	io.Closer
	// Send dispatches the given message using the underlying provider.
	Send(ctx context.Context, msg Message) error
}
