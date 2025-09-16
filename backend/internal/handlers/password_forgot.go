package handlers

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"secure-communication-ltd/backend/internal/services"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

type forgotReq struct {
	Email string `json:"email"`
}

// Always return a generic response — even if the email does not exist.
func PasswordForgot(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req forgotReq
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
		}
		email := strings.TrimSpace(req.Email)
		if email == "" || !strings.Contains(email, "@") {
			// Generic response (don't reveal anything)
			return c.JSON(http.StatusOK, map[string]string{"message": "If this email exists, a reset link has been sent."})
		}

		// Is there such a user and are they verified?
		var userID int64
		if err := db.Get(&userID, `SELECT id FROM users WHERE email = ? AND is_verified = TRUE LIMIT 1`, email); err != nil {
			// Do not reveal if not found — return generic response
			return c.JSON(http.StatusOK, map[string]string{"message": "If this email exists, a reset link has been sent."})
		}

		// Invalidate old open tokens (best-effort)
		_, _ = db.Exec(`UPDATE password_reset_tokens SET used_at = NOW() WHERE user_id = ? AND used_at IS NULL`, userID)

		// Create a new token (raw) and store SHA-1 only
		raw, err := services.NewRandomBase64URL(32) // ~43 chars
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "token error"})
		}
		sum := sha1.Sum([]byte(raw))
		sha := hex.EncodeToString(sum[:])

		exp := time.Now().Add(30 * time.Minute)
		if _, err := db.Exec(`
			INSERT INTO password_reset_tokens (user_id, token_sha1, expires_at)
			VALUES (?, ?, ?)`, userID, sha, exp); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "db error"})
		}

		// Send email (safe HTML via html/template)
		mailer, err := services.NewMailerFromEnv()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "mailer error"})
		}
		base := os.Getenv("BACKEND_PUBLIC_URL")
		if base == "" {
			base = "http://localhost:8080"
		}
		link := strings.TrimRight(base, "/") + "/api/password/reset?token=" + url.QueryEscape(raw)

		// HTML template (auto-escaping of {{.Link}})
		tpl := template.Must(template.New("pwReset").Parse(`
<h2>Password Reset</h2>
<p>We received a request to reset your password.</p>
<p>
  <a href="{{.Link}}" style="display:inline-block;padding:10px 16px;border-radius:8px;background:#4f9cff;color:#fff;text-decoration:none">
    Set a new password
  </a>
</p>
<p>If the button doesn't work, copy this URL:</p>
<p><code>{{.Link}}</code></p>
<p>This link expires in 30 minutes.</p>
`))

		var buf bytes.Buffer
		_ = tpl.Execute(&buf, struct{ Link string }{Link: link})

		// Best-effort send (don’t leak errors to user about existence)
		_ = mailer.Send(email, "Reset your password", buf.String())

		// Always generic
		return c.JSON(http.StatusOK, map[string]string{"message": "If this email exists, a reset link has been sent."})
	}
}
