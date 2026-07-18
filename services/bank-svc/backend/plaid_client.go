package backend

import (
	"context"
	"fmt"
	"strconv"

	"github.com/plaid/plaid-go/v43/plaid"
)

// plaidClientName is shown to the user inside the Plaid Link UI.
const plaidClientName = "TxAnalyzer"

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

// FetchAccounts validates accessToken against Plaid and returns the real
// accounts registered under it.
func FetchAccounts(ctx context.Context, client *plaid.APIClient, accessToken string) ([]plaid.AccountBase, error) {
	resp, _, err := client.PlaidApi.AccountsGet(ctx).AccountsGetRequest(
		*plaid.NewAccountsGetRequest(accessToken),
	).Execute()
	if err != nil {
		return nil, classifyPlaidError(err)
	}
	return resp.GetAccounts(), nil
}

// CreateLinkToken mints a link_token a frontend can use to launch Plaid
// Link for the given user.
func CreateLinkToken(ctx context.Context, client *plaid.APIClient, userID int) (plaid.LinkTokenCreateResponse, error) {
	req := plaid.NewLinkTokenCreateRequest(
		plaidClientName,
		"en",
		[]plaid.CountryCode{plaid.COUNTRYCODE_US},
	)
	req.SetUser(*plaid.NewLinkTokenCreateRequestUser(strconv.Itoa(userID)))
	req.SetProducts([]plaid.Products{plaid.PRODUCTS_TRANSACTIONS})

	resp, _, err := client.PlaidApi.LinkTokenCreate(ctx).LinkTokenCreateRequest(*req).Execute()
	if err != nil {
		return plaid.LinkTokenCreateResponse{}, classifyPlaidError(err)
	}
	return resp, nil
}

// ExchangePublicToken exchanges a short-lived Link public_token for a
// long-lived access_token. The access_token must never be sent back to a
// client after this point.
func ExchangePublicToken(ctx context.Context, client *plaid.APIClient, publicToken string) (string, error) {
	resp, _, err := client.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(
		*plaid.NewItemPublicTokenExchangeRequest(publicToken),
	).Execute()
	if err != nil {
		return "", classifyPlaidError(err)
	}
	return resp.GetAccessToken(), nil
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
