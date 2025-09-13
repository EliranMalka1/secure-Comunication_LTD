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
		// 0) User from context
		uid, err := middlewarex.UserIDFromCtx(c)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		// 1) Parse input
		var req ChangePasswordRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
		}
		if req.OldPassword == "" || req.NewPassword == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing fields"})
		}

		// 2) Policy for NEW password
		if err := services.ValidatePassword(req.NewPassword, pol); err != nil {
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		}

		// 3) Load current user secret material (+ email for notification)
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

		// 4) Verify old password
		oldHex, err := services.HashPasswordHMACHex(req.OldPassword, curSalt)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "hash error"})
		}
		if oldHex != curHash {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "old password is incorrect"})
		}

		// 5) Compute fingerprints (salt-independent)
		oldFP, err := services.HashPasswordFingerprintHex(req.OldPassword)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "fingerprint error"})
		}
		newFP, err := services.HashPasswordFingerprintHex(req.NewPassword)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "fingerprint error"})
		}

		// Block identical to current
		if newFP == oldFP {
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{
				"error": "new password must differ from current password",
			})
		}

		// 6) History check (FP first; fallback to HMAC+salt when FP is missing)
		nHistory := pol.History // <<< HISTORY COUNT HERE
		if nHistory > 0 {
			type histRow struct {
				FP   sql.NullString `db:"password_fp"`
				HMAC sql.NullString `db:"password_hmac"`
				Salt []byte         `db:"salt"`
			}
			var rows []histRow
			if err := db.Select(&rows, `
				SELECT password_fp, password_hmac, salt
				FROM password_history
				WHERE user_id = ?
				ORDER BY changed_at DESC
				LIMIT ?
			`, uid, nHistory); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
			}
			for _, r := range rows {
				// Prefer FP equality when available
				if r.FP.Valid && r.FP.String != "" {
					if r.FP.String == newFP {
						return c.JSON(http.StatusUnprocessableEntity, map[string]string{
							"error": "new password must differ from the last used passwords",
						})
					}
					continue
				}
				// Fallback: if no FP, compare HMAC(new, stored salt) with stored HMAC
				if len(r.Salt) > 0 && r.HMAC.Valid && r.HMAC.String != "" {
					hx, err := services.HashPasswordHMACHex(req.NewPassword, r.Salt)
					if err != nil {
						return c.JSON(http.StatusInternalServerError, map[string]string{"error": "hash error"})
					}
					if hx == r.HMAC.String {
						return c.JSON(http.StatusUnprocessableEntity, map[string]string{
							"error": "new password must differ from the last used passwords",
						})
					}
				}
			}
		}

		// 7) Prepare new salt+hash (only after passing checks)
		newSalt, err := services.GenerateSalt16()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "salt error"})
		}
		newHex, err := services.HashPasswordHMACHex(req.NewPassword, newSalt)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "hash error"})
		}

		// 8) Tx: push current to history (including FP+salt), update users with new hash+salt+FP, trim history
		tx, err := db.Beginx()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "tx error"})
		}
		defer tx.Rollback()

		// Save the *old/current* password to history (with FP + salt!)
		if _, err := tx.Exec(`
			INSERT INTO password_history (user_id, password_hmac, password_fp, salt)
			VALUES (?, ?, ?, ?)
		`, uid, curHash, oldFP, curSalt); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "history insert error"})
		}

		// Update users with NEW password material (hash + salt + FP)
		if _, err := tx.Exec(`
			UPDATE users
			SET password_hmac = ?, salt = ?, password_fp = ?
			WHERE id = ?
		`, newHex, newSalt, newFP, uid); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "update error"})
		}

		// Trim history to N most recent rows
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

		// 9) Best-effort email notification
		if mailer, err := services.NewMailerFromEnv(); err == nil && email != "" {
			_ = mailer.Send(email, "Your password was changed", `
				<p>Hi `+usernm+`,</p>
				<p>Your password was changed successfully. If this wasn't you, please reset your password immediately.</p>
				<p>Time: `+time.Now().UTC().Format(time.RFC3339)+` (UTC)</p>
			`)
		}

		// 10) Invalidate session (force re-login)
		cookie := &http.Cookie{
			Name:     services.CookieName,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Secure:   false, // set true behind HTTPS in production
		}
		c.SetCookie(cookie)

		return c.JSON(http.StatusOK, map[string]string{
			"message": "password changed; please sign in again",
		})
	}
}
