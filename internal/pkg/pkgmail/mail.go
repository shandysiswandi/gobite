package pkgmail

import (
	"context"
	"io"
)

// Message represents an email payload.
//
// Fields are intentionally provider-agnostic so they can be sent using SMTP or
// other delivery mechanisms.
type Message struct {
	From     string   // optional explicit sender; fallback depends on implementation
	To       []string // required recipients
	Cc       []string
	Bcc      []string
	Subject  string
	TextBody string // plain-text body; preferred when HTMLBody is empty
	HTMLBody string // optional HTML alternative
}

// Mail abstracts an email provider (SMTP, third-party API, etc).
type Mail interface {
	io.Closer
	// Send dispatches the given message using the underlying provider.
	Send(ctx context.Context, msg Message) error
}
