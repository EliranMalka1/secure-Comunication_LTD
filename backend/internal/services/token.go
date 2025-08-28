package services

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"time"
)

type RawToken struct {
	Raw       string // To build a verification link
	SHA1Hex   string // Stored in DB
	ExpiresAt time.Time
}

func NewVerificationToken(ttl time.Duration) (*RawToken, error) {
	b := make([]byte, 32) // 32B → 64 hex chars in raw token
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	raw := hex.EncodeToString(b)

	h := sha1.Sum([]byte(raw))         // 20B
	shaHex := hex.EncodeToString(h[:]) // 40 characters

	return &RawToken{
		Raw:       raw,
		SHA1Hex:   shaHex,
		ExpiresAt: time.Now().Add(ttl),
	}, nil
}
