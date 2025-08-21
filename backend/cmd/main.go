package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load environment variables from .env file (if exists)
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found, falling back to system environment")
	}

	// Create a new Echo instance (our web server)
	e := echo.New()

	// Middleware:
	// Logger -> logs each request
	// Recover -> prevents crashes from panics
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	e.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, Secure Backend!")
	})

	// Determine port from ENV (default: 8080)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Starting server on port %s...\n", port)
	log.Fatal(e.Start(":" + port))
}
