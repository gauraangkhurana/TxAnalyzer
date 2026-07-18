### Summary

tx-svc pulls transactions from Plaid for a linked bank account over a date range, and stores them so viewing them again doesn't re-hit Plaid. It shares the sqlite database with bank-svc (reads `accounts`/`banks` to find the right encrypted access token) and decrypts tokens with the same `PLAID_TOKEN_ENCRYPTION_KEY` bank-svc uses to encrypt them.

There are two operations:

1. **Pull**: given a user, bank name, account ID, and date range, call Plaid's `/transactions/get`, paginate through all results, and upsert them into the `transactions` table (deduped on Plaid's `transaction_id`, so re-pulling an overlapping range just refreshes rows instead of duplicating them).
2. **View**: return everything stored for a user across *all* their linked accounts and banks, combined into one list, sorted most recent first - no Plaid call, just sqlite.

### API Design

`POST /v1/transactions/pull`

    request_body: {
        "user_id": 124,
        "bank_name": "First Platypus Bank",
        "account_id": "...",
        "date_from": "2026-01-01",
        "date_to": "2026-07-17"
    }

    response: {
        "pulled": 42
    }

`GET /v1/transactions?user_id=124`

    response: {
        "transactions": [
            {
                "transaction_id": "...",
                "bank_name": "First Platypus Bank",
                "account_id": "...",
                "date": "2026-07-15",
                "name": "Uber",
                "merchant_name": "Uber",
                "amount": 24.50,
                "category": "TRANSPORTATION",
                "pending": false
            }
        ]
    }

### Notes

- `category` is Plaid's `personal_finance_category.primary` field.
- Not implemented (yet): full `/transactions/sync` cursor-based reconciliation (proper pending→posted transitions), a spend-by-category endpoint, duplicate/reversed-transaction detection, and an endpoint to manually edit a transaction's category - all reasonable next steps once the basic pull/view loop is in daily use.
- Required env vars: `DB_PATH`, `PORT` (defaults to `10003`), `PLAID_CLIENT_ID`, `PLAID_SECRET`, `PLAID_ENV`, `PLAID_TOKEN_ENCRYPTION_KEY` (must match bank-svc's key - see `.env.example` at the repo root).
