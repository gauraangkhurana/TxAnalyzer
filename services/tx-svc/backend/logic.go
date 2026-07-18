package backend

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/plaid/plaid-go/v43/plaid"
	"spendanalyzer.com/tx/api"
)

var FindAccountTokenQuery = "SELECT a.plaid_token FROM accounts a JOIN banks b ON a.bank_id = b.id WHERE a.user_id = ? AND b.name = ? AND a.account_id = ?"

// ErrAccountNotFound indicates no linked account matches the given
// user/bank/account_id combination.
var ErrAccountNotFound = errors.New("no linked account found for the given user, bank, and account_id")

var UpsertTransactionQuery = `INSERT INTO transactions (account_id, plaid_transaction_id, amount, tx_date, name, merchant_name, category, pending)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(plaid_transaction_id) DO UPDATE SET
	amount = excluded.amount,
	tx_date = excluded.tx_date,
	name = excluded.name,
	merchant_name = excluded.merchant_name,
	category = excluded.category,
	pending = excluded.pending`
var GetTransactionsQuery = `SELECT t.plaid_transaction_id, b.name, t.account_id, t.tx_date, t.name, t.merchant_name, t.amount, t.category, t.pending
FROM transactions t
JOIN accounts a ON t.account_id = a.account_id
JOIN banks b ON a.bank_id = b.id
WHERE a.user_id = ?
ORDER BY t.tx_date DESC`

// TxService bundles the dependencies needed to pull and serve transactions:
// the database, a Plaid API client, and the key used to decrypt access
// tokens bank-svc encrypted at rest.
type TxService struct {
	DB     *sql.DB
	Plaid  *plaid.APIClient
	EncKey []byte
}

func NewTxService(db *sql.DB, plaidClient *plaid.APIClient, encKey []byte) *TxService {
	return &TxService{DB: db, Plaid: plaidClient, EncKey: encKey}
}

// PullTransactions fetches transactions for one account over a date range
// from Plaid and upserts them into the transactions table.
func (s *TxService) PullTransactions(ctx context.Context, req api.PullTransactionsRequest) (int, error) {
	var encryptedToken string
	err := s.DB.QueryRow(FindAccountTokenQuery, req.UserId, req.BankName, req.AccountId).Scan(&encryptedToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrAccountNotFound
		}
		return 0, err
	}

	accessToken, err := Decrypt(encryptedToken, s.EncKey)
	if err != nil {
		return 0, fmt.Errorf("failed to decrypt access token: %w", err)
	}

	transactions, err := FetchTransactions(ctx, s.Plaid, accessToken, req.AccountId, req.DateFrom, req.DateTo)
	if err != nil {
		return 0, err
	}

	for _, t := range transactions {
		var category *string
		if pfc := t.GetPersonalFinanceCategory(); pfc.Primary != "" {
			category = &pfc.Primary
		}
		var merchantName *string
		if name := t.GetMerchantName(); name != "" {
			merchantName = &name
		}

		_, err := s.DB.Exec(UpsertTransactionQuery,
			t.GetAccountId(), t.GetTransactionId(), t.GetAmount(), t.GetDate(),
			t.GetName(), merchantName, category, t.GetPending(),
		)
		if err != nil {
			return 0, err
		}
	}

	return len(transactions), nil
}

// GetTransactions returns every stored transaction for a user, across all
// their linked accounts and banks, sorted most recent first.
func (s *TxService) GetTransactions(userID int) ([]api.Transaction, error) {
	rows, err := s.DB.Query(GetTransactionsQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := make([]api.Transaction, 0)
	for rows.Next() {
		var tx api.Transaction
		var merchantName, category sql.NullString
		var pendingInt int
		if err := rows.Scan(&tx.TransactionId, &tx.BankName, &tx.AccountId, &tx.Date, &tx.Name, &merchantName, &tx.Amount, &category, &pendingInt); err != nil {
			return nil, err
		}
		if merchantName.Valid {
			tx.MerchantName = &merchantName.String
		}
		if category.Valid {
			tx.Category = &category.String
		}
		tx.Pending = pendingInt != 0
		transactions = append(transactions, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}
