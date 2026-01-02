// Package mail defines the contracts for sending email messages.
//
// The main purpose is to keep the rest of the application independent from a
// specific email provider. Handlers and use cases work with the Mail interface
// and Message payload; the concrete delivery mechanism (SMTP, API provider, etc)
// is implemented elsewhere in this package.
package mail
