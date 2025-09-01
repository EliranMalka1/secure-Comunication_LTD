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

func ChangePasswordConfirm(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		raw := c.QueryParam("token")
		if raw == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing token"})
		}
		h := sha1.Sum([]byte(raw))
		shaHex := hex.EncodeToString(h[:])

		var (
			userID    int64
			newHex    string
			newSalt   []byte
			expiresAt time.Time
			usedAt    sql.NullTime
		)
		err := db.QueryRowx(`
			SELECT user_id, new_password_hmac, new_salt, expires_at, used_at
			FROM password_change_requests
			WHERE token_sha1 = ?
		`, shaHex).Scan(&userID, &newHex, &newSalt, &expiresAt, &usedAt)
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

		// move current to history
		if _, err := tx.Exec(`
			INSERT INTO password_history (user_id, password_hmac)
			SELECT id, password_hmac FROM users WHERE id = ?
		`, userID); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "history insert error"})
		}

		// update user with new password
		if _, err := tx.Exec(`
			UPDATE users SET password_hmac = ?, salt = ? WHERE id = ?
		`, newHex, newSalt, userID); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "update user error"})
		}

		// mark token used
		if _, err := tx.Exec(`
			UPDATE password_change_requests SET used_at = NOW() WHERE token_sha1 = ?
		`, shaHex); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "token update error"})
		}

		if err := tx.Commit(); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "commit error"})
		}

		// expire auth cookie to force re-login
		http.SetCookie(c.Response(), &http.Cookie{
			Name:     "auth_token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Secure:   false, // set true in HTTPS prod
		})

		// pretty landing page
		return c.HTML(http.StatusOK, verificationPage(true,
			"Password Changed",
			"Your password was updated successfully. Please sign in with your new password."))
	}
}
