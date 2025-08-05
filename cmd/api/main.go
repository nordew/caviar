package main

import (
	"caviar/internal/app"
	"context"
	"log"

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