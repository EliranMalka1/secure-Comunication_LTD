package handlers

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"net/http"
	"time"

	"secure-communication-ltd/backend/config"
	"secure-communication-ltd/backend/internal/services"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

func ChangePasswordConfirm(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		pol := config.GetPolicy()

		raw := c.QueryParam("token")
		if raw == "" {
			return RenderVerificationPage(c, http.StatusBadRequest, false,
				"Missing Token", "A confirmation token is required.")
		}

		// token SHA-1 per spec
		sum := sha1.Sum([]byte(raw))
		tokenSHA1 := hex.EncodeToString(sum[:])

		var (
			userID  int64
			newHex  string
			newSalt []byte
			newFP   string
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
			return RenderVerificationPage(c, http.StatusBadRequest, false,
				"Invalid Token", "This confirmation link is not valid.")
		}
		if err != nil {
			return RenderVerificationPage(c, http.StatusInternalServerError, false,
				"Server Error", "Could not read the pending change request.")
		}
		if usedAt.Valid {
			return RenderVerificationPage(c, http.StatusBadRequest, false,
				"Already Used", "This confirmation link was already used.")
		}
		if time.Now().After(expires) {
			return RenderVerificationPage(c, http.StatusBadRequest, false,
				"Expired Link", "This confirmation link has expired.")
		}

		tx, err := db.Beginx()
		if err != nil {
			return RenderVerificationPage(c, http.StatusInternalServerError, false,
				"Server Error", "Could not start a database transaction.")
		}
		defer tx.Rollback()

		// move CURRENT to history (store fp + salt for fallback comparisons)
		if _, err := tx.Exec(`
			INSERT INTO password_history (user_id, password_hmac, password_fp, salt)
			SELECT id, password_hmac, password_fp, salt
			FROM users
			WHERE id = ?
		`, userID); err != nil {
			return RenderVerificationPage(c, http.StatusInternalServerError, false,
				"Server Error", "Could not save previous password in history.")
		}

		// set NEW password (hash + salt + fp)
		if _, err := tx.Exec(`
			UPDATE users
			SET password_hmac = ?, salt = ?, password_fp = ?
			WHERE id = ?
		`, newHex, newSalt, newFP, userID); err != nil {
			return RenderVerificationPage(c, http.StatusInternalServerError, false,
				"Server Error", "Could not update the new password.")
		}

		// trim history to last N
		if nHistory := pol.History; nHistory > 0 {
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
				return RenderVerificationPage(c, http.StatusInternalServerError, false,
					"Server Error", "Could not trim password history.")
			}
		}

		// mark token used
		if _, err := tx.Exec(`
			UPDATE password_change_requests
			SET used_at = NOW()
			WHERE token_sha1 = ?
		`, tokenSHA1); err != nil {
			return RenderVerificationPage(c, http.StatusInternalServerError, false,
				"Server Error", "Could not mark the confirmation token as used.")
		}

		if err := tx.Commit(); err != nil {
			return RenderVerificationPage(c, http.StatusInternalServerError, false,
				"Server Error", "Could not commit the password change.")
		}

		// invalidate session cookie
		http.SetCookie(c.Response(), &http.Cookie{
			Name:     services.CookieName,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Secure:   false, // set true in HTTPS production
		})

		return RenderVerificationPage(c, http.StatusOK, true,
			"Password Changed",
			"Your password was updated successfully. Please sign in with your new password.")
	}
}
