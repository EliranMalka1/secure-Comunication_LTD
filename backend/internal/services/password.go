package services

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"regexp"
)

var (
	reUpper   = regexp.MustCompile(`[A-Z]`)
	reLower   = regexp.MustCompile(`[a-z]`)
	reDigit   = regexp.MustCompile(`\d`)
	reSpecial = regexp.MustCompile(`[^A-Za-z0-9]`)
)

type PasswordPolicy struct {
	MinLength                                                int
	RequireUpper, RequireLower, RequireDigit, RequireSpecial bool
	History                                                  int
	// New: login throttling / lockout
	MaxLoginAttempts int // e.g. 3
	LockoutMinutes   int // e.g. 15
}

func DefaultPolicy() PasswordPolicy {
	return PasswordPolicy{
		MinLength:        10,
		RequireUpper:     true,
		RequireLower:     true,
		RequireDigit:     true,
		RequireSpecial:   true,
		History:          3,
		MaxLoginAttempts: 3,
		LockoutMinutes:   15,
	}
}

func ValidatePassword(pw string, pol PasswordPolicy) error {
	if len(pw) < pol.MinLength {
		return errors.New("password too short")
	}
	if pol.RequireUpper && !reUpper.MatchString(pw) {
		return errors.New("must include uppercase letter")
	}
	if pol.RequireLower && !reLower.MatchString(pw) {
		return errors.New("must include lowercase letter")
	}
	if pol.RequireDigit && !reDigit.MatchString(pw) {
		return errors.New("must include digit")
	}
	if pol.RequireSpecial && !reSpecial.MatchString(pw) {
		return errors.New("must include special char")
	}
	return nil
}

// 16-byte random salt (raw bytes, for VARBINARY(16))
func GenerateSalt16() ([]byte, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	return b, err
}

// Returns hex(HMAC-SHA256(secret, salt||password)) – 64 hex chars.
func HashPasswordHMACHex(password string, salt []byte) (string, error) {
	secret := os.Getenv("HMAC_SECRET")
	if secret == "" {
		return "", errors.New("missing HMAC_SECRET")
	}
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(salt)
	h.Write([]byte(password))
	sum := h.Sum(nil) // 32 bytes
	return hex.EncodeToString(sum), nil
}

// HashPasswordFingerprintHex returns hex(HMAC-SHA256(history_secret, password)).
// This is salt-independent and used ONLY for password reuse checks.
func HashPasswordFingerprintHex(password string) (string, error) {
	secret := os.Getenv("HMAC_HISTORY_SECRET")
	if secret == "" {
		// Fallback to HMAC_SECRET if you prefer, but better to have a dedicated secret
		secret = os.Getenv("HMAC_SECRET")
		if secret == "" {
			return "", errors.New("missing HMAC_HISTORY_SECRET (or HMAC_SECRET)")
		}
	}
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil)), nil
}
