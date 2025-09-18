package handlers

import (
	"html"
	"net/http"
	"strings"
	"unicode/utf8"

	middlewarex "secure-communication-ltd/backend/internal/middleware"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	bluemonday "github.com/microcosm-cc/bluemonday"
)

// Global sanitation policy: complete removal of HTML/JS (plain text only)
var notesPolicy = func() *bluemonday.Policy {
	return bluemonday.StrictPolicy()
}()

type CreateCustomerRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone,omitempty"`
	Notes string `json:"notes,omitempty"`
}

type CreateCustomerResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func CreateCustomer(db *sqlx.DB) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Must be logged in
		if _, err := middlewarex.UserIDFromCtx(c); err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}

		var req CreateCustomerRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid json"})
		}

		// Basic normalization
		req.Name = strings.TrimSpace(req.Name)
		req.Email = strings.TrimSpace(req.Email)
		req.Phone = strings.TrimSpace(req.Phone)

		// Basic validation
		if req.Name == "" || req.Email == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "name and email are required"})
		}
		if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid email"})
		}
		if utf8.RuneCountInString(req.Name) > 255 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "name too long"})
		}
		if utf8.RuneCountInString(req.Email) > 255 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "email too long"})
		}
		if utf8.RuneCountInString(req.Phone) > 40 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "phone too long"})
		}

		// Light normalization for notes before sanitation
		notes := strings.ReplaceAll(req.Notes, "\r\n", "\n")
		notes = strings.ReplaceAll(notes, "\x00", "")
		notes = strings.TrimSpace(notes)

		// Server-side sanitation: removes all HTML/JS and leaves plain text
		safeNotes := notesPolicy.Sanitize(notes)

		// Unescape entities to regular text characters (", ', &) - safe because there are no more tags/JS
		safeNotes = html.UnescapeString(safeNotes)

		// Length limit after filtering
		if utf8.RuneCountInString(safeNotes) > 10000 {
			safeNotes = string([]rune(safeNotes)[:10000])
		}

		// Save to DB with a parameterized query
		res, err := db.Exec(`
			INSERT INTO customers (name, email, phone, notes)
			VALUES (?, ?, ?, ?)`,
			req.Name, req.Email, req.Phone, safeNotes,
		)
		if err != nil {
			return c.JSON(http.StatusConflict, map[string]string{"error": "could not create customer"})
		}

		id, _ := res.LastInsertId()
		return c.JSON(http.StatusCreated, CreateCustomerResponse{ID: id, Name: req.Name})
	}
}
