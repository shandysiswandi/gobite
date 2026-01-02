package mfa

// Purpose identifies the MFA encryption purpose.
type Purpose string

const (
	// PurposeOTPSeed scopes encryption to OTP seeds.
	PurposeOTPSeed Purpose = "otp_seed"
	// PurposeRecoveryKey scopes encryption to recovery keys.
	PurposeRecoveryKey Purpose = "recovery_key"
)

// Scope binds encryption to MFA-specific identifiers.
// This is used as AAD (Additional Authenticated Data) in AES-GCM.
type Scope struct {
	// UserID is the user identifier for scoping.
	UserID int64
	// Purpose is the encryption purpose.
	Purpose Purpose
}
