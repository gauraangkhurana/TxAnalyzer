//go:generate oapi-codegen -package api -generate gin,models,types,spec -o api/api.gen.go swagger.yaml

package main

import (
	"database/sql"
	"encoding/base64"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"spendanalyzer.com/bank/api"
	"spendanalyzer.com/bank/backend"
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

	encKey := loadEncryptionKey()

	plaidClientID := os.Getenv("PLAID_CLIENT_ID")
	plaidSecret := os.Getenv("PLAID_SECRET")
	if plaidClientID == "" || plaidSecret == "" {
		log.Fatalf("PLAID_CLIENT_ID and PLAID_SECRET must be set")
	}
	plaidEnv := os.Getenv("PLAID_ENV")
	plaidClient, err := backend.NewPlaidClient(plaidClientID, plaidSecret, plaidEnv)
	if err != nil {
		log.Fatalf("Failed to create Plaid client: %v", err)
	}

	bankService := backend.NewBankService(db, plaidClient, encKey)

	// gin router with default middleware
	r := gin.Default()
	bankHandler := handler.NewBankHandler(bankService)
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

// loadEncryptionKey reads and decodes the AES-256 key used to encrypt Plaid
// access tokens at rest. The process exits if the key is missing or not
// exactly 32 bytes once decoded, since that would silently produce
// unrecoverable data.
func loadEncryptionKey() []byte {
	raw := os.Getenv("PLAID_TOKEN_ENCRYPTION_KEY")
	if raw == "" {
		log.Fatalf("PLAID_TOKEN_ENCRYPTION_KEY must be set (generate with: openssl rand -base64 32)")
	}
	key, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		log.Fatalf("PLAID_TOKEN_ENCRYPTION_KEY must be valid base64: %v", err)
	}
	if len(key) != 32 {
		log.Fatalf("PLAID_TOKEN_ENCRYPTION_KEY must decode to 32 bytes, got %d", len(key))
	}
	return key
}

// TO-DO: Make sure when (saving/querying/retrieving) bank information, it is
// 1. all UPPERCASE
// 2. Trimmed
