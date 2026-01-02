// Package otp provides helpers for generating and validating one-time
// passwords (OTP), focused on TOTP (time-based OTP).
//
// This is typically used for 2FA/MFA flows: generate a secret and URI for an
// authenticator app, then validate user-provided codes.
package otp
