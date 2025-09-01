package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

	middlewarex "secure-communication-ltd/backend/internal/middleware"
	"secure-communication-ltd/backend/internal/services"
)

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func ChangePassword(db *sqlx.DB, pol services.PasswordPolicy) echo.HandlerFunc {
	return func(c echo.Context) error {
		// User ID from context (injected by the authentication middleware)
		uid, err := middlewarex.UserIDFromCtx(c)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		var req ChangePasswordRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
		}
		if req.OldPassword == "" || req.NewPassword == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing fields"})
		}

		// Policy check for the new password
		if err := services.ValidatePassword(req.NewPassword, pol); err != nil {
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		}

		// Fetch user (current hash+salt, and also email/name for notification email)
		var (
			curHash string
			curSalt []byte
			email   string
			usernm  string
		)
		err = db.QueryRowx(`
			SELECT password_hmac, salt, email, username
			FROM users
			WHERE id = ?
		`, uid).Scan(&curHash, &curSalt, &email, &usernm)
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
		}

		// Verification of the old password
		oldHex, err := services.HashPasswordHMACHex(req.OldPassword, curSalt)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "hash error"})
		}
		if oldHex != curHash {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "old password is incorrect"})
		}

		// Calculate fingerprint (salt-independent) for the new password
		newFP, err := services.HashPasswordFingerprintHex(req.NewPassword)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "fingerprint error"})
		}

		// Calculate new hash+salt for the new password
		newSalt, err := services.GenerateSalt16()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "salt error"})
		}
		newHex, err := services.HashPasswordHMACHex(req.NewPassword, newSalt)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "hash error"})
		}

		// Must be different from the current one (hash check with current/new salt)
		if newHex == curHash {
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{
				"error": "new password must differ from current password",
			})
		}

		// Must be different from the last N: we will compare by historical fingerprint
		nHistory := pol.History
		if nHistory > 0 {
			var prevFPs []sql.NullString
			if err := db.Select(&prevFPs, `
				SELECT password_fp
				FROM password_history
				WHERE user_id = ?
				ORDER BY changed_at DESC
				LIMIT ?
			`, uid, nHistory); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
			}
			for _, fp := range prevFPs {
				if fp.Valid && fp.String == newFP {
					return c.JSON(http.StatusUnprocessableEntity, map[string]string{
						"error": "new password must differ from the last used passwords",
					})
				}
			}
		}

		// Transaction: add the current password to the history (including old fingerprint), update password+salt, and clear old history
		tx, err := db.Beginx()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "tx error"})
		}
		defer tx.Rollback()

		// Fingerprint of the old password that was just provided (to be saved in history)
		oldFP, err := services.HashPasswordFingerprintHex(req.OldPassword)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "fingerprint error"})
		}

		// Add the old password to the history (including fp)
		if _, err := tx.Exec(`
			INSERT INTO password_history (user_id, password_hmac, password_fp)
			VALUES (?, ?, ?)
		`, uid, curHash, oldFP); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "history insert error"})
		}

		// Update the password and salt in the users table
		if _, err := tx.Exec(`
			UPDATE users
			SET password_hmac = ?, salt = ?
			WHERE id = ?
		`, newHex, newSalt, uid); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "update error"})
		}

		// Clear history: we will leave only the last N
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
			`, uid, uid, nHistory); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "trim history error"})
			}
		}

		if err := tx.Commit(); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "commit error"})
		}

		// Email notification about password change (best-effort)
		if mailer, err := services.NewMailerFromEnv(); err == nil && email != "" {
			_ = mailer.Send(email, "Your password was changed", `
				<p>Hi `+usernm+`,</p>
				<p>Your password was changed successfully. If this wasn't you, please reset your password immediately.</p>
				<p>Time: `+time.Now().UTC().Format(time.RFC3339)+` (UTC)</p>
			`)
		}

		// Invalidate session: clear the cookie to force a new login
		cookie := &http.Cookie{
			Name:     services.CookieName,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Secure:   false, // production: true behind HTTPS
		}
		c.SetCookie(cookie)

		return c.JSON(http.StatusOK, map[string]string{
			"message": "password changed; please sign in again",
		})
	}
}
