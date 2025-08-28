package handlers

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

func VerifyEmail(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		raw := c.QueryParam("token")
		if raw == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing token"})
		}

		h := sha1.Sum([]byte(raw))
		shaHex := hex.EncodeToString(h[:])

		var (
			userID    int64
			expiresAt time.Time
			usedAt    sql.NullTime
		)

		err := db.QueryRowx(`
			SELECT user_id, expires_at, used_at
			FROM email_verification_tokens
			WHERE token_sha1 = ?
		`, shaHex).Scan(&userID, &expiresAt, &usedAt)
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid token"})
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
		}
		if usedAt.Valid {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "token already used"})
		}
		if time.Now().After(expiresAt) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "token expired"})
		}

		tx, err := db.Beginx()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "tx error"})
		}
		defer tx.Rollback()

		if _, err := tx.Exec(`UPDATE users SET is_verified = TRUE WHERE id = ?`, userID); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "update user error"})
		}
		if _, err := tx.Exec(`UPDATE email_verification_tokens SET used_at = NOW() WHERE token_sha1 = ?`, shaHex); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "update token error"})
		}

		if err := tx.Commit(); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "commit error"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Email verified successfully"})
	}
}
