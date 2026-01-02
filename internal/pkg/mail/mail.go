package mail

import (
	"context"
	"io"
)

// Message represents an email payload.
//
// Fields are intentionally provider-agnostic so they can be sent using SMTP or
// other delivery mechanisms.
type Message struct {
	// From is an optional explicit sender; fallback depends on implementation.
	From string
	// To lists required recipients.
	To []string
	// Cc lists carbon copy recipients.
	Cc []string
	// Bcc lists blind carbon copy recipients.
	Bcc []string
	// Subject is the email subject line.
	Subject string
	// TextBody is the plain-text body; preferred when HTMLBody is empty.
	TextBody string
	// HTMLBody is the optional HTML body.
	HTMLBody string
}

// Mail abstracts an email provider (SMTP, third-party API, etc).
type Mail interface {
	io.Closer
	// Send dispatches the given message using the underlying provider.
	Send(ctx context.Context, msg Message) error
}
