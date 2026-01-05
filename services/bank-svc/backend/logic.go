package backend

import (
	"database/sql"

	"spendanalyzer.com/bank/api"
)

var GetBankInstitutionQuery = "SELECT id, name FROM banks"
var SaveBankInstitutionQuery = "INSERT INTO banks (name) VALUES (?)"

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
