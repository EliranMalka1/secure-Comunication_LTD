package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// generate a 6-digit numeric code, zero-padded (e.g. "083174")
func GenerateNumericCode6() (string, error) {
	var b [3]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	// 24-bit -> 0..16777215 ; take mod 1e6
	n := (int(b[0])<<16 | int(b[1])<<8 | int(b[2])) % 1000000
	return fmt.Sprintf("%06d", n), nil
}

func HashSHA256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// Cancel any open (unconsumed) challenges for user
func CancelOpenOTPChallenges(db *sqlx.DB, userID int64) error {
	_, err := db.Exec(`UPDATE login_otp_challenges
		SET consumed_at = NOW()
		WHERE user_id = ? AND consumed_at IS NULL`, userID)
	return err
}

type OTPConfig struct {
	TTLMinutes    int // e.g. 10
	MaxAttempts   int // e.g. 5
	ResendWindowS int // optional, if you implement /resend
}

// Start a new OTP flow: cancel old, create new, email the raw code
func StartEmailOTP(db *sqlx.DB, mailer *Mailer, userID int64, toEmail string, cfg OTPConfig) error {
	if err := CancelOpenOTPChallenges(db, userID); err != nil {
		return err
	}
	code, err := GenerateNumericCode6()
	if err != nil {
		return err
	}
	hash := HashSHA256Hex(code)
	expires := time.Now().Add(time.Duration(cfg.TTLMinutes) * time.Minute)

	if _, err := db.Exec(`
		INSERT INTO login_otp_challenges (user_id, code_sha256, expires_at)
		VALUES (?, ?, ?)`, userID, hash, expires); err != nil {
		return err
	}

	// Email content (simple HTML)
	html := fmt.Sprintf(`
		<h2>Your verification code</h2>
		<p>Enter this 6-digit code to complete your sign-in:</p>
		<p style="font-size:24px;font-weight:bold;letter-spacing:4px">%s</p>
		<p>This code expires in %d minutes.</p>
	`, code, cfg.TTLMinutes)

	return mailer.Send(toEmail, "Your verification code", html)
}
