//go:generate oapi-codegen -package api -generate gin,models,types,spec -o api/api.gen.go swagger.yaml

package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"spendanalyzer.com/bank/api"
	"spendanalyzer.com/bank/internal/handler"
)

func main() {

	db_path := os.Getenv("DB_PATH")
	// Open the database connection
	db, err := sql.Open("sqlite3", db_path)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Verify the db conn works fine
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Printf("Successfully connected to the SQLITE database")

	// gin router with default middleware
	r := gin.Default()
	bankHandler := handler.NewBankHandler(db)
	api.RegisterHandlers(r, bankHandler)

	// prepare server config
	port := os.Getenv("PORT")
	if port == "" {
		port = "10001"
	}
	addr := ":" + port

	// Start the server on localhost - DB_PORT (default: 10001)
	log.Printf("Starting the server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Error starting bank-svc")
	}
}

// TO-DO: Make sure when (saving/querying/retrieving) bank information, it is
// 1. all UPPERCASE
// 2. Trimmed
