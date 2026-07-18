package backend

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/plaid/plaid-go/v43/plaid"
	"spendanalyzer.com/bank/api"
)

var FindBankIDQuery = "SELECT id FROM banks WHERE name = (?)"
var FindBankNameQuery = "SELECT name FROM banks WHERE id = (?)"
var GetBankInstitutionQuery = "SELECT id, name FROM banks"
var SaveBankInstitutionQuery = "INSERT INTO banks (name) VALUES (?)"
var SaveBankAccountQuery = "INSERT INTO accounts (user_id, bank_id, account_id, account_type, plaid_token) VALUES (?, ?, ?, ?, ?)"
var GetUserFinancialProfileQuery = "SELECT bank_id, account_id, account_type, plaid_token FROM accounts WHERE user_id = (?)"

// BankService bundles the dependencies needed to serve bank-svc requests:
// the database, a Plaid API client, and the key used to encrypt access
// tokens at rest.
type BankService struct {
	DB     *sql.DB
	Plaid  *plaid.APIClient
	EncKey []byte
}

func NewBankService(db *sql.DB, plaidClient *plaid.APIClient, encKey []byte) *BankService {
	return &BankService{DB: db, Plaid: plaidClient, EncKey: encKey}
}

// Gets all the financial institutions in the db
func (s *BankService) GetInstitutions() (api.FinancialInstitutionsResponse, error) {
	resp := api.FinancialInstitutionsResponse{}
	var banks []api.BankItem

	rows, err := s.DB.Query(GetBankInstitutionQuery)
	if err != nil {
		return resp, err
	}
	defer rows.Close()

	for rows.Next() {
		var bankID int
		var bankName string
		if err := rows.Scan(&bankID, &bankName); err != nil {
			return resp, err
		}
		banks = append(banks, api.BankItem{BankId: bankID, BankName: bankName})
	}

	if err = rows.Err(); err != nil {
		return resp, err
	}

	resp.Banks = banks
	return resp, nil
}

// Saves the new bank and returns the auto-generated bankID
func (s *BankService) AddInstitution(bankName string) (api.SaveNewBankResponse, error) {
	var resp api.SaveNewBankResponse

	result, err := s.DB.Exec(SaveBankInstitutionQuery, bankName)
	if err != nil {
		return resp, err
	}
	bankID, err := result.LastInsertId()
	if err != nil {
		return resp, err
	}
	resp.BankId = int(bankID)

	return resp, nil
}

// Fetches a Bank ID given a Bank Name
func (s *BankService) GetBankIDByName(bankName string) (int, error) {
	var bankID int
	err := s.DB.QueryRow(FindBankIDQuery, bankName).Scan(&bankID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return -1, fmt.Errorf("bank %s has not been registered yet", bankName)
		}
		return -1, err
	}
	return bankID, nil
}

// Fetches a Bank Name given a Bank ID
func (s *BankService) GetBankNameByID(bankID int) (string, error) {
	var bankName string
	err := s.DB.QueryRow(FindBankNameQuery, bankID).Scan(&bankName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("bank id %d not found", bankID)
		}
		return "", err
	}
	return bankName, nil
}

// findOrCreateBankID looks up a bank by name, registering it if it doesn't
// exist yet. Used by flows where the bank name comes from Plaid itself
// (trusted) rather than free-text user input.
func (s *BankService) findOrCreateBankID(bankName string) (int, error) {
	bankID, err := s.GetBankIDByName(bankName)
	if err == nil {
		return bankID, nil
	}
	resp, err := s.AddInstitution(bankName)
	if err != nil {
		return -1, err
	}
	return resp.BankId, nil
}

// persistAccounts encrypts accessToken once and inserts one accounts row
// per entry in plaidAccounts, returning what was saved.
func (s *BankService) persistAccounts(userID, bankID int, accessToken string, plaidAccounts []plaid.AccountBase) ([]api.Account, error) {
	encryptedToken, err := Encrypt(accessToken, s.EncKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt access token: %w", err)
	}

	saved := make([]api.Account, 0, len(plaidAccounts))
	for _, acct := range plaidAccounts {
		accountType := string(acct.GetType())
		result, err := s.DB.Exec(SaveBankAccountQuery, userID, bankID, acct.GetAccountId(), accountType, encryptedToken)
		if err != nil {
			return nil, err
		}
		if rows, _ := result.RowsAffected(); rows == 0 {
			return nil, fmt.Errorf("no rows were saved")
		}
		saved = append(saved, api.Account{AccountId: acct.GetAccountId(), AccountType: accountType})
	}
	return saved, nil
}

