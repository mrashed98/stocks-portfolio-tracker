package main

import (
	"log"

	"github.com/joho/godotenv"
	"portfolio-app/config"
	"portfolio-app/internal/database"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := database.NewConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create seeder and run
	seeder := database.NewSeeder(db)
	if err := seeder.SeedDevelopmentData(); err != nil {
		log.Fatalf("Failed to seed development data: %v", err)
	}

	log.Println("Development data seeded successfully!")
}