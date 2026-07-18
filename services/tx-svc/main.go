//go:generate oapi-codegen -package api -generate gin,models,types,spec -o api/api.gen.go swagger.yaml

package main

import (
	"database/sql"
	"encoding/base64"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"spendanalyzer.com/tx/api"
	"spendanalyzer.com/tx/backend"
	"spendanalyzer.com/tx/internal/handler"
)

func main() {

	db_path := os.Getenv("DB_PATH")
	db, err := sql.Open("sqlite3", db_path)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

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

	txService := backend.NewTxService(db, plaidClient, encKey)

	r := gin.Default()
	txHandler := handler.NewTxHandler(txService)
	api.RegisterHandlers(r, txHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "10003"
	}
	addr := ":" + port

	log.Printf("Starting the server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Error starting tx-svc")
	}
}

// loadEncryptionKey reads and decodes the AES-256 key used to decrypt Plaid
// access tokens bank-svc encrypted at rest. The process exits if the key is
// missing or not exactly 32 bytes once decoded.
func loadEncryptionKey() []byte {
	raw := os.Getenv("PLAID_TOKEN_ENCRYPTION_KEY")
	if raw == "" {
		log.Fatalf("PLAID_TOKEN_ENCRYPTION_KEY must be set (must match bank-svc's key)")
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
