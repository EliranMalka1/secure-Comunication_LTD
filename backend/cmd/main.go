package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"secure-communication-ltd/backend/config"
	"secure-communication-ltd/backend/internal/handlers"
	middlewarex "secure-communication-ltd/backend/internal/middleware"
	"secure-communication-ltd/backend/internal/repository"
)

func main() {

	_ = godotenv.Load(".env")

	db, err := repository.NewMySQL()
	if err != nil {
		log.Fatal("db connect error: ", err)
	}
	defer db.Close()

	policyPath := os.Getenv("PASSWORD_POLICY_FILE")
	if policyPath == "" {
		policyPath = "config/password-policy.toml"
	}
	if err := config.InitRuntimePolicy(policyPath); err != nil {

		log.Printf("policy init warning: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := config.WatchPolicy(ctx, policyPath); err != nil {
		log.Printf("policy watch warning: %v", err)
	}

	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"http://localhost:5173", // Vite dev
			"http://127.0.0.1:5173",
			"http://localhost:3000", // Docker/nginx
			"http://127.0.0.1:3000",
		},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderContentType, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	e.GET("/health", func(c echo.Context) error { return c.String(http.StatusOK, "ok") })
	e.GET("/hello", func(c echo.Context) error { return c.String(http.StatusOK, "Hello, Secure Backend!") })

	e.POST("/api/register", handlers.Register(db))
	e.GET("/api/verify-email", handlers.VerifyEmail(db))
	e.POST("/api/login", handlers.Login(db))
	e.POST("/api/logout", handlers.Logout())
	e.GET("/api/me", handlers.Me(), middlewarex.RequireAuth)
	e.POST("/api/login/mfa", handlers.LoginMFA(db))
	e.POST("/api/customers", handlers.CreateCustomer(db), middlewarex.RequireAuth)
	e.GET("/api/customers/search", handlers.SearchCustomers(db), middlewarex.RequireAuth)

	// Forgot / Reset password
	e.POST("/api/password/forgot", handlers.PasswordForgot(db))
	e.GET("/api/password/reset", handlers.PasswordResetLanding())
	e.POST("/api/password/reset", handlers.PasswordReset(db))
	e.GET("/api/policy", func(c echo.Context) error {
		p := config.GetPolicy()
		return c.JSON(http.StatusOK, p)
	})
	// Change password (authenticated)
	e.POST("/api/password/change", handlers.ChangePassword(db), middlewarex.RequireAuth)
	e.GET("/api/password/change/confirm", handlers.ChangePasswordConfirm(db))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	go func() {
		log.Printf("Starting server on :%s\n", port)
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	if err := e.Shutdown(ctxShutdown); err != nil {
		log.Printf("server forced to shutdown: %v", err)
	}
}
