package handlers

import (
	"crypto/subtle"
	"database/sql"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

	"secure-communication-ltd/backend/internal/services"
)

type LoginRequest struct {
	ID       string `json:"id"` // email or username
	Password string `json:"password"`
}

type userRow struct {
	ID         int64  `db:"id"`
	Username   string `db:"username"`
	Email      string `db:"email"`
	PassHMAC   string `db:"password_hmac"`
	Salt       []byte `db:"salt"`
	IsActive   bool   `db:"is_active"`
	IsVerified bool   `db:"is_verified"`
}

func Login(db *sqlx.DB, pol services.PasswordPolicy) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req LoginRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
		}
		req.ID = strings.TrimSpace(req.ID)
		if req.ID == "" || strings.TrimSpace(req.Password) == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing fields"})
		}

		// --- Lockout window check (failed password attempts) ---
		var failCount int
		if err := db.Get(&failCount, `
			SELECT COUNT(*) FROM login_attempts
			WHERE username = ? AND success = 0
			  AND attempt_time > (NOW() - INTERVAL ? MINUTE)
		`, req.ID, pol.LockoutMinutes); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
		}
		if pol.MaxLoginAttempts > 0 && failCount >= pol.MaxLoginAttempts {
			return c.JSON(http.StatusTooManyRequests, map[string]string{"error": "account temporarily locked"})
		}

		// --- Fetch user by email OR username ---
		var u userRow
		err := db.Get(&u, `
			SELECT id, username, email, password_hmac, salt, is_active, is_verified
			FROM users
			WHERE email = ? OR username = ?
			LIMIT 1
		`, req.ID, req.ID)

		knownUser := (err == nil)
		userIDForLog := sql.NullInt64{}
		if knownUser {
			userIDForLog.Valid = true
			userIDForLog.Int64 = u.ID
		}

		// --- Constant-time style check (compute hash whether user exists or not) ---
		var computed string
		if knownUser {
			if h, e := services.HashPasswordHMACHex(req.Password, u.Salt); e == nil {
				computed = h
			}
		} else {
			// dummy work to keep timing similar
			dummySalt := make([]byte, 16)
			_, _ = services.HashPasswordHMACHex(req.Password, dummySalt)
		}

		ip := clientIP(c.Request())

		// --- Check password + statuses ---
		ok := false
		if knownUser {
			if subtle.ConstantTimeCompare([]byte(computed), []byte(u.PassHMAC)) == 1 &&
				u.IsActive && u.IsVerified {
				ok = true
			}
		}

		if !ok {
			// log failed attempt (password stage)
			_, _ = db.Exec(`
				INSERT INTO login_attempts (user_id, username, ip, success)
				VALUES (?, ?, ?, 0)
			`, userIDForLog, req.ID, ip)
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		}

		// =============== 2FA REQUIRED (Email OTP) ===============
		// Step 1 success (password ok) -> do NOT issue cookie yet.
		// Start an email OTP challenge and respond with mfa_required=true.

		mailer, err := services.NewMailerFromEnv()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "mailer error"})
		}

		// read optional config from env, with safe defaults
		ttl := 10   // minutes
		maxAtt := 5 // attempts
		if v := os.Getenv("MFA_OTP_TTL_MINUTES"); v != "" {
			if n, e := strconv.Atoi(v); e == nil && n > 0 && n <= 60 {
				ttl = n
			}
		}
		if v := os.Getenv("MFA_OTP_MAX_ATTEMPTS"); v != "" {
			if n, e := strconv.Atoi(v); e == nil && n >= 1 && n <= 10 {
				maxAtt = n
			}
		}

		cfg := services.OTPConfig{
			TTLMinutes:  ttl,
			MaxAttempts: maxAtt,
		}
		if err := services.StartEmailOTP(db, mailer, u.ID, u.Email, cfg); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "otp start error"})
		}

		// Tell client to show OTP code screen (step 2)
		return c.JSON(http.StatusOK, map[string]any{
			"mfa_required": true,
			"method":       "email_otp",
			"expires_in":   ttl,
		})
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
