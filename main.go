package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aimrintech/x-backend/api"
	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Database connection
	dbCtx := context.Background()
	neo4jUri := os.Getenv("NEO4J_URI")
	neo4jUser := os.Getenv("NEO4J_USER")
	neo4jPassword := os.Getenv("NEO4J_PASSWORD")

	driver, err := neo4j.NewDriverWithContext(neo4jUri, neo4j.BasicAuth(neo4jUser, neo4jPassword, ""))
	if err != nil {
		log.Fatalf("Failed to create Neo4j driver: %v", err)
	}
	defer driver.Close(dbCtx)

	// Verify connectivity
	err = driver.VerifyConnectivity(dbCtx)
	if err != nil {
		panic(err)
	}
	fmt.Println("Connection established.")

	// Init server
	server := api.NewServer(&driver)
	server.Start(":8080")
}
