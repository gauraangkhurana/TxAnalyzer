package backend

import (
	"context"
	"fmt"

	"github.com/plaid/plaid-go/v43/plaid"
)

// plaidPageSize is the max transactions requested per page. Plaid caps
// count at 500.
const plaidPageSize = 500

// PlaidValidationError indicates Plaid itself rejected the request (e.g. a
// bad/expired/revoked access_token), as opposed to a transport/infra failure.
type PlaidValidationError struct {
	Code    string
	Message string
}

func (e *PlaidValidationError) Error() string {
	return fmt.Sprintf("plaid rejected request: %s: %s", e.Code, e.Message)
}

// NewPlaidClient builds a Plaid API client for the given environment
// ("sandbox" or "production").
func NewPlaidClient(clientID, secret, env string) (*plaid.APIClient, error) {
	var plaidEnv plaid.Environment
	switch env {
	case "", "sandbox":
		plaidEnv = plaid.Sandbox
	case "production":
		plaidEnv = plaid.Production
	default:
		return nil, fmt.Errorf("unsupported PLAID_ENV %q (must be sandbox or production)", env)
	}

	config := plaid.NewConfiguration()
	config.AddDefaultHeader("PLAID-CLIENT-ID", clientID)
	config.AddDefaultHeader("PLAID-SECRET", secret)
	config.UseEnvironment(plaidEnv)

	return plaid.NewAPIClient(config), nil
}

// FetchTransactions returns every transaction for accountID within
// [dateFrom, dateTo] (both "YYYY-MM-DD"), paginating through Plaid's
// TotalTransactions if the range has more than one page.
func FetchTransactions(ctx context.Context, client *plaid.APIClient, accessToken, accountID, dateFrom, dateTo string) ([]plaid.Transaction, error) {
	var all []plaid.Transaction
	offset := int32(0)

	for {
		req := plaid.NewTransactionsGetRequest(accessToken, dateFrom, dateTo)
		count := int32(plaidPageSize)
		req.SetOptions(plaid.TransactionsGetRequestOptions{
			AccountIds: &[]string{accountID},
			Count:      &count,
			Offset:     &offset,
		})

		resp, _, err := client.PlaidApi.TransactionsGet(ctx).TransactionsGetRequest(*req).Execute()
		if err != nil {
			return nil, classifyPlaidError(err)
		}

		page := resp.GetTransactions()
		all = append(all, page...)

		if int32(len(all)) >= resp.GetTotalTransactions() || len(page) == 0 {
			break
		}
		offset += int32(len(page))
	}

	return all, nil
}

// classifyPlaidError distinguishes a structured Plaid API error (bad token,
// item login required, etc.) from a transport/infra failure.
func classifyPlaidError(err error) error {
	plaidErr, convErr := plaid.ToPlaidError(err)
	if convErr != nil {
		// Not a structured Plaid error - network/transport/context failure.
		return fmt.Errorf("plaid request failed: %w", err)
	}
	return &PlaidValidationError{
		Code:    plaidErr.ErrorCode,
		Message: plaidErr.ErrorMessage,
	}
}
