// @title Caviar API
// @version 1.0
// @description API for managing caviar products and orders
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and token.

package main

import (
	"caviar/internal/app"
	"context"
	"log"

	_ "caviar/docs" // Import for swagger docs

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to load .env file: %v", err)
	}
}

func main() {
	app.MustRun(context.Background())
}