// SaveBankAccount validates the client-supplied access_token against Plaid,
// confirms the requested account actually belongs to that token, and
// persists the account using Plaid's own account data (not the client's).
// The access token is encrypted before it is written to the database.
func (s *BankService) SaveBankAccount(ctx context.Context, req api.SaveNewAccountRequest) error {
	bankID, err := s.GetBankIDByName(req.BankName)
	if err != nil {
		return err
	}

	plaidAccounts, err := FetchAccounts(ctx, s.Plaid, req.AccessToken)
	if err != nil {
		return err
	}

	var matched *plaid.AccountBase
	for i := range plaidAccounts {
		if plaidAccounts[i].GetAccountId() == req.Accounts.AccountId {
			matched = &plaidAccounts[i]
			break
		}
	}
	if matched == nil {
		return &PlaidValidationError{
			Code:    "ACCOUNT_NOT_FOUND",
			Message: fmt.Sprintf("account %s does not belong to the provided access token", req.Accounts.AccountId),
		}
	}

	_, err = s.persistAccounts(req.UserId, bankID, req.AccessToken, []plaid.AccountBase{*matched})
	return err
}

// CreateLinkToken mints a Plaid link_token for the given user so a frontend
// can launch Plaid Link.
func (s *BankService) CreateLinkToken(ctx context.Context, userID int) (api.CreateLinkTokenResponse, error) {
	resp, err := CreateLinkToken(ctx, s.Plaid, userID)
	if err != nil {
		return api.CreateLinkTokenResponse{}, err
	}
	return api.CreateLinkTokenResponse{
		LinkToken:  resp.GetLinkToken(),
		Expiration: resp.GetExpiration().Format(time.RFC3339),
	}, nil
}

// ExchangePublicToken exchanges a Plaid Link public_token for an
// access_token (server-side only, never returned to the client), then
// fetches and persists every account under the resulting Item.
func (s *BankService) ExchangePublicToken(ctx context.Context, req api.ExchangePublicTokenRequest) (api.ExchangePublicTokenResponse, error) {
	var resp api.ExchangePublicTokenResponse

	accessToken, err := ExchangePublicToken(ctx, s.Plaid, req.PublicToken)
	if err != nil {
		return resp, err
	}

	bankID, err := s.findOrCreateBankID(req.BankName)
	if err != nil {
		return resp, err
	}

	plaidAccounts, err := FetchAccounts(ctx, s.Plaid, accessToken)
	if err != nil {
		return resp, err
	}

	saved, err := s.persistAccounts(req.UserId, bankID, accessToken, plaidAccounts)
	if err != nil {
		return resp, err
	}

	resp.BankId = bankID
	resp.Accounts = saved
	return resp, nil
}

// Fetches a user's financial accounts data, grouped by bank. Access tokens
// are never decrypted or included here - a user can legitimately link the
// same institution more than once (separate logins produce separate Plaid
// Items with different tokens), so accounts under one bank name are not
// assumed to share a single token.
func (s *BankService) GetUserFinancialProfile(userID int) (api.UserFinancialProfile, error) {
	var resp api.UserFinancialProfile
	resp.UserId = &userID

	var banks []api.UserAccountInfo

	rows, err := s.DB.Query(GetUserFinancialProfileQuery, userID)
	if err != nil {
		return resp, err
	}
	defer rows.Close()

	bankAccountMap := make(map[string][]api.Account)

	for rows.Next() {
		var bankID int
		var accountID string
		var accountType string
		var encryptedToken string
		if err := rows.Scan(&bankID, &accountID, &accountType, &encryptedToken); err != nil {
			return resp, fmt.Errorf("could not find account information for the user")
		}

		currentAccount := api.Account{AccountId: accountID, AccountType: accountType}
		bankName, err := s.GetBankNameByID(bankID)
		if err != nil {
			return resp, err
		}

		bankAccountMap[bankName] = append(bankAccountMap[bankName], currentAccount)
	}

	for key, accounts := range bankAccountMap {
		banks = append(banks, api.UserAccountInfo{
			BankName: key,
			Accounts: accounts,
		})
	}

	resp.Banks = banks
	return resp, nil
}
