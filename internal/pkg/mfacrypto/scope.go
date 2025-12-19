package mfacrypto

type Purpose string

const (
	PurposeOTPSeed     Purpose = "otp_seed"
	PurposeRecoveryKey Purpose = "recovery_key"
)

// Scope binds encryption to MFA-specific identifiers.
// This is used as AAD (Additional Authenticated Data) in AES-GCM.
type Scope struct {
	UserID  int64
	Purpose Purpose
}
