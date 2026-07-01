package main

import (
	"context"
	"log"
	"os"

	"github.com/echoline/echoline/backend/internal/migrate"
)

func main() {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if err := migrate.Up(context.Background(), url); err != nil {
		log.Fatal(err)
	}
	log.Println("migrations applied")
}
