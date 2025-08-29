package handlers

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

func VerifyEmail(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		raw := c.QueryParam("token")
		if raw == "" {
			return c.HTML(http.StatusBadRequest,
				verificationPage(false, "Invalid Token", "Missing token in request."))
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
			return c.HTML(http.StatusBadRequest,
				verificationPage(false, "Invalid Token", "The verification link is not valid."))
		}
		if err != nil {
			return c.HTML(http.StatusInternalServerError,
				verificationPage(false, "Server Error", "A database error occurred. Please try again later."))
		}
		if usedAt.Valid {
			return c.HTML(http.StatusBadRequest,
				verificationPage(false, "Token Already Used", "This verification link has already been used."))
		}
		if time.Now().After(expiresAt) {
			return c.HTML(http.StatusBadRequest,
				verificationPage(false, "Token Expired", "This verification link has expired. Please register again."))
		}

		tx, err := db.Beginx()
		if err != nil {
			return c.HTML(http.StatusInternalServerError,
				verificationPage(false, "Server Error", "Could not start transaction."))
		}
		defer tx.Rollback()

		if _, err := tx.Exec(`UPDATE users SET is_verified = TRUE WHERE id = ?`, userID); err != nil {
			return c.HTML(http.StatusInternalServerError,
				verificationPage(false, "Server Error", "Could not update user verification status."))
		}
		if _, err := tx.Exec(`UPDATE email_verification_tokens SET used_at = NOW() WHERE token_sha1 = ?`, shaHex); err != nil {
			return c.HTML(http.StatusInternalServerError,
				verificationPage(false, "Server Error", "Could not update verification token status."))
		}

		if err := tx.Commit(); err != nil {
			return c.HTML(http.StatusInternalServerError,
				verificationPage(false, "Server Error", "Could not finalize verification."))
		}

		return c.HTML(http.StatusOK,
			verificationPage(true, "Email Verified", "Your account is now active. You can sign in."))
	}
}

func verificationPage(success bool, title, message string) string {
	color := "#3ce37a"
	if !success {
		color = "#ff6b6b"
	}
	return fmt.Sprintf(`
		<!doctype html>
		<html lang="en">
		<head>
		  <meta charset="utf-8">
		  <title>%s</title>
		  <style>
		    body {
		      font-family: system-ui, Arial, sans-serif;
		      background: linear-gradient(135deg, #0f1221, #1b2b4b);
		      color: #f0f0f0;
		      text-align: center;
		      padding-top: 10%%;
		    }
		    .card {
		      display: inline-block;
		      padding: 32px 48px;
		      border-radius: 16px;
		      background: rgba(255,255,255,0.05);
		      border: 1px solid rgba(255,255,255,0.15);
		      box-shadow: 0 12px 24px rgba(0,0,0,0.3);
		    }
		    h2 { margin-bottom: 12px; color: %s; }
		    p { margin: 0 0 8px; }
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
		    <h2>%s</h2>
		    <p>%s</p>
		    <a href="http://localhost:3000/login">Go to Sign In</a>
		  </div>
		</body>
		</html>
	`, title, color, title, message)
}
