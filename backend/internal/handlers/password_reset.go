package handlers

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"secure-communication-ltd/backend/internal/services"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

type resetReq struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// GET /api/password/reset?token=...
// Redirects to frontend: http://localhost:3000/reset?token=...
func PasswordResetLanding() echo.HandlerFunc {
	return func(c echo.Context) error {
		raw := c.QueryParam("token")
		fe := os.Getenv("FRONTEND_PUBLIC_URL")
		if fe == "" {
			fe = "http://localhost:3000"
		}
		u := fe + "/reset?token=" + url.QueryEscape(raw)
		return c.Redirect(http.StatusTemporaryRedirect, u)
	}
}

func PasswordReset(db *sqlx.DB, pol services.PasswordPolicy) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req resetReq
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
		}
		raw := strings.TrimSpace(req.Token)
		newPw := req.NewPassword

		if raw == "" || newPw == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing fields"})
		}
		// Password policy check
		if err := services.ValidatePassword(newPw, pol); err != nil {
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		}

		h := sha1.Sum([]byte(raw))
		sha := hex.EncodeToString(h[:])

		// Find a valid open token
		var userID int64
		var expiresAt time.Time
		var usedAt sql.NullTime

		err := db.QueryRowx(`
			SELECT user_id, expires_at, used_at
			FROM password_reset_tokens
			WHERE token_sha1 = ?
			LIMIT 1
		`, sha).Scan(&userID, &expiresAt, &usedAt)
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid token"})
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
		}
		if usedAt.Valid || time.Now().After(expiresAt) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "token expired or used"})
		}

		// Password history (optional—according to requirements)
		// You can check here that the new password is not equal to one of the last ones (if you store HMAC in history).
		// We'll keep it minimal: update password + save history.

		salt, err := services.GenerateSalt16()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "salt error"})
		}
		hmacHex, err := services.HashPasswordHMACHex(newPw, salt)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "hash error"})
		}

		tx, err := db.Beginx()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "tx error"})
		}
		defer tx.Rollback()

		// Save history (store the new hash)
		if _, err := tx.Exec(`
			INSERT INTO password_history (user_id, password_hmac) VALUES (?, ?)`,
			userID, hmacHex); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "history error"})
		}

		// Update the user's password
		if _, err := tx.Exec(`
			UPDATE users SET password_hmac = ?, salt = ? WHERE id = ?`,
			hmacHex, salt, userID); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "update user error"})
		}

		// Mark the token as used (single-use)
		if _, err := tx.Exec(`
			UPDATE password_reset_tokens SET used_at = NOW() WHERE token_sha1 = ?`,
			sha); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "consume token error"})
		}

		if err := tx.Commit(); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "commit error"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Password reset successfully"})
	}
}
