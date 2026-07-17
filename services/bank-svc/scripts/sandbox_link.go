// Command sandbox_link is a standalone test helper, not part of bank-svc
// itself. It mints a fake Plaid Sandbox public_token (simulating what a
// frontend would receive from Plaid Link's onSuccess callback after a user
// logs into their bank) so you can exercise POST /v1/bank/plaid/exchange
// end to end without a frontend.
//
// Usage:
//
//	PLAID_CLIENT_ID=... PLAID_SECRET=... go run ./scripts/sandbox_link.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/plaid/plaid-go/v43/plaid"
)

func main() {
	clientID := os.Getenv("PLAID_CLIENT_ID")
	secret := os.Getenv("PLAID_SECRET")
	if clientID == "" || secret == "" {
		log.Fatalf("PLAID_CLIENT_ID and PLAID_SECRET must be set")
	}

	config := plaid.NewConfiguration()
	config.AddDefaultHeader("PLAID-CLIENT-ID", clientID)
	config.AddDefaultHeader("PLAID-SECRET", secret)
	config.UseEnvironment(plaid.Sandbox)
	client := plaid.NewAPIClient(config)

	ctx := context.Background()

	// ins_109508 is Plaid's well-known Sandbox test institution
	// ("First Platypus Bank").
	sandboxResp, _, err := client.PlaidApi.SandboxPublicTokenCreate(ctx).SandboxPublicTokenCreateRequest(
		*plaid.NewSandboxPublicTokenCreateRequest(
			"ins_109508",
			[]plaid.Products{plaid.PRODUCTS_TRANSACTIONS},
		),
	).Execute()
	if err != nil {
		log.Fatalf("failed to create sandbox public token: %v", err)
	}
	publicToken := sandboxResp.GetPublicToken()

	fmt.Println("Sandbox public_token:", publicToken)
	fmt.Println()
	fmt.Println("This simulates what your frontend would get back from Plaid Link's onSuccess callback.")
	fmt.Println("Feed it into POST /v1/bank/plaid/exchange to see bank-svc do the server-side exchange + persist:")
	fmt.Println()
	fmt.Printf(`curl -s -X POST http://localhost:10001/v1/bank/plaid/exchange -H "Content-Type: application/json" -d '{
  "user_id": 1,
  "bank_name": "First Platypus Bank",
  "public_token": %q
}'
`, publicToken)
	fmt.Println()
	fmt.Println("Note: public_token is short-lived (expires after ~30 minutes) - exchange it soon after generating it.")
}
