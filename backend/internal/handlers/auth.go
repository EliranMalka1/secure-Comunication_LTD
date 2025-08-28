package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

	"secure-communication-ltd/backend/internal/services"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Register(db *sqlx.DB, pol services.PasswordPolicy) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req RegisterRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
		}
		req.Username = strings.TrimSpace(req.Username)
		req.Email = strings.TrimSpace(req.Email)

		if req.Username == "" || req.Email == "" || req.Password == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing fields"})
		}
		if !looksLikeEmail(req.Email) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid email"})
		}
		if err := services.ValidatePassword(req.Password, pol); err != nil {
			return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
		}

		var exists int
		if err := db.Get(&exists, `SELECT COUNT(*) FROM users WHERE username = ? OR email = ?`, req.Username, req.Email); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
		}
		if exists > 0 {
			return c.JSON(http.StatusConflict, map[string]string{"error": "username or email already exists"})
		}

		salt, err := services.GenerateSalt16()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "salt error"})
		}
		hashHex, err := services.HashPasswordHMACHex(req.Password, salt)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "hash error"})
		}

		res, err := db.Exec(`
			INSERT INTO users (username, email, password_hmac, salt, is_verified)
			VALUES (?, ?, ?, ?, FALSE)`,
			req.Username, req.Email, hashHex, salt)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "insert error"})
		}
		uid, _ := res.LastInsertId()

		// Create verification token and save to DB
		vTok, err := services.NewVerificationToken(24 * time.Hour)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "token error"})
		}
		_, err = db.Exec(`
			INSERT INTO email_verification_tokens (user_id, token_sha1, expires_at)
			VALUES (?, ?, ?)`,
			uid, vTok.SHA1Hex, vTok.ExpiresAt)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "token save error"})
		}

		// Send email
		mailer, err := services.NewMailerFromEnv()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "mailer error"})
		}
		base := os.Getenv("BACKEND_PUBLIC_URL")
		if base == "" {
			base = "http://localhost:8080"
		}
		link := fmt.Sprintf("%s/api/verify-email?token=%s", strings.TrimRight(base, "/"), vTok.Raw)

		html := fmt.Sprintf(`
			<h2>Verify your email</h2>
			<p>Hi %s, thanks for registering.</p>
			<p>Please click the button below to verify your email address:</p>
			<p><a href="%s" style="display:inline-block;padding:10px 16px;border-radius:8px;background:#4f9cff;color:#fff;text-decoration:none">Verify Email</a></p>
			<p>If the button doesn't work, copy this URL:</p>
			<p><code>%s</code></p>
		`, req.Username, link, link)

		if err := mailer.Send(req.Email, "Verify your email", html); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "send mail error"})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"message": "User registered. Please check your email to verify your account.",
		})
	}
}

// Basic email validation (client already does basic validation; this is server-side reinforcement)
func looksLikeEmail(s string) bool {
	return strings.Count(s, "@") == 1 && len(s) >= 6 && strings.Contains(s, ".")
}
