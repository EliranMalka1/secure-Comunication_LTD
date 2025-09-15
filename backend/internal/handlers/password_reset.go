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

		if err := services.ValidatePassword(newPw, pol); err != nil {
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		}

		sum := sha1.Sum([]byte(raw))
		sha := hex.EncodeToString(sum[:])

		var (
			userID    int64
			expiresAt time.Time
			usedAt    sql.NullTime
		)
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

		var currentFP string
		var curHash string
		var curSalt []byte
		err = db.QueryRowx(`
			SELECT password_fp, password_hmac, salt
			FROM users
			WHERE id = ?
		`, userID).Scan(&currentFP, &curHash, &curSalt)
		if err == sql.ErrNoRows {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user"})
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
		}

		newFP, err := services.HashPasswordFingerprintHex(newPw)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "fingerprint error"})
		}

		if currentFP != "" && newFP == currentFP {
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{
				"error": "new password must differ from current password",
			})
		}

		hxCur, err := services.HashPasswordHMACHex(newPw, curSalt)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "hash error"})
		}
		if hxCur == curHash {
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{
				"error": "new password must differ from current password",
			})
		}

		nHistory := pol.History
		if nHistory > 0 {
			type histRow struct {
				FP   sql.NullString `db:"password_fp"`
				HMAC sql.NullString `db:"password_hmac"`
				Salt []byte         `db:"salt"`
			}
			var prev []histRow
			if err := db.Select(&prev, `
				SELECT password_fp, password_hmac, salt
				FROM password_history
				WHERE user_id = ?
				ORDER BY changed_at DESC
				LIMIT ?
			`, userID, nHistory); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
			}
			for _, r := range prev {

				if r.FP.Valid && r.FP.String != "" && r.FP.String == newFP {
					return c.JSON(http.StatusUnprocessableEntity, map[string]string{
						"error": "new password must differ from the last used passwords",
					})
				}

				if len(r.Salt) > 0 && r.HMAC.Valid && r.HMAC.String != "" {
					hx, err := services.HashPasswordHMACHex(newPw, r.Salt)
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

		newSalt, err := services.GenerateSalt16()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "salt error"})
		}
		newHex, err := services.HashPasswordHMACHex(newPw, newSalt)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "hash error"})
		}

		tx, err := db.Beginx()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "tx error"})
		}
		defer tx.Rollback()

		if _, err := tx.Exec(`
			INSERT INTO password_history (user_id, password_hmac, password_fp, salt)
			SELECT id, password_hmac, password_fp, salt
			FROM users
			WHERE id = ?
		`, userID); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "history insert error"})
		}

		if _, err := tx.Exec(`
			UPDATE users
			SET password_hmac = ?, salt = ?, password_fp = ?
			WHERE id = ?
		`, newHex, newSalt, newFP, userID); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "update user error"})
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
			`, userID, userID, nHistory); err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "trim history error"})
			}
		}

		if _, err := tx.Exec(`
			UPDATE password_reset_tokens
			SET used_at = NOW()
			WHERE token_sha1 = ?
		`, sha); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "consume token error"})
		}

		if err := tx.Commit(); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "commit error"})
		}

		return c.JSON(http.StatusOK, map[string]string{"message": "Password reset successfully"})
	}
}
