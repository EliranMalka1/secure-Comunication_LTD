package handlers

import (
	"bytes"
	"database/sql"
	"errors"
	"html/template"
	"net/http"
	"time"

	"secure-communication-ltd/backend/config"
	middlewarex "secure-communication-ltd/backend/internal/middleware"
	"secure-communication-ltd/backend/internal/services"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func ChangePassword(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		pol := config.GetPolicy()

		// 1) User from context
		uid, err := middlewarex.UserIDFromCtx(c)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		// 2) Parse input
		var req ChangePasswordRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
		}
		if req.OldPassword == "" || req.NewPassword == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing fields"})
		}

		// 3) Policy for NEW password
		if err := services.ValidatePassword(req.NewPassword, pol); err != nil {
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		}

		// 4) Load current user secrets (+ email for notification)
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

		// 5) Verify old password
		oldHex, err := services.HashPasswordHMACHex(req.OldPassword, curSalt)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "hash error"})
		}
		if oldHex != curHash {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "old password is incorrect"})
		}

		// 6) Fingerprints (salt-independent)
		oldFP, err := services.HashPasswordFingerprintHex(req.OldPassword)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "fingerprint error"})
		}
		newFP, err := services.HashPasswordFingerprintHex(req.NewPassword)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "fingerprint error"})
		}

		// 7) Block identical to current
		if newFP == oldFP {
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{
				"error": "new password must differ from current password",
			})
		}

		// 8) History check (FP first; fallback to HMAC+salt)
		nHistory := pol.History
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
				if r.FP.Valid && r.FP.String != "" {
					if r.FP.String == newFP {
						return c.JSON(http.StatusUnprocessableEntity, map[string]string{
							"error": "new password must differ from the last used passwords",
						})
					}
					continue
				}
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

		// 9) Prepare new salt+hash
		newSalt, err := services.GenerateSalt16()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "salt error"})
		}
		newHex, err := services.HashPasswordHMACHex(req.NewPassword, newSalt)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "hash error"})
		}

		// 10) Tx: push current to history (FP+salt), update users (hash+salt+FP), trim history
		tx, err := db.Beginx()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "tx error"})
		}
		defer tx.Rollback()

		if _, err := tx.Exec(`
			INSERT INTO password_history (user_id, password_hmac, password_fp, salt)
			VALUES (?, ?, ?, ?)
		`, uid, curHash, oldFP, curSalt); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "history insert error"})
		}

		if _, err := tx.Exec(`
			UPDATE users
			SET password_hmac = ?, salt = ?, password_fp = ?
			WHERE id = ?
		`, newHex, newSalt, newFP, uid); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "update error"})
		}

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

		// 11) Best-effort email (SAFE HTML via template)
		if mailer, err := services.NewMailerFromEnv(); err == nil && email != "" {
			type mailData struct {
				Username string
				WhenUTC  string
			}
			md := mailData{
				Username: usernm,
				WhenUTC:  time.Now().UTC().Format(time.RFC3339),
			}
			tpl := template.Must(template.New("pwChanged").Parse(`
<p>Hi {{.Username}},</p>
<p>Your password was changed successfully. If this wasn't you, please reset your password immediately.</p>
<p>Time: {{.WhenUTC}} (UTC)</p>
`))
			var buf bytes.Buffer
			if err := tpl.Execute(&buf, md); err == nil {
				_ = mailer.Send(email, "Your password was changed", buf.String())
			}
		}

		// 12) Invalidate session (force re-login)
		cookie := &http.Cookie{
			Name:     services.CookieName,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Secure:   false, // set true in production behind HTTPS
		}
		c.SetCookie(cookie)

		return c.JSON(http.StatusOK, map[string]string{
			"message": "password changed; please sign in again",
		})
	}
}
