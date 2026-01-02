package otp

import (
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// OTP defines the contract for TOTP operations.
type OTP interface {
	// Generate creates a secret and provisioning URI for an account name.
	Generate(accountName string) (secret string, uri string, err error)
	// Validate checks whether a code is valid at the given time.
	Validate(code, secret string, at time.Time) bool
	// GenerateCode creates a TOTP code for the given secret and time.
	GenerateCode(secret string, at time.Time) (string, error)
}

// TOTP implements OTP using the Time-based One-Time Password algorithm.
type TOTP struct {
	issuer string
	period uint
	skew   uint
	digits otp.Digits
}

// NewTOTP constructs a TOTP instance with sensible defaults.
//
// If digits is not 6 or 8, it falls back to 6 digits. If period is 0, it uses
// the common 30-second period.
func NewTOTP(issuer string, period, skew uint, digits otp.Digits) *TOTP {
	if digits != otp.DigitsSix && digits != otp.DigitsEight {
		digits = otp.DigitsSix
	}

	if period == 0 {
		period = 30
	}

	if skew == 0 {
		skew = 1
	}

	return &TOTP{
		issuer: issuer,
		period: period,
		skew:   skew,
		digits: digits,
	}
}

// Generate creates a secret and provisioning URI for an account name.
func (o *TOTP) Generate(accountName string) (secret string, uri string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      o.issuer,
		AccountName: accountName,
		Period:      o.period,
		SecretSize:  20, // RFC 4226/6238 recommendation
		Digits:      o.digits,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return "", "", err
	}

	return key.Secret(), key.URL(), nil
}

// Validate checks whether a code is valid at the given time.
func (o *TOTP) Validate(code, secret string, at time.Time) bool {
	rv, err := totp.ValidateCustom(code, secret, at, totp.ValidateOpts{
		Period:    o.period,
		Skew:      o.skew,
		Digits:    o.digits,
		Algorithm: otp.AlgorithmSHA1,
	})

	return rv && err == nil
}

// GenerateCode creates a TOTP code for the given secret and time.
func (o *TOTP) GenerateCode(secret string, at time.Time) (string, error) {
	return totp.GenerateCodeCustom(secret, at, totp.ValidateOpts{
		Period:    o.period,
		Skew:      o.skew,
		Digits:    o.digits,
		Algorithm: otp.AlgorithmSHA1,
	})
}
