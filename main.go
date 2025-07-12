package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aimrintech/x-backend/api"
	"github.com/aimrintech/x-backend/config"
	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"golang.org/x/oauth2"
)

var AuthConfig *oauth2.Config

func init() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
		panic(err)
	}

	// Initialize auth config
	AuthConfig = config.InitAuthConfig()
}

func main() {
	// Database connection
	dbCtx := context.Background()
	neo4jUri := os.Getenv("NEO4J_URI")
	neo4jUser := os.Getenv("NEO4J_USERNAME")
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
	fmt.Println("Database connection established.")

	// Init server
	server := api.NewServer(&driver, &dbCtx, AuthConfig)
	fmt.Println("Server listening on port 8080")
	server.Start(":8080")
}
