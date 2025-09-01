package handlers

import (
	"net/http"
	"strings"

	middlewarex "secure-communication-ltd/backend/internal/middleware"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

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
		req.Name = strings.TrimSpace(req.Name)
		req.Email = strings.TrimSpace(req.Email)
		req.Phone = strings.TrimSpace(req.Phone)
		// Short validation
		if req.Name == "" || req.Email == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "name and email are required"})
		}
		if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid email"})
		}
		// Secure insertion (prepared)
		res, err := db.Exec(`
			INSERT INTO customers (name, email, phone, notes)
			VALUES (?, ?, ?, ?)`,
			req.Name, req.Email, req.Phone, req.Notes,
		)
		if err != nil {
			// Possible DUPLICATE email
			return c.JSON(http.StatusConflict, map[string]string{"error": "could not create customer"})
		}
		id, _ := res.LastInsertId()
		return c.JSON(http.StatusCreated, CreateCustomerResponse{ID: id, Name: req.Name})
	}
}
