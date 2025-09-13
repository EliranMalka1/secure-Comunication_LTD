package handlers

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

	"secure-communication-ltd/backend/internal/services"
)

func ChangePasswordConfirm(db *sqlx.DB, pol services.PasswordPolicy) echo.HandlerFunc {
	return func(c echo.Context) error {
		raw := c.QueryParam("token")
		if raw == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing token"})
		}

		// token SHA-1 per project spec
		sum := sha1.Sum([]byte(raw))
		tokenSHA1 := hex.EncodeToString(sum[:])

		var (
			userID  int64
			newHex  string
			newSalt []byte
			newFP   string // requires password_change_requests.new_password_fp in schema!
			expires time.Time
			usedAt  sql.NullTime
		)

		// load pending request
		err := db.QueryRowx(`
			SELECT user_id, new_password_hmac, new_salt, new_password_fp, expires_at, used_at
			FROM password_change_requests
			WHERE token_sha1 = ?
		`, tokenSHA1).Scan(&userID, &newHex, &newSalt, &newFP, &expires, &usedAt)
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid token"})
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
		}
		if usedAt.Valid {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "token already used"})
		}
		if time.Now().After(expires) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "token expired"})
		}

		tx, err := db.Beginx()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "tx error"})
		}
		defer tx.Rollback()

		// push CURRENT to history (include fp + salt for fallback)
		if _, err := tx.Exec(`
			INSERT INTO password_history (user_id, password_hmac, password_fp, salt)
			SELECT id, password_hmac, password_fp, salt
			FROM users
			WHERE id = ?
		`, userID); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "history insert error"})
		}

		// update user with NEW (hash + salt + fp)
		if _, err := tx.Exec(`
			UPDATE users
			SET password_hmac = ?, salt = ?, password_fp = ?
			WHERE id = ?
		`, newHex, newSalt, newFP, userID); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "update user error"})
		}

		// trim to last N
		nHistory := pol.History // <<< HISTORY COUNT HERE
		if nHistory > 0 {
			if _, err := tx.Exec(`
				DELETE FROM password_history
				WHERE user_id = ?
				  AND id NOT IN (
					SELECT id FROM (
						SELECT id
						FROM password_history
						WHERE user_id = ?
						ORDER BY changed_at DESC
						LIMIT ?
					) AS keep_rows
				  )
			`, userID, userID, nHistory); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "trim history error"})
			}
		}

		// mark token used
		if _, err := tx.Exec(`
			UPDATE password_change_requests
			SET used_at = NOW()
			WHERE token_sha1 = ?
		`, tokenSHA1); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "token update error"})
		}

		if err := tx.Commit(); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "commit error"})
		}

		// invalidate session
		http.SetCookie(c.Response(), &http.Cookie{
			Name:     services.CookieName,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Secure:   false, // set true in HTTPS prod
		})

		return c.HTML(http.StatusOK, verificationPage(true,
			"Password Changed",
			"Your password was updated successfully. Please sign in with your new password."))
	}
}
