package services

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"time"
)

type RawToken struct {
	Raw       string
	SHA1Hex   string
	ExpiresAt time.Time
}

func NewVerificationToken(ttl time.Duration) (*RawToken, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	raw := hex.EncodeToString(b)

	h := sha1.Sum([]byte(raw))
	shaHex := hex.EncodeToString(h[:])

	return &RawToken{
		Raw:       raw,
		SHA1Hex:   shaHex,
		ExpiresAt: time.Now().Add(ttl),
	}, nil
}

func NewRandomBase64URL(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
