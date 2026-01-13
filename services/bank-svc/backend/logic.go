package backend

import (
	"database/sql"
	"errors"
	"fmt"

	"spendanalyzer.com/bank/api"
)

var FindBankIDQuery = "SELECT id FROM banks WHERE name = (?)"
var FindBankNameQuery = "SELECT name FROM banks WHERE id = (?)"
var GetBankInstitutionQuery = "SELECT id, name FROM banks"
var SaveBankInstitutionQuery = "INSERT INTO banks (name) VALUES (?)"
var SaveBankAccountQuery = "INSERT INTO accounts (user_id, bank_id, account_id, account_type, plaid_token) VALUES (?, ?, ?, ?, ?)"
var GetUserFinancialProfileQuery = "SELECT bank_id, account_id, account_type, plaid_token FROM accounts WHERE user_id = (?)"

// Gets all the financial institutions in the db
func GetInstitutions(db *sql.DB) (api.FinancialInstitutionsResponse, error) {
	resp := api.FinancialInstitutionsResponse{}
	var banks []api.BankItem

	// Query to call the database
	rows, err := db.Query(GetBankInstitutionQuery)
	if err != nil {
		// Error encountered while getting institutions
		return resp, err
	}
	defer rows.Close()

	// Parse the response in the format
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
	// return the response
	return resp, nil
}

// Saves the new bank and returns the auto-generated bankID
func AddInstitution(db *sql.DB, bankName string) (api.SaveNewBankResponse, error) {
	var resp api.SaveNewBankResponse

	// Use query to save inside the database
	result, err := db.Exec(SaveBankInstitutionQuery, bankName)
	if err != nil {
		// Error encountered while saving institution
		return resp, err
	}
	// Get the auto-generated bankID
	bankID, err := result.LastInsertId()
	if err != nil {
		return resp, err
	}
	resp.BankId = int(bankID)

	return resp, nil
}

// Fetches a Bank ID given a Bank Name
func GetBankIDByName(db *sql.DB, bankName string) (int, error) {
	var bankID int
	// Query a single row for ID since bankNames are unique
	err := db.QueryRow(FindBankIDQuery, bankName).Scan(&bankID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No Bank ID was captured for this bank name
			return -1, fmt.Errorf("bank %s has not been registered yet", bankName)
		}
		return -1, err
	}
	return bankID, nil
}

// Fetches a Bank Name given a Bank ID
func GetBankNameByID(db *sql.DB, bankID int) (string, error) {
	var bankName string
	// Query a single row for ID since bankNames are unique
	err := db.QueryRow(FindBankNameQuery, bankID).Scan(&bankName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No Bank ID was captured for this bank name
			return "", fmt.Errorf("bank id %s not found", bankID)
		}
		return "", err
	}
	return bankName, nil
}

// Save a bank account for a user
func SaveBankAcount(db *sql.DB, req api.SaveNewAccountRequest) error {
	// Fetch bank_id as that will be used to save the entry
	bank_id, err := GetBankIDByName(db, req.BankName)
	if err != nil {
		return err
	}
	// Execute SQL to save
	result, err := db.Exec(SaveBankAccountQuery, req.UserId, bank_id, req.Accounts.AccountId, req.Accounts.AccountType, req.AccessToken)
	if err != nil {
		// This is a 500 error
		return err
	}
	// Verify that some rows were modified
	if rows, _ := result.RowsAffected(); rows == 0 {
		// No rows were updated
		return fmt.Errorf("No rows were saved")
	}
	return nil
}

// Fetches a user's financial accounts data
func GetUserFinancialProfile(db *sql.DB, userID int) (api.UserFinancialProfile, error) {
	var resp api.UserFinancialProfile
	resp.UserId = &userID

	// Find all banks user is associated
	var banks []api.UserAccountInfo

	rows, err := db.Query(GetUserFinancialProfileQuery, userID)
	if err != nil {
		return resp, err
	}
	defer rows.Close()

	// Created to store bank to account mapping
	var bankAccountMap map[string][]api.Account
	bankAccountMap = make(map[string][]api.Account)

	// TO-DO: BASED ON PLAID RETURNS EITHER VALUE IS ARRAY TOO
	var bankPlaidMap map[string]string
	bankPlaidMap = make(map[string]string)

	// Parse the rows and store in intermediate map
	for rows.Next() {
		var bankID int
		var accountID string
		var accountType string
		var plaid_token string
		if err := rows.Scan(&bankID, &accountID, &accountType, &plaid_token); err != nil {
			return resp, fmt.Errorf("could not find account information for the user")
		}
		currentAccount := api.Account{AccountId: accountID, AccountType: accountType}
		bankName, err := GetBankNameByID(db, bankID)
		if err != nil {
			return resp, err
		}

		// CREATE BANK <-> ACCOUNT MAPPING HERE
		value, exists := bankAccountMap[bankName]
		// if bank entry already in the map
		if exists {
			// append account info to value
			value = append(value, currentAccount)
		} else {
			// otherwise create a new array for account
			value := make([]api.Account, 0)
			value = append(value, currentAccount)
			// and create entry in map
			bankAccountMap[bankName] = value
		}

		// CREATE BANK <-> PLAID-TOKEN MAPPING HERE
		accessToken, exists := bankPlaidMap[bankName]
		if exists {
			if accessToken != plaid_token {
				return resp, fmt.Errorf("Multiple access tokens found")
			}
		} else {
			bankPlaidMap[bankName] = plaid_token
		}
	}

	// use the maps to construct our output
	for key, value := range bankAccountMap {
		banks = append(banks, api.UserAccountInfo{
			AccessToken: bankPlaidMap[key],
			BankName:    key,
			Accounts:    value,
		})
	}

	resp.Banks = banks
	return resp, nil
}
