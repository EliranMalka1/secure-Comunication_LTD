package handlers

import (
	"bytes"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"

	"secure-communication-ltd/backend/config"
	"secure-communication-ltd/backend/internal/services"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Register(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		pol := config.GetPolicy()

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
		fpHex, err := services.HashPasswordFingerprintHex(req.Password) // HMAC(pepper, password)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "fingerprint error"})
		}

		res, err := db.Exec(`
			INSERT INTO users (username, email, password_hmac, salt, password_fp, is_verified)
			VALUES (?, ?, ?, ?, ?, FALSE)
		`, req.Username, req.Email, hashHex, salt, fpHex)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "insert error"})
		}
		uid, _ := res.LastInsertId()

		// verification token
		vTok, err := services.NewVerificationToken(24 * time.Hour)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "token error"})
		}
		if _, err := db.Exec(`
			INSERT INTO email_verification_tokens (user_id, token_sha1, expires_at)
			VALUES (?, ?, ?)
		`, uid, vTok.SHA1Hex, vTok.ExpiresAt); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "token save error"})
		}

		// email (HTML built safely via html/template)
		mailer, err := services.NewMailerFromEnv()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "mailer error"})
		}
		base := os.Getenv("BACKEND_PUBLIC_URL")
		if base == "" {
			base = "http://localhost:8080"
		}
		verifyURL := strings.TrimRight(base, "/") + "/api/verify-email?token=" + url.QueryEscape(vTok.Raw)

		// HTML template with automatic escaping
		htmlTpl := template.Must(template.New("verifyEmail").Parse(`
<h2>Verify your email</h2>
<p>Hi {{.Username}}, thanks for registering.</p>
<p>
  <a href="{{.Link}}" style="display:inline-block;padding:10px 16px;border-radius:8px;background:#4f9cff;color:#fff;text-decoration:none">
    Verify Email
  </a>
</p>
<p>If the button doesn't work, copy this URL:</p>
<p><code>{{.Link}}</code></p>
`))

		data := struct {
			Username string
			Link     string
		}{
			Username: req.Username,
			Link:     verifyURL,
		}

		var buf bytes.Buffer
		if err := htmlTpl.Execute(&buf, data); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "template error"})
		}

		if err := mailer.Send(req.Email, "Verify your email", buf.String()); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "send mail error"})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"message": "User registered. Please check your email to verify your account.",
		})
	}
}

func looksLikeEmail(s string) bool {
	return strings.Count(s, "@") == 1 && len(s) >= 6 && strings.Contains(s, ".")
}
