package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"secure-communication-ltd/backend/internal/services"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

type MFALoginRequest struct {
	ID   string `json:"id"`   // email or username (same identifier as in the password step)
	Code string `json:"code"` // 6 digits
}

type otpRow struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	Attempts  int       `db:"attempts"`
	ExpiresAt time.Time `db:"expires_at"`
}

func LoginMFA(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req MFALoginRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
		}
		id := strings.TrimSpace(req.ID)
		code := strings.TrimSpace(req.Code)
		if id == "" || code == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing fields"})
		}

		// find user
		var userID int64
		if err := db.Get(&userID, `
			SELECT id FROM users WHERE email = ? OR username = ? LIMIT 1
		`, id, id); err != nil {
			if err == sql.ErrNoRows {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid code"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
		}

		// get last open challenge
		var ch otpRow
		err := db.Get(&ch, `
			SELECT id, user_id, attempts, expires_at
			FROM login_otp_challenges
			WHERE user_id = ? AND consumed_at IS NULL AND expires_at > NOW()
			ORDER BY id DESC
			LIMIT 1
		`, userID)
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "no active code"})
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
		}

		// optional: lock after too many attempts
		if ch.Attempts >= 5 {
			_, _ = db.Exec(`UPDATE login_otp_challenges SET consumed_at = NOW() WHERE id = ?`, ch.ID)
			return c.JSON(http.StatusTooManyRequests, map[string]string{"error": "too many attempts"})
		}

		// check code hash
		hash := services.HashSHA256Hex(code)
		var ok int
		if err := db.Get(&ok, `
			SELECT COUNT(*) FROM login_otp_challenges
			WHERE id = ? AND code_sha256 = ? AND consumed_at IS NULL AND expires_at > NOW()
			LIMIT 1
		`, ch.ID, hash); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
		}

		if ok == 0 {
			_, _ = db.Exec(`UPDATE login_otp_challenges SET attempts = attempts + 1 WHERE id = ?`, ch.ID)
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid code"})
		}

		// consume on success (single-use)
		if _, err := db.Exec(`UPDATE login_otp_challenges SET consumed_at = NOW() WHERE id = ?`, ch.ID); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "consume error"})
		}

		// === SUCCESS: issue JWT + cookie (using your existing helper/signature) ===
		// Your CreateJWT function receives (userID, username, ttl). We don't have username here,
		// so we will fetch it briefly or send an empty string if the function allows. Preferably fetch:
		var username string
		_ = db.Get(&username, `SELECT username FROM users WHERE id = ?`, userID)

		token, err := services.CreateJWT(userID, username, 24*time.Hour)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "token error"})
		}

		cookie := &http.Cookie{
			Name:     services.CookieName, // As used in login.go
			Value:    token,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Secure:   false, // Set to true in HTTPS
		}
		c.SetCookie(cookie)

		return c.JSON(http.StatusOK, map[string]string{"message": "ok"})
	}
}
