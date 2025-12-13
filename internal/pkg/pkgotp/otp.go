package pkgotp

import (
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type TOTP struct {
	issuer string
	period uint
	skew   uint
	digits otp.Digits
}

// OTP defines the contract for TOTP operations.
type OTP interface {
	Generate(accountName string) (secret string, uri string, err error)
	Validate(code, secret string, at time.Time) bool
	GenerateCode(secret string, at time.Time) (string, error)
}

func NewTOTP(issuer string, period, skew uint, digits otp.Digits) *TOTP {
	if digits != otp.DigitsSix && digits != otp.DigitsEight {
		digits = otp.DigitsSix
	}

	if period == 0 {
		period = 30
	}

	return &TOTP{
		issuer: issuer,
		period: period,
		skew:   skew,
		digits: digits,
	}
}

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

func (o *TOTP) Validate(code, secret string, at time.Time) bool {
	rv, err := totp.ValidateCustom(code, secret, at, totp.ValidateOpts{
		Period:    o.period,
		Skew:      o.skew,
		Digits:    o.digits,
		Algorithm: otp.AlgorithmSHA1,
	})

	return rv && err == nil
}

func (o *TOTP) GenerateCode(secret string, at time.Time) (string, error) {
	return totp.GenerateCodeCustom(secret, at, totp.ValidateOpts{
		Period:    o.period,
		Skew:      o.skew,
		Digits:    o.digits,
		Algorithm: otp.AlgorithmSHA1,
	})
}
