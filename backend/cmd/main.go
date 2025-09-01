package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"secure-communication-ltd/backend/config"
	"secure-communication-ltd/backend/internal/handlers"
	middlewarex "secure-communication-ltd/backend/internal/middleware"
	"secure-communication-ltd/backend/internal/repository"
	"secure-communication-ltd/backend/internal/services"
)

func main() {
	// Local .env (in Docker loaded from env_file)
	_ = godotenv.Load(".env")

	// DB
	db, err := repository.NewMySQL()
	if err != nil {
		log.Fatal("db connect error: ", err)
	}
	defer db.Close()

	// Echo
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// CORS for development; in production, restrict to your domain
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"http://localhost:5173", // Vite dev
			"http://127.0.0.1:5173", // Vite dev
			"http://localhost:3000", // Docker nginx
			"http://127.0.0.1:3000", // Docker nginx
		},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderContentType, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	// Load password policy from config file
	policyPath := os.Getenv("PASSWORD_POLICY_FILE")
	if policyPath == "" {
		policyPath = "config/password-policy.toml"
	}
	pol, err := config.LoadPasswordPolicy(policyPath)
	if err != nil {
		log.Println("Warning: using default password policy:", err)
		pol = services.DefaultPolicy()
	}

	// Health / Hello
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	e.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, Secure Backend!")
	})

	// Auth
	e.POST("/api/register", handlers.Register(db, pol))
	e.GET("/api/verify-email", handlers.VerifyEmail(db))
	e.POST("/api/login", handlers.Login(db, pol))
	e.POST("/api/logout", handlers.Logout())
	e.GET("/api/me", handlers.Me(), middlewarex.RequireAuth)
	e.POST("/api/login/mfa", handlers.LoginMFA(db))
	e.POST("/api/customers", handlers.CreateCustomer(db), middlewarex.RequireAuth)
	// Forgot / Reset password
	e.POST("/api/password/forgot", handlers.PasswordForgot(db))
	e.GET("/api/password/reset", handlers.PasswordResetLanding()) // redirect ל-frontend
	e.POST("/api/password/reset", handlers.PasswordReset(db, pol))
	// Change password (authenticated)
	e.POST("/api/password/change", handlers.ChangePassword(db, pol), middlewarex.RequireAuth)
	e.GET("/api/customers/search", handlers.SearchCustomers(db), middlewarex.RequireAuth)

	e.GET("/api/password/change/confirm", handlers.ChangePasswordConfirm(db))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on :%s\n", port)
	log.Fatal(e.Start(":" + port))
}
