//go:generate oapi-codegen -package api -generate gin,models,types,spec -o api/api.gen.go swagger.yaml

package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"spendanalyzer.com/bank/api"
	"spendanalyzer.com/bank/internal/handler"
)

func main() {

	// Open the database connection
	db, err := sql.Open("sqlite3", "../../db/data/database.db")
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

	// Start the server on 8080
	// Server will listen on localhost
	r.Run("0.0.0.0:10001")
}
