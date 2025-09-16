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
			return RenderVerificationPage(c, http.StatusBadRequest, false,
				"Invalid Token", "Missing token in request.")
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
			return RenderVerificationPage(c, http.StatusBadRequest, false,
				"Invalid Token", "The verification link is not valid.")
		}
		if err != nil {
			return RenderVerificationPage(c, http.StatusInternalServerError, false,
				"Server Error", "A database error occurred. Please try again later.")
		}
		if usedAt.Valid {
			return RenderVerificationPage(c, http.StatusBadRequest, false,
				"Token Already Used", "This verification link has already been used.")
		}
		if time.Now().After(expiresAt) {
			return RenderVerificationPage(c, http.StatusBadRequest, false,
				"Token Expired", "This verification link has expired. Please register again.")
		}

		tx, err := db.Beginx()
		if err != nil {
			return RenderVerificationPage(c, http.StatusInternalServerError, false,
				"Server Error", "Could not start transaction.")
		}
		defer tx.Rollback()

		if _, err := tx.Exec(`UPDATE users SET is_verified = TRUE WHERE id = ?`, userID); err != nil {
			return RenderVerificationPage(c, http.StatusInternalServerError, false,
				"Server Error", "Could not update user verification status.")
		}
		if _, err := tx.Exec(`UPDATE email_verification_tokens SET used_at = NOW() WHERE token_sha1 = ?`, shaHex); err != nil {
			return RenderVerificationPage(c, http.StatusInternalServerError, false,
				"Server Error", "Could not update verification token status.")
		}

		if err := tx.Commit(); err != nil {
			return RenderVerificationPage(c, http.StatusInternalServerError, false,
				"Server Error", "Could not finalize verification.")
		}

		return RenderVerificationPage(c, http.StatusOK, true,
			"Email Verified", "Your account is now active. You can sign in.")
	}
}
