package mfa

import (
	"crypto/rand"
	"math/big"
	"strings"
)

// RecoveryCodeGenerator defines an interface for generating MFA recovery codes.
type RecoveryCodeGenerator interface {
	// Generate returns a slice of unique recovery codes or an error if the
	// random source fails.
	Generate() ([]string, error)
}

// alphabet is the character set used for recovery code generation.
//
// It includes digits, uppercase letters, and lowercase letters for a total
// of 62 characters, providing high entropy while remaining user-friendly.
const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// RecoveryCode generates cryptographically secure MFA recovery codes.
//
// It produces recovery codes formatted as:
//
//	XXXX-XXXX-XXXX
//
// Each X is selected uniformly at random from the alphabet constant.
type RecoveryCode struct{}

// NewRecoveryCode returns a new RecoveryCode generator.
func NewRecoveryCode() *RecoveryCode {
	return &RecoveryCode{}
}

// Generate produces a set of unique recovery codes.
//
// It returns exactly 10 codes. Each code is randomly generated using
// crypto/rand for cryptographic security.
func (rc *RecoveryCode) Generate() ([]string, error) {
	out := make([]string, 0, 10)
	seen := make(map[string]struct{}, 10)

	for len(out) < 10 {
		code, err := rc.generateStrictCode()
		if err != nil {
			return nil, err
		}

		// extremely unlikely, but prevents accidental duplicates
		if _, ok := seen[code]; ok {
			continue
		}

		seen[code] = struct{}{}
		out = append(out, code)
	}

	return out, nil
}

func (rc *RecoveryCode) generateStrictCode() (string, error) {
	raw, err := rc.randomStrictString(12)
	if err != nil {
		return "", err
	}
	return raw[0:4] + "-" + raw[4:8] + "-" + raw[8:12], nil
}

func (rc *RecoveryCode) randomStrictString(n int) (string, error) {
	var sb strings.Builder
	sb.Grow(n)

	for i := 0; i < n; i++ {
		idx, err := rc.randIntStrict(len(alphabet))
		if err != nil {
			return "", err
		}
		sb.WriteByte(alphabet[idx])
	}

	return sb.String(), nil
}

func (rc *RecoveryCode) randIntStrict(max int) (int, error) {
	num, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(num.Int64()), nil
}
