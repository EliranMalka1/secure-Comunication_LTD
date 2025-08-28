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

		return c.HTML(http.StatusOK, `
  <!doctype html>
  <html lang="en">
  <head>
    <meta charset="utf-8">
    <title>Email Verified</title>
    <style>
      body {
        font-family: system-ui, Arial, sans-serif;
        background: linear-gradient(135deg, #0f1221, #1b2b4b);
        color: #f0f0f0;
        text-align: center;
        padding-top: 10%;
      }
      .card {
        display: inline-block;
        padding: 32px 48px;
        border-radius: 16px;
        background: rgba(255,255,255,0.05);
        border: 1px solid rgba(255,255,255,0.15);
        box-shadow: 0 12px 24px rgba(0,0,0,0.3);
      }
      h2 { margin-bottom: 12px; }
      a {
        display:inline-block;
        margin-top: 16px;
        padding: 10px 20px;
        border-radius: 8px;
        background: linear-gradient(135deg,#6c8bff,#55e7ff);
        color: #0b1120;
        text-decoration: none;
        font-weight: 600;
      }
      a:hover { box-shadow: 0 4px 12px rgba(110,160,255,0.5); }
    </style>
  </head>
  <body>
    <div class="card">
      <h2>✅ Email Verified</h2>
      <p>Your account is now active. You can sign in.</p>
      <a href="http://localhost:3000/login">Go to Sign In</a>
    </div>
  </body>
  </html>
`)

	}
}
